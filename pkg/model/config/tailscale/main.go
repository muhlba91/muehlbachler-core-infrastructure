package tailscale

// Config defines configuration data for Tailscale.
type Config struct {
	// AuthKey is the Tailscale auth key.
	AuthKey *string `yaml:"authKey,omitempty"`
}
