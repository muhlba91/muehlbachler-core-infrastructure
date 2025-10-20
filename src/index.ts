import { all } from '@pulumi/pulumi';
import { stringify } from 'yaml';

import { createDNSRecords } from './lib/dns/record';
import { installDocker } from './lib/docker';
import { createFRRResources } from './lib/frr';
import { installFRR } from './lib/frr/install';
import { installGCloud } from './lib/gcloud';
import { createHetznerInstance } from './lib/hetzner';
import { createServiceAccount } from './lib/iam';
import { installTraefik } from './lib/traefik';
import { createDir } from './lib/util/create_dir';
import { createSSHKey } from './lib/util/ssh_key';
import { writeFilePulumiAndUploadToS3 } from './lib/util/storage';
import { configureVault, createVaultResources } from './lib/vault';
import { installVault } from './lib/vault/install';
import { createWireguardResources } from './lib/wireguard';
import { installWireguard } from './lib/wireguard/install';

export = async () => {
  createDir('outputs');

  // instance
  const sshKey = createSSHKey('core', {});
  const instance = await createHetznerInstance(sshKey.publicKeyOpenssh);

  // dns
  const dnsEntries = createDNSRecords(instance.publicIPv4, instance.publicIPv6);

  // docker
  const docker = installDocker(instance.sshIPv4, sshKey.privateKeyPem, [
    instance.resource,
  ]);

  // google cloud
  const serviceAccount = createServiceAccount();
  const gcloud = installGCloud(
    instance.sshIPv4,
    sshKey.privateKeyPem,
    serviceAccount,
    [docker, instance.resource],
  );

  // traefik
  const traefik = gcloud.apply((gcloudInstall) =>
    installTraefik(instance.sshIPv4, sshKey.privateKeyPem, [
      ...dnsEntries,
      docker,
      gcloudInstall,
      instance.resource,
    ]),
  );

  // vault
  const vaultData = createVaultResources(serviceAccount);
  const vault = all([gcloud, traefik, vaultData.bucket.id]).apply(
    ([gcloudInstall, traefikInstall, bucket]) =>
      installVault(instance.sshIPv4, sshKey.privateKeyPem, bucket, [
        ...dnsEntries,
        docker,
        gcloudInstall,
        traefikInstall,
        instance.resource,
      ]),
  );
  const vaultInstanceData = all([traefik, vault, vaultData.bucket.id]).apply(
    ([traefikInstall, vaultInstall, bucket]) =>
      configureVault(instance.sshIPv4, sshKey.privateKeyPem, bucket, [
        ...dnsEntries,
        traefikInstall,
        vaultInstall,
      ]),
  );

  // wireguard
  // FIXME: add backups?
  const wireguardData = createWireguardResources();
  const wireguard = all([traefik]).apply(([traefikInstall]) =>
    installWireguard(instance.sshIPv4, sshKey.privateKeyPem, wireguardData, [
      ...dnsEntries,
      traefikInstall,
    ]),
  );

  // frr
  const frrData = createFRRResources(instance.hostname);
  all([wireguard]).apply(([wireguardInstall]) =>
    installFRR(instance.sshIPv4, sshKey.privateKeyPem, frrData, [
      wireguardInstall,
    ]),
  );

  // write output files
  writeFilePulumiAndUploadToS3('ssh.key', sshKey.privateKeyPem, {
    permissions: '0600',
  });
  writeFilePulumiAndUploadToS3(
    'vault.yml',
    all([vaultInstanceData.address, vaultInstanceData.keys]).apply(
      ([address, keys]) =>
        stringify({
          address: address,
          keys: keys,
        }),
    ),
    {
      permissions: '0600',
    },
  );

  return {
    server: {
      ipv4: instance.publicIPv4,
      ipv6: instance.publicIPv6,
    },
    vault: {
      address: vaultInstanceData.address,
      storage: {
        type: 'gcs',
        bucket: vaultData.bucket.id,
      },
      keys: vaultInstanceData.keys,
      ownedSecrets: {
        mount: vaultInstanceData.ownedSecrets.mount.path,
        keys: vaultInstanceData.ownedSecrets.keys.path,
      },
    },
    wireguard: {
      adminPassword: wireguardData.adminPassword,
    },
  };
};

/**

pulumi config set --path 'bgp.neighbors[0].remoteAsn' 65011
pulumi config set --path 'bgp.neighbors[1].remoteAsn' 65021
pulumi config set --path 'bgp.neighbors[2].remoteAsn' 65031

 */