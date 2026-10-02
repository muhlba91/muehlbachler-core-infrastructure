package netbirdclient

import (
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/netbird/setupkey"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	rModel "github.com/muhlba91/pulumi-shared-library/pkg/model/rotation"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/config"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/netbird"
)

const (
	// setupKeyRotationDays is the rotation of the setup key in days (it expires after twice the rotation).
	setupKeyRotationDays = 90
)

// createSetupKey creates the setup key to register the NetBird client.
// The key is only needed for the first registration, but kept valid to register again if required.
// ctx: Pulumi context.
// netbirdInstance: The NetBird (server) instance.
func createSetupKey(ctx *pulumi.Context, netbirdInstance *netbird.Instance) (pulumi.StringOutput, error) {
	key, err := setupkey.Create(ctx, config.GlobalName, &setupkey.CreateOptions{
		Type:     pulumi.String("reusable"),
		Rotation: &rModel.Options{Days: setupKeyRotationDays},
		PulumiOptions: []pulumi.ResourceOption{
			pulumi.Provider(netbirdInstance.Provider),
			pulumi.DependsOn([]pulumi.Resource{netbirdInstance.Ready, netbirdInstance.Settings}),
		},
	})
	if err != nil {
		return pulumi.StringOutput{}, err
	}

	secretKey, _ := pulumi.ToSecret(key.Key).(pulumi.StringOutput)
	return secretKey, nil
}
