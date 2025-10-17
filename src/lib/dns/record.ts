import { Output, Resource } from '@pulumi/pulumi';

import { createVaultDNSRecords } from '../vault/record';

/**
 * Creates the DNS records.
 *
 * @param {Output<string>} ipv4Address the IPv4 address of the Vault instance
 * @param {Output<string>} ipv6Address the IPv6 address of the Vault instance
 * @returns {Output<Resource[]>} the DNS records
 */
export const createDNSRecords = (
  ipv4Address: Output<string>,
  ipv6Address: Output<string>,
): Resource[] => createVaultDNSRecords(ipv4Address, ipv6Address);
