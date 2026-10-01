package traefik

import (
	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/template"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/dns"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/util/install"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/util/remotefile"
)

// Install Traefik on the remote server via SSH.
// ctx: Pulumi context.
// sshIPv4: The IPv4 address of the server to connect to via SSH.
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// dnsConfig: DNS configuration.
// dependsOn: Pulumi resource option to specify dependencies.
func Install(
	ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	dnsConfig *dns.Config,
	dependsOn pulumi.ResourceOrInvokeOption,
) (*remote.Command, error) {
	conn := &remote.ConnectionArgs{
		Host:       sshIPv4,
		PrivateKey: privateKeyPem,
		User:       pulumi.String("root"),
	}

	opts := []pulumi.ResourceOption{dependsOn}

	opts, prepErr := install.Prepare(ctx, "traefik", conn, opts...)
	if prepErr != nil {
		return nil, prepErr
	}

	dockerCompose, dcErr := template.Render("./assets/traefik/docker-compose.yml.j2", map[string]any{
		"gcpProject": dnsConfig.Project,
	})
	if dcErr != nil {
		return nil, dcErr
	}
	dockerComposeHash, dockerComposeCopy := remotefile.CopyRendered(
		ctx,
		"remote-copy-traefik-docker-compose",
		"./outputs/traefik_docker-compose.yml",
		pulumi.String("/opt/traefik/docker-compose.yml"),
		pulumi.String(dockerCompose),
		conn,
		opts...,
	)

	traefikYaml, dcErr := template.Render("./assets/traefik/traefik.yml.j2", map[string]any{
		"acmeEmail": dnsConfig.Email,
	})
	if dcErr != nil {
		return nil, dcErr
	}
	traefikYmlHash, traefikYmlCopy := remotefile.CopyRendered(
		ctx,
		"remote-copy-traefik-config",
		"./outputs/traefik_traefik.yml",
		pulumi.String("/opt/traefik/traefik.yml"),
		pulumi.String(traefikYaml),
		conn,
		opts...,
	)

	opts, systemdServiceHash, shErr := install.SystemDService(ctx, "traefik", conn, opts...)
	if shErr != nil {
		return nil, shErr
	}

	installFn, iErr := file.ReadContents("./assets/traefik/install.sh")
	if iErr != nil {
		return nil, iErr
	}
	return remote.NewCommand(ctx, "remote-command-install-traefik", &remote.CommandArgs{
		Create:     pulumi.StringPtr(installFn),
		Update:     pulumi.StringPtr(installFn),
		Triggers:   pulumi.Array{dockerComposeHash, pulumi.String(*systemdServiceHash), traefikYmlHash},
		Connection: conn,
	}, append(opts, remotefile.DependsOnAll(dockerComposeCopy, traefikYmlCopy))...)
}
