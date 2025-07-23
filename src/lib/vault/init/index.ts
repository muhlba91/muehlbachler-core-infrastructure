import { remote } from '@pulumi/command';
import { Output } from '@pulumi/pulumi';
import { parse } from 'yaml';

import { ServerData } from '../../../model/server';
import { VaultKeysData } from '../../../model/vault';
import { username } from '../../configuration';
import { readFileContents } from '../../util/file';

/**
 * Initializes the Hashicorp Vault instance.
 *
 * @param {ServerData} server the server data
 * @param {string} sshPrivateKey the SSH private key (PEM)
 * @returns {Output<VaultKeysData>} the Vault keys
 */
export const initVault = (
  server: ServerData,
  sshPrivateKey: string,
): Output<VaultKeysData> => {
  const initScript = readFileContents('./assets/vault/init.sh');
  const vaultInit = new remote.Command(
    'vault-init',
    {
      connection: {
        host: server.ipv4Address,
        user: username,
        privateKey: sshPrivateKey,
      },
      create: initScript,
    },
    {
      customTimeouts: {
        create: '40m',
        update: '40m',
      },
      dependsOn: [server.resource],
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
