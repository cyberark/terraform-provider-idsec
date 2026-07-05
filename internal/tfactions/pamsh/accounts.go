// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package pamsh

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/pamsh/pamshaccounts/actions"
	accountsmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/pamsh/pamshaccounts/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "pamsh-accounts",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "pamsh-account", ActionDescription: "Manage Privilege Cloud account information, metadata, and credentials", ActionVersion: 1, Schemas: actions.ActionToSchemaMap, Enabled: boolPtr(false),
					},
					ExtraRequiredAttributes: []string{},
					ComputedAttributes: []string{
						"id",
						"status",
						"created_time",
						"category_modification_time",
						"secret_management.last_modified_time",
					},
					ImmutableAttributes: []string{
						"safe_name",
						"secret_type",
					},
					SensitiveAttributes: []string{"secret"},
					StateSchema:         &accountsmodels.IdsecPamshAccount{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "create", tfactions.ReadOperation: "get", tfactions.UpdateOperation: "update", tfactions.DeleteOperation: "delete"},
				ImportID:            "id",
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "pamsh-account", ActionDescription: "Privilege Cloud account data source, reads account information and metadata, based on the account ID.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap, Enabled: boolPtr(false),
					},
					ExtraRequiredAttributes: []string{"id"},
					StateSchema:             &accountsmodels.IdsecPamshAccount{},
				},
				DataSourceAction: "get",
			},
		},
	})
}

func boolPtr(b bool) *bool {
	return &b
}
