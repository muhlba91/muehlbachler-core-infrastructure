package wireguard

import (
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/dns"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/oidc"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/wireguard"
)

// Install WireGuard on the remote server via SSH and create necessary resources.
// ctx: Pulumi context.
// sshIPv4: The IPv4 address of the server to connect to via SSH.
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// dnsConfig: DNS configuration.
// oidcConfig: OIDC configuration.
// dependsOn: List of Pulumi resources that this installation depends on.
func Install(ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	dnsConfig *dns.Config,
	oidcConfig *oidc.Config,
	dependsOn []pulumi.Resource,
) (*wireguard.Data, *remote.Command, error) {
	wireguardData, wdErr := createResources(ctx, oidcConfig)
	if wdErr != nil {
		return nil, nil, wdErr
	}
	wireguardInstall, wiErr := installer(
		ctx,
		sshIPv4,
		privateKeyPem,
		wireguardData,
		dnsConfig,
		pulumi.DependsOn(dependsOn),
	)
	if wiErr != nil {
		return nil, nil, wiErr
	}

	return wireguardData, wireguardInstall, nil
}
