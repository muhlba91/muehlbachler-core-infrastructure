package server

import (
	"fmt"

	"github.com/muhlba91/pulumi-shared-library/pkg/lib/hetzner/network/subnet"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/hetzner/primaryip"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/hetzner/server"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/hetzner/sshkey"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/hetzner/location"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/pulumi/convert"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/config"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/hetzner/firewall"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/hetzner/network"
	networkConf "github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/network"
	serverConf "github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/server"
	serverModel "github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/server"
)

// PrimaryIPs holds the primary IP addresses of a Hetzner server.
type PrimaryIPs struct {
	IPv4 pulumi.StringOutput
	IPv6 pulumi.StringOutput
}

// Create creates a new Hetzner server.
// ctx: Pulumi context
// publicSSHKey: Public SSH key to be added to the server for access.
func Create(
	ctx *pulumi.Context,
	publicSSHKey pulumi.StringOutput,
	serverConfig *serverConf.Config,
	networkConfig *networkConf.Config,
) (*serverModel.Data, error) {
	// location & datacenter
	dc := location.ToDatacenter(serverConfig.Location)

	// SSH Key
	hetznerSSHKey, hErr := sshkey.Create(ctx, config.GlobalName, &sshkey.CreateOptions{
		Name:      fmt.Sprintf("%s-%s", config.GlobalName, config.Environment),
		PublicKey: publicSSHKey,
		Labels:    config.CommonLabels(),
	})
	if hErr != nil {
		return nil, hErr
	}

	// network
	network, nErr := network.GetOrCreate(ctx, networkConfig)
	if nErr != nil {
		return nil, nErr
	}
	_, _ = subnet.Create(ctx, config.GlobalName, &subnet.CreateOptions{
		NetworkID: network,
		Cidr:      *networkConfig.SubnetCIDR,
	})

	firewall, fErr := firewall.Create(ctx, networkConfig, serverConfig)
	if fErr != nil {
		return nil, fErr
	}

	// primary IPs
	primaryIPv4, pv4Err := primaryip.Create(ctx, config.GlobalName, &primaryip.CreateOptions{
		Name:       fmt.Sprintf("%s-%s", config.GlobalName, config.Environment),
		IPType:     "ipv4",
		Datacenter: dc,
		AutoDelete: pulumi.Bool(false),
		Labels:     config.CommonLabels(),
	})
	if pv4Err != nil {
		return nil, pv4Err
	}
	primaryIPv6, pv6Err := primaryip.Create(ctx, config.GlobalName, &primaryip.CreateOptions{
		Name:       fmt.Sprintf("%s-%s", config.GlobalName, config.Environment),
		IPType:     "ipv6",
		Datacenter: dc,
		AutoDelete: pulumi.Bool(false),
		Labels:     config.CommonLabels(),
	})
	if pv6Err != nil {
		return nil, pv6Err
	}

	// server
	enableIPv6 := false
	server, sErr := server.Create(
		ctx,
		fmt.Sprintf("%s-%s", config.GlobalName, *serverConfig.Location),
		&server.CreateOptions{
			Hostname: pulumi.Sprintf(
				"%s-%s-%s",
				config.GlobalName,
				config.Environment,
				*serverConfig.Location,
			),
			ServerType:         pulumi.String(*serverConfig.Type),
			Image:              pulumi.String("ubuntu-24.04"),
			SSHKeys:            []pulumi.StringInput{hetznerSSHKey.ID().ToStringOutput()},
			Location:           pulumi.String(*serverConfig.Location),
			NetworkID:          network,
			IPAddress:          pulumi.String(*serverConfig.IPv4),
			PrimaryIPv4Address: primaryIPv4,
			PrimaryIPv6Address: primaryIPv6,
			EnableIPv6:         &enableIPv6,
			Firewalls:          []pulumi.IntInput{convert.IDToInt(firewall.ID())},
			Backups:            pulumi.Bool(false),
			Protection:         true,
			Labels:             config.CommonLabels(),
			PublicSSH:          *serverConfig.PublicSSH,
		},
	)
	if sErr != nil {
		return nil, sErr
	}

	sshIP := pulumi.String(*serverConfig.IPv4).ToStringOutput()
	if *serverConfig.PublicSSH {
		sshIP = primaryIPv4.IpAddress
	}
	return &serverModel.Data{
		Resource:    server.Resource,
		Hostname:    server.Hostname,
		PrivateIPv4: pulumi.String(*serverConfig.IPv4).ToStringOutput(),
		PublicIPv4:  primaryIPv4.IpAddress,
		PublicIPv6:  pulumi.Sprintf("%s1", primaryIPv6.IpAddress),
		SSHIPv4:     sshIP,
		Network:     pulumi.String(*networkConfig.Name).ToStringOutput(),
	}, nil
}
