import { WireGuardData } from '../../model/wireguard';
import { oidcConfig } from '../configuration';
import { createRandomPassword } from '../util/random';

/**
 * Creates the WireGuard (Portal) resources.
 *
 * @returns {WireGuardData} the WireGuard data
 */
export const createWireguardResources = (): WireGuardData => {
  const adminPassword = createRandomPassword('wireguard-admin-password', {
    special: false,
  });

  const databaseEncryptionPassphrase = createRandomPassword(
    'wireguard-database-encryption-passphrase',
    {
      length: 32,
      special: false,
    },
  );

  const sessionSecret = createRandomPassword('wireguard-web-session-secret', {
    length: 32,
    special: false,
  });
  const csrfSecret = createRandomPassword('wireguard-web-csrf-secret', {
    length: 32,
    special: false,
  });

  return {
    adminPassword: adminPassword.password,
    database: {
      encryptionPassphrase: databaseEncryptionPassphrase.password,
    },
    oidc: {
      baseUrl: oidcConfig.discoveryUrl,
      clientId: oidcConfig.clients['wireguard'].clientId,
      clientSecret: oidcConfig.clients['wireguard'].clientSecret,
    },
    web: {
      sessionSecret: sessionSecret.password,
      csrfSecret: csrfSecret.password,
    },
  };
};
