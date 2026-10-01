package netbird

// Config defines configuration data for NetBird.
type Config struct {
	// StoreEncryptionKey is the base64 encoded 32 byte key encrypting sensitive data at rest in the database.
	// Losing it makes the restored database unreadable.
	StoreEncryptionKey *string `yaml:"storeEncryptionKey,omitempty"`
	// Admin is the initial admin of the NetBird instance.
	Admin *AdminConfig `yaml:"admin,omitempty"`
}

// AdminConfig defines configuration data for the initial NetBird admin.
type AdminConfig struct {
	// Name is the display name.
	Name *string `yaml:"name,omitempty"`
	// Email is the email address (login).
	Email *string `yaml:"email,omitempty"`
}
