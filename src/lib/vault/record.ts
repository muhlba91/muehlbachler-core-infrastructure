import { Output } from '@pulumi/pulumi';

import { dnsConfig } from '../configuration';
import { createRecord } from '../google/dns/record';

/**
 * Creates the DNS records.
 */
export const createVaultDNSRecords = (
  ipv4Address: Output<string>,
  ipv6Address: Output<string>,
) => {
  createRecord(
    dnsConfig.entries.vault.domain,
    dnsConfig.entries.vault.zoneId,
    'A',
    [ipv4Address],
    {
      project: dnsConfig.project,
    },
  );

  createRecord(
    dnsConfig.entries.vault.domain,
    dnsConfig.entries.vault.zoneId,
    'AAAA',
    [ipv6Address],
    {
      project: dnsConfig.project,
    },
  );
};
