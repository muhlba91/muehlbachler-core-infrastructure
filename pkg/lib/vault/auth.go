package vault

import (
	"github.com/pulumi/pulumi-vault/sdk/v7/go/vault"
	"github.com/pulumi/pulumi-vault/sdk/v7/go/vault/jwt"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Enables the AppRole authentication method in Vault.
// ctx: Pulumi context.
// provider: Vault provider.
func enableAppRole(
	ctx *pulumi.Context,
	provider *vault.Provider,
) error {
	_, err := vault.NewAuthBackend(ctx, "vault-auth-backend-approle", &vault.AuthBackendArgs{
		Type:        pulumi.String("approle"),
		Description: pulumi.String("App Role Backend"),
		Tune:        &vault.AuthBackendTuneArgs{},
	}, pulumi.Provider(provider))
	return err
}

// Enables the GitHub authentication method in Vault.
// ctx: Pulumi context.
// provider: Vault provider.
func enableGitHubAuth(
	ctx *pulumi.Context,
	provider *vault.Provider,
) error {
	_, err := jwt.NewAuthBackend(ctx, "vault-auth-jwt-github", &jwt.AuthBackendArgs{
		Path:             pulumi.String("github"),
		BoundIssuer:      pulumi.String("https://token.actions.githubusercontent.com"),
		OidcDiscoveryUrl: pulumi.String("https://token.actions.githubusercontent.com"),
		Description:      pulumi.String("GitHub JWT Trust for Actions"),
	}, pulumi.Provider(provider))
	return err
}
