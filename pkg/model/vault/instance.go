package vault

// Instance holds references to resources of the Vault instance.
type Instance struct {
	// The GCS bucket used by Vault for storage.
	Bucket string
	// The Vault server address.
	Address string
	// The Vault keys.
	Keys *Keys
	// The Vault owned secrets.
	OwnedSecrets *OwnedSecrets
}
