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

- [Go](https://golang.org/dl/)
- [Pulumi](https://www.pulumi.com/docs/install/)

## Creating the Infrastructure

To create the services, a [Pulumi Stack](https://www.pulumi.com/docs/concepts/stack/) with the correct configuration needs to exists.

The stack can be deployed via:

```bash
pulumi up
```

## Destroying the Infrastructure

The entire infrastructure can be destroyed via:

```bash
pulumi destroy
```

## Environment Variables

To successfully run, and configure the Pulumi plugins, you need to set a list of environment variables. Alternatively, refer to the used Pulumi provider's configuration documentation.

- `CLOUDSDK_COMPUTE_REGION` the Google Cloud (GCP) region
- `GOOGLE_APPLICATION_CREDENTIALS`: reference to a file containing the Google Cloud (GCP) service account credentials
- `HCLOUD_TOKEN`: the Hetzner Cloud API token

---

## Configuration

The following section describes the configuration which must be set in the Pulumi Stack.

***Attention:*** do use [Secrets Encryption](https://www.pulumi.com/docs/concepts/secrets/#:~:text=Pulumi%20never%20sends%20authentication%20secrets,“secrets”%20for%20extra%20protection.) provided by Pulumi for secret values!

### Bucket

```yaml
bucketId: the bucket identifier to store output assets in
backupBucketId: the backup bucket identifier
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
  dnsSuffix: the DNS suffix for internal DNS entries
  firewallRules: a map containing the firewall rules
    <name>:
      description: the description of the rule
      protocol: the protocol (tcp/udp/icmp/..., optional)
      port: the port or port range
      sourceIps: the source IPs (for inbound rules, optional)
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

### BGP

```yaml
bgp:
  localAsn: the local ASN
  advertisedIPv4Networks: a list of IPv4 networks to advertise
  advertisedIPv6Networks: a list of IPv6 networks to advertise
  neighbors: a map of BGP neighbors
    <name>:
      asn: the ASN (optional due to usage of peer groups)
      address: the neighbor address
      interfaceName: the BGP interface
      isPublic: whether the neighbor is a public peer
      password: the BGP neighbor password (optional, will be set automatically for internal peers if not specified)
      gre: the GRE tunnel configuration for this neighbor, if applicable
        remoteIp: the GRE neighbor address
        tunnelIp: the GRE tunnel IP address
        type: the type of the GRE network interface (optional, default: "gre")
  internalNetworks: the internal networks to be advertised
    ipv4: a list of IPv4 networks to advertise internally
    ipv6: a list of IPv6 networks to advertise internally
  publicNetworks: the public networks to be advertised
    ipv4: a list of IPv4 networks to advertise publicly
    ipv6: a list of IPv6 networks to advertise publicly
```

### Tailscale

```yaml
tailscale:
  authKey: the Tailscale auth key
```

> [!IMPORTANT]  
> The Tailscale auth key must be recreated periodically, if reused due to expiration.

### NetBird

Only the NetBird **server** is installed by Pulumi (DNS entry `netbird` is required, see [DNS](#dns)).
The `netbird-stun` firewall rule (UDP `3478`) and the `netbird-client` firewall rule (UDP, the same port as `netbird.client.wireguardPort`) must exist, see [Network](#network).

```yaml
netbird:
  storeEncryptionKey: the base64 encoded 32 byte key encrypting sensitive data in the database (e.g. openssl rand -base64 32)
  admin: the initial owner, created once via the setup API
    name: the display name
    email: the login email
  networkRange: the IPv4 range (CIDR) the peers are addressed from (e.g. 10.253.0.0/16): it must not overlap with other network routed over the overlay
  client: the NetBird client running on this server
    wireguardPort: the fixed WireGuard (UDP) port of the client, must match the netbird-client firewall rule (e.g. 65500)
    mtu: the MTU of the client interface (e.g. 1420: the maximum for WireGuard on a 1500 byte link, 1370 remain for VXLAN on top)
```

> [!IMPORTANT]  
> `netbird.storeEncryptionKey` **must stay the same** across restores and full re-bootstraps:
> the restored database cannot be decrypted without the key.
> After a full re-bootstrap with lost Pulumi state, the owner already
> exists in the restored data, so a newly generated password does not apply: reset it in NetBird.

> [!IMPORTANT]  
> The Pulumi provider does not know the IPv6 settings of the NetBird account, and every update of the account settings (including the first apply, e.g., a changed `networkRange`) resets them:
> afterwards, add the `All` group to the IPv6 enabled groups again in NetBird, otherwise the peers lose their IPv6 addresses.

> [!IMPORTANT]  
> The owner created by the setup endpoint has no domain, so SSO users would each open their own account instead of joining it.
> Directly after the setup, the initialization sets the domain `netbird.selfhosted` (the one hard-coded by the NetBird server), the category `private`, and the primary flag on the owner account.
> This is only done on the first installation: for an existing instance, set `domain`, `domain_category`, and `is_domain_primary_account` of the owner account manually, and restart NetBird.

> [!IMPORTANT]  
> If the NetBird database is lost, all clients must be registered again with a new setup key.

---

## Continuous Integration and Automations

- [GitHub Actions](https://docs.github.com/en/actions) are linting, and verifying the code.
- [Renovate Bot](https://github.com/renovatebot/renovate) is updating Go modules, and GitHub Actions.
