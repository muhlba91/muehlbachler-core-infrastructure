package wireguard

import (
	"fmt"

	"github.com/muhlba91/pulumi-shared-library/pkg/lib/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/config"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/oidc"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/wireguard"
)

// Length of the generated WireGuard secrets.
const wireguardSecretLength = 32

// CreateResources creates resources for WireGuard based on the provided configuration.
// ctx: The Pulumi context for resource creation.
// oidcConfig: Configuration related to OIDC (OpenID Connect).
func createResources(
	ctx *pulumi.Context,
	oidcConfig *oidc.Config,
) (*wireguard.Data, error) {
	adminPassword, apErr := random.CreatePassword(
		ctx,
		fmt.Sprintf("password-wireguard-admin-password-%s", config.Environment),
		&random.PasswordOptions{
			Special: false,
		},
	)
	if apErr != nil {
		return nil, apErr
	}

	databaseEncryptionPassphrase, depErr := random.CreatePassword(
		ctx,
		fmt.Sprintf("password-wireguard-database-encryption-passphrase-%s", config.Environment),
		&random.PasswordOptions{
			Length:  wireguardSecretLength,
			Special: false,
		},
	)
	if depErr != nil {
		return nil, depErr
	}

	sessionSecret, ssErr := random.CreatePassword(
		ctx,
		fmt.Sprintf("password-wireguard-web-session-secret-%s", config.Environment),
		&random.PasswordOptions{
			Length:  wireguardSecretLength,
			Special: false,
		},
	)
	if ssErr != nil {
		return nil, ssErr
	}

	csrfSecret, csErr := random.CreatePassword(
		ctx,
		fmt.Sprintf("password-wireguard-web-csrf-secret-%s", config.Environment),
		&random.PasswordOptions{
			Length:  wireguardSecretLength,
			Special: false,
		},
	)
	if csErr != nil {
		return nil, csErr
	}

	return &wireguard.Data{
		AdminPassword: adminPassword.Password,
		Database: &wireguard.Database{
			EncryptionPassphrase: databaseEncryptionPassphrase.Password,
		},
		OIDC: &wireguard.OIDC{
			BaseURL:      *oidcConfig.DiscoveryURL,
			ClientID:     *oidcConfig.Clients["wireguard"].ClientID,
			ClientSecret: *oidcConfig.Clients["wireguard"].ClientSecret,
		},
		Web: &wireguard.Web{
			SessionSecret: sessionSecret.Password,
			CSRFSecret:    csrfSecret.Password,
		},
	}, nil
}
