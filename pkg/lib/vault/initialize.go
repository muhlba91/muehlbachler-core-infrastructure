package vault

import (
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/vault"
)

// Initializes Vault on the remote server via SSH.
// ctx: Pulumi context.
// sshIPv4: The IPv4 address of the server to connect to via SSH.
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// bucket: The GCS bucket to be used by Vault for storage.
// dnsConfig: DNS configuration.
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

	script, sErr := file.ReadContents("./assets/vault/init.sh")
	if sErr != nil {
		return nil, sErr
	}

	cmd, cErr := remote.NewCommand(ctx, "vault-init", &remote.CommandArgs{
		Create:     pulumi.StringPtr(script),
		Connection: conn,
	}, dependsOn, pulumi.Timeouts(&pulumi.CustomTimeouts{
		Create: "40m",
		Update: "40m",
	}))
	if cErr != nil {
		return nil, cErr
	}

	keys, _ := cmd.Stdout.ApplyT(func(stdout string) *vault.Keys {
		startBlock := strings.LastIndex(stdout, "--START TOKENS--")
		endBlock := strings.LastIndex(stdout, "--END TOKENS--")
		tokens := stdout[startBlock+len("--START TOKENS--") : endBlock]
		parsedTokens := parse(tokens)

		var recoveryKeys []string
		if rk, ok := parsedTokens["recovery_keys"].([]any); ok {
			for _, key := range rk {
				recoveryKeys = append(recoveryKeys, key.(string))
			}
		}

		return &vault.Keys{
			RootToken:    parsedTokens["root_token"].(string),
			RecoveryKeys: recoveryKeys,
		}
	}).(pulumi.AnyOutput)

	return &keys, nil
}

// parse parses a YAML-formatted string into a map[string]interface{}.
// On error it returns an empty map.
func parse(s string) map[string]interface{} {
	var out map[string]interface{}
	if err := yaml.Unmarshal([]byte(s), &out); err != nil {
		return map[string]interface{}{}
	}
	return out
}
