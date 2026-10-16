package bgp

// EVPNConfig defines configuration data for EVPN (L3VNI) over the NetBird overlay.
// The site routers are dynamic BGP neighbors from the NetBird network range.
type EVPNConfig struct {
	// VRF is the name of the VRF carrying the L3VNI.
	VRF *string `yaml:"vrf,omitempty"`
	// VNI is the VXLAN network identifier of the L3VNI: it must be the same on all sites.
	VNI *int `yaml:"vni,omitempty"`
	// Table is the kernel routing table of the VRF.
	Table *int `yaml:"table,omitempty"`
	// MTU is the MTU of the VXLAN devices: the NetBird client MTU minus 50 bytes VXLAN overhead.
	MTU *int `yaml:"mtu,omitempty"`
	// LeakNetworks are the networks leaked between the default VRF and the EVPN VRF (both directions).
	// The public IPv6 networks (PublicNetworks) are always leaked, and the NetBird network range never.
	LeakNetworks *AdvertisedNetworksConfig `yaml:"leakNetworks,omitempty"`
	// ClientNAT are the host addresses in the EVPN VRF which the traffic of NetBird clients is masqueraded to.
	// The server routes the clients into the sites, and the sites route the replies back to these addresses (EVPN):
	// the masquerade only picks a source address of the VRF the traffic leaves into.
	ClientNAT *AdvertisedNetworksConfig `yaml:"clientNat,omitempty"`
	// Egress is the Internet egress of the sites through the upstream neighbors (optional, disabled if unset).
	Egress *EgressConfig `yaml:"egress,omitempty"`
}

// EgressConfig defines the Internet egress of the sites for sources of the public networks.
// The providers of the sites drop these sources: the replies of exposed services leave through the hub instead.
// The upstream neighbors provide the default route, which is only used by the EVPN VRF (the sites).
type EgressConfig struct {
	// MaximumPaths is the number of upstream default routes used in parallel (ECMP, default: 2).
	MaximumPaths *int `yaml:"maximumPaths,omitempty"`
}
