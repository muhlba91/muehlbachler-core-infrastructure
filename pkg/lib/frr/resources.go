package frr

import (
	"fmt"

	"github.com/muhlba91/pulumi-shared-library/pkg/lib/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/config"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/network"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/frr"
)

// CreateResources creates resources for FRR based on the provided configuration.
// ctx: The Pulumi context for resource creation.
// hostname: The hostname for the FRR instance.
// networkConfig: The network configuration details.
func createResources(
	ctx *pulumi.Context,
	hostname pulumi.StringOutput,
	networkConfig *network.Config,
) (*frr.Data, error) {
	neighborPassword, err := random.CreatePassword(
		ctx,
		fmt.Sprintf("password-frr-neighbor-password-%s", config.Environment),
		&random.PasswordOptions{
			Special: false,
		},
	)
	if err != nil {
		return nil, err
	}

	return &frr.Data{
		Hostname:         pulumi.Sprintf("%s.%s", hostname, *networkConfig.DNSSuffix),
		NeighborPassword: neighborPassword.Password,
	}, nil
}
