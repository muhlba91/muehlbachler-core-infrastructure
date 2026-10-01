package vault

import (
	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/template"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/config"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/dns"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/google"
	vaultData "github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/vault"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/util/install"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/util/remotefile"
)

// Install Vault on the remote server via SSH.
// ctx: Pulumi context.
// sshIPv4: The IPv4 address of the server to connect to via SSH.
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// vaultData: Vault configuration data.
// dnsConfig: DNS configuration.
// googleConfig: Google Cloud configuration.
// dependsOn: Pulumi resource option to specify dependencies.
func installer(
	ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	vaultData *vaultData.Data,
	googleConfig *google.Config,
	dnsConfig *dns.Config,
	dependsOn pulumi.ResourceOrInvokeOption,
) (*remote.Command, error) {
	conn := &remote.ConnectionArgs{
		Host:       sshIPv4,
		PrivateKey: privateKeyPem,
		User:       pulumi.String("root"),
	}

	opts := []pulumi.ResourceOption{dependsOn}

	opts, prepErr := install.Prepare(ctx, "vault", conn, opts...)
	if prepErr != nil {
		return nil, prepErr
	}

	dockerCompose, dcErr := template.Render("./assets/vault/docker-compose.yml.j2", map[string]any{
		"domain": dnsConfig.Entries["vault"].Domain,
	})
	if dcErr != nil {
		return nil, dcErr
	}
	dockerComposeHash, dockerComposeCopy := remotefile.CopyRendered(
		ctx,
		"remote-copy-vault-docker-compose",
		"./outputs/vault_docker-compose.yml",
		pulumi.String("/opt/vault/docker-compose.yml"),
		pulumi.String(dockerCompose),
		conn,
		opts...,
	)

	vaultConfig, _ := pulumi.All(vaultData.ScalewayBucket.Name, vaultData.Application.Key.AccessKey, vaultData.Application.Key.SecretKey).ApplyT(func(args []any) string {
		scalewayBucket, _ := args[0].(string)
		accessKey, _ := args[1].(string)
		secretKey, _ := args[2].(string)

		tpl, _ := template.Render("./assets/vault/config.hcl.j2", map[string]any{
			"gcp": googleConfig,
			"scaleway": map[string]string{
				"bucket":    scalewayBucket,
				"region":    config.ScalewayDefaultRegion,
				"accessKey": accessKey,
				"secretKey": secretKey,
			},
		})
		return tpl
	}).(pulumi.StringOutput)
	vaultConfigHash, vaultConfigCopy := remotefile.CopyRendered(
		ctx,
		"remote-copy-vault-config",
		"./outputs/vault_vault-config.hcl",
		pulumi.String("/opt/vault/config/vault-config.hcl"),
		vaultConfig,
		conn,
		opts...,
	)

	opts, systemdServiceHash, shErr := install.SystemDService(ctx, "vault", conn, opts...)
	if shErr != nil {
		return nil, shErr
	}

	installFn, iErr := file.ReadContents("./assets/vault/install.sh")
	if iErr != nil {
		return nil, iErr
	}
	return remote.NewCommand(ctx, "remote-command-install-vault", &remote.CommandArgs{
		Create:     pulumi.StringPtr(installFn),
		Update:     pulumi.StringPtr(installFn),
		Triggers:   pulumi.Array{dockerComposeHash, pulumi.String(*systemdServiceHash), vaultConfigHash},
		Connection: conn,
	}, append(opts, remotefile.DependsOnAll(dockerComposeCopy, vaultConfigCopy))...)
}
