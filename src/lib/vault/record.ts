import { Output, Resource } from '@pulumi/pulumi';

import { dnsConfig } from '../configuration';
import { createRecord } from '../google/dns/record';

/**
 * Creates the DNS records.
 *
 * @param {Output<string>} ipv4Address the IPv4 address of the Vault instance
 * @param {Output<string>} ipv6Address the IPv6 address of the Vault instance
 * @returns {Output<Resource[]>} the DNS records
 */
export const createVaultDNSRecords = (
  ipv4Address: Output<string>,
  ipv6Address: Output<string>,
): Resource[] => [
  createRecord(
    dnsConfig.entries.vault.domain,
    dnsConfig.entries.vault.zoneId,
    'A',
    [ipv4Address],
    {
      project: dnsConfig.project,
    },
  ),

  createRecord(
    dnsConfig.entries.vault.domain,
    dnsConfig.entries.vault.zoneId,
    'AAAA',
    [ipv6Address],
    {
      project: dnsConfig.project,
    },
  ),
];
