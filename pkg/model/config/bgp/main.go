package bgp

// Config defines configuration data for BGP.
type Config struct {
	// RouterID is the BGP router identifier.
	RouterID *string `yaml:"routerId,omitempty"`
	// LocalASN is the BGP local autonomous system number.
	LocalASN uint32 `yaml:"localAsn,omitempty"`
	// InterfaceName is the name of the interface.
	InterfaceName *string `yaml:"interfaceName,omitempty"`
	// Neighbors are the BGP neighbors.
	Neighbors []NeighborConfig `yaml:"neighbors,omitempty"`
	// AdvertisedIPv4Networks are the advertised IPv4 networks.
	AdvertisedIPv4Networks []string `yaml:"advertisedIPv4Networks,omitempty"`
	// AdvertisedIPv6Networks are the advertised IPv6 networks.
	AdvertisedIPv6Networks []string `yaml:"advertisedIPv6Networks,omitempty"`
}

// NeighborConfig defines configuration data for a BGP neighbor.
type NeighborConfig struct {
	// Address is the BGP neighbor address.
	Address *string `yaml:"address,omitempty"`
	// RemoteASN is the BGP neighbor remote autonomous system number.
	RemoteASN uint32 `yaml:"remoteAsn,omitempty"`
	// InterfaceName is the name of the interface.
	InterfaceName *string `yaml:"interfaceName,omitempty"`
}
