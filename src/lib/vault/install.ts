import { remote } from '@pulumi/command';
import { all, Output, Resource } from '@pulumi/pulumi';
import { FileAsset } from '@pulumi/pulumi/asset';

import { dnsConfig, gcpConfig } from '../configuration';
import { getFileHash, readFileContents, writeFileContents } from '../util/file';
import { renderTemplate } from '../util/template';

/**
 * Installs Hashicorp Vault.
 *
 * @param {Output<string>} ipv4Address the IPv4 address
 * @param {Output<string>} sshKey the SSH key
 * @param {string} bucket the bucket name for Vault
 * @param {readonly Resource[]} dependsOn the resources this command depends on
 * @returns {Output<remote.Command>} the remote command
 */
export const installVault = (
  ipv4Address: Output<string>,
  sshKey: Output<string>,
  bucket: string,
  dependsOn: readonly Resource[],
): Output<remote.Command> => {
  const connection = {
    host: ipv4Address,
    privateKey: sshKey,
    user: 'root',
  };

  const prepare = new remote.Command(
    'remote-command-prepare-vault',
    {
      create: readFileContents('./assets/vault/prepare.sh'),
      connection: connection,
    },
    {
      dependsOn: [...dependsOn],
    },
  );

  const dockerComposeHash = Output.create(
    renderTemplate('./assets/vault/docker-compose.yml.j2', {
      domain: dnsConfig.entries.vault.domain,
    }),
  )
    .apply((content) =>
      writeFileContents('./outputs/vault_docker-compose.yml', content, {}),
    )
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    .apply((_) => getFileHash('./outputs/vault_docker-compose.yml'));
  const dockerComposeCopy = dockerComposeHash.apply(
    (hash) =>
      new remote.CopyToRemote(
        'remote-copy-vault-docker-compose',
        {
          source: new FileAsset('./outputs/vault_docker-compose.yml'),
          remotePath: '/opt/vault/docker-compose.yml',
          triggers: [Output.create(hash)],
          connection: connection,
        },
        {
          dependsOn: [...dependsOn, prepare],
        },
      ),
  );

  const vaultConfigHash = Output.create(
    renderTemplate('./assets/vault/config.hcl.j2', {
      gcp: gcpConfig,
      bucket: bucket,
    }),
  )
    .apply((content) =>
      writeFileContents('./outputs/vault_vault-config.hcl', content, {}),
    )
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    .apply((_) => getFileHash('./outputs/vault_vault-config.hcl'));
  const vaultConfigCopy = vaultConfigHash.apply(
    (hash) =>
      new remote.CopyToRemote(
        'remote-copy-vault-config',
        {
          source: new FileAsset('./outputs/vault_vault-config.hcl'),
          remotePath: '/opt/vault/config/vault-config.hcl',
          triggers: [Output.create(hash)],
          connection: connection,
        },
        {
          dependsOn: [...dependsOn, prepare],
        },
      ),
  );

  const systemdServiceHash = getFileHash('./assets/vault/vault.service');
  const systemdServiceCopy = new remote.CopyToRemote(
    'remote-copy-vault-service',
    {
      source: new FileAsset('./assets/vault/vault.service'),
      remotePath: '/etc/systemd/system/vault.service',
      triggers: [Output.create(systemdServiceHash)],
      connection: connection,
    },
    {
      dependsOn: [...dependsOn, prepare],
    },
  );

  return all([dockerComposeCopy, vaultConfigCopy]).apply(
    ([composeCopy, vaultCopy]) =>
      new remote.Command(
        'remote-command-install-vault',
        {
          create: readFileContents('./assets/vault/install.sh'),
          update: readFileContents('./assets/vault/install.sh'),
          triggers: [dockerComposeHash, systemdServiceHash, vaultConfigHash],
          connection: connection,
        },
        {
          dependsOn: [...dependsOn, composeCopy, vaultCopy, systemdServiceCopy],
        },
      ),
  );
};
