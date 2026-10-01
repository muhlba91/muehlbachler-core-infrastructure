package gcloud

import (
	"github.com/muhlba91/pulumi-shared-library/pkg/model/google/iam/serviceaccount"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/encoding"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/util/install"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/util/remotefile"
)

// Install gcloud on the remote server via SSH.
// ctx: Pulumi context.
// sshIPv4: The IPv4 address of the server to connect to via SSH.
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// serviceAccount: The Google service account to use for authentication.
// dependsOn: Pulumi resource option to specify dependencies.
func Install(
	ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	serviceAccount *serviceaccount.User,
	dependsOn pulumi.ResourceOrInvokeOption,
) (*remote.Command, error) {
	conn := &remote.ConnectionArgs{
		Host:       sshIPv4,
		PrivateKey: privateKeyPem,
		User:       pulumi.String("root"),
	}

	opts := []pulumi.ResourceOption{dependsOn}

	opts, prepErr := install.Prepare(ctx, "gcloud", conn, opts...)
	if prepErr != nil {
		return nil, prepErr
	}

	privateKey, _ := serviceAccount.Key.PrivateKey.ApplyT(func(key string) string {
		decKey, _ := encoding.B64Decode(key)
		return decKey
	}).(pulumi.StringOutput)
	gcpCredentialsHash, gcpCredentialsCopy := remotefile.CopyRendered(
		ctx,
		"remote-copy-gcloud-service-account",
		"./outputs/google_credentials.json",
		pulumi.String("/opt/google/credentials.json"),
		privateKey,
		conn,
		opts...,
	)

	installFn, iErr := file.ReadContents("./assets/gcloud/install.sh")
	if iErr != nil {
		return nil, iErr
	}
	return remote.NewCommand(ctx, "remote-command-install-gcloud", &remote.CommandArgs{
		Create:     pulumi.StringPtr(installFn),
		Update:     pulumi.StringPtr(installFn),
		Triggers:   pulumi.Array{gcpCredentialsHash},
		Connection: conn,
	}, append(opts, remotefile.DependsOnAll(gcpCredentialsCopy))...)
}
