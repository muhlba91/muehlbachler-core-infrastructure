package netbird

import (
	"github.com/muhlba91/pulumi-shared-library/pkg/util/template"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/dns"
)

// waitReady waits on the remote server via SSH until NetBird is reachable via its public address.
// Resources talking to the NetBird API from outside depend on it: DNS, Traefik, and the certificate must work.
// The command is only run on creation, and never re-run.
// ctx: Pulumi context.
// sshIPv4: The IPv4 address of the server to connect to via SSH.
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// dnsConfig: DNS configuration.
// dependsOn: Pulumi resource option to specify dependencies.
func waitReady(
	ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	dnsConfig *dns.Config,
	dependsOn pulumi.ResourceOrInvokeOption,
) (*remote.Command, error) {
	readyScript, rErr := template.Render("./assets/netbird/ready.sh.j2", map[string]any{
		"domain": *dnsConfig.Entries["netbird"].Domain, //nolint:goconst // template key
	})
	if rErr != nil {
		return nil, rErr
	}

	return remote.NewCommand(ctx, "netbird-ready", &remote.CommandArgs{
		Create: pulumi.StringPtr(readyScript),
		Connection: &remote.ConnectionArgs{
			Host:       sshIPv4,
			PrivateKey: privateKeyPem,
			User:       pulumi.String("root"),
		},
	}, dependsOn, pulumi.Timeouts(&pulumi.CustomTimeouts{
		Create: "10m",
	}))
}
