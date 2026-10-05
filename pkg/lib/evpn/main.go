package evpn

import (
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/bgp"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/netbird"
)

// Install creates the EVPN (L3VNI) devices, and firewall rules on the remote server via SSH.
// The BGP (EVPN) configuration is part of the FRR configuration.
// The VXLAN device uses the NetBird IP of the server as local address (VTEP): the NetBird client must be registered.
// Returns the VTEP (NetBird IPv4 address of the server), which the site routers peer with.
// ctx: The Pulumi context for resource creation.
// sshIPv4: The IPv4 address of the server to connect to via SSH (the BGP router ID).
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// hostname: The hostname of the server, which is the name of the NetBird client (peer).
// evpnConfig: The EVPN configuration.
// netbirdInstance: The NetBird (server) instance.
// netbirdClient: The installation of the NetBird client.
// dependsOn: List of Pulumi resources that this installation depends on.
func Install(ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	hostname pulumi.StringOutput,
	evpnConfig *bgp.EVPNConfig,
	netbirdInstance *netbird.Instance,
	netbirdClient *remote.Command,
	dependsOn []pulumi.Resource,
) (pulumi.StringOutput, *remote.Command, error) {
	vtep := lookupVTEP(ctx, hostname, netbirdInstance, netbirdClient)

	evpnInstall, iErr := installer(
		ctx,
		sshIPv4,
		privateKeyPem,
		vtep,
		evpnConfig,
		pulumi.DependsOn(append([]pulumi.Resource{netbirdClient}, dependsOn...)),
	)
	if iErr != nil {
		return pulumi.StringOutput{}, nil, iErr
	}

	return vtep, evpnInstall, nil
}
