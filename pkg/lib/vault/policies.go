package vault

import (
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/vault/policy"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/pulumi/pulumi-vault/sdk/v7/go/vault"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Creates the default Vault policies.
// ctx: Pulumi context.
// provider: Vault provider.
func createDefaultPolicies(
	ctx *pulumi.Context,
	provider *vault.Provider,
) error {
	policyDoc, rErr := file.ReadContents("./assets/vault/policies/admin.hcl")
	if rErr != nil {
		return rErr
	}
	_, pErr := policy.Create(ctx, &policy.CreateOptions{
		Name:   "admin",
		Policy: pulumi.String(policyDoc),
		PulumiOptions: []pulumi.ResourceOption{
			pulumi.Provider(provider),
		},
	})
	if pErr != nil {
		return pErr
	}

	policyDoc, rErr = file.ReadContents("./assets/vault/policies/manager.hcl")
	if rErr != nil {
		return rErr
	}
	_, pErr = policy.Create(ctx, &policy.CreateOptions{
		Name:   "manager",
		Policy: pulumi.String(policyDoc),
		PulumiOptions: []pulumi.ResourceOption{
			pulumi.Provider(provider),
		},
	})
	if pErr != nil {
		return pErr
	}

	policyDoc, rErr = file.ReadContents("./assets/vault/policies/reader.hcl")
	if rErr != nil {
		return rErr
	}
	_, pErr = policy.Create(ctx, &policy.CreateOptions{
		Name:   "reader",
		Policy: pulumi.String(policyDoc),
		PulumiOptions: []pulumi.ResourceOption{
			pulumi.Provider(provider),
		},
	})
	if pErr != nil {
		return pErr
	}

	return nil
}
