import { Config, getStack } from '@pulumi/pulumi';

import { DNSConfig } from '../model/config/dns';
import { GCPConfig } from '../model/config/google';
import { NetworkConfig } from '../model/config/network';
import { OIDCConfig } from '../model/config/oidc';
import { ServerConfig } from '../model/config/server';

export const environment = getStack();

const config = new Config();
export const gcpConfig = config.requireObject<GCPConfig>('gcp');
export const serverConfig = config.requireObject<ServerConfig>('server');
export const networkConfig = config.requireObject<NetworkConfig>('network');
export const oidcConfig = config.requireObject<OIDCConfig>('oidc');
export const bucketId = config.require<string>('bucketId');
export const dnsConfig = config.requireObject<DNSConfig>('dns');

export const globalName = 'core';

export const commonLabels = {
  environment: environment,
  application: globalName,
};
