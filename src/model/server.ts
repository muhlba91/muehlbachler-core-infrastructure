/**
 * Defines a server.
 */
import { Output, Resource } from '@pulumi/pulumi';

export interface ServerData {
  readonly resource: Resource;
  readonly hostname: Output<string>;
  readonly privateIPv4: Output<string>;
  readonly publicIPv4: Output<string>;
  readonly publicIPv6: Output<string>;
  readonly sshIPv4: Output<string>;
  readonly network?: Output<string>;
}
