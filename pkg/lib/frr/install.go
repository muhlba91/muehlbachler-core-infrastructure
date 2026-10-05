package frr

import (
	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/template"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/bgp"
	netbirdConf "github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/netbird"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/frr"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/util/install"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/util/remotefile"
)

// Install FRR on the remote server via SSH.
// ctx: Pulumi context.
// sshIPv4: The IPv4 address of the server to connect to via SSH.
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// frrData: The FRR configuration data.
// bgpConfig: The BGP configuration details.
// netbirdConfig: NetBird configuration (its network range is the EVPN underlay).
// dependsOn: Pulumi resource option to specify dependencies.
func installer(
	ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	frrData *frr.Data,
	bgpConfig *bgp.Config,
	netbirdConfig *netbirdConf.Config,
	dependsOn pulumi.ResourceOrInvokeOption,
) (*remote.Command, error) {
	conn := &remote.ConnectionArgs{
		Host:       sshIPv4,
		PrivateKey: privateKeyPem,
		User:       pulumi.String("root"),
	}

	opts := []pulumi.ResourceOption{dependsOn}

	opts, prepErr := install.Prepare(ctx, "frr", conn, opts...)
	if prepErr != nil {
		return nil, prepErr
	}

	dockerComposeHash, dcErr := file.Hash("./assets/frr/docker-compose.yml")
	if dcErr != nil {
		return nil, dcErr
	}
	dockerComposeCopy, ccErr := remote.NewCopyToRemote(
		ctx,
		"remote-copy-frr-docker-compose",
		&remote.CopyToRemoteArgs{
			Source:     pulumi.NewFileAsset("./assets/frr/docker-compose.yml"),
			RemotePath: pulumi.String("/opt/frr/docker-compose.yml"),
			Triggers:   pulumi.Array{pulumi.String(*dockerComposeHash)},
			Connection: conn,
		},
		opts...)
	if ccErr != nil {
		return nil, ccErr
	}
	opts = append(opts, pulumi.DependsOn([]pulumi.Resource{dockerComposeCopy}))

	configResources, configHashes, cErr := createConfigs(
		ctx,
		frrData,
		bgpConfig,
		netbirdConfig,
		sshIPv4,
		conn,
		opts...,
	)
	if cErr != nil {
		return nil, cErr
	}

	opts, systemdServiceHash, shErr := install.SystemDService(ctx, "frr", conn, opts...)
	if shErr != nil {
		return nil, shErr
	}

	installFn, iErr := file.ReadContents("./assets/frr/install.sh")
	if iErr != nil {
		return nil, iErr
	}
	return remote.NewCommand(ctx, "remote-command-install-frr", &remote.CommandArgs{
		Create:     pulumi.StringPtr(installFn),
		Update:     pulumi.StringPtr(installFn),
		Triggers:   append(configHashes, pulumi.String(*dockerComposeHash), pulumi.String(*systemdServiceHash)),
		Connection: conn,
	}, append(opts, remotefile.DependsOnAll(configResources...))...)
}

// createConfigs generates the FRR configuration files and uploads them to the remote server.
// ctx: Pulumi context.
// frrData: The FRR configuration data.
// bgpConfig: The BGP configuration details.
// netbirdConfig: NetBird configuration (its network range is the EVPN underlay).
// publicIP: The public IP address to be used in the configuration.
// conn: The remote connection arguments.
// opts: Additional Pulumi resource options.
func createConfigs(
	ctx *pulumi.Context,
	frrData *frr.Data,
	bgpConfig *bgp.Config,
	netbirdConfig *netbirdConf.Config,
	publicIP pulumi.StringOutput,
	conn *remote.ConnectionArgs,
	opts ...pulumi.ResourceOption,
) ([]pulumi.ResourceOutput, pulumi.Array, error) {
	// a rendering error fails the deployment: an empty configuration would be copied, and applied otherwise
	frrConfig, _ := pulumi.All(frrData.Hostname, frrData.NeighborPassword, publicIP).ApplyT(func(args []any) (string, error) {
		hostname, _ := args[0].(string)
		neighborPassword, _ := args[1].(string)
		ip, _ := args[2].(string)

		for name := range bgpConfig.Neighbors {
			neighbor := bgpConfig.Neighbors[name]
			if neighbor.Password == nil && !neighbor.IsPublic {
				neighbor.Password = &neighborPassword
			}
		}
		return template.Render("./assets/frr/config/frr.conf.j2", map[string]any{
			"hostname": hostname,
			"publicIp": ip,
			"bgp":      bgpConfig,
			"underlay": *netbirdConfig.NetworkRange,
		})
	}).(pulumi.StringOutput)
	frrConfigHash, frrConfigCopy := remotefile.CopyRendered(
		ctx,
		"remote-copy-frr-config",
		"./outputs/frr_frr.conf",
		pulumi.String("/opt/frr/config/frr.conf"),
		frrConfig,
		conn,
		opts...,
	)

	vtyshConfigHash, vtErr := file.Hash("./assets/frr/config/vtysh.conf")
	if vtErr != nil {
		return nil, nil, vtErr
	}
	vtyshConfigCopy, vcErr := remote.NewCopyToRemote(ctx, "remote-copy-frr-vtysh", &remote.CopyToRemoteArgs{
		Source:     pulumi.NewFileAsset("./assets/frr/config/vtysh.conf"),
		RemotePath: pulumi.String("/opt/frr/config/vtysh.conf"),
		Triggers:   pulumi.Array{pulumi.String(*vtyshConfigHash)},
		Connection: conn,
	}, opts...)
	if vcErr != nil {
		return nil, nil, vcErr
	}

	daemonsHash, dhErr := file.Hash("./assets/frr/config/daemons")
	if dhErr != nil {
		return nil, nil, dhErr
	}
	daemonsCopy, dcErr := remote.NewCopyToRemote(ctx, "remote-copy-frr-daemons", &remote.CopyToRemoteArgs{
		Source:     pulumi.NewFileAsset("./assets/frr/config/daemons"),
		RemotePath: pulumi.String("/opt/frr/config/daemons"),
		Triggers:   pulumi.Array{pulumi.String(*daemonsHash)},
		Connection: conn,
	}, opts...)
	if dcErr != nil {
		return nil, nil, dcErr
	}

	return []pulumi.ResourceOutput{
		frrConfigCopy,
		pulumi.NewResourceOutput(vtyshConfigCopy),
		pulumi.NewResourceOutput(daemonsCopy),
	}, pulumi.Array{
		frrConfigHash,
		pulumi.String(*vtyshConfigHash),
		pulumi.String(*daemonsHash),
	}, nil
}
