package netbird

import (
	nbProvider "github.com/KitStream/netbird-pulumi-provider/sdk/go/netbird"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/netbird/group"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/netbird/policy"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	pModel "github.com/muhlba91/pulumi-shared-library/pkg/model/netbird/policy"
)

const (
	// backboneName is the name of the backbone group.
	backboneName = "backbone"
	// backbonePolicyName is the name of the backbone policy and its rule.
	backbonePolicyName = "backbone-mesh"
	// backbonePolicyDescription is the description of the backbone policy.
	backbonePolicyDescription = "inter-site routing: BGP/EVPN/VXLAN between hub and site routers"
)

// createBackbone creates the backbone group and the policy allowing its peers to reach each other.
// Returns the ID of the backbone group.
// ctx: Pulumi context.
// provider: The NetBird provider.
// dependsOn: Pulumi resources the backbone depends on.
func createBackbone(
	ctx *pulumi.Context,
	provider *nbProvider.Provider,
	dependsOn []pulumi.Resource,
) (pulumi.StringOutput, error) {
	opts := []pulumi.ResourceOption{pulumi.Provider(provider), pulumi.DependsOn(dependsOn)}

	backbone, gErr := group.Create(ctx, backboneName, &group.CreateOptions{PulumiOptions: opts})
	if gErr != nil {
		return pulumi.StringOutput{}, gErr
	}
	groupID := backbone.ID().ToStringOutput()

	desc, accept, all, yes := backbonePolicyDescription, "accept", "all", true
	_, pErr := policy.Create(ctx, backbonePolicyName, &policy.CreateOptions{
		Description: &desc,
		Rules: []pModel.Rule{{
			Name:          backbonePolicyName,
			Action:        &accept,
			Bidirectional: &yes,
			Protocol:      &all,
			Sources:       pulumi.StringArray{groupID},
			Destinations:  pulumi.StringArray{groupID},
		}},
		PulumiOptions: opts,
	})
	if pErr != nil {
		return pulumi.StringOutput{}, pErr
	}

	return groupID, nil
}
