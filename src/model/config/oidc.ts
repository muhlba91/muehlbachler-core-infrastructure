import { StringMap } from '../map';

/**
 * Defines configuration data for OIDC.
 */
export interface OIDCConfig {
  readonly discoveryUrl: string;
  readonly clients: StringMap<OIDCClientConfig>;
}

/**
 * Defines configuration data for an OIDC client.
 */
export interface OIDCClientConfig {
  readonly clientId: string;
  readonly clientSecret: string;
}
