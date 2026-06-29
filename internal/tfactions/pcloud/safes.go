// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package pcloud

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/pcloud/safes/actions"
	safesmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/pcloud/safes/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "pcloud-safes",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "pcloud-safe", ActionDescription: "Privilege Cloud Safe resource, manages Privilege Cloud Safes information and metadata.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ComputedAttributes: []string{
						"safe_number",
						"creator",
						"creation_time",
						"last_modification_time",
						"is_expired_member",
					},
					ImmutableAttributes: []string{
						"auto_purge_enabled",
					},
					StateSchema: &safesmodels.IdsecPCloudSafe{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "create", tfactions.ReadOperation: "get", tfactions.UpdateOperation: "update", tfactions.DeleteOperation: "delete"},
				ImportID:            "safe_id",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "pcloud-safe-member", ActionDescription: "Privilege Cloud safe member resource, manages Privilege Cloud Safe members and their relevant permissions.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ComputedAttributes: []string{
						"member_id",
						"safe_number",
						"is_expired_membership_enabled",
						"is_predefined_user",
						"is_read_only",
					},
					ImmutableAttributes: []string{
						"search_in",
						"member_name",
						"member_type",
					},
					StateSchema: &safesmodels.IdsecPCloudSafeMember{},
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
						ActionName: "pcloud-safe", ActionDescription: "Privilege Cloud Safe data source, reads safe information and metadata, based on the Safe ID.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{"safe_id"},
					StateSchema:             &safesmodels.IdsecPCloudSafe{},
				},
				DataSourceAction: "get",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "pcloud-safe-member", ActionDescription: "Privilege Cloud Safe Member data source, reads Safe member information and metadata, based on the Safe ID and the member name.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{"safe_id", "member_name"},
					StateSchema:             &safesmodels.IdsecPCloudSafeMember{},
				},
				DataSourceAction: "get-member",
			},
		},
	})
}
