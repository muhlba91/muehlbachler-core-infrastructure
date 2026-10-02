package netbird

import (
	"github.com/muhlba91/pulumi-shared-library/pkg/util/template"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/config"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/dns"
	netbirdData "github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/netbird"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/util/install"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/util/remotefile"
)

// Install NetBird on the remote server via SSH.
// ctx: Pulumi context.
// sshIPv4: The IPv4 address of the server to connect to via SSH.
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// netbirdData: NetBird configuration data.
// dnsConfig: DNS configuration.
// dependsOn: Pulumi resource option to specify dependencies.
func installer(
	ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	netbirdData *netbirdData.Data,
	dnsConfig *dns.Config,
	dependsOn pulumi.ResourceOrInvokeOption,
) (*remote.Command, error) {
	conn := &remote.ConnectionArgs{
		Host:       sshIPv4,
		PrivateKey: privateKeyPem,
		User:       pulumi.String("root"),
	}

	opts := []pulumi.ResourceOption{dependsOn}

	opts, prepErr := install.Prepare(ctx, "netbird", conn, opts...)
	if prepErr != nil {
		return nil, prepErr
	}

	configResources, configHashes := createConfigs(ctx, netbirdData, dnsConfig, conn, opts...)

	cronResources, cronErr := install.Cron(ctx, "netbird", conn, opts...)
	if cronErr != nil {
		return nil, cronErr
	}

	opts, systemdServiceHash, shErr := install.SystemDService(ctx, "netbird", conn, opts...)
	if shErr != nil {
		return nil, shErr
	}

	installFn, iErr := template.Render("./assets/netbird/install.sh.j2", map[string]any{
		"bucket": map[string]string{
			"id":   config.BackupBucketID,
			"path": config.BackupBucketPath,
		},
	})
	if iErr != nil {
		return nil, iErr
	}

	return remote.NewCommand(ctx, "remote-command-install-netbird", &remote.CommandArgs{
		Create:     pulumi.StringPtr(installFn),
		Update:     pulumi.StringPtr(installFn),
		Triggers:   append(configHashes, pulumi.String(*systemdServiceHash)),
		Connection: conn,
	}, append(opts, remotefile.DependsOnAll(append(cronResources, configResources...)...))...)
}

// createConfigs generates the NetBird docker compose and configuration files (both contain secrets)
// and uploads them to the remote server.
// ctx: Pulumi context.
// netbirdData: The NetBird configuration data.
// dnsConfig: The DNS configuration.
// conn: The remote connection arguments.
// opts: Additional Pulumi resource options.
func createConfigs(
	ctx *pulumi.Context,
	netbirdData *netbirdData.Data,
	dnsConfig *dns.Config,
	conn *remote.ConnectionArgs,
	opts ...pulumi.ResourceOption,
) ([]pulumi.ResourceOutput, pulumi.Array) {
	domain := *dnsConfig.Entries["netbird"].Domain

	dockerCompose, _ := netbirdData.DatabasePassword.ApplyT(func(password string) (string, error) {
		return template.Render("./assets/netbird/docker-compose.yml.j2", map[string]any{
			"domain": domain, //nolint:goconst // template key
			"db": map[string]string{
				"database": netbirdDatabaseName,
				"user":     netbirdDatabaseUser,
				"password": password, //nolint:goconst // template key
			},
		})
	}).(pulumi.StringOutput)
	dockerComposeHash, dockerComposeCopy := remotefile.CopyRendered(
		ctx,
		"remote-copy-netbird-docker-compose",
		"./outputs/netbird_docker-compose.yml",
		pulumi.String("/opt/netbird/docker-compose.yml"),
		dockerCompose,
		conn,
		opts...,
	)

	netbirdConfig, _ := pulumi.All(
		netbirdData.AuthSecret,
		netbirdData.SessionCookieEncryptionKey,
		netbirdData.StoreEncryptionKey,
		netbirdData.DatabasePassword,
	).ApplyT(func(args []any) (string, error) {
		authSecret, _ := args[0].(string)
		sessionCookieEncryptionKey, _ := args[1].(string)
		storeEncryptionKey, _ := args[2].(string)
		databasePassword, _ := args[3].(string)

		return template.Render("./assets/netbird/config.yml.j2", map[string]any{
			"domain":                     domain,
			"authSecret":                 authSecret,
			"sessionCookieEncryptionKey": sessionCookieEncryptionKey,
			"storeEncryptionKey":         storeEncryptionKey,
			"db": map[string]string{
				"database": netbirdDatabaseName,
				"user":     netbirdDatabaseUser,
				"password": databasePassword,
			},
		})
	}).(pulumi.StringOutput)
	netbirdConfigHash, netbirdConfigCopy := remotefile.CopyRendered(
		ctx,
		"remote-copy-netbird-config",
		"./outputs/netbird_config.yml",
		pulumi.String("/opt/netbird/config.yml"),
		netbirdConfig,
		conn,
		opts...,
	)

	// note: the hashes derive from new (secret) values, which are unknown during previews:
	// unknown secrets are not accepted as triggers, and the hash of secret content is not sensitive
	return []pulumi.ResourceOutput{dockerComposeCopy, netbirdConfigCopy}, pulumi.Array{
		pulumi.Unsecret(dockerComposeHash),
		pulumi.Unsecret(netbirdConfigHash),
	}
}
