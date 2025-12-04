package frr

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

// Data holds the outputs from the FRR resource creation.
type Data struct {
	// Hostname is the hostname of the FRR instance.
	Hostname pulumi.StringOutput
	// NeighborPassword is the BGP neighbor password.
	NeighborPassword pulumi.StringOutput
}
