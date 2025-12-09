package wireguard

import (
	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/google/project"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/template"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/config"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/dns"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/google"
	wireguardData "github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/wireguard"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/util/install"
)

// Install WireGuard on the remote server via SSH.
// ctx: Pulumi context.
// sshIPv4: The IPv4 address of the server to connect to via SSH.
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// wireguardData: WireGuard configuration data.
// dnsConfig: DNS configuration.
// gcpConfig: GCP configuration.
// dependsOn: Pulumi resource option to specify dependencies.
func installer(
	ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	wireguardData *wireguardData.Data,
	dnsConfig *dns.Config,
	gcpConfig *google.Config,
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
	dockerComposeHash := file.WritePulumi("./outputs/wireguard_docker-compose.yml", pulumi.String(dockerCompose)).
		ApplyT(func(_ string) string {
			hash, _ := file.Hash("./outputs/wireguard_docker-compose.yml")
			return *hash
		})
	dockerComposeCopy := dockerComposeHash.ApplyT(func(_ string) pulumi.ResourceOption {
		cmd, _ := remote.NewCopyToRemote(
			ctx,
			"remote-copy-wireguard-docker-compose",
			&remote.CopyToRemoteArgs{
				Source:     pulumi.NewFileAsset("./outputs/wireguard_docker-compose.yml"),
				RemotePath: pulumi.String("/opt/wireguard/docker-compose.yml"),
				Triggers:   pulumi.Array{dockerComposeHash},
				Connection: conn,
			},
			opts...)
		return pulumi.DependsOn([]pulumi.Resource{cmd})
	})

	configResources, configHashes := createConfigs(ctx, wireguardData, dnsConfig, conn, opts...)

	cronResources, cronErr := install.Cron(ctx, "wireguard", conn, gcpConfig, opts...)
	if cronErr != nil {
		return nil, cronErr
	}

	opts, systemdServiceHash, shErr := install.SystemDService(ctx, "wireguard", conn, opts...)
	if shErr != nil {
		return nil, shErr
	}

	installFn, iErr := template.Render("./assets/wireguard/install.sh.j2", map[string]any{
		"project": project.GetOrDefault(ctx, gcpConfig.Project),
		"bucket": map[string]string{
			"id":   config.BackupBucketID,
			"path": config.BucketPath,
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
	}, append(opts, install.CollectResourceOptions(append(append(cronResources, configResources...), dockerComposeCopy))...)...)
}

// createConfigs generates the FRR configuration files and uploads them to the remote server.
// ctx: Pulumi context.
// frrData: The FRR configuration data.
// bgpConfig: The BGP configuration details.
// conn: The remote connection arguments.
// opts: Additional Pulumi resource options.
func createConfigs(
	ctx *pulumi.Context,
	wireguardData *wireguardData.Data,
	dnsConfig *dns.Config,
	conn *remote.ConnectionArgs,
	opts ...pulumi.ResourceOption,
) ([]pulumi.Output, pulumi.Array) {
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
	wireguardConfigHash := file.WritePulumi("./outputs/wireguard_config.yml", wireguardConfig).
		ApplyT(func(_ string) string {
			hash, _ := file.Hash("./outputs/wireguard_config.yml")
			return *hash
		})
	wireguardConfigCopy := wireguardConfigHash.ApplyT(func(_ string) pulumi.ResourceOption {
		cmd, _ := remote.NewCopyToRemote(ctx, "remote-copy-wireguard-config", &remote.CopyToRemoteArgs{
			Source:     pulumi.NewFileAsset("./outputs/wireguard_config.yml"),
			RemotePath: pulumi.String("/opt/wireguard/config/config.yml"),
			Triggers:   pulumi.Array{wireguardConfigHash},
			Connection: conn,
		}, opts...)
		return pulumi.DependsOn([]pulumi.Resource{cmd})
	})

	return []pulumi.Output{
			wireguardConfigCopy,
		}, pulumi.Array{
			wireguardConfigHash,
		}
}
