package netbird

// Config defines configuration data for NetBird.
type Config struct {
	// StoreEncryptionKey is the base64 encoded 32 byte key encrypting sensitive data at rest in the database.
	// Losing it makes the restored database unreadable.
	StoreEncryptionKey *string `yaml:"storeEncryptionKey,omitempty"`
	// Admin is the initial admin of the NetBird instance.
	Admin *AdminConfig `yaml:"admin,omitempty"`
	// NetworkRange is the IPv4 range (CIDR) the NetBird peers are addressed from.
	// It must not overlap Tailscale (100.64.0.0/10), nor any other network routed with it.
	NetworkRange *string `yaml:"networkRange,omitempty"`
	// Client is the NetBird client of this server.
	Client *ClientConfig `yaml:"client,omitempty"`
}

// AdminConfig defines configuration data for the initial NetBird admin.
type AdminConfig struct {
	// Name is the display name.
	Name *string `yaml:"name,omitempty"`
	// Email is the email address (login).
	Email *string `yaml:"email,omitempty"`
}

// ClientConfig defines configuration data for the NetBird client.
type ClientConfig struct {
	// WireguardPort is the fixed WireGuard (UDP) port of the client: it must be allowed in the firewall.
	WireguardPort *int `yaml:"wireguardPort,omitempty"`
	// MTU is the MTU of the client interface.
	// 1420 is the maximum for WireGuard on a 1500 byte link (1370 remain for VXLAN on top).
	MTU *int `yaml:"mtu,omitempty"`
}
