package gre

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/bgp"
)

// Install creates resources for GRE based tunnels to facilitate BGP peerings.
// ctx: The Pulumi context for resource creation.
// bgpConfig: The BGP configuration.
// dependsOn: List of Pulumi resources that this installation depends on.
func Install(ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	bgpConfig *bgp.Config,
	dependsOn []pulumi.Resource,
) (pulumi.Resource, error) {
	return installer(
		ctx,
		sshIPv4,
		privateKeyPem,
		bgpConfig,
		pulumi.DependsOn(dependsOn),
	)
}
