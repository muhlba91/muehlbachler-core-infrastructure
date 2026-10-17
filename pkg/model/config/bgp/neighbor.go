package bgp

import (
	"fmt"
	"strings"
)

// NeighborConfig defines configuration data for a BGP neighbor.
type NeighborConfig struct {
	// Addresses is the list of BGP neighbor addresses: without any, no BGP session is configured (the tunnel is kept).
	Addresses []string `yaml:"addresses,omitempty"`
	// ASN is the BGP neighbor autonomous system number.
	ASN *uint32 `yaml:"asn,omitempty"`
	// InterfaceName is the name of the interface.
	InterfaceName *string `yaml:"interfaceName,omitempty"`
	// Password is the BGP neighbor password.
	Password *string `yaml:"password,omitempty"`
	// GRE contains GRE tunnel configuration for this neighbor, if applicable.
	GRE *GreConfig `yaml:"gre,omitempty"`
	// LocalPref is the local preference of the default route learned from this neighbor (default: 100):
	// equal values share the traffic (ECMP), the highest is preferred. Only used with the EVPN egress.
	LocalPref *int `yaml:"localPref,omitempty"`
	// Prepend is the number of times the local ASN is additionally prepended to the announced networks (default: 0),
	// which makes this neighbor less preferred for incoming traffic.
	Prepend *int `yaml:"prepend,omitempty"`
}

// PrependPath returns the AS path prepended to the announced networks: the local ASN, repeated.
// localASN: The local ASN.
func (n *NeighborConfig) PrependPath(localASN uint32) string {
	if n.Prepend == nil {
		return ""
	}
	return strings.TrimSpace(strings.Repeat(fmt.Sprintf("%d ", localASN), *n.Prepend))
}

// GreConfig defines configuration data for a GRE tunnel associated with a BGP neighbor.
type GreConfig struct {
	// RemoteIP is the GRE neighbor address.
	RemoteIP *string `yaml:"remoteIp,omitempty"`
	// TunnelIP is the GRE tunnel IP address.
	TunnelIP *string `yaml:"tunnelIp,omitempty"`
	// LinkLocalIP is the GRE link-local IP address.
	LinkLocalIP *string `yaml:"linkLocalIp,omitempty"`
	// Type is the type of the GRE network interface: "gre" (default), or "gretap".
	Type *string `yaml:"type,omitempty"`
	// Key is the optional GRE key identifying the tunnel (input and output): required by peers
	// if there are multiple tunnels to the same endpoint.
	Key *uint32 `yaml:"key,omitempty"`
	// MTU is the MTU of the GRE network interface (default: 1472, max. 1476).
	MTU *int `yaml:"mtu,omitempty"`
}
