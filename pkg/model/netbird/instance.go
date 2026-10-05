package netbird

import (
	nbProvider "github.com/KitStream/netbird-pulumi-provider/sdk/go/netbird"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Instance defines a NetBird (server) instance which is installed, initialized, and reachable.
type Instance struct {
	// Data contains the NetBird configuration data.
	Data *Data
	// Token is the personal access token of the initial owner.
	Token pulumi.StringOutput
	// Provider is the NetBird provider to manage resources (e.g., setup keys) of the instance.
	Provider *nbProvider.Provider
	// Ready is the resource which is completed once the instance is reachable via its public address.
	Ready pulumi.Resource
	// Settings is the resource configuring the account: resources of the instance (e.g., setup keys) depend on it.
	Settings pulumi.Resource
	// BackboneGroupID is the ID of the group of the backbone peers, which may reach each other.
	BackboneGroupID pulumi.StringOutput
}
