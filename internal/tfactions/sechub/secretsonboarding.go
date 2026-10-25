// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package sechub

import (
	onboardingactions "github.com/cyberark/idsec-sdk-golang/pkg/services/sechub/secretsonboarding/actions"
	onboardingmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/sechub/secretsonboarding/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "sechub-secretsonboarding",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "sechub-secret-onboarding",
						ActionDescription: "Secrets Hub secret onboarding resource, onboards a secret discovered in an external secret store to a PAM account in a Safe.",
						ActionVersion:     1,
						Schemas:           onboardingactions.ActionToSchemaMap,
					},
					// Onboarding has no update API: every input is immutable, including the
					// leaves of the nested pam_account object and its properties map.
					ImmutableAttributes: []string{
						"source_store_type",
						"target_store_type",
						"secret_id",
						"secret_value_type",
						"safe_name",
						"name",
						"platform_id",
						"template",
						"automatic_management_enabled",
						"type",
						"value",
						"key_ref",
					},
					StateSchema: &onboardingmodels.IdsecSecHubOnboardSecret{},
					PageNotes: []tfactions.DocNote{
						{
							Severity: tfactions.DocNoteDanger,
							Body: "Destroying this resource only removes it from the Terraform state. The PAM account is not " +
								"deleted and the secret stays onboarded in Secrets Hub, because there is no offboarding API.",
						},
						{
							Severity: tfactions.DocNoteInfo,
							Body: "This resource does not refresh its state and cannot be imported: an onboarding changed or " +
								"removed outside Terraform is not detected. No attribute can be changed after creation.",
						},
					},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{
					tfactions.CreateOperation,
				},
				ActionsMappings: map[tfactions.IdsecServiceActionOperation]string{
					tfactions.CreateOperation: "create",
				},
			},
		},
	})
}
