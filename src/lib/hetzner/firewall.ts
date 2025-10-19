import * as hcloud from '@pulumi/hcloud';

import {
  commonLabels,
  environment,
  globalName,
  networkConfig,
  serverConfig,
} from '../configuration';

const NETWORK_ALL_CIDR = ['0.0.0.0/0', '::/0'];

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
            ? NETWORK_ALL_CIDR
            : [networkConfig.cidr],
        },
        ...Object.values(networkConfig.firewallRules).map((rule) => ({
          description: rule.description,
          direction: 'in',
          port: rule.port,
          protocol: rule.protocol,
          sourceIps: rule.sourceIps ?? NETWORK_ALL_CIDR,
        })),
      ],
      labels: commonLabels,
    },
    {},
  );
