package netbirdclient

import (
	"github.com/muhlba91/pulumi-shared-library/pkg/util/template"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/config"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/dns"
	netbirdConf "github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/netbird"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/util/install"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/util/remotefile"
)

// Install the NetBird client on the remote server via SSH.
// ctx: Pulumi context.
// sshIPv4: The IPv4 address of the server to connect to via SSH.
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// hostname: The hostname of the server, which is the name of the client (peer).
// setupKey: The setup key to register the client.
// dnsConfig: DNS configuration.
// netbirdConfig: NetBird configuration.
// dependsOn: Pulumi resource option to specify dependencies.
func installer(
	ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	hostname pulumi.StringOutput,
	setupKey pulumi.StringOutput,
	dnsConfig *dns.Config,
	netbirdConfig *netbirdConf.Config,
	dependsOn pulumi.ResourceOrInvokeOption,
) (*remote.Command, error) {
	conn := &remote.ConnectionArgs{
		Host:       sshIPv4,
		PrivateKey: privateKeyPem,
		User:       pulumi.String("root"),
	}

	opts := []pulumi.ResourceOption{dependsOn}

	opts, prepErr := install.Prepare(ctx, "netbird-client", conn, opts...)
	if prepErr != nil {
		return nil, prepErr
	}

	dockerCompose, _ := pulumi.All(hostname, setupKey).ApplyT(func(args []any) (string, error) {
		hostnameValue, _ := args[0].(string)
		setupKeyValue, _ := args[1].(string)

		return template.Render("./assets/netbird-client/docker-compose.yml.j2", map[string]any{
			"domain":        *dnsConfig.Entries["netbird"].Domain,
			"hostname":      hostnameValue,
			"wireguardPort": *netbirdConfig.Client.WireguardPort,
			"mtu":           *netbirdConfig.Client.MTU,
			"setupKey":      setupKeyValue,
		})
	}).(pulumi.StringOutput)
	dockerComposeHash, dockerComposeCopy := remotefile.CopyRendered(
		ctx,
		"remote-copy-netbird-client-docker-compose",
		"./outputs/netbird-client_docker-compose.yml",
		pulumi.String("/opt/netbird-client/docker-compose.yml"),
		dockerCompose,
		conn,
		opts...,
	)

	cronResources, cronErr := install.Cron(ctx, "netbird-client", conn, opts...)
	if cronErr != nil {
		return nil, cronErr
	}

	opts, systemdServiceHash, shErr := install.SystemDService(ctx, "netbird-client", conn, opts...)
	if shErr != nil {
		return nil, shErr
	}

	installFn, iErr := template.Render("./assets/netbird-client/install.sh.j2", map[string]any{
		"bucket": map[string]string{
			"id":   config.BackupBucketID,
			"path": config.BackupBucketPath,
		},
	})
	if iErr != nil {
		return nil, iErr
	}

	return remote.NewCommand(ctx, "remote-command-install-netbird-client", &remote.CommandArgs{
		Create:     pulumi.StringPtr(installFn),
		Update:     pulumi.StringPtr(installFn),
		Triggers:   pulumi.Array{pulumi.Unsecret(dockerComposeHash), pulumi.String(*systemdServiceHash)},
		Connection: conn,
	}, append(opts, remotefile.DependsOnAll(append(cronResources, dockerComposeCopy)...))...)
}
