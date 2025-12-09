package frr

import (
	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/template"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/bgp"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/frr"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/util/install"
)

// Install FRR on the remote server via SSH.
// ctx: Pulumi context.
// sshIPv4: The IPv4 address of the server to connect to via SSH.
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// frrData: The FRR configuration data.
// bgpConfig: The BGP configuration details.
// dependsOn: Pulumi resource option to specify dependencies.
func installer(
	ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	frrData *frr.Data,
	bgpConfig *bgp.Config,
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

	configResources, configHashes, cErr := createConfigs(ctx, frrData, bgpConfig, conn, opts...)
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
	}, append(opts, install.CollectResourceOptions(configResources)...)...)
}

// createConfigs generates the FRR configuration files and uploads them to the remote server.
// ctx: Pulumi context.
// frrData: The FRR configuration data.
// bgpConfig: The BGP configuration details.
// conn: The remote connection arguments.
// opts: Additional Pulumi resource options.
func createConfigs(
	ctx *pulumi.Context,
	frrData *frr.Data,
	bgpConfig *bgp.Config,
	conn *remote.ConnectionArgs,
	opts ...pulumi.ResourceOption,
) ([]pulumi.Output, pulumi.Array, error) {
	frrConfig, _ := pulumi.All(frrData.Hostname, frrData.NeighborPassword).ApplyT(func(args []any) string {
		hostname, _ := args[0].(string)
		neighborPassword, _ := args[1].(string)

		neighbors := []map[string]any{}
		for _, neighbor := range bgpConfig.Neighbors {
			n := map[string]any{
				"address":   neighbor.Address,
				"asn":       neighbor.RemoteASN,
				"interface": neighbor.InterfaceName,
				"password":  neighborPassword,
			}
			if n["interface"] == nil || n["interface"] == "" {
				n["interface"] = bgpConfig.InterfaceName
			}
			neighbors = append(neighbors, n)
		}
		tpl, _ := template.Render("./assets/frr/config/frr.conf.j2", map[string]any{
			"hostname": hostname,
			"bgp": map[string]any{
				"localAsn":               bgpConfig.LocalASN,
				"routerId":               bgpConfig.RouterID,
				"interface":              bgpConfig.InterfaceName,
				"advertisedIPv4Networks": bgpConfig.AdvertisedIPv4Networks,
				"advertisedIPv6Networks": bgpConfig.AdvertisedIPv6Networks,
				"neighbors":              neighbors,
			},
		})
		return tpl
	}).(pulumi.StringOutput)
	frrConfigHash, _ := file.WritePulumi("./outputs/frr_frr.conf", frrConfig).
		ApplyT(func(_ string) string {
			hash, _ := file.Hash("./outputs/frr_frr.conf")
			return *hash
		}).(pulumi.StringOutput)
	frrConfigCopy := frrConfigHash.ApplyT(func(_ string) pulumi.ResourceOption {
		cmd, _ := remote.NewCopyToRemote(ctx, "remote-copy-frr-config", &remote.CopyToRemoteArgs{
			Source:     pulumi.NewFileAsset("./outputs/frr_frr.conf"),
			RemotePath: pulumi.String("/opt/frr/config/frr.conf"),
			Triggers:   pulumi.Array{frrConfigHash},
			Connection: conn,
		}, opts...)
		return pulumi.DependsOn([]pulumi.Resource{cmd})
	})

	vtyshConfigHash, vtErr := file.Hash("./assets/frr/config/vtysh.conf")
	if vtErr != nil {
		return nil, nil, vtErr
	}
	vtyshConfigCopy := pulumi.String(*vtyshConfigHash).ToStringOutput().ApplyT(func(_ string) pulumi.ResourceOption {
		cmd, _ := remote.NewCopyToRemote(ctx, "remote-copy-frr-vtysh", &remote.CopyToRemoteArgs{
			Source:     pulumi.NewFileAsset("./assets/frr/config/vtysh.conf"),
			RemotePath: pulumi.String("/opt/frr/config/vtysh.conf"),
			Triggers:   pulumi.Array{pulumi.String(*vtyshConfigHash)},
			Connection: conn,
		}, opts...)
		return pulumi.DependsOn([]pulumi.Resource{cmd})
	})

	daemonsHash, dhErr := file.Hash("./assets/frr/config/daemons")
	if dhErr != nil {
		return nil, nil, dhErr
	}
	daemonsCopy := pulumi.String(*daemonsHash).ToStringOutput().ApplyT(func(_ string) pulumi.ResourceOption {
		cmd, _ := remote.NewCopyToRemote(ctx, "remote-copy-frr-daemons", &remote.CopyToRemoteArgs{
			Source:     pulumi.NewFileAsset("./assets/frr/config/daemons"),
			RemotePath: pulumi.String("/opt/frr/config/daemons"),
			Triggers:   pulumi.Array{pulumi.String(*daemonsHash)},
			Connection: conn,
		}, opts...)
		return pulumi.DependsOn([]pulumi.Resource{cmd})
	})

	return []pulumi.Output{
			frrConfigCopy,
			vtyshConfigCopy,
			daemonsCopy,
		}, pulumi.Array{
			frrConfigHash,
			pulumi.String(*vtyshConfigHash),
			pulumi.String(*daemonsHash),
		}, nil
}
