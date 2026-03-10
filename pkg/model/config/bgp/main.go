package bgp

// Config defines configuration data for BGP.
type Config struct {
	// RouterID is the BGP router identifier.
	RouterID *string `yaml:"routerId,omitempty"`
	// LocalASN is the BGP local autonomous system number.
	LocalASN uint32 `yaml:"localAsn,omitempty"`
	// Neighbors are the BGP neighbors.
	Neighbors []NeighborConfig `yaml:"neighbors,omitempty"`
	// InternalNetworks are the internal networks to be advertised.
	InternalNetworks *AdvertisedNetworksConfig `yaml:"internalNetworks,omitempty"`
	// PublicNetworks are the public networks to be advertised.
	PublicNetworks *AdvertisedNetworksConfig `yaml:"publicNetworks,omitempty"`
}

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
}

// AdvertisedNetworksConfig defines configuration data for advertised networks in BGP.
type AdvertisedNetworksConfig struct {
	// IPv4 are the advertised IPv4 networks.
	IPv4 []string `yaml:"ipv4,omitempty"`
	// IPv6 are the advertised IPv6 networks.
	IPv6 []string `yaml:"ipv6,omitempty"`
}
