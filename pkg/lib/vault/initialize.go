package vault

import (
	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/vault"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/util/script"
)

// Initializes Vault on the remote server via SSH.
// ctx: Pulumi context.
// sshIPv4: The IPv4 address of the server to connect to via SSH.
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// dependsOn: Pulumi resource option to specify dependencies.
func initialize(
	ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	dependsOn pulumi.ResourceOrInvokeOption,
) (*pulumi.AnyOutput, error) {
	conn := &remote.ConnectionArgs{
		Host:       sshIPv4,
		PrivateKey: privateKeyPem,
		User:       pulumi.String("root"),
	}

	initScript, sErr := file.ReadContents("./assets/vault/init.sh")
	if sErr != nil {
		return nil, sErr
	}

	cmd, cErr := remote.NewCommand(ctx, "vault-init", &remote.CommandArgs{
		Create:     pulumi.StringPtr(initScript),
		Connection: conn,
	}, dependsOn, pulumi.Timeouts(&pulumi.CustomTimeouts{
		Create: "40m",
		Update: "40m",
	}))
	if cErr != nil {
		return nil, cErr
	}

	keys, _ := cmd.Stdout.ApplyT(func(stdout string) (*vault.Keys, error) {
		parsedTokens, err := script.ParseTokens(stdout)
		if err != nil {
			return nil, err
		}

		rootToken, _ := parsedTokens["root_token"].(string)
		var recoveryKeys []string
		if rk, ok := parsedTokens["recovery_keys"].([]any); ok {
			for _, key := range rk {
				recoveryKeys = append(recoveryKeys, key.(string))
			}
		}

		return &vault.Keys{
			RootToken:    rootToken,
			RecoveryKeys: recoveryKeys,
		}, nil
	}).(pulumi.AnyOutput)

	return &keys, nil
}
