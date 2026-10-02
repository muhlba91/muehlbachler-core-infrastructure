package netbird

import (
	nbProvider "github.com/KitStream/netbird-pulumi-provider/sdk/go/netbird"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	netbirdConf "github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/netbird"
)

// configureAccount configures the settings of the NetBird account via its API.
// Settings which are not set here keep their current value.
// ctx: Pulumi context.
// netbirdConfig: NetBird configuration.
// provider: The NetBird provider.
// dependsOn: Pulumi resource option to specify dependencies.
func configureAccount(
	ctx *pulumi.Context,
	netbirdConfig *netbirdConf.Config,
	provider *nbProvider.Provider,
	dependsOn pulumi.ResourceOrInvokeOption,
) (*nbProvider.AccountSettings, error) {
	return nbProvider.NewAccountSettings(ctx, "netbird-account-settings", &nbProvider.AccountSettingsArgs{
		// must not overlap Tailscale: its firewall drops (and routes) traffic of its range
		NetworkRange: pulumi.String(*netbirdConfig.NetworkRange),
		// the tunnels must stay up for the routing protocols running on top
		LazyConnectionEnabled: pulumi.Bool(false),
		// the client image is updated via its version, not by the server
		AutoUpdateVersion: pulumi.String("disabled"),
		// no services are exposed via the reverse proxy
		PeerExposeEnabled: pulumi.Bool(false),
		// routing is done by BGP: routing peers do not resolve DNS names of network resources
		RoutingPeerDnsResolutionEnabled: pulumi.Bool(false),
		// only the owner administers the instance
		RegularUsersViewBlocked: pulumi.Bool(true),
	}, pulumi.Provider(provider), dependsOn)
}
