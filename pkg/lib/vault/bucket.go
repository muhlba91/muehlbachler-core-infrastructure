package vault

import (
	"fmt"

	gcs "github.com/muhlba91/pulumi-shared-library/pkg/lib/google/storage/bucket"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/google/storage/iam"
	scw "github.com/muhlba91/pulumi-shared-library/pkg/lib/scaleway/storage/bucket"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway/object"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/config"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/google"
)

// defaultOneZoneTransitionDays is the default number of days before transitioning objects to One Zone storage class in Scaleway.
const defaultOneZoneTransitionDays = 3650

// createBucket creates a Google Cloud Storage bucket for Vault.
// ctx: The Pulumi context for resource creation.
// serviceAccount: The email of the service account to grant access to the bucket.
// googleConfig: The Google configuration containing project and other settings.
func createBucket(
	ctx *pulumi.Context,
	serviceAccount pulumi.StringInput,
	googleConfig *google.Config,
) (*storage.Bucket, *object.Bucket, error) {
	gBucket, gcsErr := createGCSBucket(ctx, serviceAccount, googleConfig)
	if gcsErr != nil {
		return nil, nil, gcsErr
	}

	scwBucket, scwErr := createScalewayBucket(ctx)
	if scwErr != nil {
		return nil, nil, scwErr
	}

	return gBucket, scwBucket, nil
}

// createGCSBucket creates a Google Cloud Storage bucket for Vault.
// ctx: The Pulumi context for resource creation.
// serviceAccount: The email of the service account to grant access to the bucket.
// googleConfig: The Google configuration containing project and other settings.
func createGCSBucket(
	ctx *pulumi.Context,
	serviceAccount pulumi.StringInput,
	googleConfig *google.Config,
) (*storage.Bucket, error) {
	bucket, bErr := gcs.Create(ctx, "vault", &gcs.CreateOptions{
		Location: pulumi.String(*googleConfig.Region),
		Labels:   config.CommonLabels(),
	})
	if bErr != nil {
		return nil, bErr
	}

	_ = pulumi.All(serviceAccount, bucket.ID().ToStringOutput()).ApplyT(func(args []any) error {
		email, _ := args[0].(string)
		bucketID, _ := args[1].(string)

		_, iamErr := iam.CreateIAMMember(
			ctx,
			&iam.MemberOptions{
				BucketID: bucketID,
				Member:   fmt.Sprintf("serviceAccount:%s", email),
				Role:     "roles/storage.objectAdmin",
			},
		)
		return iamErr
	})

	return bucket, nil
}

// createScalewayBucket creates a Scaleway Storage bucket for Vault.
// ctx: The Pulumi context for resource creation.
// serviceAccount: The email of the service account to grant access to the bucket.
func createScalewayBucket(
	ctx *pulumi.Context,
) (*object.Bucket, error) {
	oneZoneTransitionDays := defaultOneZoneTransitionDays

	bucket, bErr := scw.Create(ctx, "vault", &scw.CreateOptions{
		Location:              pulumi.String(config.ScalewayDefaultRegion),
		OneZoneTransitionDays: &oneZoneTransitionDays,
		Labels:                config.CommonLabels(),
	})
	if bErr != nil {
		return nil, bErr
	}

	return bucket, nil
}
