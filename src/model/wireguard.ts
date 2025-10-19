import { Output } from '@pulumi/pulumi';

/**
 * Defines WireGuard data.
 */
export interface WireGuardData {
  readonly adminPassword: Output<string>;
  readonly database: WireGuardDatabaseData;
  readonly oidc: WireGuardOIDCData;
  readonly web: WireGuardWebData;
}

/**
 * Defines WireGuard database data.
 */
export interface WireGuardDatabaseData {
  readonly encryptionPassphrase: Output<string>;
}

/**
 * Defines WireGuard OIDC data.
 */
export interface WireGuardOIDCData {
  readonly baseUrl: string;
  readonly clientId: string;
  readonly clientSecret: string;
}

/**
 * Defines WireGuard web data.
 */
export interface WireGuardWebData {
  readonly sessionSecret: Output<string>;
  readonly csrfSecret: Output<string>;
}
