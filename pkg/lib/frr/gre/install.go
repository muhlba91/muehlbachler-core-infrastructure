package gre

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/muhlba91/pulumi-shared-library/pkg/util/defaults"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/template"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/bgp"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/util/install"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/util/remotefile"
)

// Install GRE tunnels, and their firewall rules on the remote server via SSH.
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
	firewallTunnels := []tunnel{}
	configResources := []pulumi.ResourceOutput{}
	configHashes := pulumi.Array{}
	keys := slices.Collect(maps.Keys(bgpConfig.Neighbors))
	slices.Sort(keys)
	for _, key := range keys {
		neighbor := bgpConfig.Neighbors[key]
		if neighbor.GRE == nil {
			continue
		}

		tunnels = append(tunnels, *neighbor.InterfaceName)
		firewallTunnels = append(firewallTunnels, tunnel{name: *neighbor.InterfaceName, addresses: neighbor.Addresses})

		hash, resource := copyConfig(ctx, sshIPv4, neighbor, conn, opts...)
		configHashes = append(configHashes, hash)
		configResources = append(configResources, resource)
	}

	installFn, iErr := template.Render("./assets/frr/gre/run.sh.j2", map[string]any{
		"tunnels": strings.Join(tunnels, " "),
	})
	if iErr != nil {
		return nil, iErr
	}
	greInstall, gErr := remote.NewCommand(ctx, "remote-command-install-gre", &remote.CommandArgs{
		Create:     pulumi.StringPtr(installFn),
		Update:     pulumi.StringPtr(installFn),
		Triggers:   configHashes,
		Connection: conn,
	}, append(opts, remotefile.DependsOnAll(configResources...))...)
	if gErr != nil {
		return nil, gErr
	}

	return installFirewall(
		ctx,
		firewallTunnels,
		bgpConfig.PublicNetworks.IPv6,
		conn,
		append(opts, pulumi.DependsOn([]pulumi.Resource{greInstall}))...,
	)
}

// copyConfig renders the GRE netplan configuration file, and copies it to the remote server.
// A rendering error fails the deployment: an empty configuration would be copied, and applied otherwise.
// ctx: Pulumi context.
// localIP: The local IP address to be used in the GRE configuration.
// neighbor: The BGP neighbor for which the GRE configuration is being created.
// conn: The remote connection arguments.
// opts: Additional Pulumi resource options.
func copyConfig(
	ctx *pulumi.Context,
	localIP pulumi.StringOutput,
	neighbor *bgp.NeighborConfig,
	conn *remote.ConnectionArgs,
	opts ...pulumi.ResourceOption,
) (pulumi.StringOutput, pulumi.ResourceOutput) {
	name := *neighbor.InterfaceName

	netplanConfig, _ := localIP.ApplyT(func(localIP string) (string, error) {
		return template.Render("./assets/frr/gre/config/netplan.yml.j2", map[string]any{
			"Name":        name,
			"LocalIP":     localIP,
			"RemoteIP":    *neighbor.GRE.RemoteIP,
			"TunnelIP":    *neighbor.GRE.TunnelIP,
			"LinkLocalIP": *neighbor.GRE.LinkLocalIP,
			"Type":        defaults.GetOrDefault(neighbor.GRE.Type, bgp.GreTypeGre),
			"Key":         defaults.GetOrDefault(neighbor.GRE.Key, 0),
			"MTU":         defaults.GetOrDefault(neighbor.GRE.MTU, bgp.DefaultGreMTU),
		})
	}).(pulumi.StringOutput)

	return remotefile.CopyRendered(
		ctx,
		fmt.Sprintf("remote-copy-gre-netplan-%s", name),
		fmt.Sprintf("./outputs/gre_netplan_%s.yaml", name),
		pulumi.String(fmt.Sprintf("/etc/netplan/%s.yaml", name)),
		netplanConfig,
		conn,
		opts...,
	)
}
