package bgp

// NeighborConfig defines configuration data for a BGP neighbor.
type NeighborConfig struct {
	// Address is the BGP neighbor address.
	Address *string `yaml:"address,omitempty"`
	// ASN is the BGP neighbor autonomous system number.
	ASN *uint32 `yaml:"asn,omitempty"`
	// InterfaceName is the name of the interface.
	InterfaceName *string `yaml:"interfaceName,omitempty"`
	// IsPublic indicates if the neighbor is a public peer.
	IsPublic bool `yaml:"isPublic,omitempty"`
	// Password is the BGP neighbor password. Will be set automatically for internal peers, if not specified.
	Password *string `yaml:"password,omitempty"`
	// GRE contains GRE tunnel configuration for this neighbor, if applicable.
	GRE *GreConfig `yaml:"gre,omitempty"`
}

// GreConfig defines configuration data for a GRE tunnel associated with a BGP neighbor.
type GreConfig struct {
	// RemoteIP is the GRE neighbor address.
	RemoteIP *string `yaml:"remoteIp,omitempty"`
	// TunnelIP is the GRE tunnel IP address.
	TunnelIP *string `yaml:"tunnelIp,omitempty"`
	// Type is the type of the GRE network interface.
	Type *string `yaml:"type,omitempty"`
}
