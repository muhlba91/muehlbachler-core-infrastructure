package firewall

import (
	"fmt"
	"strconv"

	slFirewall "github.com/muhlba91/pulumi-shared-library/pkg/lib/hetzner/firewall"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/config"
	networkConf "github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/network"
	serverConf "github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/server"
)

// networkAllCIDR defines the CIDR blocks that represent all IP addresses.
//
//nolint:gochecknoglobals // global is acceptable here
var networkAllCIDR = []pulumi.StringInput{pulumi.String("0.0.0.0/0"), pulumi.String("::/0")}

// Create gets or creates a Hetzner firewall based on the provided configuration.
// ctx: Pulumi context
// networkConfig: Configuration for the Hetzner network.
// serverConfig: Configuration for the Hetzner server.
func Create(
	ctx *pulumi.Context,
	networkConfig *networkConf.Config,
	serverConfig *serverConf.Config,
) (*hcloud.Firewall, error) {
	sshSourceIps := networkAllCIDR
	if !*serverConfig.PublicSSH {
		sshSourceIps = []pulumi.StringInput{pulumi.String(*networkConfig.SubnetCIDR)}
	}
	sshRule := slFirewall.Rule{
		Description: pulumi.String("Allow incoming SSH traffic"),
		Direction:   "in",
		Port:        "22",
		Protocol:    "tcp",
		SourceIPs:   sshSourceIps,
	}

	rules := []slFirewall.Rule{sshRule}
	for _, rule := range networkConfig.FirewallRules {
		rSourceIps := networkAllCIDR
		if rule.SourceIPs != nil {
			rSourceIps = []pulumi.StringInput{}
			for _, ip := range rule.SourceIPs {
				rSourceIps = append(rSourceIps, pulumi.String(ip))
			}
		}
		rules = append(rules, slFirewall.Rule{
			Description: pulumi.String(*rule.Description),
			Direction:   "in",
			Port:        strconv.Itoa(*rule.Port),
			Protocol:    *rule.Protocol,
			SourceIPs:   rSourceIps,
		})
	}

	return slFirewall.Create(ctx, config.GlobalName, &slFirewall.CreateOptions{
		Name:   fmt.Sprintf("%s-%s", config.GlobalName, config.Environment),
		Labels: config.CommonLabels(),
		Rules:  rules,
	})
}
