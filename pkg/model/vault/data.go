package vault

import (
	"github.com/muhlba91/pulumi-shared-library/pkg/model/google/iam/serviceaccount"
	"github.com/muhlba91/pulumi-shared-library/pkg/model/scaleway/iam/application"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/storage"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway/object"
)

// Data holds references to resources created for Vault.
type Data struct {
	// Google Service Account used by Vault to access GCP resources
	ServiceAccount *serviceaccount.User
	// Scaleway Application used by Vault to access Scaleway resources
	Application *application.Application
	// GCS Bucket used by Vault for storage backend
	GCSBucket *storage.Bucket
	// Scaleway S3 Bucket used by Vault for storage backend
	ScalewayBucket *object.Bucket
}
