# Bootstrap (day 1)

This repository creates Vault, which every other repository relies on. It must therefore be deployable from a
developer machine (Mac) **without Vault and without CI**. This document lists everything that has to exist before
the first `pulumi up`, and the order of the initial bootstrap.

- **Day 1** is the initial bootstrap of everything, run from the Mac. Nothing in it depends on resources this
  repository or `github-infrastructure` create.
- **Day 2** is everything after. CI works from day 2 on: its Google identity (workload identity) is created by
  `github-infrastructure`, which runs after this repository.

## Prerequisites

Created by hand, outside of any repository:

1. **Google Cloud:** the project `public-services-402810` with
   - the KMS key `europe/infrastructure-encryption/infrastructure-encryption` (Pulumi's secrets provider and Vault's
     auto-unseal),
   - the Secret Manager API enabled (`gcloud services enable secretmanager.googleapis.com --project public-services-402810`).
2. **Your Google login** on the Mac: `gcloud auth login` and `gcloud auth application-default login`, with the rights
   this program needs in the project.
3. **Pulumi Cloud:** the account and `pulumi login` on the Mac.
4. **Scaleway:** an IAM application used only by this repository, with an API key and a policy granting
   - `IAMManager` on the organization (the program creates the Vault server's IAM application, policy and API key),
   - `ObjectStorageFullAccess` on the projects holding the Vault, assets and backup buckets.
5. **Hetzner Cloud:** an API token for the project the servers run in.
6. **mise** on the Mac (it installs Go, Pulumi and fnox from `mise.toml`).

## Secrets

The secrets live in Google Secret Manager in `public-services-402810`, one secret per environment variable, named
`muehlbachler-core-infrastructure--<VARIABLE>`:

| Secret | Content |
| --- | --- |
| `muehlbachler-core-infrastructure--SCW_ACCESS_KEY` | Scaleway API key: access key |
| `muehlbachler-core-infrastructure--SCW_SECRET_KEY` | Scaleway API key: secret key |
| `muehlbachler-core-infrastructure--HCLOUD_TOKEN` | Hetzner Cloud API token |

Seed or rotate them with the script (it prompts for values without echo and never prints them):

```bash
bootstrap/seed-secrets.sh --check   # lists missing secrets
bootstrap/seed-secrets.sh           # creates missing secrets and prompts for values
```

Non-secret configuration (regions, Scaleway organization and project) is set as defaults in `fnox.toml`; the
remaining secret configuration (Tailscale auth key, NetBird store encryption key) is encrypted in the stack
configuration with the KMS key.

## Order (day 1)

1. Prerequisites above.
2. `bootstrap/seed-secrets.sh`, then `bootstrap/seed-secrets.sh --check`.
3. This repository: `fnox exec -- pulumi up --stack prod`.
4. `github-infrastructure` (it reads Vault's root token from this stack's outputs and creates the repositories,
   Vault's mounts and the CI identities, including this repository's).
5. Day 2: CI of this repository works.

## Notes

- `mise.toml` sets `FNOX_PROFILE=default` in this repository, so a developer machine's `dev` fnox profile (which reads
  from Vault) is never used here.
- Local overrides go into `fnox.local.toml` (gitignored).
