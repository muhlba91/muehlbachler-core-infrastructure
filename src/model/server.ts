/**
 * Defines a server.
 */
import { Output, Resource } from '@pulumi/pulumi';

export interface ServerData {
  readonly resource: Resource;
  readonly hostname: string;
  readonly serverId: Output<number | undefined>;
  readonly ipv4Address: string;
  readonly ipv6Address: string;
  // TODO: new
  // readonly resource: Resource;
  readonly privateIPv4: Output<string>;
  readonly publicIPv4: Output<string>;
  readonly publicIPv6: Output<string>;
  readonly sshIPv4: Output<string>;
  readonly network?: Output<string>;
}
