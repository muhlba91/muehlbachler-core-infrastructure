package google

// Config defines configuration data for GCP.
type Config struct {
	// Project is the GCP project ID.
	Project *string `yaml:"project,omitempty"`
	// Region is the GCP region.
	Region *string `yaml:"region,omitempty"`
	// EncryptionKey is the GCP encryption key configuration.
	EncryptionKey *EncryptionKeyConfig `yaml:"encryptionKey,omitempty"`
}

// EncryptionKeyConfig defines encryption key configuration data for GCP.
type EncryptionKeyConfig struct {
	// Location is the location of the encryption key.
	Location *string `yaml:"location,omitempty"`
	// KeyringID is the keyring ID of the encryption key.
	KeyringID *string `yaml:"keyringId,omitempty"`
	// CryptoKeyID is the crypto key ID of the encryption key.
	CryptoKeyID *string `yaml:"cryptoKeyId,omitempty"`
}
