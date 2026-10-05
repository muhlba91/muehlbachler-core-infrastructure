package evpn

import (
	"fmt"
	"net"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	nbLib "github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/netbird"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/netbird"
)

// lookupVTEP looks up the NetBird IPv4 address of the server, which is the local VXLAN address (VTEP).
// The lookup runs once the NetBird client is installed (registered), and fails if the peer name is not unique.
// ctx: Pulumi context.
// hostname: The hostname of the server, which is the name of the NetBird client (peer).
// netbirdInstance: The NetBird (server) instance.
// netbirdClient: The installation of the NetBird client.
func lookupVTEP(
	ctx *pulumi.Context,
	hostname pulumi.StringOutput,
	netbirdInstance *netbird.Instance,
	netbirdClient *remote.Command,
) pulumi.StringOutput {
	return nbLib.LookupPeer(ctx, hostname, netbirdInstance, netbirdClient).MapIndex(pulumi.String(nbLib.PeerIPv4))
}

// routerMAC derives the router MAC address of the EVPN bridge from the BGP router ID (IPv4 address):
// a locally administered address with the prefix 02:5e followed by the four address bytes.
// routerID: The BGP router ID (IPv4 address).
func routerMAC(routerID string) (string, error) {
	ip := net.ParseIP(routerID).To4()
	if ip == nil {
		return "", fmt.Errorf("invalid router id (IPv4 address): %s", routerID)
	}
	return fmt.Sprintf("02:5e:%02x:%02x:%02x:%02x", ip[0], ip[1], ip[2], ip[3]), nil
}
