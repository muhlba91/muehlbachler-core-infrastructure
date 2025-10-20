import { interpolate, Output } from '@pulumi/pulumi';

import { FRRData } from '../../model/frr';
import { networkConfig } from '../configuration';
import { createRandomPassword } from '../util/random';

/**
 * Creates the FRR resources.
 *
 * @param {Output<string>} hostname the hostname
 * @returns {FRRData} the FRR data
 */
export const createFRRResources = (hostname: Output<string>): FRRData => {
  const neighborPassword = createRandomPassword('frr-neighbor-password', {
    special: false,
  });

  return {
    hostname: interpolate`${hostname}.${networkConfig.dnsSuffix}`,
    neighborPassword: neighborPassword.password,
  };
};
