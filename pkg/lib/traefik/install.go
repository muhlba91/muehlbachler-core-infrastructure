package traefik

import (
	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/template"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/dns"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/util/install"
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

	dockerCompose, dcErr := template.Render("./assets/traefik/docker-compose.yml.j2", map[string]interface{}{
		"gcpProject": dnsConfig.Project,
	})
	if dcErr != nil {
		return nil, dcErr
	}
	dockerComposeHash := file.WritePulumi("./outputs/traefik_docker-compose.yml", pulumi.String(dockerCompose)).
		ApplyT(func(_ string) string {
			hash, _ := file.Hash("./outputs/traefik_docker-compose.yml")
			return *hash
		})
	dockerComposeCopy := dockerComposeHash.ApplyT(func(_ string) pulumi.ResourceOption {
		cmd, _ := remote.NewCopyToRemote(
			ctx,
			"remote-copy-traefik-docker-compose",
			&remote.CopyToRemoteArgs{
				Source:     pulumi.NewFileAsset("./outputs/traefik_docker-compose.yml"),
				RemotePath: pulumi.String("/opt/traefik/docker-compose.yml"),
				Triggers:   pulumi.Array{dockerComposeHash},
				Connection: conn,
			},
			opts...)
		return pulumi.DependsOn([]pulumi.Resource{cmd})
	})

	traefikYaml, dcErr := template.Render("./assets/traefik/traefik.yml.j2", map[string]interface{}{
		"acmeEmail": dnsConfig.Email,
	})
	if dcErr != nil {
		return nil, dcErr
	}
	traefikYmlHash := file.WritePulumi("./outputs/traefik_traefik.yml", pulumi.String(traefikYaml)).
		ApplyT(func(_ string) string {
			hash, _ := file.Hash("./outputs/traefik_traefik.yml")
			return *hash
		})
	traefikYmlCopy := traefikYmlHash.ApplyT(func(_ string) pulumi.ResourceOption {
		cmd, _ := remote.NewCopyToRemote(ctx, "remote-copy-traefik-config", &remote.CopyToRemoteArgs{
			Source:     pulumi.NewFileAsset("./outputs/traefik_traefik.yml"),
			RemotePath: pulumi.String("/opt/traefik/traefik.yml"),
			Triggers:   pulumi.Array{traefikYmlHash},
			Connection: conn,
		}, opts...)
		return pulumi.DependsOn([]pulumi.Resource{cmd})
	})

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
	}, append(opts, install.CollectResourceOptions([]pulumi.Output{dockerComposeCopy, traefikYmlCopy})...)...)
}
