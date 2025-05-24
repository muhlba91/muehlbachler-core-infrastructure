import * as vault from '@pulumi/vault';

/**
 * Enables the AppRole authentication method.
 *
 * @param {vault.Provider} provider the Vault provider
 * @returns {vault.AuthBackend} the AppRole backend
 */
export const enableAppRole = (provider: vault.Provider): vault.AuthBackend =>
  new vault.AuthBackend(
    'vault-auth-backend-approle',
    {
      type: 'approle',
      description: 'App Role Backend',
      tune: {},
    },
    {
      provider: provider,
    },
  );
