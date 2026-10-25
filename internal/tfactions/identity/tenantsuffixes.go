// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

// Package identity registers Terraform actions for CyberArk Identity services,
// including users, roles, policies, directories, web apps, authentication profiles,
// tenant suffixes and federations, with the provider action registry.
package identity

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	api "github.com/cyberark/idsec-sdk-golang/pkg"
	"github.com/cyberark/idsec-sdk-golang/pkg/services/identity/tenantsuffixes/actions"
	tenantsuffixesmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/identity/tenantsuffixes/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
	"github.com/cyberark/terraform-provider-idsec/internal/schemas"
)

// tenantSuffixRenameIDValidator plans id as unknown when an update renames the suffix. id is the
// server's row key, <tenant-id>.<suffix> on a suffixless tenant, so a rename changes it -- but as a
// computed-only attribute it is pinned by UseStateForUnknown, which would plan the old key and then
// keep it in state after the apply. A case-only change is left alone: suffix compares
// case-insensitively, so it plans no update at all.
type tenantSuffixRenameIDValidator struct{}

// ValidatePlan implements provider.IdsecPlanValidator.
func (v tenantSuffixRenameIDValidator) ValidatePlan(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
	_ *api.IdsecAPI,
) {
	// Create has no prior id to pin, and destroy has no plan.
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	suffixPath := path.Root("suffix")
	var planned, prior types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, suffixPath, &planned)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, suffixPath, &prior)...)
	if resp.Diagnostics.HasError() || planned.IsUnknown() {
		return
	}
	if strings.EqualFold(strings.TrimSpace(planned.ValueString()), strings.TrimSpace(prior.ValueString())) {
		return
	}
	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("id"), types.StringUnknown())...)
}

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "identity-tenant-suffixes",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "identity-tenant-suffix",
						ActionDescription: "The Identity service tenant suffix resource.",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
					},
					StateSchema: &tenantsuffixesmodels.IdsecIdentityTenantSuffix{},
					// id is the server's row key and is what an update renames from. It must also be the
					// ImportID: that is the only attribute an update carries from prior state, since computed
					// attributes are otherwise stripped from the update request.
					ComputedAttributes: []string{"id"},
					SemanticEqualityAttributes: map[string]schemas.SemanticEqualityKind{
						"suffix": schemas.SemanticEqualityCaseInsensitive,
					},
					PageNotes: []tfactions.DocNote{
						{
							Severity: tfactions.DocNoteDanger,
							Body:     "Deleting a tenant suffix that is set as the tenant's default suffix also clears the default suffix setting for the tenant. Set a different default before destroying this resource.",
						},
						{
							Severity: tfactions.DocNoteWarning,
							Body:     "A suffix that is still in use by any user cannot be deleted; the API rejects the request. Move or remove those users first, or terraform destroy will fail.",
						},
						{
							Severity: tfactions.DocNoteWarning,
							Body:     "Changing `suffix` renames it in place. On most tenants, every Cloud Directory user whose login is on the old suffix is moved to the new one, so `user@old.example` becomes `user@new.example`, and the default suffix follows if it was the old one. If any user cannot be moved, the server keeps the old suffix as well; the apply then reports an error, and the old suffix stays in the tenant outside Terraform's management. On tenants with the suffixless configuration enabled, the rename does not move any users: they keep their logins on the old suffix, which the rename deletes. Move them to the new suffix after the apply.",
						},
						{
							Severity: tfactions.DocNoteInfo,
							Body:     "To read a suffix this resource manages, use the `idsec_identity_tenant_suffix` data source. Do not rely on the `idsec_identity_tenant_suffixes` data source for this: on tenants with the suffixless configuration enabled, the server stores a created suffix under the key `<tenant-id>.<suffix>`, and that data source omits every suffix whose key differs from the suffix itself, so a suffix created here does not appear in it.",
						},
					},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings: map[tfactions.IdsecServiceActionOperation]string{
					tfactions.CreateOperation: "create",
					tfactions.ReadOperation:   "get",
					tfactions.UpdateOperation: "update",
					tfactions.DeleteOperation: "delete",
				},
				PlanValidators: []tfactions.IdsecPlanValidator{tenantSuffixRenameIDValidator{}},
				// The import value may be the suffix itself; the SDK's Get resolves it to the row.
				ImportID: "id",
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "identity-tenant-suffix",
						ActionDescription: "The Identity service tenant suffix data source. It reads a single tenant suffix and fails if the tenant does not have it.",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
					},
					StateSchema: &tenantsuffixesmodels.IdsecIdentityTenantSuffix{},
				},
				DataSourceAction: "get",
			},
		},
	})
}
