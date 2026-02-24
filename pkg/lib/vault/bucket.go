package vault

import (
	scw "github.com/muhlba91/pulumi-shared-library/pkg/lib/scaleway/storage/bucket"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway/object"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/config"
)

// defaultOneZoneTransitionDays is the default number of days before transitioning objects to One Zone storage class in Scaleway.
const defaultOneZoneTransitionDays = 3650

// createBucket creates a Google Cloud Storage bucket for Vault.
// ctx: The Pulumi context for resource creation.
func createBucket(
	ctx *pulumi.Context,
) (*object.Bucket, error) {
	scwBucket, scwErr := createScalewayBucket(ctx)
	if scwErr != nil {
		return nil, scwErr
	}

	return scwBucket, nil
}

// createScalewayBucket creates a Scaleway Storage bucket for Vault.
// ctx: The Pulumi context for resource creation.
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
