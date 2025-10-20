import { Output } from '@pulumi/pulumi';

/**
 * Defines FRR data.
 */
export interface FRRData {
  readonly hostname: Output<string>;
  readonly neighborPassword: Output<string>;
}
