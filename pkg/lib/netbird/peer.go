package netbird

import (
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/netbird/peer"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/netbird"
)

const (
	// PeerID is the key of the peer ID in the result of LookupPeer.
	PeerID = "id"
	// PeerIPv4 is the key of the NetBird IPv4 address in the result of LookupPeer.
	PeerIPv4 = "ipv4"
)

// LookupPeer looks up the NetBird peer of the server: its ID (PeerID), and NetBird IPv4 address (PeerIPv4).
// The lookup runs once the NetBird client is installed (registered), and fails if the peer name is not unique.
// ctx: Pulumi context.
// hostname: The hostname of the server, which is the name of the NetBird client (peer).
// netbirdInstance: The NetBird (server) instance.
// netbirdClient: The installation of the NetBird client.
func LookupPeer(
	ctx *pulumi.Context,
	hostname pulumi.StringOutput,
	netbirdInstance *netbird.Instance,
	netbirdClient *remote.Command,
) pulumi.StringMapOutput {
	// the ID of the NetBird client installation is only used to run the lookup after it (registered):
	// direct invokes ignore the DependsOn option
	nbPeer, _ := pulumi.All(hostname, netbirdClient.ID()).ApplyT(func(args []any) (map[string]string, error) {
		name, _ := args[0].(string)

		res, err := peer.Get(ctx, &peer.GetOptions{
			Name:          &name,
			PulumiOptions: []pulumi.InvokeOption{pulumi.Provider(netbirdInstance.Provider)},
		})
		if err != nil {
			return nil, err
		}
		return map[string]string{PeerID: res.Id, PeerIPv4: res.Ip}, nil
	}).(pulumi.StringMapOutput)

	return nbPeer
}
