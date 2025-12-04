package vault

import (
	"github.com/muhlba91/pulumi-shared-library/pkg/model/google/iam/serviceaccount"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/storage"
)

// Data holds references to resources created for Vault.
type Data struct {
	// Google Service Account used by Vault to access GCP resources
	ServiceAccount *serviceaccount.User
	// GCS Bucket used by Vault for storage backend
	Bucket *storage.Bucket
}
