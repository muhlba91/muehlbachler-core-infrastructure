import * as hcloud from '@pulumi/hcloud';

import {
  commonLabels,
  environment,
  globalName,
  networkConfig,
  serverConfig,
} from '../configuration';

/**
 * Creates a Hetzner firewall.
 *
 * @returns {hcloud.Firewall} the generated network
 */
export const createFirewall = (): hcloud.Firewall =>
  new hcloud.Firewall(
    `hcloud-firewall-${globalName}`,
    {
      name: `${globalName}-${environment}`,
      rules: [
        {
          description: 'Allow incoming SSH traffic',
          direction: 'in',
          port: '22',
          protocol: 'tcp',
          sourceIps: serverConfig.publicSsh
            ? ['0.0.0.0/0', '::/0']
            : [networkConfig.cidr],
        },
        {
          description: 'Allow incoming Vault traffic',
          direction: 'in',
          port: '8200',
          protocol: 'tcp',
          sourceIps: ['0.0.0.0/0', '::/0'],
        },
        {
          description: 'Allow incoming HTTP traffic',
          direction: 'in',
          port: '80',
          protocol: 'tcp',
          sourceIps: ['0.0.0.0/0', '::/0'],
        },
        {
          description: 'Allow incoming HTTPS traffic',
          direction: 'in',
          port: '443',
          protocol: 'tcp',
          sourceIps: ['0.0.0.0/0', '::/0'],
        },
      ],
      labels: commonLabels,
    },
    {},
  );
