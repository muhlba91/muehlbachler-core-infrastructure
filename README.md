# muehlbachler: Core Infrastructure

[![Build status](https://img.shields.io/github/actions/workflow/status/muhlba91/muehlbachler-core-infrastructure/pipeline.yml?style=for-the-badge)](https://github.com/muhlba91/muehlbachler-core-infrastructure/actions/workflows/pipeline.yml)
[![License](https://img.shields.io/github/license/muhlba91/muehlbachler-core-infrastructure?style=for-the-badge)](LICENSE.md)
[![](https://api.scorecard.dev/projects/github.com/muhlba91/muehlbachler-core-infrastructure/badge?style=for-the-badge)](https://scorecard.dev/viewer/?uri=github.com/muhlba91/muehlbachler-core-infrastructure)

This repository contains the infrastructure as code (IaC) for the Core Infrastructure using [Pulumi](http://pulumi.com).

> **ℹ️ Info:**  
> This repository was previously named `vault` as it initially only contained the Vault deployment.
> There are leftover references to this name in the code, and documentation, which cannot be changed at this point.

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
- `HCLOUD_TOKEN`: the Hetzner Cloud API token

---

## Services

The following services are deployed:

- [HashiCorp Vault](https://www.vaultproject.io) as the secret management solution

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
  name: the Hetzner Cloud network name
  cidr: the CIDR of the internal network
  subnetCidr: the CIDR of the internal network subnet
```

### OIDC

The OIDC configuration to connect the instance to for login.

```yaml
oidc:
  discoveryUrl: the OIDC discovery url (without ".well-known")
  clients: a map containing the OIDC clients
    <name>:
      clientId: the client id
      clientSecret: the client secret
```

### Server

The Hetzner server configuration.

```yaml
server:
  location: the Hetzner Cloud server location
  type: the Hetzner Cloud server type
  ipv4: the IPv4 address of the server
  publicSsh: whether to allow public SSH access
```

### DNS

```yaml
dns:
  project: the Google Cloud project
  email: the ACME email address
  entries: a map containing the DNS entries to create
    <name>:
      domain: the domain name
      zoneId: the Google Cloud DNS zone identifier
```

---

## Continuous Integration and Automations

- [GitHub Actions](https://docs.github.com/en/actions) are linting, and verifying the code.
- [Renovate Bot](https://github.com/renovatebot/renovate) is updating NodeJS packages, and GitHub Actions.
