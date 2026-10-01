package dns

import (
	"maps"
	"slices"

	"github.com/muhlba91/pulumi-shared-library/pkg/lib/google/dns/record"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	dnsConf "github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/dns"
)

// Create creates DNS records for the given IP addresses for every entry of the provided DNS configuration.
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

	for _, name := range slices.Sorted(maps.Keys(dnsConfig.Entries)) {
		resources = append(resources, createRecords(ctx, dnsConfig, name, ipv4, ipv6)...)
	}

	return resources
}

// createRecords creates the A and AAAA records for a DNS entry.
// ctx: The Pulumi context for resource creation.
// dnsConfig: The DNS configuration containing domain and record details.
// name: The name of the DNS entry.
// ipv4: The public IPv4 address to create the record for.
// ipv6: The public IPv6 address to create the record for.
func createRecords(
	ctx *pulumi.Context,
	dnsConfig *dnsConf.Config,
	name string,
	ipv4 pulumi.StringOutput,
	ipv6 pulumi.StringOutput,
) []pulumi.Resource {
	dnsEntry := dnsConfig.Entries[name]

	v4, v4Err := record.Create(ctx, &record.CreateOptions{
		Domain:     *dnsEntry.Domain,
		ZoneID:     pulumi.String(*dnsEntry.ZoneID),
		RecordType: "A",
		Records:    pulumi.StringArray([]pulumi.StringInput{ipv4}),
		Project:    dnsConfig.Project,
	})
	if v4Err != nil {
		return nil
	}

	v6, v6Err := record.Create(ctx, &record.CreateOptions{
		Domain:     *dnsEntry.Domain,
		ZoneID:     pulumi.String(*dnsEntry.ZoneID),
		RecordType: "AAAA",
		Records:    pulumi.StringArray([]pulumi.StringInput{ipv6}),
		Project:    dnsConfig.Project,
	})
	if v6Err != nil {
		return nil
	}

	return []pulumi.Resource{v4, v6}
}
