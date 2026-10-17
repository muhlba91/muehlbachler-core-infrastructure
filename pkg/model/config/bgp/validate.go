package bgp

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
)

const (
	// greInterfacePrefix is the required name prefix of GRE interfaces.
	greInterfacePrefix = "gre-"
	// GreTypeGre is the GRE type for a plain (layer 3) GRE tunnel.
	GreTypeGre = "gre"
	// greTypeGretap is the GRE type for a GRETAP (layer 2) tunnel.
	greTypeGretap = "gretap"
	// DefaultGreMTU is the MTU of GRE interfaces if none is configured.
	DefaultGreMTU = 1472
	// maxGreMTU is the maximum MTU of GRE interfaces: 1500 bytes minus the IPv4 (20), and GRE (4) headers.
	maxGreMTU = 1476
	// maxInterfaceNameLength is the maximum length of a Linux interface name.
	maxInterfaceNameLength = 15
	// defaultLocalPref is the local preference of an upstream default route if none is configured.
	defaultLocalPref = 100
	// maxPrepend is the maximum number of prepended ASNs.
	maxPrepend = 10
	// defaultMaximumPaths is the number of upstream default routes used in parallel if none is configured.
	defaultMaximumPaths = 2
	// maxMaximumPaths is the maximum number of upstream default routes used in parallel.
	maxMaximumPaths = 8
)

// Validate checks the configuration, and replaces unset network lists with empty ones.
// It fails with all problems found, so the templates, and installers never meet nil values.
func (c *Config) Validate() error {
	errs := []error{}

	if c.PublicNetworks == nil {
		c.PublicNetworks = &AdvertisedNetworksConfig{}
	}

	for _, name := range slices.Sorted(maps.Keys(c.Neighbors)) {
		if err := c.Neighbors[name].validate(); err != nil {
			errs = append(errs, fmt.Errorf("bgp.neighbors.%s: %w", name, err))
		}
	}

	if err := c.EVPN.validate(); err != nil {
		errs = append(errs, fmt.Errorf("bgp.evpn: %w", err))
	}
	if c.EVPN != nil && c.EVPN.Egress != nil && len(c.PublicNetworks.IPv6) == 0 {
		errs = append(errs, errors.New("bgp.evpn.egress: publicNetworks.ipv6 is required"))
	}

	return errors.Join(errs...)
}

// validate checks a neighbor.
func (n *NeighborConfig) validate() error {
	if n == nil {
		return errors.New("must not be empty")
	}

	errs := []error{}
	if n.ASN == nil {
		errs = append(errs, errors.New("asn is required"))
	}
	if n.GRE != nil {
		errs = append(errs, n.validateGRE())
	}
	n.LocalPref = orDefault(n.LocalPref, defaultLocalPref)
	n.Prepend = orDefault(n.Prepend, 0)
	if *n.LocalPref < 1 {
		errs = append(errs, errors.New("localPref must be positive"))
	}
	if *n.Prepend < 0 || *n.Prepend > maxPrepend {
		errs = append(errs, fmt.Errorf("prepend must be between 0 and %d", maxPrepend))
	}

	return errors.Join(errs...)
}

// validateGRE checks the GRE tunnel of a neighbor.
func (n *NeighborConfig) validateGRE() error {
	errs := []error{}

	switch {
	case n.InterfaceName == nil:
		errs = append(errs, errors.New("interfaceName is required for a GRE tunnel"))
	case !strings.HasPrefix(*n.InterfaceName, greInterfacePrefix):
		errs = append(errs, fmt.Errorf("interfaceName must start with %q", greInterfacePrefix))
	case len(*n.InterfaceName) > maxInterfaceNameLength:
		errs = append(errs, fmt.Errorf("interfaceName must not exceed %d characters", maxInterfaceNameLength))
	}
	if n.GRE.RemoteIP == nil {
		errs = append(errs, errors.New("gre.remoteIp is required"))
	}
	if n.GRE.TunnelIP == nil {
		errs = append(errs, errors.New("gre.tunnelIp is required"))
	}
	if n.GRE.LinkLocalIP == nil {
		errs = append(errs, errors.New("gre.linkLocalIp is required"))
	}
	if n.GRE.Type != nil && *n.GRE.Type != GreTypeGre && *n.GRE.Type != greTypeGretap {
		errs = append(errs, fmt.Errorf("gre.type must be %q or %q", GreTypeGre, greTypeGretap))
	}
	if n.GRE.Key != nil && *n.GRE.Key == 0 {
		errs = append(errs, errors.New("gre.key must be positive"))
	}
	if n.GRE.MTU != nil && (*n.GRE.MTU < 1280 || *n.GRE.MTU > maxGreMTU) {
		errs = append(errs, fmt.Errorf("gre.mtu must be between 1280 and %d", maxGreMTU))
	}

	return errors.Join(errs...)
}

// validate checks the EVPN configuration, which the NetBird, FRR, and EVPN installations require.
func (e *EVPNConfig) validate() error {
	if e == nil {
		return errors.New("is required")
	}

	errs := []error{}
	if e.VRF == nil {
		errs = append(errs, errors.New("vrf is required"))
	}
	if e.VNI == nil {
		errs = append(errs, errors.New("vni is required"))
	}
	if e.Table == nil {
		errs = append(errs, errors.New("table is required"))
	}
	if e.MTU == nil {
		errs = append(errs, errors.New("mtu is required"))
	}
	if e.LeakNetworks == nil {
		e.LeakNetworks = &AdvertisedNetworksConfig{}
	}
	if e.ClientNAT == nil {
		e.ClientNAT = &AdvertisedNetworksConfig{}
	}
	if e.Egress != nil {
		e.Egress.MaximumPaths = orDefault(e.Egress.MaximumPaths, defaultMaximumPaths)
		if *e.Egress.MaximumPaths < 1 || *e.Egress.MaximumPaths > maxMaximumPaths {
			errs = append(errs, fmt.Errorf("egress.maximumPaths must be between 1 and %d", maxMaximumPaths))
		}
	}

	return errors.Join(errs...)
}

// orDefault returns the value, or a pointer to the default if it is unset.
// value: The configured value.
// def: The default value.
func orDefault(value *int, def int) *int {
	if value == nil {
		return &def
	}
	return value
}
