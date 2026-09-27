// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package pcloud

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/pcloud/localgroups/actions"
	localgroupsmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/pcloud/localgroups/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
	"github.com/cyberark/terraform-provider-idsec/internal/schemas"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "pcloud-localgroups",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "pcloud-local-group",
						ActionDescription: "Privilege Cloud local group resource, manages Vault local groups for CP/CCP workloads.",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
					},
					ComputedAttributes: []string{
						"group_id",
						"group_type",
						"created_time",
						"last_modified_time",
					},
					// PVWA's EditUserGroup only applies group_name; description and location are
					// accepted in the payload but silently ignored, so changing either requires
					// destroy+recreate.
					ImmutableAttributes: []string{
						"description",
						"location",
					},
					// The Vault matches group names and locations case-insensitively.
					SemanticEqualityAttributes: map[string]schemas.SemanticEqualityKind{
						"group_name": schemas.SemanticEqualityCaseInsensitive,
						"location":   schemas.SemanticEqualityCaseInsensitive,
					},
					StateSchema: &localgroupsmodels.IdsecPCloudLocalGroup{},
					PageNotes: []tfactions.DocNote{
						{
							Severity: tfactions.DocNoteInfo,
							Body: "Only `group_name` can be changed after the group is created. The Vault " +
								"does not apply a change to `description` or `location`, so changing " +
								"either fails the plan and the group has to be recreated instead.",
						},
					},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{
					tfactions.CreateOperation,
					tfactions.ReadOperation,
					tfactions.UpdateOperation,
					tfactions.DeleteOperation,
					tfactions.StateOperation,
				},
				ActionsMappings: map[tfactions.IdsecServiceActionOperation]string{
					tfactions.CreateOperation: "create",
					tfactions.ReadOperation:   "get",
					tfactions.UpdateOperation: "update",
					tfactions.DeleteOperation: "delete",
				},
				ImportID: "group_id",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "pcloud-local-group-member",
						ActionDescription: "Privilege Cloud local group member resource, manages membership of a user or group in a Vault local group.",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
					},
					ComputedAttributes: []string{
						"membership_id",
					},
					// Membership has no updatable part: all identifying attributes are immutable.
					ImmutableAttributes: []string{
						"group_id",
						"member_name",
						"member_type",
					},
					// The Vault matches member names case-insensitively.
					SemanticEqualityAttributes: map[string]schemas.SemanticEqualityKind{
						"member_name": schemas.SemanticEqualityCaseInsensitive,
					},
					StateSchema: &localgroupsmodels.IdsecPCloudLocalGroupMember{},
					PageNotes: []tfactions.DocNote{
						{
							Severity: tfactions.DocNoteInfo,
							Body: "This resource does not refresh its state and cannot be imported: a " +
								"membership removed outside Terraform is not detected, and the resource " +
								"has to be recreated to restore it.",
						},
					},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{
					tfactions.CreateOperation,
					tfactions.DeleteOperation,
				},
				ActionsMappings: map[tfactions.IdsecServiceActionOperation]string{
					tfactions.CreateOperation: "add-member",
					tfactions.DeleteOperation: "delete-member",
				},
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "pcloud-local-group",
						ActionDescription: "Privilege Cloud local group data source, reads a Vault local group by its ID.",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{"group_id"},
					StateSchema:             &localgroupsmodels.IdsecPCloudLocalGroup{},
				},
				DataSourceAction: "get",
			},
		},
	})
}
