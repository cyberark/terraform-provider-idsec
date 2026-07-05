// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package pamsh

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/pamsh/pamshsafes/actions"
	safesmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/pamsh/pamshsafes/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "pamsh-safes",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "pamsh-safe", ActionDescription: "Privilege Cloud Safe resource, manages Privilege Cloud Safes information and metadata.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap, Enabled: boolPtr(false),
					},
					ComputedAttributes: []string{
						"safe_number",
						"creator",
						"creation_time",
						"last_modification_time",
						"safe_id",
						"is_expired_member",
					},
					ImmutableAttributes: []string{
						"auto_purge_enabled",
					},
					StateSchema: &safesmodels.IdsecPamshSafe{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "create", tfactions.ReadOperation: "get", tfactions.UpdateOperation: "update", tfactions.DeleteOperation: "delete"},
				ImportID:            "safe_id",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "pamsh-safe-member", ActionDescription: "Privilege Cloud safe member resource, manages Privilege Cloud Safe members and their relevant permissions.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap, Enabled: boolPtr(false),
					},
					ComputedAttributes: []string{
						"member_id",
						"safe_number",
						"safe_name",
						"is_expired_membership_enabled",
						"is_predefined_user",
						"is_read_only",
					},
					ImmutableAttributes: []string{
						"search_in",
						"member_name",
						"member_type",
						"safe_id",
					},
					StateSchema: &safesmodels.IdsecPamshSafeMember{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "add-member", tfactions.ReadOperation: "get-member", tfactions.UpdateOperation: "update-member", tfactions.DeleteOperation: "delete-member"},
				ImportID:            "safe_id:member_name",
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "pamsh-safe", ActionDescription: "Privilege Cloud Safe data source, reads safe information and metadata, based on the Safe ID.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap, Enabled: boolPtr(false),
					},
					ExtraRequiredAttributes: []string{"safe_id"},
					StateSchema:             &safesmodels.IdsecPamshSafe{},
				},
				DataSourceAction: "get",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "pamsh-safe-member", ActionDescription: "Privilege Cloud Safe Member data source, reads Safe member information and metadata, based on the Safe ID and the member name.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap, Enabled: boolPtr(false),
					},
					ExtraRequiredAttributes: []string{"safe_id", "member_name"},
					StateSchema:             &safesmodels.IdsecPamshSafeMember{},
				},
				DataSourceAction: "get-member",
			},
		},
	})
}
