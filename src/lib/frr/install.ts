import { remote } from '@pulumi/command';
import { all, Output, Resource } from '@pulumi/pulumi';
import { FileAsset } from '@pulumi/pulumi/asset';

import { FRRData } from '../../model/frr';
import { bgpConfig } from '../configuration';
import { getFileHash, readFileContents, writeFileContents } from '../util/file';
import { renderTemplate } from '../util/template';

/**
 * Installs FRR.
 *
 * @param {Output<string>} ipv4Address the IPv4 address
 * @param {Output<string>} sshKey the SSH key
 * @param {FRRData} frrData the FRR data
 * @param {readonly Resource[]} dependsOn the resources this command depends on
 * @returns {Output<remote.Command>} the remote command
 */
export const installFRR = (
  ipv4Address: Output<string>,
  sshKey: Output<string>,
  frrData: FRRData,
  dependsOn: readonly Resource[],
): Output<remote.Command> => {
  const connection = {
    host: ipv4Address,
    privateKey: sshKey,
    user: 'root',
  };

  const prepare = new remote.Command(
    'remote-command-prepare-frr',
    {
      create: readFileContents('./assets/frr/prepare.sh'),
      connection: connection,
    },
    {
      dependsOn: [...dependsOn],
    },
  );

  const dockerComposeHash = getFileHash('./assets/frr/docker-compose.yml');
  const dockerComposeCopy = new remote.CopyToRemote(
    'remote-copy-frr-docker-compose',
    {
      source: new FileAsset('./assets/frr/docker-compose.yml'),
      remotePath: '/opt/frr/docker-compose.yml',
      triggers: [Output.create(dockerComposeHash)],
      connection: connection,
    },
    {
      dependsOn: [...dependsOn, prepare],
    },
  );

  const frrConfigHash = all([frrData.hostname, frrData.neighborPassword])
    .apply(([hostname, password]) =>
      renderTemplate('./assets/frr/config/frr.conf.j2', {
        hostname: hostname,
        bgp: {
          localAsn: bgpConfig.localAsn,
          routerId: bgpConfig.routerId,
          interface: bgpConfig.interface,
          advertisedIPv4Networks: bgpConfig.advertisedIPv4Networks,
          advertisedIPv6Networks: bgpConfig.advertisedIPv6Networks,
          neighbors: bgpConfig.neighbors.map((neighbor) => ({
            address: neighbor.address,
            asn: neighbor.remoteAsn,
            interface: neighbor.interface || bgpConfig.interface,
            password: password,
          })),
        },
      }),
    )
    .apply((content) =>
      writeFileContents('./outputs/frr_frr.conf', content, {}),
    )
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    .apply((_) => getFileHash('./outputs/frr_frr.conf'));
  const frrConfigCopy = frrConfigHash.apply(
    (hash) =>
      new remote.CopyToRemote(
        'remote-copy-frr-config',
        {
          source: new FileAsset('./outputs/frr_frr.conf'),
          remotePath: '/opt/frr/config/frr.conf',
          triggers: [Output.create(hash)],
          connection: connection,
        },
        {
          dependsOn: [...dependsOn, prepare],
        },
      ),
  );

  const vtyshConfigHash = getFileHash('./assets/frr/config/vtysh.conf');
  const vtyshConfigCopy = new remote.CopyToRemote(
    'remote-copy-frr-vtysh',
    {
      source: new FileAsset('./assets/frr/config/vtysh.conf'),
      remotePath: '/opt/frr/config/vtysh.conf',
      triggers: [Output.create(vtyshConfigHash)],
      connection: connection,
    },
    {
      dependsOn: [...dependsOn, prepare],
    },
  );

  const daemonsHash = getFileHash('./assets/frr/config/daemons');
  const daemonsCopy = new remote.CopyToRemote(
    'remote-copy-frr-daemons',
    {
      source: new FileAsset('./assets/frr/config/daemons'),
      remotePath: '/opt/frr/config/daemons',
      triggers: [Output.create(daemonsHash)],
      connection: connection,
    },
    {
      dependsOn: [...dependsOn, prepare],
    },
  );

  const systemdServiceHash = getFileHash('./assets/frr/frr.service');
  const systemdServiceCopy = new remote.CopyToRemote(
    'remote-copy-frr-service',
    {
      source: new FileAsset('./assets/frr/frr.service'),
      remotePath: '/etc/systemd/system/frr.service',
      triggers: [Output.create(systemdServiceHash)],
      connection: connection,
    },
    {
      dependsOn: [...dependsOn, prepare],
    },
  );

  return all([frrConfigCopy]).apply(
    ([frrCopy]) =>
      new remote.Command(
        'remote-command-install-frr',
        {
          create: readFileContents('./assets/frr/install.sh'),
          update: readFileContents('./assets/frr/install.sh'),
          triggers: [
            dockerComposeHash,
            systemdServiceHash,
            frrConfigHash,
            vtyshConfigHash,
            daemonsHash,
          ],
          connection: connection,
        },
        {
          dependsOn: [
            ...dependsOn,
            dockerComposeCopy,
            frrCopy,
            systemdServiceCopy,
            vtyshConfigCopy,
            daemonsCopy,
          ],
        },
      ),
  );
};
