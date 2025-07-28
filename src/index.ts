import { all } from '@pulumi/pulumi';
import { stringify } from 'yaml';

import { installDocker } from './lib/docker';
import { installGCloud } from './lib/gcloud';
import { createHetznerInstance } from './lib/hetzner';
import { createServiceAccount } from './lib/iam';
import { installTraefik } from './lib/traefik';
import { createDir } from './lib/util/create_dir';
import { createSSHKey } from './lib/util/ssh_key';
import { writeFilePulumiAndUploadToS3 } from './lib/util/storage';
import { configureVault, createVaultResources } from './lib/vault';
import { installVault } from './lib/vault/install';
import { createVaultDNSRecords } from './lib/vault/record';

export = async () => {
  createDir('outputs');

  // Keys, IAM, ...
  const serviceAccount = createServiceAccount();
  const sshKey = createSSHKey('vault', {});
  const vaultData = createVaultResources(serviceAccount);

  // Instance
  const instance = await createHetznerInstance(sshKey.publicKeyOpenssh);
  const dnsEntries = createVaultDNSRecords(
    instance.publicIPv4,
    instance.publicIPv6,
  );
  const docker = installDocker(instance.sshIPv4, sshKey.privateKeyPem, [
    instance.resource,
  ]);
  const gcloud = installGCloud(
    instance.sshIPv4,
    sshKey.privateKeyPem,
    serviceAccount,
    [docker, instance.resource],
  );
  const traefik = gcloud.apply((gcloudInstall) =>
    installTraefik(instance.sshIPv4, sshKey.privateKeyPem, [
      ...dnsEntries,
      docker,
      gcloudInstall,
      instance.resource,
    ]),
  );
  const vault = all([gcloud, traefik, vaultData.bucket.id]).apply(
    ([gcloudInstall, traefikInstall, bucket]) =>
      installVault(instance.sshIPv4, sshKey.privateKeyPem, bucket, [
        ...dnsEntries,
        docker,
        gcloudInstall,
        traefikInstall,
        instance.resource,
      ]),
  );
  const vaultInstanceData = all([traefik, vault, vaultData.bucket.id]).apply(
    ([traefikInstall, vaultInstall, bucket]) =>
      configureVault(instance.sshIPv4, sshKey.privateKeyPem, bucket, [
        ...dnsEntries,
        traefikInstall,
        vaultInstall,
      ]),
  );

  // Write output files
  writeFilePulumiAndUploadToS3('ssh.key', sshKey.privateKeyPem, {
    permissions: '0600',
  });
  writeFilePulumiAndUploadToS3(
    'vault.yml',
    all([vaultInstanceData.address, vaultInstanceData.keys]).apply(
      ([address, keys]) =>
        stringify({
          address: address,
          keys: keys,
        }),
    ),
    {
      permissions: '0600',
    },
  );

  return {
    server: {
      ipv4: instance.publicIPv4,
      ipv6: instance.publicIPv6,
    },
    vault: {
      address: vaultInstanceData.address,
      storage: {
        type: 'gcs',
        bucket: vaultData.bucket.id,
      },
      keys: vaultInstanceData.keys,
      ownedSecrets: {
        mount: vaultInstanceData.ownedSecrets.mount.path,
        keys: vaultInstanceData.ownedSecrets.keys.path,
      },
    },
  };
};
