package netbirdclient

import (
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/dns"
	netbirdConf "github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/netbird"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/netbird"
)

// Install the NetBird client on the remote server via SSH and create necessary resources.
// The NetBird (server) instance must be installed, and reachable.
// ctx: Pulumi context.
// sshIPv4: The IPv4 address of the server to connect to via SSH.
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// hostname: The hostname of the server, which is the name of the client (peer).
// dnsConfig: DNS configuration.
// netbirdConfig: NetBird configuration.
// netbirdInstance: The NetBird (server) instance.
// dependsOn: List of Pulumi resources that this installation depends on.
func Install(
	ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	hostname pulumi.StringOutput,
	dnsConfig *dns.Config,
	netbirdConfig *netbirdConf.Config,
	netbirdInstance *netbird.Instance,
	dependsOn []pulumi.Resource,
) (*remote.Command, error) {
	setupKey, skErr := createSetupKey(ctx, netbirdInstance)
	if skErr != nil {
		return nil, skErr
	}

	return installer(
		ctx,
		sshIPv4,
		privateKeyPem,
		hostname,
		setupKey,
		dnsConfig,
		netbirdConfig,
		pulumi.DependsOn(append([]pulumi.Resource{netbirdInstance.Ready, netbirdInstance.Settings}, dependsOn...)),
	)
}
