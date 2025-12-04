package vault

import (
	"github.com/pulumi/pulumi-vault/sdk/v7/go/vault"
	"github.com/pulumi/pulumi-vault/sdk/v7/go/vault/kv"
)

// OwnedSecrets holds the Vault owned secrets.
type OwnedSecrets struct {
	// Mount is the Vault mount.
	Mount *vault.Mount
	// Keys are the Vault secrets.
	Keys *kv.SecretV2Output
}
