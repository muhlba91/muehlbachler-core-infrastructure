package main

import (
	"fmt"
	"os"

	"github.com/pulumi/pulumi-vault/sdk/v7/go/vault/kv"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gopkg.in/yaml.v3"

	"github.com/muhlba91/pulumi-shared-library/pkg/lib/tls"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/dir"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/storage"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/storage/google"
	tlsProv "github.com/pulumi/pulumi-tls/sdk/v5/go/tls"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/config"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/docker"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/frr"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/gcloud"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/google/dns"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/google/serviceaccount"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/hetzner/server"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/scaleway"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/scaleway/application"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/tailscale"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/traefik"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/vault"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/wireguard"
	serverModel "github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/server"
	vaultModel "github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/vault"
	wireguardModel "github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/wireguard"
)

//nolint:gocognit,funlen // main is the entry point of the Pulumi program.
func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		dErr := dir.Create("outputs")
		if dErr != nil {
			return dErr
		}

		// configuration
		googleConfig, scalewayConfig, serverConfig, networkConfig, oidcConfig, dnsConfig, bgpConfig, tailscaleConfig, err := config.LoadConfig(
			ctx,
		)
		if err != nil {
			return err
		}

		// instance
		sshKey, sErr := tls.CreateSSHKey(ctx, fmt.Sprintf("core-%s", config.Environment), 0)
		if sErr != nil {
			return sErr
		}
		instance, iErr := server.Create(ctx, sshKey.PublicKeyOpenssh, serverConfig, networkConfig)
		if iErr != nil {
			return iErr
		}
		dependsOn := []pulumi.Resource{instance.Resource}

		// dns
		dnsEntries := dns.Create(ctx, dnsConfig, instance.PublicIPv4, instance.PublicIPv6)
		dependsOn = append(dependsOn, dnsEntries...)

		// docker
		dockerInstall, doErr := docker.Install(ctx, instance.SSHIPv4, sshKey.PrivateKeyPem, pulumi.DependsOn(dependsOn))
		if doErr != nil {
			return doErr
		}
		dependsOn = append(dependsOn, dockerInstall)

		// google cloud
		serviceAccount, saErr := serviceaccount.Create(ctx, googleConfig, dnsConfig)
		if saErr != nil {
			return saErr
		}
		gcloudInstall, gcErr := gcloud.Install(
			ctx,
			instance.SSHIPv4,
			sshKey.PrivateKeyPem,
			serviceAccount,
			pulumi.DependsOn(dependsOn),
		)
		if gcErr != nil {
			return gcErr
		}
		dependsOn = append(dependsOn, gcloudInstall)

		// scaleway
		scwApplication, scaErr := application.Create(ctx, scalewayConfig)
		if scaErr != nil {
			return scaErr
		}
		scalewayInstall, gcErr := scaleway.Install(
			ctx,
			instance.SSHIPv4,
			sshKey.PrivateKeyPem,
			scwApplication,
			scalewayConfig,
			pulumi.DependsOn(dependsOn),
		)
		if gcErr != nil {
			return gcErr
		}
		dependsOn = append(dependsOn, scalewayInstall)

		// traefik
		traefikInstall, tErr := traefik.Install(
			ctx,
			instance.SSHIPv4,
			sshKey.PrivateKeyPem,
			dnsConfig,
			pulumi.DependsOn(dependsOn),
		)
		if tErr != nil {
			return tErr
		}
		dependsOn = append(dependsOn, traefikInstall)

		// vault
		vaultData, vaultInstanceData, _, vdErr := vault.Install(
			ctx,
			instance.SSHIPv4,
			sshKey.PrivateKeyPem,
			serviceAccount,
			dnsConfig,
			googleConfig,
			dependsOn,
		)
		if vdErr != nil {
			return vdErr
		}

		// wireguard
		wireguardData, _, wiErr := wireguard.Install(
			ctx,
			instance.SSHIPv4,
			sshKey.PrivateKeyPem,
			dnsConfig,
			oidcConfig,
			dependsOn,
		)
		if wiErr != nil {
			return wiErr
		}

		// frr
		_, _, frrErr := frr.Install(
			ctx,
			instance.SSHIPv4,
			sshKey.PrivateKeyPem,
			instance.Hostname,
			networkConfig,
			bgpConfig,
			dependsOn,
		)
		if frrErr != nil {
			return frrErr
		}

		// tailscale
		_, tsErr := tailscale.Install(
			ctx,
			instance.SSHIPv4,
			sshKey.PrivateKeyPem,
			tailscaleConfig,
			pulumi.DependsOn(dependsOn),
		)
		if tsErr != nil {
			return tsErr
		}

		// write output files
		writeOutputFiles(ctx, sshKey, vaultInstanceData)

		// outputs
		exportPulumiOutputs(ctx, instance, vaultData, vaultInstanceData, wireguardData)

		return nil
	})
}

// writeOutputFiles writes the SSH key and Vault configuration files to the specified storage.
// ctx: The Pulumi context.
// sshKey: The SSH private key resource.
// vaultInstanceData: The Vault instance data output.
func writeOutputFiles(ctx *pulumi.Context, sshKey *tlsProv.PrivateKey, vaultInstanceData *pulumi.AnyOutput) {
	google.WriteFileAndUpload(ctx, &storage.WriteFileAndUploadOptions{
		BucketID:    config.BucketID,
		BucketPath:  fmt.Sprintf("%s/", config.BucketPath),
		OutputPath:  "./outputs",
		Name:        "ssh.key",
		Content:     sshKey.PrivateKeyPem,
		Labels:      config.CommonLabels(),
		Permissions: []os.FileMode{0o600},
	})
	vaultYaml, _ := vaultInstanceData.ApplyT(func(data any) string {
		b, _ := yaml.Marshal(map[string]any{
			"address": data.(*vaultModel.Instance).Address,
			"keys":    data.(*vaultModel.Instance).Keys,
		})
		return string(b)
	}).(pulumi.StringOutput)
	google.WriteFileAndUpload(ctx, &storage.WriteFileAndUploadOptions{
		BucketID:    config.BucketID,
		BucketPath:  fmt.Sprintf("%s/", config.BucketPath),
		OutputPath:  "./outputs",
		Name:        "vault.yml",
		Content:     vaultYaml,
		Labels:      config.CommonLabels(),
		Permissions: []os.FileMode{0o600},
	})
}

// exportPulumiOutputs exports the necessary Pulumi outputs.
// ctx: The Pulumi context.
// instance: The Hetzner server instance data.
// vaultData: The Vault resources data.
// vaultInstanceData: The Vault instance data output.
// wireguardData: The WireGuard resources data.
func exportPulumiOutputs(
	ctx *pulumi.Context,
	instance *serverModel.Data,
	vaultData *vaultModel.Data,
	vaultInstanceData *pulumi.AnyOutput,
	wireguardData *wireguardModel.Data,
) {
	ctx.Export("server", pulumi.ToMap(map[string]any{
		"ipv4": instance.PublicIPv4,
		"ipv6": instance.PublicIPv6,
	}))

	ctx.Export("vault", vaultInstanceData.ApplyT(func(data any) map[string]any {
		instanceData, _ := data.(*vaultModel.Instance)

		return map[string]any{
			"storage": map[string]any{
				"type":   "gcs",
				"bucket": vaultData.Bucket.ID(),
			},
			"address": instanceData.Address,
			"keys": pulumi.ToSecret(map[string]any{
				"rootToken":    instanceData.Keys.RootToken,
				"recoveryKeys": instanceData.Keys.RecoveryKeys,
			}),
			"ownedSecrets": map[string]any{
				"mount": instanceData.OwnedSecrets.Mount.Path,
				"keys": instanceData.OwnedSecrets.Keys.ApplyT(func(keys any) pulumi.StringOutput {
					k, _ := keys.(*kv.SecretV2)
					return k.Path
				}),
			},
		}
	}))

	ctx.Export("wireguard", pulumi.ToMap(map[string]any{
		"adminPassword": wireguardData.AdminPassword,
	}))
}
