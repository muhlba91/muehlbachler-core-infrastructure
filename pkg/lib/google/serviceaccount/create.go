package serviceaccount

import (
	"fmt"

	"github.com/muhlba91/pulumi-shared-library/pkg/lib/google/iam/role"
	kmsIam "github.com/muhlba91/pulumi-shared-library/pkg/lib/google/kms/iam"
	gmodel "github.com/muhlba91/pulumi-shared-library/pkg/model/google/iam/serviceaccount"
	slServiceAccount "github.com/muhlba91/pulumi-shared-library/pkg/util/google/iam/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/lib/config"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/dns"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/config/google"
)

// Create a Google Cloud Service Account with necessary IAM roles.
// ctx: Pulumi context for resource management.
// googleConfig: Configuration details for Google Cloud.
// dnsConfig: Configuration details for DNS management.
func Create(ctx *pulumi.Context, googleConfig *google.Config, dnsConfig *dns.Config) (*gmodel.User, error) {
	iam, err := slServiceAccount.CreateServiceAccountUser(ctx, &slServiceAccount.CreateOptions{
		Name:    fmt.Sprintf("%s-%s-%s", config.GlobalName, config.GlobalName, config.Environment),
		Project: pulumi.String(*googleConfig.Project),
	})
	if err != nil {
		return nil, err
	}

	iam.ServiceAccount.Email.ApplyT(func(email string) error {
		keyringID := fmt.Sprintf(
			"%s/%s/%s",
			*googleConfig.Project,
			*googleConfig.EncryptionKey.Location,
			*googleConfig.EncryptionKey.KeyringID,
		)
		_, _ = kmsIam.CreateKeyringBinding(ctx, &kmsIam.KeyringBindingOptions{
			KeyRingID: keyringID,
			Member:    fmt.Sprintf("serviceAccount:%s", email),
			Role:      "roles/cloudkms.cryptoKeyEncrypterDecrypter",
		})
		_, _ = kmsIam.CreateKeyringBinding(ctx, &kmsIam.KeyringBindingOptions{
			KeyRingID: keyringID,
			Member:    fmt.Sprintf("serviceAccount:%s", email),
			Role:      "roles/cloudkms.signerVerifier",
		})
		_, _ = kmsIam.CreateKeyringBinding(ctx, &kmsIam.KeyringBindingOptions{
			KeyRingID: keyringID,
			Member:    fmt.Sprintf("serviceAccount:%s", email),
			Role:      "roles/cloudkms.viewer",
		})

		_, _ = role.CreateMember(ctx, fmt.Sprintf("%s-dns-admin", email), &role.MemberOptions{
			Member:  pulumi.Sprintf("serviceAccount:%s", email),
			Roles:   []string{"roles/dns.admin"},
			Project: pulumi.String(*dnsConfig.Project),
		})

		return nil
	})

	return iam, nil
}
