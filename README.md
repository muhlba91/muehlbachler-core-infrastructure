# muehlbachler: Hashicorp Vault - Infrastructure

[![Build status](https://img.shields.io/github/actions/workflow/status/muhlba91/muehlbachler-hashicorp-vault-infrastructure/pipeline.yml?style=for-the-badge)](https://github.com/muhlba91/muehlbachler-hashicorp-vault-infrastructure/actions/workflows/pipeline.yml)
[![License](https://img.shields.io/github/license/muhlba91/muehlbachler-hashicorp-vault-infrastructure?style=for-the-badge)](LICENSE.md)
[![](https://api.scorecard.dev/projects/github.com/muhlba91/muehlbachler-hashicorp-vault-infrastructure/badge?style=for-the-badge)](https://scorecard.dev/viewer/?uri=github.com/muhlba91/muehlbachler-hashicorp-vault-infrastructure)

This repository contains the infrastructure as code (IaC) for the [Hashicorp Vault instance](https://www.vaultproject.io) using [Pulumi](http://pulumi.com).

---

## Requirements

- [NodeJS](https://nodejs.org/en), and [yarn](https://yarnpkg.com)
- [Pulumi](https://www.pulumi.com/docs/install/)

## Creating the Infrastructure

To create the infrastructure and deploy the cluster, a [Pulumi Stack](https://www.pulumi.com/docs/concepts/stack/) with the correct configuration needs to exists.

The stack can be deployed via:

```bash
yarn install
yarn build; pulumi up
```

## Destroying the Infrastructure

The entire infrastructure can be destroyed via:

```bash
yarn install
yarn build; pulumi destroy
```

## Environment Variables

To successfully run, and configure the Pulumi plugins, you need to set a list of environment variables. Alternatively, refer to the used Pulumi provider's configuration documentation.

- `CLOUDSDK_COMPUTE_REGION` the Google Cloud (GCP) region
- `GOOGLE_APPLICATION_CREDENTIALS`: reference to a file containing the Google Cloud (GCP) service account credentials
- `GITHUB_TOKEN`: the GitHub Personal Access Token (PAT)
- `PROXMOX_VE_USERNAME`: the Proxmox username
- `PROXMOX_VE_PASSWORD`: the Proxmox password
- `PROXMOX_VE_ENDPOINT`: the endpoint to connect to Proxmox
- `PROXMOX_VE_INSECURE`: turn on/off insecure connections to Proxmox

---

## Configuration

The following section describes the configuration which must be set in the Pulumi Stack.

***Attention:*** do use [Secrets Encryption](https://www.pulumi.com/docs/concepts/secrets/#:~:text=Pulumi%20never%20sends%20authentication%20secrets,“secrets”%20for%20extra%20protection.) provided by Pulumi for secret values!

### Bucket Identifier

```yaml
bucketId: the bucket identifier to upload assets to
```

### Google Cloud (GCP)

Flux deployed applications can reference secrets being encrypted with [sops](https://github.com/mozilla/sops).
We need to specify, and allow access to this encryption stored in [Google KMS](https://cloud.google.com/security-key-management).

```yaml
gcp:
  project: the GCP project to create all resources in
  region: the GCP region to create resources in
  encryptionKey: references the sops encryption key
    cryptoKeyId: the CryptoKey identifier
    keyringId: the KeyRing identifier
    location: the location of the key
```

### Network

General configuration about the local network.

```yaml
network:
  domain: the internal DNS domain
  ipv4:
    cidrMask: the CIDR mask of the internal network
    enabled: enables IPv4 networking
    gateway: the IPv4 gateway
  ipv6:
    cidrMask: the CIDR mask of the internal network
    enabled: enables IPv6 networking
    gateway: the IPv6 gateway
  nameservers: a list of all nameservers to set (IPv4, IPv6)
```

### OIDC

The OIDC configuration to connect the Vault instance to for login.

```yaml
oidc:
  discoveryUrl: the OIDC discovery url (without ".well-known")
  clientId: the client id
  clientSecret: the client secret
  redirectUrls: a list of redirect URLs to set
```

### Proxmox VE (pve)

General configuration about the Proxmox environment.

***Attention:*** you must download the specifief `imageName` to each Proxmox host!

```yaml
pve:
  cpuType: the default CPU type to assign to machines
  imageName: the reference to the locally installed image
  localStoragePool: the storage pool used for snippets
  networkBridge: the network bridge to use for server connectivity
  storagePool: the storage pool used for machine disks
```

### Server

The Proxmox server configuration.

```yaml
server:
  cpu: the CPU allocation
  diskSize: the disk size to use
  memory: memory configuration (enables or disables ballooning automatically)
    min: the minimum memory to assign
    max: the maximum memory to assign
  host: the Proxmox host to create the node on
  ipv4Address: the internal IPv4 address
  ipv6Address: the internal IPv6 address (optional)
  startupOrder: the order when the VM should be automatically started
```

### Username

```yaml
username: the username to use for interacting with the servers
```

---

## Continuous Integration and Automations

- [GitHub Actions](https://docs.github.com/en/actions) are linting, and verifying the code.
- [Renovate Bot](https://github.com/renovatebot/renovate) is updating NodeJS packages, and GitHub Actions.
