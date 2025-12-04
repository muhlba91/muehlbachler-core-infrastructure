package dns

// Config defines configuration data for DNS.
type Config struct {
	// Project is the DNS project identifier.
	Project *string `yaml:"project,omitempty"`
	// Email is the DNS contact email.
	Email *string `yaml:"email,omitempty"`
	// Entries are the DNS entries.
	Entries map[string]EntryConfig `yaml:"entries,omitempty"`
}

// EntryConfig defines configuration data for a DNS entry.
type EntryConfig struct {
	// Domain is the DNS entry domain.
	Domain *string `yaml:"domain,omitempty"`
	// ZoneID is the DNS entry zone ID.
	ZoneID *string `yaml:"zoneId,omitempty"`
}
