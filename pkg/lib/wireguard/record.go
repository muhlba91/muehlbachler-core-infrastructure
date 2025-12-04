package wireguard

import (
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/google/dns/record"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	dnsConf "github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/dns"
)

// CreateDNSRecords creates DNS records for Wireguard based on the provided DNS configuration.
// ctx: The Pulumi context for resource creation.
// dnsConfig: The DNS configuration containing domain and record details.
// ipv4: The public IPv4 address to create DNS records for.
// ipv6: The public IPv6 address to create DNS records for.
func CreateDNSRecords(
	ctx *pulumi.Context,
	dnsConfig *dnsConf.Config,
	ipv4 pulumi.StringOutput,
	ipv6 pulumi.StringOutput,
) []pulumi.Resource {
	dnsEntry := dnsConfig.Entries["wireguard"]

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
