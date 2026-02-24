package vault

import (
	"github.com/muhlba91/pulumi-shared-library/pkg/model/google/iam/serviceaccount"
	"github.com/muhlba91/pulumi-shared-library/pkg/model/scaleway/iam/application"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/vault"
)

// CreateResources creates resources for Vault based on the provided configuration.
// ctx: The Pulumi context for resource creation.
// serviceAccount: The Google service account used for authentication.
// application: The Scaleway application used for authentication.
func createResources(
	ctx *pulumi.Context,
	serviceAccount *serviceaccount.User,
	application *application.Application,
) (*vault.Data, error) {
	scwBucket, err := createBucket(ctx)
	if err != nil {
		return nil, err
	}

	return &vault.Data{
		ServiceAccount: serviceAccount,
		Application:    application,
		ScalewayBucket: scwBucket,
	}, nil
}
