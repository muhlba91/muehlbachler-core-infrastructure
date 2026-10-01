package netbird

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

// Data defines NetBird data.
type Data struct {
	// AuthSecret is the relay authentication secret.
	AuthSecret pulumi.StringOutput
	// SessionCookieEncryptionKey is the key encrypting the identity provider session cookies.
	SessionCookieEncryptionKey pulumi.StringOutput
	// StoreEncryptionKey is the key encrypting sensitive data at rest.
	StoreEncryptionKey pulumi.StringOutput
	// DatabasePassword is the database password.
	DatabasePassword pulumi.StringOutput
	// Admin contains the initial owner data.
	Admin *Admin
}

// Admin defines the NetBird initial owner data.
type Admin struct {
	// Name is the display name.
	Name string
	// Email is the email address.
	Email string
	// Password is the password.
	Password pulumi.StringOutput
}
