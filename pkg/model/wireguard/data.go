package wireguard

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

// Data defines WireGuard data.
type Data struct {
	// AdminPassword is the admin password.
	AdminPassword pulumi.StringOutput
	// Database contains database related data.
	Database *Database
	// OIDC contains OIDC related data.
	OIDC *OIDC
	// Web contains web related data.
	Web *Web
}

// Database defines WireGuard database data.
type Database struct {
	// EncryptionPassphrase is the encryption passphrase.
	EncryptionPassphrase pulumi.StringOutput
}

// OIDC defines WireGuard OIDC data.
type OIDC struct {
	// BaseURL is the base URL.
	BaseURL string
	// ClientID is the client ID.
	ClientID string
	// ClientSecret is the client secret.
	//nolint:gosec // This is a configuration value, not a hardcoded secret.
	ClientSecret string
}

// Web defines WireGuard web data.
type Web struct {
	// SessionSecret is the session secret.
	SessionSecret pulumi.StringOutput
	// CSRFSecret is the CSRF secret.
	CSRFSecret pulumi.StringOutput
}
