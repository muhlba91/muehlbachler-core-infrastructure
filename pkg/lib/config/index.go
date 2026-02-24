package config

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/bgp"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/dns"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/google"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/network"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/oidc"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/scaleway"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/server"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/tailscale"
)

//nolint:gochecknoglobals // global configuration is acceptable here
var (
	// Environment holds the current deployment environment (e.g., dev, staging, prod).
	Environment string
	// GlobalName is a constant name used across resources.
	GlobalName = "core"
	// ScalewayDefaultRegion is the default Scaleway region for deployments.
	ScalewayDefaultRegion = "fr-par"
	// BucketPath is the path within the buckets for this project.
	BucketPath string
	// BackupBucketPath is the path within the backup buckets for this project.
	BackupBucketPath string
	// BucketID is the ID of the main storage bucket.
	BucketID string
	// BackupBucketID is the ID of the backup storage bucket.
	BackupBucketID string
)

// LoadConfig loads the configuration for the given Pulumi context.
// ctx: The Pulumi context.
func LoadConfig(
	ctx *pulumi.Context,
) (*google.Config, *scaleway.Config, *server.Config, *network.Config, *oidc.Config, *dns.Config, *bgp.Config, *tailscale.Config, error) {
	Environment = ctx.Stack()

	cfg := config.New(ctx, "")

	BucketID = cfg.Require("bucketId")
	BackupBucketID = cfg.Require("backupBucketId")
	BucketPath = fmt.Sprintf("%s/%s", GlobalName, Environment)
	BackupBucketPath = fmt.Sprintf("%s/backup", BucketPath)

	var googleConfig google.Config
	cfg.RequireObject("gcp", &googleConfig)

	var scalewayConfig scaleway.Config
	cfg.RequireObject("scaleway", &scalewayConfig)

	var serverConfig server.Config
	cfg.RequireObject("server", &serverConfig)

	var networkConfig network.Config
	cfg.RequireObject("network", &networkConfig)

	var oidcConfig oidc.Config
	cfg.RequireObject("oidc", &oidcConfig)

	var dnsConfig dns.Config
	cfg.RequireObject("dns", &dnsConfig)

	var bgpConfig bgp.Config
	cfg.RequireObject("bgp", &bgpConfig)

	var tailscaleConfig tailscale.Config
	cfg.RequireObject("tailscale", &tailscaleConfig)

	return &googleConfig, &scalewayConfig, &serverConfig, &networkConfig, &oidcConfig, &dnsConfig, &bgpConfig, &tailscaleConfig, nil
}

// CommonLabels returns a map of common labels to be used across resources.
func CommonLabels() map[string]string {
	return map[string]string{
		"environment": Environment,
		"application": GlobalName,
	}
}
