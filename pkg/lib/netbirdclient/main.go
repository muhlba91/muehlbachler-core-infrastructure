package netbirdclient

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/netbird"
)

// Install the NetBird client on the remote server via SSH and create necessary resources.
// The NetBird (server) instance must be installed, and reachable.
// ctx: Pulumi context.
// netbirdInstance: The NetBird (server) instance.
// Returns the setup key to register the client.
func Install(
	ctx *pulumi.Context,
	netbirdInstance *netbird.Instance,
) (pulumi.StringOutput, error) {
	return createSetupKey(ctx, netbirdInstance)
}
