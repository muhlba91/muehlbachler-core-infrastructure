import { remote } from '@pulumi/command';
import { all, getProject, Output, Resource } from '@pulumi/pulumi';
import { FileAsset } from '@pulumi/pulumi/asset';

import { WireGuardData } from '../../model/wireguard';
import { backupBucketId, dnsConfig } from '../configuration';
import { getFileHash, readFileContents, writeFileContents } from '../util/file';
import { BUCKET_PATH } from '../util/storage';
import { renderTemplate } from '../util/template';

/**
 * Installs WireGuard (Portal).
 *
 * @param {Output<string>} ipv4Address the IPv4 address
 * @param {Output<string>} sshKey the SSH key
 * @param {WireGuardData} wireguardData the WireGuard configuration data
 * @param {readonly Resource[]} dependsOn the resources this command depends on
 * @returns {Output<remote.Command>} the remote command
 */
export const installWireguard = (
  ipv4Address: Output<string>,
  sshKey: Output<string>,
  wireguardData: WireGuardData,
  dependsOn: readonly Resource[],
): Output<remote.Command> => {
  const connection = {
    host: ipv4Address,
    privateKey: sshKey,
    user: 'root',
  };

  const prepare = new remote.Command(
    'remote-command-prepare-wireguard',
    {
      create: readFileContents('./assets/wireguard/prepare.sh'),
      connection: connection,
    },
    {
      dependsOn: [...dependsOn],
    },
  );

  const cronFileHash = getFileHash('./assets/wireguard/cron/cron');
  const cronFileCopy = new remote.CopyToRemote(
    'remote-copy-wireguard-cron',
    {
      source: new FileAsset('./assets/wireguard/cron/cron'),
      remotePath: '/etc/cron.d/wireguard',
      triggers: [Output.create(cronFileHash)],
      connection: connection,
    },
    {
      dependsOn: [...dependsOn, prepare],
    },
  );

  const backupFileHash = Output.create(
    renderTemplate('./assets/wireguard/cron/wireguard-backup.j2', {
      project: getProject(),
      bucket: {
        id: backupBucketId,
        path: BUCKET_PATH,
      },
    }),
  )
    .apply((content) =>
      writeFileContents('./outputs/wireguard_backup', content, {}),
    )
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    .apply((_) => getFileHash('./outputs/wireguard_backup'));
  const backupFileCopy = backupFileHash.apply(
    (hash) =>
      new remote.CopyToRemote(
        'remote-copy-wireguard-backup',
        {
          source: new FileAsset('./outputs/wireguard_backup'),
          remotePath: '/bin/wireguard-backup',
          triggers: [Output.create(hash)],
          connection: connection,
        },
        {
          dependsOn: [...dependsOn, prepare],
        },
      ),
  );

  const cronInstall = all([cronFileCopy, backupFileCopy]).apply(
    ([cronCopy, backupCopy]) =>
      new remote.Command(
        'remote-command-install-wireguard-cron',
        {
          create: readFileContents('./assets/wireguard/cron/install.sh'),
          update: readFileContents('./assets/wireguard/cron/install.sh'),
          triggers: [cronFileHash, backupFileHash],
          connection: connection,
        },
        {
          dependsOn: [...dependsOn, prepare, cronCopy, backupCopy],
        },
      ),
  );

  const dockerComposeHash = Output.create(
    renderTemplate('./assets/wireguard/docker-compose.yml.j2', {
      domain: dnsConfig.entries.wireguard.domain,
    }),
  )
    .apply((content) =>
      writeFileContents('./outputs/wireguard_docker-compose.yml', content, {}),
    )
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    .apply((_) => getFileHash('./outputs/wireguard_docker-compose.yml'));
  const dockerComposeCopy = dockerComposeHash.apply(
    (hash) =>
      new remote.CopyToRemote(
        'remote-copy-wireguard-docker-compose',
        {
          source: new FileAsset('./outputs/wireguard_docker-compose.yml'),
          remotePath: '/opt/wireguard/docker-compose.yml',
          triggers: [Output.create(hash)],
          connection: connection,
        },
        {
          dependsOn: [...dependsOn, prepare],
        },
      ),
  );

  const wireguardConfigHash = all([
    wireguardData.adminPassword,
    wireguardData.database.encryptionPassphrase,
    wireguardData.web.sessionSecret,
    wireguardData.web.csrfSecret,
  ])
    .apply(([adminPassword, encryptionPassphrase, sessionSecret, csrfSecret]) =>
      renderTemplate('./assets/wireguard/config.yml.j2', {
        domain: dnsConfig.entries.wireguard.domain,
        ...wireguardData,
        adminPassword: adminPassword,
        database: {
          ...wireguardData.database,
          encryptionPassphrase: encryptionPassphrase,
        },
        web: {
          ...wireguardData.web,
          sessionSecret: sessionSecret,
          csrfSecret: csrfSecret,
        },
      }),
    )
    .apply((content) =>
      writeFileContents('./outputs/wireguard_config.yml', content, {}),
    )
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    .apply((_) => getFileHash('./outputs/wireguard_config.yml'));
  const wireguardConfigCopy = wireguardConfigHash.apply(
    (hash) =>
      new remote.CopyToRemote(
        'remote-copy-wireguard-config',
        {
          source: new FileAsset('./outputs/wireguard_config.yml'),
          remotePath: '/opt/wireguard/config/config.yml',
          triggers: [Output.create(hash)],
          connection: connection,
        },
        {
          dependsOn: [...dependsOn, prepare],
        },
      ),
  );

  const systemdServiceHash = getFileHash(
    './assets/wireguard/wireguard.service',
  );
  const systemdServiceCopy = new remote.CopyToRemote(
    'remote-copy-wireguard-service',
    {
      source: new FileAsset('./assets/wireguard/wireguard.service'),
      remotePath: '/etc/systemd/system/wireguard.service',
      triggers: [Output.create(systemdServiceHash)],
      connection: connection,
    },
    {
      dependsOn: [...dependsOn, prepare],
    },
  );

  return all([dockerComposeCopy, wireguardConfigCopy, cronInstall]).apply(
    ([composeCopy, wireguardCopy, cronInstaller]) => {
      const installScript = renderTemplate('./assets/wireguard/install.sh.j2', {
        project: getProject(),
        bucket: {
          id: backupBucketId,
          path: BUCKET_PATH,
        },
      });

      return new remote.Command(
        'remote-command-install-wireguard',
        {
          create: installScript,
          update: installScript,
          triggers: [
            dockerComposeHash,
            systemdServiceHash,
            wireguardConfigHash,
          ],
          connection: connection,
        },
        {
          dependsOn: [
            ...dependsOn,
            composeCopy,
            wireguardCopy,
            systemdServiceCopy,
            cronInstaller,
          ],
        },
      );
    },
  );
};
