package tailscale

// Config defines configuration data for Tailscale.
type Config struct {
	// AuthKey is the Tailscale auth key.
	//nolint:gosec // This is a configuration value, not a hardcoded secret.
	AuthKey *string `yaml:"authKey,omitempty"`
}
