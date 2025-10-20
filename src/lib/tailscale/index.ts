import { remote } from '@pulumi/command';
import { all, Output, Resource } from '@pulumi/pulumi';
import { FileAsset } from '@pulumi/pulumi/asset';

import { tailscaleConfig } from '../configuration';
import { getFileHash, readFileContents, writeFileContents } from '../util/file';
import { renderTemplate } from '../util/template';

/**
 * Installs Tailscale.
 *
 * @param {Output<string>} ipv4Address the IPv4 address
 * @param {Output<string>} sshKey the SSH key
 * @param {readonly Resource[]} dependsOn the resources this command depends on
 * @returns {Output<remote.Command>} the remote command
 */
export const installTailscale = (
  ipv4Address: Output<string>,
  sshKey: Output<string>,
  dependsOn: readonly Resource[],
): Output<remote.Command> => {
  const connection = {
    host: ipv4Address,
    privateKey: sshKey,
    user: 'root',
  };

  const prepare = new remote.Command(
    'remote-command-prepare-tailscale',
    {
      create: readFileContents('./assets/tailscale/prepare.sh'),
      connection: connection,
    },
    {
      dependsOn: [...dependsOn],
    },
  );

  const dockerComposeHash = Output.create(
    renderTemplate('./assets/tailscale/docker-compose.yml.j2', {
      authKey: tailscaleConfig.authKey,
    }),
  )
    .apply((content) =>
      writeFileContents('./outputs/tailscale_docker-compose.yml', content, {}),
    )
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    .apply((_) => getFileHash('./outputs/tailscale_docker-compose.yml'));
  const dockerComposeCopy = dockerComposeHash.apply(
    (hash) =>
      new remote.CopyToRemote(
        'remote-copy-tailscale-docker-compose',
        {
          source: new FileAsset('./outputs/tailscale_docker-compose.yml'),
          remotePath: '/opt/tailscale/docker-compose.yml',
          triggers: [Output.create(hash)],
          connection: connection,
        },
        {
          dependsOn: [...dependsOn, prepare],
        },
      ),
  );

  const systemdServiceHash = getFileHash(
    './assets/tailscale/tailscale.service',
  );
  const systemdServiceCopy = new remote.CopyToRemote(
    'remote-copy-tailscale-service',
    {
      source: new FileAsset('./assets/tailscale/tailscale.service'),
      remotePath: '/etc/systemd/system/tailscale.service',
      triggers: [Output.create(systemdServiceHash)],
      connection: connection,
    },
    {
      dependsOn: [...dependsOn, prepare],
    },
  );

  return all([dockerComposeCopy]).apply(
    ([composeCopy]) =>
      new remote.Command(
        'remote-command-install-tailscale',
        {
          create: readFileContents('./assets/tailscale/install.sh'),
          update: readFileContents('./assets/tailscale/install.sh'),
          triggers: [dockerComposeHash, systemdServiceHash],
          connection: connection,
        },
        {
          dependsOn: [...dependsOn, composeCopy, systemdServiceCopy],
        },
      ),
  );
};
