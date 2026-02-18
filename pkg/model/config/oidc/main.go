package oidc

// Config defines configuration data for OIDC.
type Config struct {
	// DiscoveryURL is the OIDC discovery URL.
	DiscoveryURL *string `yaml:"discoveryUrl,omitempty"`
	// Clients is a map of OIDC client configurations.
	Clients map[string]*ClientConfig `yaml:"clients,omitempty"`
}

// ClientConfig defines configuration data for an OIDC client.
type ClientConfig struct {
	// ClientID is the OIDC client ID.
	ClientID *string `yaml:"clientId,omitempty"`
	// ClientSecret is the OIDC client secret.
	//nolint:gosec // This is a configuration value, not a hardcoded secret.
	ClientSecret *string `yaml:"clientSecret,omitempty"`
}
