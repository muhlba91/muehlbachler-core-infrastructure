package netbird

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/dns"
	netbirdConf "github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/netbird"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/netbird"
)

// Install NetBird (server) on the remote server via SSH and create necessary resources.
// The personal access token of the initial owner is stored in Vault, which must be installed before.
// The returned instance is usable (e.g., via its provider) once its ready resource is completed.
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
) (*netbird.Instance, error) {
	netbirdData, ndErr := createResources(ctx, netbirdConfig)
	if ndErr != nil {
		return nil, ndErr
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
		return nil, niErr
	}

	token, initCmd, nErr := configure(
		ctx,
		sshIPv4,
		privateKeyPem,
		netbirdData,
		vaultInstanceData,
		pulumi.DependsOn(append([]pulumi.Resource{netbirdInstall}, dependsOn...)),
	)
	if nErr != nil {
		return nil, nErr
	}

	ready, rErr := waitReady(
		ctx,
		sshIPv4,
		privateKeyPem,
		dnsConfig,
		pulumi.DependsOn([]pulumi.Resource{initCmd}),
	)
	if rErr != nil {
		return nil, rErr
	}

	provider, pErr := createProvider(ctx, dnsConfig, token)
	if pErr != nil {
		return nil, pErr
	}

	settings, sErr := configureAccount(ctx, netbirdConfig, provider, pulumi.DependsOn([]pulumi.Resource{ready}))
	if sErr != nil {
		return nil, sErr
	}

	backboneGroupID, bErr := createBackbone(ctx, provider, []pulumi.Resource{ready, settings})
	if bErr != nil {
		return nil, bErr
	}

	return &netbird.Instance{
		Data:            netbirdData,
		Token:           token,
		Provider:        provider,
		Ready:           ready,
		Settings:        settings,
		BackboneGroupID: backboneGroupID,
	}, nil
}
