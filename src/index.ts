import { all } from '@pulumi/pulumi';
import { stringify } from 'yaml';

import { installDocker } from './lib/docker';
import { installGCloud } from './lib/gcloud';
import { createHetznerInstance } from './lib/hetzner';
import { createServiceAccount } from './lib/iam';
import { installTraefik } from './lib/traefik';
import { createDir } from './lib/util/create_dir';
import { createRandomPassword } from './lib/util/random';
import { createSSHKey } from './lib/util/ssh_key';
import { writeFilePulumiAndUploadToS3 } from './lib/util/storage';
import { createVaultInstance, createVaultResources } from './lib/vault';
import { installVault } from './lib/vault/install';
import { createVaultDNSRecords } from './lib/vault/record';

export = async () => {
  createDir('outputs');

  // Keys, IAM, ...
  const serviceAccount = createServiceAccount();
  const userPassword = createRandomPassword('server', {});
  const sshKey = createSSHKey('vault', {});
  const vaultData = createVaultResources(serviceAccount);

  // Instance
  const instance = await createHetznerInstance(sshKey.publicKeyOpenssh);
  const docker = installDocker(instance.sshIPv4, sshKey.privateKeyPem, [
    instance.resource,
  ]);
  const gcloud = installGCloud(
    instance.sshIPv4,
    sshKey.privateKeyPem,
    serviceAccount,
    [docker, instance.resource],
  );
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const traefik = gcloud.apply((gcloudInstall) =>
    installTraefik(instance.sshIPv4, sshKey.privateKeyPem, [
      docker,
      gcloudInstall,
      instance.resource,
    ]),
  );
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const vault = all([gcloud, vaultData.bucket.id]).apply(
    ([gcloudInstall, bucket]) =>
      installVault(instance.sshIPv4, sshKey.privateKeyPem, bucket, [
        docker,
        gcloudInstall,
        instance.resource,
      ]),
  );
  // TODO: init vault
  createVaultDNSRecords(instance.publicIPv4, instance.publicIPv6);

  // Vault instance
  const vaultInstance = all([
    userPassword.password,
    sshKey.publicKeyOpenssh,
    sshKey.privateKeyPem,
    serviceAccount.key.privateKey,
    vaultData.bucket.id,
  ]).apply(
    ([
      userPasswordPlain,
      sshPublicKey,
      sshPrivateKey,
      vaultServiceAccountKey,
      bucket,
    ]) =>
      createVaultInstance(
        userPasswordPlain,
        sshPublicKey.trim(),
        sshPrivateKey.trim(),
        vaultServiceAccountKey.trim(),
        bucket,
      ),
  );

  // Write output files
  writeFilePulumiAndUploadToS3('ssh.key', sshKey.privateKeyPem, {
    permissions: '0600',
  });
  writeFilePulumiAndUploadToS3(
    'vault.yml',
    all([vaultInstance.address, vaultInstance.keys]).apply(([address, keys]) =>
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
      address: vaultInstance.address,
      storage: {
        type: 'gcs',
        bucket: vaultInstance.bucket,
      },
      keys: vaultInstance.keys,
      ownedSecrets: {
        mount: vaultInstance.ownedSecrets.mount.path,
        keys: vaultInstance.ownedSecrets.keys.path,
      },
    },
  };
};
