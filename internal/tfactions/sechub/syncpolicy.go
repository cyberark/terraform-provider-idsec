// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

// Package sechub registers Terraform actions for IDIRA services,
// including secret stores and sync policies, with the provider action registry.
package sechub

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/sechub/syncpolicies/actions"
	syncpoliciesmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/sechub/syncpolicies/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "sechub-syncpolicies",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "sechub-sync-policy",
						ActionDescription: "Manage Sync Policy resource",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{},
					ComputedAttributes: []string{

						"created_at",
						"created_by",
						"creation_details",
						"state",
						"status",
						"updated_at",
						"updated_by",

						"filter.id",

						"source.behaviors",
						"source.created_at",
						"source.created_by",
						"source.creation_details",
						"source.data",
						"source.data.url",
						"source.data.user_name",
						"source.description",
						"source.name",
						"source.state",
						"source.type",
						"source.updated_at",
						"source.updated_by",

						"state.current",
						"state.state_details",
						"state.state_details.from_state",
						"state.state_details.status",
						"state.state_details.to_state",

						"status.id",
						"status.is_running",
						"status.last_run",
						"status.last_success_time",
						"status.policy_status",

						"target.behaviors",
						"target.created_at",
						"target.created_by",
						"target.creation_details",
						"target.data",
						"target.data.account_alias",
						"target.data.account_id",
						"target.data.authentication_method",
						"target.data.connection_config.connection_type",
						"target.data.region_id",
						"target.data.role_name",
						"target.description",
						"target.name",
						"target.state",
						"target.type",
						"target.updated_at",
						"target.updated_by",

						"transformation.id",
					},
					// The only thing that can be changed on a sync policy via the API is the state
					// (enabled/disabled) and not via Terraform, so all other attributes are immutable.
					// ImmutableAttributes are matched by leaf field name at every nesting level.
					ImmutableAttributes: []string{
						"behaviors",
						"created_at",
						"created_by",
						"creation_details",
						"current",
						"data",
						"description",
						"from_state",
						"id",
						"is_running",
						"last_run",
						"last_success_time",
						"name",
						"policy_status",
						"predefined",
						"safe_name",
						"state",
						"state_details",
						"status",
						"to_state",
						"type",
						"updated_at",
						"updated_by",
					},
					StateSchema: &syncpoliciesmodels.IdsecSecHubPolicy{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{
					tfactions.CreateOperation,
					tfactions.ReadOperation,
					tfactions.DeleteOperation,
					tfactions.UpdateOperation,
					tfactions.StateOperation,
				},
				ActionsMappings: map[tfactions.IdsecServiceActionOperation]string{
					tfactions.CreateOperation: "create",
					tfactions.ReadOperation:   "get",
					tfactions.DeleteOperation: "delete",
					tfactions.UpdateOperation: "update",
				},
				ImportID: "id",
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "sechub-sync-policy",
						ActionDescription: "Manage Sync Policy resource",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{
						"id",
					},
					StateSchema: &syncpoliciesmodels.IdsecSecHubPolicy{},
				},
				DataSourceAction: "get",
			},
		},
	})
}
