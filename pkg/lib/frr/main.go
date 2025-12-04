package frr

import (
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/bgp"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/network"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/frr"
)

// Install creates resources for FRR based on the provided configuration.
// ctx: The Pulumi context for resource creation.
// hostname: The hostname of the server where FRR will be installed.
// networkConfig: The network configuration.
// bgpConfig: The BGP configuration.
// dependsOn: List of Pulumi resources that this installation depends on.
func Install(ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	hostname pulumi.StringOutput,
	networkConfig *network.Config,
	bgpConfig *bgp.Config,
	dependsOn []pulumi.Resource,
) (*frr.Data, *remote.Command, error) {
	frrData, frrErr := createResources(ctx, hostname, networkConfig)
	if frrErr != nil {
		return nil, nil, frrErr
	}
	frrInstall, frrErr := installer(
		ctx,
		sshIPv4,
		privateKeyPem,
		frrData,
		bgpConfig,
		pulumi.DependsOn(dependsOn),
	)
	if frrErr != nil {
		return nil, nil, frrErr
	}

	return frrData, frrInstall, nil
}
