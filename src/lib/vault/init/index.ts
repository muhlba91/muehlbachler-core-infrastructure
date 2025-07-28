import { remote } from '@pulumi/command';
import { Output, Resource } from '@pulumi/pulumi';
import { parse } from 'yaml';

import { VaultKeysData } from '../../../model/vault';
import { readFileContents } from '../../util/file';

/**
 * Initializes the Hashicorp Vault instance.
 *
 * @param {Output<string>} ipv4Address the IPv4 address of the server
 * @param {Output<string>} sshKey the SSH key (PEM format)
 * @param {readonly Resource[]} dependsOn the resources this command depends on
 * @returns {Output<VaultKeysData>} the Vault keys
 */
export const initVault = (
  ipv4Address: Output<string>,
  sshKey: Output<string>,
  dependsOn: readonly Resource[],
): Output<VaultKeysData> => {
  const initScript = readFileContents('./assets/vault/init.sh');
  const vaultInit = new remote.Command(
    'vault-init',
    {
      connection: {
        host: ipv4Address,
        privateKey: sshKey,
      },
      create: initScript,
    },
    {
      customTimeouts: {
        create: '40m',
        update: '40m',
      },
      dependsOn: dependsOn.map((resource) => resource),
    },
  );

  const keys: Output<VaultKeysData> = vaultInit.stdout.apply((stdout) => {
    const startBlock = stdout.lastIndexOf('--START TOKENS--');
    const endBlock = stdout.lastIndexOf('--END TOKENS--');
    const tokens = stdout.substring(
      startBlock + '--START TOKENS--'.length,
      endBlock,
    );
    const parsedTokens = parse(tokens);
    return {
      rootToken: parsedTokens['root_token'] as string,
      recoveryKeys: parsedTokens['recovery_keys'] as string[],
    };
  });

  return keys;
};
