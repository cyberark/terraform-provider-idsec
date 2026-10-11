// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package pcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	api "github.com/cyberark/idsec-sdk-golang/pkg"
	"github.com/cyberark/idsec-sdk-golang/pkg/services/pcloud/accounts/actions"
	accountsmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/pcloud/accounts/models"
	platformsmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/pcloud/platforms/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

// platformExistsValidator checks at plan time that the platform_id specified on an account
// refers to a platform that already exists in Privilege Cloud. Platforms are never managed
// by Terraform, so there is no risk of a false positive from same-plan creation.
type platformExistsValidator struct{}

// ValidatePlan implements provider.IdsecPlanValidator.
func (v platformExistsValidator) ValidatePlan(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
	idsecAPI *api.IdsecAPI,
) {
	// Only validate on create (state is null).
	if !req.State.Raw.IsNull() {
		return
	}

	var platformID types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("platform_id"), &platformID)...)
	if resp.Diagnostics.HasError() || platformID.IsNull() || platformID.IsUnknown() || platformID.ValueString() == "" {
		return
	}

	platformsService, err := idsecAPI.PcloudPlatforms()
	if err != nil {
		resp.Diagnostics.AddWarning("Platform Validation Skipped", fmt.Sprintf("Could not initialise platforms service: %s", err))
		return
	}

	_, err = platformsService.Get(&platformsmodels.IdsecPCloudGetPlatform{PlatformID: platformID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("platform_id"),
			"Platform Does Not Exist",
			fmt.Sprintf("Platform %q was not found in Privilege Cloud.", platformID.ValueString()),
		)
	}
}

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "pcloud-accounts",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "pcloud-account",
						ActionDescription: "Manage Privilege Cloud account information, metadata, and credentials",
						ActionVersion:     2,
						Schemas:           actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{},
					ComputedAttributes: []string{
						"account_id",
						"status",
						"created_time",
						"category_modification_time",
						"last_modified_time",
					},
					ImmutableAttributes: []string{
						"safe_name",
						"secret_type",
					},
					StateSchema: &accountsmodels.IdsecPCloudAccount{},
					AttributeNotes: map[string][]tfactions.DocNote{
						"secret": {{
							Severity: tfactions.DocNoteWarning,
							Body: "`secret` is a write-only attribute — its value is never stored in plain text in state " +
								"and always reads back as null, even when set in your configuration. " +
								"Any outputs or locals that reference the state value, e.g. " +
								"`idsec_pcloud_account.example_resource.secret`, will also be null. " +
								"After `terraform import`, a subsequent `terraform apply` will show a diff like: " +
								"`+ secret_write_only_hash               = (sensitive value)`. " +
								"This is expected: the import records the configured value as a hash in state " +
								"but does not write it to the vault. Once state and configuration agree, " +
								"no further apply will push the credential — even if the vault still holds an older value. " +
								"To rotate the credential, change `secret` to a new value and apply.",
						},
							{
								Severity: tfactions.DocNoteWarning,
								Body: "** When upgrading to version 0.13 or later a one-time state migration is needed:** " +
									"After upgrading the provider, run terraform " +
									"apply (or terraform apply -refresh-only) once. The plan shows no changes, " +
									"but the apply rewrites state so the previously stored plain-text secret is " +
									"removed and replaced by secret_write_only_hash. " +
									"No credential is sent to the vault.",
							},
						},
						"secret_file": {{
							Severity: tfactions.DocNoteWarning,
							Body: "Editing the contents of the file at this path does not rotate the credential: " +
								"Terraform compares the path, not the file's contents. To rotate, change the path, " +
								"or set the `secret` attribute instead.",
						}},
					},
				},
				SupportedOperations:       []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:           map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "create", tfactions.ReadOperation: "get", tfactions.UpdateOperation: "update", tfactions.DeleteOperation: "delete"},
				ImportID:                  "account_id",
				PlanValidators:            []tfactions.IdsecPlanValidator{platformExistsValidator{}},
				WriteOnlyHashedAttributes: []string{"secret"},
				UpdateOnlyModified:        true,
				StateUpgrades: map[int64]tfactions.StateUpgrade{
					1: {}, // "secret" became write-only (hashed)
				},
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "pcloud-account", ActionDescription: "Privilege Cloud account data source, reads account information and metadata, based on the account ID.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{"account_id"},
					StateSchema:             &accountsmodels.IdsecPCloudAccount{},
				},
				DataSourceAction: "get",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "pcloud-account-credentials", ActionDescription: "Privilege Cloud account credentials data source, reads account credentials from vault, based on the account ID.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{"account_id"},
					StateSchema:             &accountsmodels.IdsecPCloudAccountCredentials{},
				},
				DataSourceAction: "get-credentials",
			},
		},
	})
}
