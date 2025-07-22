import { all } from '@pulumi/pulumi';
import { stringify } from 'yaml';

import { createHetznerInstance } from './lib/hetzner';
import { createDir } from './lib/util/create_dir';
import { createRandomPassword } from './lib/util/random';
import { createSSHKey } from './lib/util/ssh_key';
import { writeFilePulumiAndUploadToS3 } from './lib/util/storage';
import { createVaultInstance, createVaultResources } from './lib/vault';

export = async () => {
  createDir('outputs');

  // Keys, IAM, ...
  const userPassword = createRandomPassword('server', {});
  const sshKey = createSSHKey('vault', {});
  const vaultData = createVaultResources();

  // Hetzner resources
  const hetznerVaultInstance = await createHetznerInstance(
    sshKey.publicKeyOpenssh,
  );

  // Vault instance
  const vaultInstance = all([
    userPassword.password,
    sshKey.publicKeyOpenssh,
    sshKey.privateKeyPem,
    vaultData.serviceAccount.key.privateKey,
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
    hetzner: {
      ipv4: hetznerVaultInstance.publicIPv4,
      ipv6: hetznerVaultInstance.publicIPv6,
    },
    server: {
      ipv4: vaultInstance.server.ipv4Address,
      ipv6: vaultInstance.server.ipv6Address,
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
