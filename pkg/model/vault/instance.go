package vault

import "github.com/pulumi/pulumi-vault/sdk/v7/go/vault"

// Instance holds references to resources of the Vault instance.
type Instance struct {
	// The Scaleway bucket used by Vault for storage.
	Bucket string
	// The Vault server address.
	Address string
	// The Vault keys.
	Keys *Keys
	// The Vault owned secrets.
	OwnedSecrets *OwnedSecrets
	// The Vault provider authenticated with the root token.
	Provider *vault.Provider
}
