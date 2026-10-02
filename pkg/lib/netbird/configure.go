package netbird

import (
	"encoding/json"

	"github.com/muhlba91/pulumi-shared-library/pkg/lib/vault/secret"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/vault/store"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi-vault/sdk/v7/go/vault"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/netbird"
	vaultModel "github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/vault"
)

// Configure configures a NetBird instance on a server: it initializes it, and stores the personal access token in Vault.
// The returned token resolves once it is stored in Vault, and the returned command is the initialization.
// ctx: Pulumi context.
// sshIPv4: The IPv4 address of the server to connect to via SSH.
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// netbirdData: NetBird configuration data.
// vaultInstanceData: The Vault instance data output.
// dependsOn: Pulumi resource option to specify dependencies.
func configure(
	ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	netbirdData *netbird.Data,
	vaultInstanceData *pulumi.AnyOutput,
	dependsOn pulumi.ResourceOrInvokeOption,
) (pulumi.StringOutput, *remote.Command, error) {
	token, initCmd, iErr := initialize(ctx, sshIPv4, privateKeyPem, netbirdData, dependsOn)
	if iErr != nil {
		return pulumi.StringOutput{}, nil, iErr
	}

	stored, _ := pulumi.All(token, *vaultInstanceData).ApplyT(func(vs []any) (string, error) {
		nbToken, _ := vs[0].(string)
		vaultInstance, _ := vs[1].(*vaultModel.Instance)

		return nbToken, storeToken(ctx, nbToken, vaultInstance.Provider)
	}).(pulumi.StringOutput)

	return stored, initCmd, nil
}

// Stores the NetBird personal access token in Vault's KV secrets engine.
// ctx: Pulumi context
// token: The NetBird personal access token.
// provider: Vault provider
func storeToken(ctx *pulumi.Context, token string, provider *vault.Provider) error {
	prefix := "mount"
	mount, err := store.Create(ctx, "kv-netbird", &store.CreateOptions{
		NamePrefix:    &prefix,
		Path:          pulumi.String("netbird"),
		Description:   pulumi.String("NetBird related secrets"),
		PulumiOptions: []pulumi.ResourceOption{pulumi.Provider(provider)},
	})
	if err != nil {
		return err
	}

	value, err := json.Marshal(map[string]string{"token": token})
	if err != nil {
		return err
	}
	secretValue, _ := pulumi.ToSecret(pulumi.String(value)).(pulumi.StringOutput)
	_, err = secret.Create(ctx, &secret.CreateOptions{
		Path:          "netbird",
		Key:           "pat",
		Value:         secretValue,
		PulumiOptions: []pulumi.ResourceOption{pulumi.Provider(provider), pulumi.DependsOn([]pulumi.Resource{mount})},
	})
	return err
}
