package dns

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/vault"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/wireguard"
	dnsConf "github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/dns"
)

// Create creates DNS records for the given IP addresses based on the provided DNS configuration.
// ctx: The Pulumi context for resource creation.
// dnsConfig: The DNS configuration containing domain and record details.
// ipv4: The public IPv4 address to create DNS records for.
// ipv6: The public IPv6 address to create DNS records for.
func Create(
	ctx *pulumi.Context,
	dnsConfig *dnsConf.Config,
	ipv4 pulumi.StringOutput,
	ipv6 pulumi.StringOutput,
) []pulumi.Resource {
	var resources []pulumi.Resource

	resources = append(resources, vault.CreateDNSRecords(ctx, dnsConfig, ipv4, ipv6)...)
	resources = append(resources, wireguard.CreateDNSRecords(ctx, dnsConfig, ipv4, ipv6)...)

	return resources
}
