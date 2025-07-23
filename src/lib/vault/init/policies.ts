import * as vault from '@pulumi/vault';

import { readFileContents } from '../../util/file';

/**
 * Creates the default Hashicorp Vault policies.
 *
 * @param {vault.Provider} provider the Vault provider
 * @returns {vault.Policy[]} the policies
 */
export const createDefaultPolicies = (
  provider: vault.Provider,
): vault.Policy[] => [
  new vault.Policy(
    'vault-policy-admin',
    {
      name: 'admin',
      policy: readFileContents('./assets/vault/policies/admin.hcl'),
    },
    {
      provider: provider,
    },
  ),
  new vault.Policy(
    'vault-policy-manager',
    {
      name: 'manager',
      policy: readFileContents('./assets/vault/policies/manager.hcl'),
    },
    {
      provider: provider,
    },
  ),
  new vault.Policy(
    'vault-policy-reader',
    {
      name: 'reader',
      policy: readFileContents('./assets/vault/policies/reader.hcl'),
    },
    {
      provider: provider,
    },
  ),
];
