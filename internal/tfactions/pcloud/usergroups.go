// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package pcloud

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/pcloud/usergroups/actions"
	usergroupsmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/pcloud/usergroups/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
	"github.com/cyberark/terraform-provider-idsec/internal/schemas"
)

// userGroupDeprecationNote is the documentation counterpart of the DeprecationMessage carried by
// both user group resources, kept in one place so the two stay in sync.
var userGroupDeprecationNote = tfactions.DocNote{
	Severity: tfactions.DocNoteWarning,
	Body: "This resource is deprecated and will be removed in a future release. " +
		"Vault-native user groups are no longer the recommended way to group principals.",
}

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "pcloud-usergroups",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "pcloud-user-group",
						ActionDescription: "Privilege Cloud user group resource, manages Vault user groups.",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
					},
					ComputedAttributes: []string{
						"group_id",
						"group_type",
					},
					// The name is the only part of a group PVWA's PUT actually changes. It
					// accepts a description and a location in the payload and answers 200, but
					// a group re-read afterwards still reports the stored values, so honouring
					// either here would record a change in state that the next refresh undoes
					// and never converge. Acceptance testing confirmed this for the
					// description; for the location only some PVWA versions accept a change at
					// all, so moving a group across the Vault hierarchy is not something the
					// provider can promise either.
					ImmutableAttributes: []string{
						"description",
						"location",
					},
					// The Vault treats group names and locations case-insensitively and echoes
					// back its own spelling, so without this a config that differs from the
					// stored value only in case would look like a change on the refresh after
					// create: a pointless update for the name, and a failed plan for the
					// immutable location. The same coercion is why the SDK matches group names
					// with strings.EqualFold.
					SemanticEqualityAttributes: map[string]schemas.SemanticEqualityKind{
						"group_name": schemas.SemanticEqualityCaseInsensitive,
						"location":   schemas.SemanticEqualityCaseInsensitive,
					},
					StateSchema: &usergroupsmodels.IdsecPCloudUserGroup{},
					PageNotes: []tfactions.DocNote{
						userGroupDeprecationNote,
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
				DeprecationMessage: "idsec_pcloud_user_group is deprecated and will be removed in a future release. " +
					"Vault-native user groups are no longer the recommended way to group principals; use Identity roles instead.",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "pcloud-user-group-member",
						ActionDescription: "Privilege Cloud user group member resource, manages the membership of a Vault user in a Vault user group.",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
					},
					// The Vault assigns the member its numeric ID and reports it in the add
					// response; it is not an input to either membership call.
					ComputedAttributes: []string{
						"member_id",
					},
					// A membership has no updatable part: all three configurable attributes
					// identify it, so a change means removing the membership and adding it back.
					ImmutableAttributes: []string{
						"group_id",
						"member_name",
						"member_type",
					},
					// The Vault matches user and group names case-insensitively, so a config that
					// respells the member name is the same membership and must not recreate it.
					SemanticEqualityAttributes: map[string]schemas.SemanticEqualityKind{
						"member_name": schemas.SemanticEqualityCaseInsensitive,
					},
					StateSchema: &usergroupsmodels.IdsecPCloudUserGroupMember{},
					PageNotes: []tfactions.DocNote{
						userGroupDeprecationNote,
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
				DeprecationMessage: "idsec_pcloud_user_group_member is deprecated and will be removed in a future release. " +
					"Vault-native user groups are no longer the recommended way to group principals; use Identity roles instead.",
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "pcloud-user-group",
						ActionDescription: "Privilege Cloud user group data source, reads Vault user group information based on the group name, optionally narrowed by the group ID.",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
					},
					// The lookup resolves through the filtered list endpoint, which searches by
					// name, so the name is required and the ID only disambiguates the match.
					ExtraRequiredAttributes: []string{"group_name"},
					StateSchema:             &usergroupsmodels.IdsecPCloudUserGroup{},
					PageNotes: []tfactions.DocNote{
						{
							Severity: tfactions.DocNoteWarning,
							Body: "This data source is deprecated and will be removed in a future release. " +
								"Vault-native user groups are no longer the recommended way to group principals.",
						},
					},
				},
				DataSourceAction: "get",
			},
		},
	})
}
