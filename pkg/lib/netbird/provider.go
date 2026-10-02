package netbird

import (
	"fmt"

	nbProvider "github.com/KitStream/netbird-pulumi-provider/sdk/go/netbird"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/dns"
)

// createProvider creates the NetBird provider to manage resources of the NetBird instance via its API.
// ctx: Pulumi context.
// dnsConfig: DNS configuration.
// token: The personal access token of the initial owner.
func createProvider(
	ctx *pulumi.Context,
	dnsConfig *dns.Config,
	token pulumi.StringOutput,
) (*nbProvider.Provider, error) {
	return nbProvider.NewProvider(ctx, "netbird", &nbProvider.ProviderArgs{
		ManagementUrl: pulumi.String(fmt.Sprintf("https://%s", *dnsConfig.Entries["netbird"].Domain)),
		Token:         token,
	})
}
