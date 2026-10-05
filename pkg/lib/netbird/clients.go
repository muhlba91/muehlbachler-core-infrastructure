package netbird

import (
	"regexp"
	"slices"

	"github.com/muhlba91/pulumi-shared-library/pkg/lib/netbird/group"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/netbird/network"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/netbird/policy"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	pModel "github.com/muhlba91/pulumi-shared-library/pkg/model/netbird/policy"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/bgp"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/netbird"
)

const (
	// sitesName is the name of the network of the sites, and of the group of its resources.
	sitesName = "sites"
	// sitesDescription is the description of the network of the sites.
	sitesDescription = "the sites (EVPN/BGP routed): this server routes the clients into them, masqueraded"
	// adminsName is the name of the admins group, and of its policy.
	adminsName = "admins"
	// adminsPolicyDescription is the description of the admins policy.
	adminsPolicyDescription = "admins reach the sites"
)

// CreateClientAccess creates the access of NetBird clients (e.g., laptops) to the sites: this server is the routing peer
// of the network of the sites, and masquerades the clients to the client NAT addresses (EVPN VRF).
// The routed networks are the networks leaked into the EVPN VRF, and the public IPv6 networks: the sites appear
// dynamically through BGP, NetBird only routes the clients to this server.
// The group memberships of the peers are managed in NetBird, not here.
// ctx: Pulumi context.
// hostname: The hostname of the server, which is the name of the NetBird client (peer).
// bgpConfig: BGP configuration.
// netbirdInstance: The NetBird (server) instance.
// netbirdClient: The installation of the NetBird client.
func CreateClientAccess(
	ctx *pulumi.Context,
	hostname pulumi.StringOutput,
	bgpConfig *bgp.Config,
	netbirdInstance *netbird.Instance,
	netbirdClient *remote.Command,
) error {
	opts := []pulumi.ResourceOption{
		pulumi.Provider(netbirdInstance.Provider),
		pulumi.DependsOn([]pulumi.Resource{netbirdInstance.Ready, netbirdInstance.Settings}),
	}

	// the members are managed in NetBird (peers by the dashboard or setup keys, resources by the network resources)
	membersOpts := append(slices.Clone(opts), pulumi.IgnoreChanges([]string{"peers", "resources"}))
	sites, sErr := group.Create(ctx, sitesName, &group.CreateOptions{PulumiOptions: membersOpts})
	if sErr != nil {
		return sErr
	}
	admins, aErr := group.Create(ctx, adminsName, &group.CreateOptions{PulumiOptions: membersOpts})
	if aErr != nil {
		return aErr
	}

	desc := sitesDescription
	sitesNetwork, nErr := network.Create(ctx, sitesName, &network.CreateOptions{
		Description:   &desc,
		PulumiOptions: opts,
	})
	if nErr != nil {
		return nErr
	}

	for _, address := range routedNetworks(bgpConfig) {
		_, rErr := network.CreateResource(ctx, resourceName(address), &network.CreateResourceOptions{
			Address:       address,
			Groups:        pulumi.StringArray{sites.ID().ToStringOutput()},
			NetworkID:     sitesNetwork.ID().ToStringOutput(),
			PulumiOptions: opts,
		})
		if rErr != nil {
			return rErr
		}
	}

	masquerade := true
	nbPeer := LookupPeer(ctx, hostname, netbirdInstance, netbirdClient)
	_, rtErr := network.CreateRouter(ctx, sitesName, &network.CreateRouterOptions{
		NetworkID:     sitesNetwork.ID().ToStringOutput(),
		Peer:          nbPeer.MapIndex(pulumi.String(PeerID)),
		Masquerade:    &masquerade,
		PulumiOptions: opts,
	})
	if rtErr != nil {
		return rtErr
	}

	policyDesc, accept, all, no := adminsPolicyDescription, "accept", "all", false
	_, pErr := policy.Create(ctx, adminsName, &policy.CreateOptions{
		Description: &policyDesc,
		Rules: []pModel.Rule{{
			Name:          adminsName,
			Action:        &accept,
			Bidirectional: &no,
			Protocol:      &all,
			Sources:       pulumi.StringArray{admins.ID().ToStringOutput()},
			Destinations:  pulumi.StringArray{sites.ID().ToStringOutput()},
		}},
		PulumiOptions: opts,
	})
	return pErr
}

// routedNetworks returns the networks routed to the sites: the networks leaked into the EVPN VRF, and the public IPv6 networks.
// bgpConfig: BGP configuration.
func routedNetworks(bgpConfig *bgp.Config) []string {
	return slices.Concat(
		bgpConfig.EVPN.LeakNetworks.IPv4,
		bgpConfig.EVPN.LeakNetworks.IPv6,
		bgpConfig.PublicNetworks.IPv6,
	)
}

// nonAlphanumeric matches the separators of an address (e.g., ".", ":", "/").
var nonAlphanumeric = regexp.MustCompile(`[^0-9a-zA-Z]+`)

// resourceName returns the name of the network resource of a network (e.g., sites-fc00-7 for fc00::/7).
// address: The network (CIDR).
func resourceName(address string) string {
	return sitesName + "-" + nonAlphanumeric.ReplaceAllString(address, "-")
}
