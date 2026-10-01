package remotefile

import (
	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// CopyRendered writes the rendered content to a local file, and copies it to the remote server
// whenever the content (hash) changes.
// The copy can only be created once the content is known: it is returned as resource output,
// to be used as dependency via DependsOnAll.
// It returns the hash of the content (to be used as trigger), and the copy resource output.
// ctx: Pulumi context.
// name: The name of the copy resource.
// localPath: The local path the content is written to.
// remotePath: The path on the remote server.
// content: The rendered content.
// conn: The remote connection arguments.
// opts: Additional Pulumi resource options.
func CopyRendered(
	ctx *pulumi.Context,
	name string,
	localPath string,
	remotePath pulumi.StringInput,
	content pulumi.StringInput,
	conn *remote.ConnectionArgs,
	opts ...pulumi.ResourceOption,
) (pulumi.StringOutput, pulumi.ResourceOutput) {
	hash, _ := file.WritePulumi(localPath, content).
		ApplyT(func(_ string) string {
			h, _ := file.Hash(localPath)
			return *h
		}).(pulumi.StringOutput)
	copyResource, _ := hash.ApplyT(func(_ string) (pulumi.Resource, error) {
		return remote.NewCopyToRemote(ctx, name, &remote.CopyToRemoteArgs{
			Source:     pulumi.NewFileAsset(localPath),
			RemotePath: remotePath,
			Triggers:   pulumi.Array{hash},
			Connection: conn,
		}, opts...)
	}).(pulumi.ResourceOutput)

	return hash, copyResource
}

// DependsOnAll declares explicit dependencies on resources (outputs), which are only available
// once their values are known (e.g., the copies created via CopyRendered).
// Dependencies which are not known yet (e.g., during previews) are ignored.
// resources: The resource outputs to depend on.
func DependsOnAll(resources ...pulumi.ResourceOutput) pulumi.ResourceOrInvokeOption {
	return pulumi.DependsOnInputs(pulumi.NewResourceArrayOutput(resources...))
}
