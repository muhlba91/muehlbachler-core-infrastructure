package vault

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/muhlba91/pulumi-shared-library/pkg/lib/vault/secret"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/vault/store"
	"github.com/pulumi/pulumi-vault/sdk/v7/go/vault"
	"github.com/pulumi/pulumi-vault/sdk/v7/go/vault/kv"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/dns"
	vaultModel "github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/vault"
)

// Configure configures a Vault instance on a server.
// ctx: Pulumi context.
// sshIPv4: The IPv4 address of the server to connect to via SSH.
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// bucket: The GCS bucket to be used by Vault for storage.
// dnsConfig: DNS configuration.
// dependsOn: Pulumi resource option to specify dependencies.
func configure(
	ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	bucket pulumi.IDOutput,
	dnsConfig *dns.Config,
	dependsOn pulumi.ResourceOrInvokeOption,
) (*pulumi.AnyOutput, error) {
	address := fmt.Sprintf("https://%s", net.JoinHostPort(*dnsConfig.Entries["vault"].Domain, "8200"))

	keys, iErr := initialize(ctx, sshIPv4, privateKeyPem, dependsOn)
	if iErr != nil {
		return nil, iErr
	}

	rootToken, _ := keys.ApplyT(func(k any) string {
		return k.(*vaultModel.Keys).RootToken
	}).(pulumi.StringOutput)
	provider, pErr := vault.NewProvider(ctx, "vault", &vault.ProviderArgs{
		Address: pulumi.StringPtr(address),
		Token:   rootToken,
	})
	if pErr != nil {
		return nil, pErr
	}

	polErr := createDefaultPolicies(ctx, provider)
	if polErr != nil {
		return nil, polErr
	}

	arErr := enableAppRole(ctx, provider)
	if arErr != nil {
		return nil, arErr
	}

	ghErr := enableGitHubAuth(ctx, provider)
	if ghErr != nil {
		return nil, ghErr
	}

	data, _ := pulumi.All(bucket.ToStringOutput(), address, keys).ApplyT(func(vs []any) *vaultModel.Instance {
		vBucket, _ := vs[0].(string)
		vAddress, _ := vs[1].(string)
		vKeys, _ := vs[2].(*vaultModel.Keys)

		ownedSecrets, _ := storeVaultSecrets(ctx, vKeys, provider)

		return &vaultModel.Instance{
			Bucket:       vBucket,
			Address:      vAddress,
			Keys:         vKeys,
			OwnedSecrets: ownedSecrets,
		}
	}).(pulumi.AnyOutput)

	return &data, nil
}

// Stores the Vault secrets (root token and unseal key) in Vault's KV secrets engine.
// ctx: Pulumi context
// keys: Vault keys containing the root token and unseal key
// provider: Vault provider
func storeVaultSecrets(
	ctx *pulumi.Context,
	keys *vaultModel.Keys,
	provider *vault.Provider,
) (*vaultModel.OwnedSecrets, error) {
	prefix := "mount"
	mount, err := store.Create(ctx, "kv-vault", &store.CreateOptions{
		NamePrefix:    &prefix,
		Path:          pulumi.String("vault"),
		Description:   pulumi.String("Vault related secrets"),
		PulumiOptions: []pulumi.ResourceOption{pulumi.Provider(provider)},
	})
	if err != nil {
		return nil, err
	}

	value, _ := json.Marshal(map[string]string{
		"rootToken":    keys.RootToken,
		"recoveryKey1": keys.RecoveryKeys[0],
		"recoveryKey2": keys.RecoveryKeys[1],
		"recoveryKey3": keys.RecoveryKeys[2],
		"recoveryKey4": keys.RecoveryKeys[3],
		"recoveryKey5": keys.RecoveryKeys[4],
	})
	secret, _ := mount.Path.ApplyT(func(path string) *kv.SecretV2 {
		kv, _ := secret.Create(ctx, &secret.CreateOptions{
			Path:          path,
			Key:           "keys",
			Value:         pulumi.String(value),
			PulumiOptions: []pulumi.ResourceOption{pulumi.Provider(provider)},
		})
		return kv
	}).(kv.SecretV2Output)

	return &vaultModel.OwnedSecrets{
		Mount: mount,
		Keys:  &secret,
	}, nil
}
