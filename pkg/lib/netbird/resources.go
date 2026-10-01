package netbird

import (
	"errors"
	"fmt"

	"github.com/muhlba91/pulumi-shared-library/pkg/lib/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/config"
	netbirdConf "github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/netbird"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/netbird"
)

const (
	// netbirdSecretLength is the length of the generated NetBird secrets.
	netbirdSecretLength = 32
	// netbirdKeyBytes is the number of random bytes of generated encryption keys (AES-256).
	netbirdKeyBytes = 32
	// netbirdDatabaseName is the name of the database.
	netbirdDatabaseName = "netbird"
	// netbirdDatabaseUser is the user of the database.
	netbirdDatabaseUser = "netbird"
)

// createResources creates resources for NetBird based on the provided configuration.
// ctx: The Pulumi context for resource creation.
// netbirdConfig: The NetBird configuration.
func createResources(
	ctx *pulumi.Context,
	netbirdConfig *netbirdConf.Config,
) (*netbird.Data, error) {
	if err := validateConfig(netbirdConfig); err != nil {
		return nil, err
	}

	authSecret, asErr := random.CreatePassword(
		ctx,
		fmt.Sprintf("password-netbird-auth-secret-%s", config.Environment),
		&random.PasswordOptions{
			Length:  netbirdSecretLength,
			Special: false,
		},
	)
	if asErr != nil {
		return nil, asErr
	}

	databasePassword, dpErr := random.CreatePassword(
		ctx,
		fmt.Sprintf("password-netbird-database-password-%s", config.Environment),
		&random.PasswordOptions{
			Length:  netbirdSecretLength,
			Special: false,
		},
	)
	if dpErr != nil {
		return nil, dpErr
	}

	sessionCookieKey, sckErr := random.CreateBytes(
		ctx,
		fmt.Sprintf("bytes-netbird-session-cookie-encryption-key-%s", config.Environment),
		&random.BytesOptions{
			Length: netbirdKeyBytes,
		},
	)
	if sckErr != nil {
		return nil, sckErr
	}

	adminPassword, apErr := random.CreatePassword(
		ctx,
		fmt.Sprintf("password-netbird-admin-password-%s", config.Environment),
		&random.PasswordOptions{
			Length:  netbirdSecretLength,
			Special: false,
		},
	)
	if apErr != nil {
		return nil, apErr
	}

	return &netbird.Data{
		AuthSecret:                 authSecret.Password,
		SessionCookieEncryptionKey: sessionCookieKey.Base64Std,
		StoreEncryptionKey:         secretString(*netbirdConfig.StoreEncryptionKey),
		DatabasePassword:           databasePassword.Password,
		Admin: &netbird.Admin{
			Name:     *netbirdConfig.Admin.Name,
			Email:    *netbirdConfig.Admin.Email,
			Password: adminPassword.Password,
		},
	}, nil
}

// validateConfig verifies the required NetBird configuration is present.
// netbirdConfig: The NetBird configuration.
func validateConfig(netbirdConfig *netbirdConf.Config) error {
	if netbirdConfig == nil || netbirdConfig.StoreEncryptionKey == nil || *netbirdConfig.StoreEncryptionKey == "" {
		return errors.New("netbird: storeEncryptionKey must be configured")
	}
	admin := netbirdConfig.Admin
	if admin == nil || admin.Name == nil || admin.Email == nil {
		return errors.New("netbird: admin.name and admin.email must be configured")
	}
	return nil
}

// secretString wraps a plain string into a Pulumi secret output.
// value: The secret value.
func secretString(value string) pulumi.StringOutput {
	out, _ := pulumi.ToSecret(pulumi.String(value)).(pulumi.StringOutput)
	return out
}
