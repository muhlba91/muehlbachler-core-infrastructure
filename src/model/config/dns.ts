import { StringMap } from '../map';

/**
 * Defines configuration data for DNS.
 */
export interface DNSConfig {
  readonly project: string;
  readonly email: string;
  readonly entries: StringMap<string>;
}
