package vault

import (
	"github.com/muhlba91/pulumi-shared-library/pkg/model/google/iam/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/google"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/vault"
)

// CreateResources creates resources for Vault based on the provided configuration.
// CreateResources creates resources for Vault based on the provided configuration.
// ctx: The Pulumi context for resource creation.
// serviceAccount: The Google service account used for authentication.
// googleConfig: The Google configuration containing project and other settings.
func createResources(
	ctx *pulumi.Context,
	serviceAccount *serviceaccount.User,
	googleConfig *google.Config,
) (*vault.Data, error) {
	bucket, err := createBucket(ctx, serviceAccount.ServiceAccount.Email, googleConfig)
	if err != nil {
		return nil, err
	}

	return &vault.Data{
		ServiceAccount: serviceAccount,
		Bucket:         bucket,
	}, nil
}
