package netbird

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/dns"
	netbirdConf "github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/netbird"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/netbird"
)

// Install NetBird (server) on the remote server via SSH and create necessary resources.
// The personal access token of the initial owner is stored in Vault, which must be installed before.
// ctx: Pulumi context.
// sshIPv4: The IPv4 address of the server to connect to via SSH.
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// dnsConfig: DNS configuration.
// netbirdConfig: NetBird configuration.
// vaultInstanceData: The Vault instance data output.
// dependsOn: List of Pulumi resources that this installation depends on.
func Install(ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	dnsConfig *dns.Config,
	netbirdConfig *netbirdConf.Config,
	vaultInstanceData *pulumi.AnyOutput,
	dependsOn []pulumi.Resource,
) (*netbird.Data, pulumi.StringOutput, pulumi.Resource, error) {
	netbirdData, ndErr := createResources(ctx, netbirdConfig)
	if ndErr != nil {
		return nil, pulumi.StringOutput{}, nil, ndErr
	}

	netbirdInstall, niErr := installer(
		ctx,
		sshIPv4,
		privateKeyPem,
		netbirdData,
		dnsConfig,
		pulumi.DependsOn(dependsOn),
	)
	if niErr != nil {
		return nil, pulumi.StringOutput{}, nil, niErr
	}

	token, nErr := configure(
		ctx,
		sshIPv4,
		privateKeyPem,
		netbirdData,
		vaultInstanceData,
		pulumi.DependsOn(append([]pulumi.Resource{netbirdInstall}, dependsOn...)),
	)
	if nErr != nil {
		return nil, pulumi.StringOutput{}, nil, nErr
	}

	return netbirdData, token, netbirdInstall, nil
}
