package vault

import (
	"github.com/muhlba91/pulumi-shared-library/pkg/model/google/iam/serviceaccount"
	"github.com/muhlba91/pulumi-shared-library/pkg/model/scaleway/iam/application"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway/object"
)

// Data holds references to resources created for Vault.
type Data struct {
	// Google Service Account used by Vault to access GCP resources
	ServiceAccount *serviceaccount.User
	// Scaleway Application used by Vault to access Scaleway resources
	Application *application.Application
	// Scaleway S3 Bucket used by Vault for storage backend
	ScalewayBucket *object.Bucket
}
