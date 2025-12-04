package vault

import (
	"github.com/muhlba91/pulumi-shared-library/pkg/model/google/iam/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/dns"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/google"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/vault"
)

// Install Vault on the remote server via SSH and create necessary resources.
// ctx: Pulumi context.
// sshIPv4: The IPv4 address of the server to connect to via SSH.
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// dnsConfig: DNS configuration.
// oidcConfig: OIDC configuration.
// dependsOn: List of Pulumi resources that this installation depends on.
func Install(ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	serviceAccount *serviceaccount.User,
	dnsConfig *dns.Config,
	googleConfig *google.Config,
	dependsOn []pulumi.Resource,
) (*vault.Data, *pulumi.AnyOutput, pulumi.Resource, error) {
	vaultData, vdErr := createResources(ctx, serviceAccount, googleConfig)
	if vdErr != nil {
		return nil, nil, nil, vdErr
	}

	vaultInstall, vErr := installer(
		ctx,
		sshIPv4,
		privateKeyPem,
		vaultData,
		googleConfig,
		dnsConfig,
		pulumi.DependsOn(dependsOn),
	)
	if vErr != nil {
		return nil, nil, nil, vErr
	}

	vaultInstanceData, viErr := configure(
		ctx,
		sshIPv4,
		privateKeyPem,
		vaultData.Bucket.ID(),
		dnsConfig,
		pulumi.DependsOn(append([]pulumi.Resource{vaultInstall}, dependsOn...)),
	)
	if viErr != nil {
		return nil, nil, nil, viErr
	}

	return vaultData, vaultInstanceData, vaultInstall, nil
}
