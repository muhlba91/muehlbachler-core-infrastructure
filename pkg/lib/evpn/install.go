package evpn

import (
	"fmt"
	"maps"

	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/template"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/bgp"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/util/install"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/util/remotefile"
)

const (
	// netbirdInterface is the interface of the NetBird client (see assets/netbird-client/docker-compose.yml.j2).
	netbirdInterface = "wt0"
)

// Install the EVPN devices, and firewall rules on the remote server via SSH.
// ctx: Pulumi context.
// sshIPv4: The IPv4 address of the server to connect to via SSH (the BGP router ID).
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// vtep: The local VXLAN address (NetBird IPv4 address of the server).
// evpnConfig: The EVPN configuration.
// dependsOn: Pulumi resource option to specify dependencies.
func installer(
	ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	vtep pulumi.StringOutput,
	evpnConfig *bgp.EVPNConfig,
	dependsOn pulumi.ResourceOrInvokeOption,
) (*remote.Command, error) {
	conn := &remote.ConnectionArgs{
		Host:       sshIPv4,
		PrivateKey: privateKeyPem,
		User:       pulumi.String("root"),
	}

	opts := []pulumi.ResourceOption{dependsOn}

	opts, prepErr := install.Prepare(ctx, "evpn", conn, opts...)
	if prepErr != nil {
		return nil, prepErr
	}

	// the device names are derived from the VNI
	templateData := map[string]any{
		"vrf":              *evpnConfig.VRF,
		"table":            *evpnConfig.Table,
		"vni":              *evpnConfig.VNI,
		"mtu":              *evpnConfig.MTU,
		"bridge":           fmt.Sprintf("br%d", *evpnConfig.VNI),
		"vxlan":            fmt.Sprintf("vni%d", *evpnConfig.VNI),
		"netbirdInterface": netbirdInterface,
	}

	configResources, configHashes, cErr := createConfigs(ctx, templateData, vtep, sshIPv4, conn, opts...)
	if cErr != nil {
		return nil, cErr
	}

	opts, systemdServiceHash, shErr := install.SystemDService(ctx, "evpn", conn, opts...)
	if shErr != nil {
		return nil, shErr
	}

	installFn, _ := vtep.ApplyT(func(vtepValue string) (string, error) {
		return template.Render("./assets/evpn/install.sh.j2", withValues(templateData, map[string]any{
			"vtep": vtepValue,
		}))
	}).(pulumi.StringOutput)
	return remote.NewCommand(ctx, "remote-command-install-evpn", &remote.CommandArgs{
		Create:     installFn,
		Update:     installFn,
		Triggers:   append(configHashes, pulumi.String(*systemdServiceHash)),
		Connection: conn,
	}, append(opts, remotefile.DependsOnAll(configResources...))...)
}

// createConfigs generates the EVPN configuration files and uploads them to the remote server.
// ctx: Pulumi context.
// templateData: The values rendered into the configuration files.
// vtep: The local VXLAN address (NetBird IPv4 address of the server).
// routerID: The BGP router ID (IPv4 address), which the router MAC address is derived from.
// conn: The remote connection arguments.
// opts: Additional Pulumi resource options.
func createConfigs(
	ctx *pulumi.Context,
	templateData map[string]any,
	vtep pulumi.StringOutput,
	routerID pulumi.StringOutput,
	conn *remote.ConnectionArgs,
	opts ...pulumi.ResourceOption,
) ([]pulumi.ResourceOutput, pulumi.Array, error) {
	networkdConfigHash, nhErr := file.Hash("./assets/evpn/config/networkd.conf")
	if nhErr != nil {
		return nil, nil, nhErr
	}
	networkdConfigCopy, ncErr := remote.NewCopyToRemote(ctx, "remote-copy-evpn-networkd", &remote.CopyToRemoteArgs{
		Source:     pulumi.NewFileAsset("./assets/evpn/config/networkd.conf"),
		RemotePath: pulumi.String("/etc/systemd/networkd.conf.d/50-routing-daemons.conf"),
		Triggers:   pulumi.Array{pulumi.String(*networkdConfigHash)},
		Connection: conn,
	}, opts...)
	if ncErr != nil {
		return nil, nil, ncErr
	}

	netplanConfig, _ := pulumi.All(vtep, routerID).ApplyT(func(args []any) (string, error) {
		vtepValue, _ := args[0].(string)
		routerIDValue, _ := args[1].(string)

		routerMac, mErr := routerMAC(routerIDValue)
		if mErr != nil {
			return "", mErr
		}
		return template.Render("./assets/evpn/config/netplan.yml.j2", withValues(templateData, map[string]any{
			"vtep":      vtepValue,
			"routerMac": routerMac,
		}))
	}).(pulumi.StringOutput)
	// the digit prefix keeps the configuration from being removed as orphaned GRE tunnel configuration
	netplanConfigHash, netplanConfigCopy := remotefile.CopyRendered(
		ctx,
		"remote-copy-evpn-netplan",
		"./outputs/evpn_netplan.yaml",
		pulumi.String("/etc/netplan/60-evpn.yaml"),
		netplanConfig,
		conn,
		opts...,
	)

	firewallScript, fwErr := template.Render("./assets/evpn/firewall.sh.j2", templateData)
	if fwErr != nil {
		return nil, nil, fwErr
	}
	firewallScriptHash, firewallScriptCopy := remotefile.CopyRendered(
		ctx,
		"remote-copy-evpn-firewall",
		"./outputs/evpn_firewall.sh",
		pulumi.String("/opt/evpn/firewall.sh"),
		pulumi.String(firewallScript),
		conn,
		opts...,
	)

	return []pulumi.ResourceOutput{
		pulumi.NewResourceOutput(networkdConfigCopy),
		netplanConfigCopy,
		firewallScriptCopy,
	}, pulumi.Array{
		pulumi.String(*networkdConfigHash),
		netplanConfigHash,
		firewallScriptHash,
	}, nil
}

// withValues returns a copy of the template data with additional values.
// templateData: The values rendered into the configuration files.
// values: The additional values.
func withValues(templateData map[string]any, values map[string]any) map[string]any {
	data := maps.Clone(templateData)
	maps.Copy(data, values)
	return data
}
