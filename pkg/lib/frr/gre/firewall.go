package gre

import (
	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/template"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/util/install"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/util/remotefile"
)

// tunnel is a GRE interface with the addresses of its BGP neighbors.
type tunnel struct {
	// name is the name of the interface.
	name string
	// addresses are the addresses of the BGP neighbors.
	addresses []string
}

// installFirewall installs the firewall rules of the GRE interfaces on the remote server via SSH:
// the host only accepts BGP from the neighbors, ICMPv6, and replies on them,
// and forwards from them to the public networks only.
// ctx: Pulumi context.
// tunnels: The GRE interfaces with their BGP neighbors.
// publicNetworks: The public IPv6 networks, the only destinations forwarded from the GRE interfaces.
// conn: The remote connection arguments.
// opts: Additional Pulumi resource options.
func installFirewall(
	ctx *pulumi.Context,
	tunnels []tunnel,
	publicNetworks []string,
	conn *remote.ConnectionArgs,
	opts ...pulumi.ResourceOption,
) (*remote.Command, error) {
	opts, prepErr := install.Prepare(ctx, "gre-firewall", conn, opts...)
	if prepErr != nil {
		return nil, prepErr
	}

	data := make([]map[string]any, 0, len(tunnels))
	for _, t := range tunnels {
		data = append(data, map[string]any{"name": t.name, "addresses": t.addresses})
	}
	script, sErr := template.Render("./assets/gre-firewall/firewall.sh.j2", map[string]any{
		"tunnels":        data,
		"publicNetworks": publicNetworks,
	})
	if sErr != nil {
		return nil, sErr
	}
	scriptHash, scriptCopy := remotefile.CopyRendered(
		ctx,
		"remote-copy-gre-firewall",
		"./outputs/gre_firewall.sh",
		pulumi.String("/opt/gre-firewall/firewall.sh"),
		pulumi.String(script),
		conn,
		opts...,
	)

	opts, serviceHash, shErr := install.SystemDService(ctx, "gre-firewall", conn, opts...)
	if shErr != nil {
		return nil, shErr
	}

	installFn, iErr := file.ReadContents("./assets/gre-firewall/install.sh")
	if iErr != nil {
		return nil, iErr
	}
	return remote.NewCommand(ctx, "remote-command-install-gre-firewall", &remote.CommandArgs{
		Create:     pulumi.StringPtr(installFn),
		Update:     pulumi.StringPtr(installFn),
		Triggers:   pulumi.Array{scriptHash, pulumi.String(*serviceHash)},
		Connection: conn,
	}, append(opts, remotefile.DependsOnAll(scriptCopy))...)
}
