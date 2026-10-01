package wireguard

import (
	"github.com/muhlba91/pulumi-shared-library/pkg/util/template"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/config"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/dns"
	wireguardData "github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/wireguard"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/util/install"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/util/remotefile"
)

// Install WireGuard on the remote server via SSH.
// ctx: Pulumi context.
// sshIPv4: The IPv4 address of the server to connect to via SSH.
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// wireguardData: WireGuard configuration data.
// dnsConfig: DNS configuration.
// dependsOn: Pulumi resource option to specify dependencies.
func installer(
	ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	wireguardData *wireguardData.Data,
	dnsConfig *dns.Config,
	dependsOn pulumi.ResourceOrInvokeOption,
) (*remote.Command, error) {
	conn := &remote.ConnectionArgs{
		Host:       sshIPv4,
		PrivateKey: privateKeyPem,
		User:       pulumi.String("root"),
	}

	opts := []pulumi.ResourceOption{dependsOn}

	opts, prepErr := install.Prepare(ctx, "wireguard", conn, opts...)
	if prepErr != nil {
		return nil, prepErr
	}

	dockerCompose, dcErr := template.Render("./assets/wireguard/docker-compose.yml.j2", map[string]any{
		"domain": dnsConfig.Entries["wireguard"].Domain,
	})
	if dcErr != nil {
		return nil, dcErr
	}
	dockerComposeHash, dockerComposeCopy := remotefile.CopyRendered(
		ctx,
		"remote-copy-wireguard-docker-compose",
		"./outputs/wireguard_docker-compose.yml",
		pulumi.String("/opt/wireguard/docker-compose.yml"),
		pulumi.String(dockerCompose),
		conn,
		opts...,
	)

	configResources, configHashes := createConfigs(ctx, wireguardData, dnsConfig, conn, opts...)

	cronResources, cronErr := install.Cron(ctx, "wireguard", conn, opts...)
	if cronErr != nil {
		return nil, cronErr
	}

	opts, systemdServiceHash, shErr := install.SystemDService(ctx, "wireguard", conn, opts...)
	if shErr != nil {
		return nil, shErr
	}

	installFn, iErr := template.Render("./assets/wireguard/install.sh.j2", map[string]any{
		"bucket": map[string]string{
			"id":   config.BackupBucketID,
			"path": config.BackupBucketPath,
		},
	})
	if iErr != nil {
		return nil, iErr
	}

	return remote.NewCommand(ctx, "remote-command-install-wireguard", &remote.CommandArgs{
		Create:     pulumi.StringPtr(installFn),
		Update:     pulumi.StringPtr(installFn),
		Triggers:   append(configHashes, dockerComposeHash, pulumi.String(*systemdServiceHash)),
		Connection: conn,
	}, append(opts, remotefile.DependsOnAll(append(append(cronResources, configResources...), dockerComposeCopy)...))...)
}

// createConfigs generates the WireGuard configuration files and uploads them to the remote server.
// ctx: Pulumi context.
// wireguardData: The WireGuard configuration data.
// dnsConfig: The DNS configuration.
// conn: The remote connection arguments.
// opts: Additional Pulumi resource options.
func createConfigs(
	ctx *pulumi.Context,
	wireguardData *wireguardData.Data,
	dnsConfig *dns.Config,
	conn *remote.ConnectionArgs,
	opts ...pulumi.ResourceOption,
) ([]pulumi.ResourceOutput, pulumi.Array) {
	wireguardConfig, _ := pulumi.All(wireguardData.AdminPassword, wireguardData.Database.EncryptionPassphrase, wireguardData.Web.SessionSecret, wireguardData.Web.CSRFSecret).ApplyT(func(args []any) string {
		adminPassword, _ := args[0].(string)
		encryptionPassphrase, _ := args[1].(string)
		sessionSecret, _ := args[2].(string)
		csrfSecret, _ := args[3].(string)

		tpl, _ := template.Render("./assets/wireguard/config.yml.j2", map[string]any{
			"domain":        dnsConfig.Entries["wireguard"].Domain,
			"adminPassword": adminPassword,
			"database": map[string]string{
				"encryptionPassphrase": encryptionPassphrase,
			},
			"web": map[string]string{
				"sessionSecret": sessionSecret,
				"csrfSecret":    csrfSecret,
			},
			"oidc": map[string]string{
				"baseUrl":      wireguardData.OIDC.BaseURL,
				"clientId":     wireguardData.OIDC.ClientID,
				"clientSecret": wireguardData.OIDC.ClientSecret,
			},
		})
		return tpl
	}).(pulumi.StringOutput)
	wireguardConfigHash, wireguardConfigCopy := remotefile.CopyRendered(
		ctx,
		"remote-copy-wireguard-config",
		"./outputs/wireguard_config.yml",
		pulumi.String("/opt/wireguard/config/config.yml"),
		wireguardConfig,
		conn,
		opts...,
	)

	return []pulumi.ResourceOutput{wireguardConfigCopy}, pulumi.Array{
		wireguardConfigHash,
	}
}
