package bgp

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
)

const (
	// GreInterfacePrefix is the required name prefix of GRE interfaces.
	GreInterfacePrefix = "gre-"
	// GreTypeGre is the GRE type for a plain (layer 3) GRE tunnel.
	GreTypeGre = "gre"
	// GreTypeGretap is the GRE type for a GRETAP (layer 2) tunnel.
	GreTypeGretap = "gretap"
	// DefaultGreMTU is the MTU of GRE interfaces if none is configured.
	DefaultGreMTU = 1472
	// maxInterfaceNameLength is the maximum length of a Linux interface name.
	maxInterfaceNameLength = 15
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

	return errors.Join(errs...)
}

// validateGRE checks the GRE tunnel of a neighbor.
func (n *NeighborConfig) validateGRE() error {
	errs := []error{}

	switch {
	case n.InterfaceName == nil:
		errs = append(errs, errors.New("interfaceName is required for a GRE tunnel"))
	case !strings.HasPrefix(*n.InterfaceName, GreInterfacePrefix):
		errs = append(errs, fmt.Errorf("interfaceName must start with %q", GreInterfacePrefix))
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
	if n.GRE.Type != nil && *n.GRE.Type != GreTypeGre && *n.GRE.Type != GreTypeGretap {
		errs = append(errs, fmt.Errorf("gre.type must be %q or %q", GreTypeGre, GreTypeGretap))
	}
	if n.GRE.MTU != nil && (*n.GRE.MTU < 1280 || *n.GRE.MTU > DefaultGreMTU) {
		errs = append(errs, fmt.Errorf("gre.mtu must be between 1280 and %d", DefaultGreMTU))
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

	return errors.Join(errs...)
}
