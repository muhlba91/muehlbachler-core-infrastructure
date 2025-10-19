import { Output, Resource } from '@pulumi/pulumi';

import { dnsConfig } from '../configuration';
import { createRecord } from '../google/dns/record';

/**
 * Creates the DNS records.
 *
 * @param {Output<string>} ipv4Address the IPv4 address of the WireGuard instance
 * @param {Output<string>} ipv6Address the IPv6 address of the WireGuard instance
 * @returns {Output<Resource[]>} the DNS records
 */
export const createWireguardDNSRecords = (
  ipv4Address: Output<string>,
  ipv6Address: Output<string>,
): Resource[] => [
  createRecord(
    dnsConfig.entries.wireguard.domain,
    dnsConfig.entries.wireguard.zoneId,
    'A',
    [ipv4Address],
    {
      project: dnsConfig.project,
    },
  ),

  createRecord(
    dnsConfig.entries.wireguard.domain,
    dnsConfig.entries.wireguard.zoneId,
    'AAAA',
    [ipv6Address],
    {
      project: dnsConfig.project,
    },
  ),
];
