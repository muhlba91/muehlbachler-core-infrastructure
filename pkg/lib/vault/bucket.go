package vault

import (
	"fmt"

	"github.com/muhlba91/pulumi-shared-library/pkg/lib/google/storage/bucket"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/google/storage/iam"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/config"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/google"
)

// createBucket creates a Google Cloud Storage bucket for Vault.
// ctx: The Pulumi context for resource creation.
// serviceAccount: The email of the service account to grant access to the bucket.
// googleConfig: The Google configuration containing project and other settings.
func createBucket(
	ctx *pulumi.Context,
	serviceAccount pulumi.StringInput,
	googleConfig *google.Config,
) (*storage.Bucket, error) {
	bucket, bErr := bucket.Create(ctx, "vault", &bucket.CreateOptions{
		Location: pulumi.String(*googleConfig.Region),
		Labels:   config.CommonLabels(),
	})
	if bErr != nil {
		return nil, bErr
	}

	_ = pulumi.All(serviceAccount, bucket.ID().ToStringOutput()).ApplyT(func(args []interface{}) error {
		email, _ := args[0].(string)
		bucketID, _ := args[1].(string)

		_, iamErr := iam.CreateIAMMember(
			ctx,
			&iam.MemberArgs{
				BucketID: bucketID,
				Member:   fmt.Sprintf("serviceAccount:%s", email),
				Role:     "roles/storage.objectAdmin",
			},
		)
		return iamErr
	})

	return bucket, nil
}
