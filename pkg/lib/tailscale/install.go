package tailscale

import (
	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/template"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	gcpConfig "github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/config"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/config"
	tailscaleConf "github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/tailscale"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/util/install"
)

// Install Tailscale on the remote server via SSH.
// ctx: Pulumi context.
// sshIPv4: The IPv4 address of the server to connect to via SSH.
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// tailscaleConfig: Configuration for Tailscale installation.
// dependsOn: Pulumi resource option to specify dependencies.
func Install(
	ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	tailscaleConfig *tailscaleConf.Config,
	dependsOn pulumi.ResourceOrInvokeOption,
) (*remote.Command, error) {
	conn := &remote.ConnectionArgs{
		Host:       sshIPv4,
		PrivateKey: privateKeyPem,
		User:       pulumi.String("root"),
	}

	opts := []pulumi.ResourceOption{dependsOn}

	opts, prepErr := install.Prepare(ctx, "tailscale", conn, opts...)
	if prepErr != nil {
		return nil, prepErr
	}

	dockerCompose, dcErr := template.Render("./assets/tailscale/docker-compose.yml.j2", map[string]interface{}{
		"authKey": tailscaleConfig.AuthKey,
	})
	if dcErr != nil {
		return nil, dcErr
	}
	dockerComposeHash := file.WritePulumi("./outputs/tailscale_docker-compose.yml", pulumi.String(dockerCompose)).
		ApplyT(func(_ string) string {
			hash, _ := file.Hash("./outputs/tailscale_docker-compose.yml")
			return *hash
		})
	dockerComposeCopy := dockerComposeHash.ApplyT(func(_ string) pulumi.ResourceOption {
		cmd, _ := remote.NewCopyToRemote(
			ctx,
			"remote-copy-tailscale-docker-compose",
			&remote.CopyToRemoteArgs{
				Source:     pulumi.NewFileAsset("./outputs/tailscale_docker-compose.yml"),
				RemotePath: pulumi.String("/opt/tailscale/docker-compose.yml"),
				Triggers:   pulumi.Array{dockerComposeHash},
				Connection: conn,
			},
			opts...)
		return pulumi.DependsOn([]pulumi.Resource{cmd})
	})

	cronResources, cronErr := install.Cron(ctx, "tailscale", conn, opts...)
	if cronErr != nil {
		return nil, cronErr
	}

	opts, systemdServiceHash, shErr := install.SystemDService(ctx, "tailscale", conn, opts...)
	if shErr != nil {
		return nil, shErr
	}

	installFn, iErr := template.Render("./assets/tailscale/install.sh.j2", map[string]interface{}{
		"project": gcpConfig.GetProject(ctx),
		"bucket": map[string]string{
			"id":   config.BackupBucketID,
			"path": config.BucketPath,
		},
	})
	if iErr != nil {
		return nil, iErr
	}
	return remote.NewCommand(ctx, "remote-command-install-tailscale", &remote.CommandArgs{
		Create:     pulumi.StringPtr(installFn),
		Update:     pulumi.StringPtr(installFn),
		Triggers:   pulumi.Array{dockerComposeHash, pulumi.String(*systemdServiceHash)},
		Connection: conn,
	}, append(opts, install.CollectResourceOptions(append(cronResources, dockerComposeCopy))...)...)
}
