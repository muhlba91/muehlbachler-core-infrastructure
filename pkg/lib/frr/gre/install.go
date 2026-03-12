package gre

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/muhlba91/pulumi-shared-library/pkg/util/defaults"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/template"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/bgp"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/util/install"
)

// Install GRE tunnels on the remote server via SSH.
// ctx: Pulumi context.
// sshIPv4: The IPv4 address of the server to connect to via SSH.
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// bgpConfig: The BGP configuration details.
// dependsOn: Pulumi resource option to specify dependencies.
func installer(
	ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	bgpConfig *bgp.Config,
	dependsOn pulumi.ResourceOrInvokeOption,
) (pulumi.Resource, error) {
	conn := &remote.ConnectionArgs{
		Host:       sshIPv4,
		PrivateKey: privateKeyPem,
		User:       pulumi.String("root"),
	}

	opts := []pulumi.ResourceOption{dependsOn}

	opts, prepErr := install.Prepare(ctx, "frr/gre", conn, opts...)
	if prepErr != nil {
		return nil, prepErr
	}

	tunnels := []string{}
	configResources := []pulumi.Output{}
	configHashes := pulumi.Array{}
	keys := slices.Collect(maps.Keys(bgpConfig.Neighbors))
	slices.Sort(keys)
	for _, key := range keys {
		neighbor := bgpConfig.Neighbors[key]
		if neighbor.GRE == nil {
			continue
		}

		tunnels = append(tunnels, *neighbor.InterfaceName)

		hash := createConfig(sshIPv4, neighbor)
		configHashes = append(configHashes, hash)

		resource := writeToRemote(ctx, hash, neighbor.InterfaceName, conn, opts...)
		configResources = append(configResources, resource)
	}

	installFn, iErr := template.Render("./assets/frr/gre/run.sh.j2", map[string]any{
		"tunnels": strings.Join(tunnels, " "),
	})
	if iErr != nil {
		return nil, iErr
	}
	return remote.NewCommand(ctx, "remote-command-install-gre", &remote.CommandArgs{
		Create:     pulumi.StringPtr(installFn),
		Update:     pulumi.StringPtr(installFn),
		Delete:     pulumi.StringPtr(installFn),
		Triggers:   configHashes,
		Connection: conn,
	}, append(opts, install.CollectResourceOptions(configResources)...)...)
}

// createConfig generates the GRE netplan configuration file.
// localIP: The local IP address to be used in the GRE configuration.
// neighbor: The BGP neighbor for which the GRE configuration is being created.
func createConfig(
	localIP pulumi.StringOutput,
	neighbor *bgp.NeighborConfig,
) pulumi.StringOutput {
	netplanConfig, _ := localIP.ApplyT(func(localIP string) string {
		netplanData := struct {
			Name     string
			LocalIP  string
			RemoteIP string
			TunnelIP string
			Type     string
		}{
			Name:     *neighbor.InterfaceName,
			LocalIP:  localIP,
			RemoteIP: *neighbor.GRE.RemoteIP,
			TunnelIP: *neighbor.GRE.TunnelIP,
			Type:     defaults.GetOrDefault(neighbor.GRE.Type, "gre"),
		}

		config, _ := template.Render("./assets/frr/gre/config/netplan.yml.j2", netplanData)
		return config
	}).(pulumi.StringOutput)

	netplanConfigHash, _ := file.WritePulumi(fmt.Sprintf("./outputs/gre_netplan_%s.yaml", *neighbor.InterfaceName), netplanConfig).
		ApplyT(func(_ string) string {
			hash, _ := file.Hash(fmt.Sprintf("./outputs/gre_netplan_%s.yaml", *neighbor.InterfaceName))
			return *hash
		}).(pulumi.StringOutput)

	return netplanConfigHash
}

// writeToRemote uploads the GRE netplan config file to the remote server.
// ctx: Pulumi context.
// bgpConfig: The BGP configuration details.
// conn: The remote connection arguments.
// opts: Additional Pulumi resource options.
func writeToRemote(
	ctx *pulumi.Context,
	hash pulumi.StringOutput,
	interfaceName *string,
	conn *remote.ConnectionArgs,
	opts ...pulumi.ResourceOption,
) pulumi.Output {
	netplanConfigCopy := hash.ApplyT(func(_ string) pulumi.ResourceOption {
		cmd, _ := remote.NewCopyToRemote(
			ctx,
			fmt.Sprintf("remote-copy-gre-netplan-%s", *interfaceName),
			&remote.CopyToRemoteArgs{
				Source:     pulumi.NewFileAsset(fmt.Sprintf("./outputs/gre_netplan_%s.yaml", *interfaceName)),
				RemotePath: pulumi.String(fmt.Sprintf("/etc/netplan/%s.yaml", *interfaceName)),
				Triggers:   pulumi.Array{hash},
				Connection: conn,
			},
			opts...)
		return pulumi.DependsOn([]pulumi.Resource{cmd})
	})

	return netplanConfigCopy
}
