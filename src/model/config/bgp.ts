/**
 * Defines configuration data for BGP.
 */
export interface BGPConfig {
  readonly routerId: string;
  readonly localAsn: number;
  readonly interface: string;
  readonly neighbors: BGPNeighborConfig[];
  readonly advertisedIPv4Networks?: string[];
  readonly advertisedIPv6Networks?: string[];
}

/**
 * Defines configuration data for a BGP neighbor.
 */
export interface BGPNeighborConfig {
  readonly address: string;
  readonly remoteAsn: number;
  readonly interface?: string;
}
