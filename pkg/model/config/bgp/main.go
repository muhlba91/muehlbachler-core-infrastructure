package bgp

// Config defines configuration data for BGP.
type Config struct {
	// LocalASN is the BGP local autonomous system number.
	LocalASN uint32 `yaml:"localAsn,omitempty"`
	// Neighbors are the BGP neighbors.
	Neighbors map[string]*NeighborConfig `yaml:"neighbors,omitempty"`
	// InternalNetworks are the internal networks to be advertised.
	InternalNetworks *AdvertisedNetworksConfig `yaml:"internalNetworks,omitempty"`
	// PublicNetworks are the public networks to be advertised.
	PublicNetworks *AdvertisedNetworksConfig `yaml:"publicNetworks,omitempty"`
}

// AdvertisedNetworksConfig defines configuration data for advertised networks in BGP.
type AdvertisedNetworksConfig struct {
	// IPv4 are the advertised IPv4 networks.
	IPv4 []string `yaml:"ipv4,omitempty"`
	// IPv6 are the advertised IPv6 networks.
	IPv6 []string `yaml:"ipv6,omitempty"`
}
