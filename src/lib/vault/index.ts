import { all, Output, Resource } from '@pulumi/pulumi';
import * as vault from '@pulumi/vault';

import { ServiceAccountData } from '../../model/google/service_account_data';
import {
  VaultData,
  VaultInstanceData,
  VaultKeysData,
  VaultOwnedSecretsData,
} from '../../model/vault';
import { dnsConfig } from '../configuration';

import { enableAppRole } from './auth/approle';
import { enableGithubAuth } from './auth/github_auth';
import { initVault } from './init';
import { createDefaultPolicies } from './init/policies';
import { createStorageBucket } from './storage';

/**
 * Creates the Hashicorp Vault resources.
 *
 * @param {ServiceAccountData} serviceAccount the service account data
 * @returns {VaultData} the Vault data
 */
export const createVaultResources = (
  serviceAccount: ServiceAccountData,
): VaultData => {
  const bucket = createStorageBucket(serviceAccount.serviceAccount.email);

  return {
    serviceAccount: serviceAccount,
    bucket: bucket,
  };
};

/**
 * Configures the Hashicorp Vault instance.
 *
 * @param {Output<string>} ipv4Address the IPv4 address of the server
 * @param {Output<string>} sshKey the SSH key (PEM format)
 * @param {string} bucket the bucket name
 * @param {readonly Resource[]} dependsOn the resources this command depends on
 * @returns {Output<VaultInstanceData>} the instance data
 */
export const configureVault = (
  ipv4Address: Output<string>,
  sshKey: Output<string>,
  bucket: string,
  dependsOn: readonly Resource[],
): Output<VaultInstanceData> => {
  const address = `https://${dnsConfig.entries.vault.domain}:8200`;
  const keys = initVault(ipv4Address, sshKey, dependsOn);
  const vaultProvider = new vault.Provider('vault', {
    address: address,
    token: keys.rootToken,
  });
  createDefaultPolicies(vaultProvider);
  enableAppRole(vaultProvider);
  enableGithubAuth(vaultProvider);

  const ownedSecrets = keys.apply((keys) =>
    storeVaultSecrets(keys, vaultProvider),
  );

  return all([keys, ownedSecrets]).apply(([vaultKeys, vaultOwnedSecrets]) => ({
    bucket: bucket,
    address: address,
    keys: vaultKeys,
    ownedSecrets: vaultOwnedSecrets,
  }));
};

/**
 * Stores the Vault secrets.
 *
 * @param {VaultKeysData} keys the Vault keys
 * @param {vault.Provider} provider the Vault provider
 * @returns {VaultOwnedSecretsData} the owned secrets
 */
export const storeVaultSecrets = (
  keys: VaultKeysData,
  provider: vault.Provider,
): VaultOwnedSecretsData => {
  const mount = new vault.Mount(
    'vault-mount-kv-vault',
    {
      path: 'vault',
      type: 'kv',
      options: {
        version: '2',
      },
      description: 'Vault related secrets',
    },
    {
      provider: provider,
    },
  );

  const keysSecret = new vault.kv.SecretV2(
    'vault-secret-kv-vault-keys',
    {
      mount: mount.path,
      name: 'keys',
      dataJson: JSON.stringify({
        rootToken: keys.rootToken,
        recoveryKey1: keys.recoveryKeys[0],
        recoveryKey2: keys.recoveryKeys[1],
        recoveryKey3: keys.recoveryKeys[2],
        recoveryKey4: keys.recoveryKeys[3],
        recoveryKey5: keys.recoveryKeys[4],
      }),
    },
    {
      provider: provider,
    },
  );

  return {
    mount: mount,
    keys: keysSecret,
  };
};
