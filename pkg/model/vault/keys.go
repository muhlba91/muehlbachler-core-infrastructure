package vault

// Keys holds the Vault keys.
type Keys struct {
	// RootToken is the Vault root token.
	RootToken string `yaml:"rootToken"`
	// RecoveryKeys are the Vault recovery keys.
	RecoveryKeys []string `yaml:"recoveryKeys"`
}
