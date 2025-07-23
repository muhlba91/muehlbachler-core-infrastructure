import { StringMap } from '../map';

/**
 * Defines configuration data for DNS.
 */
export interface DNSConfig {
  readonly project: string;
  readonly email: string;
  readonly entries: StringMap<DNSEntryConfig>;
}

/**
 * Defines configuration data for one DNS entry.
 */
export interface DNSEntryConfig {
  readonly domain: string;
  readonly zoneId: string;
}
