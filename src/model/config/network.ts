import { StringMap } from '../map';

/**
 * Defines network configuration.
 */
export interface NetworkConfig {
  readonly name: string;
  readonly dnsSuffix: string;
  readonly cidr: string;
  readonly subnetCidr: string;
  readonly firewallRules: StringMap<NetworkFirewallRule>;
}

export interface NetworkFirewallRule {
  readonly description: string;
  readonly port?: string;
  readonly protocol: 'tcp' | 'udp' | 'icmp';
  readonly sourceIps?: string[];
}
