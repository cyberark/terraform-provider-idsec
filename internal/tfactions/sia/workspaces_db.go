// Copyright (c) CyberArk
// SPDX-License-Identifier: Apache-2.0

package sia

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/sia/workspacesdb/actions"
	workspacesdbmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/sia/workspacesdb/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "sia-workspaces-db",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-workspaces-db", ActionDescription: "The SIA workspaces database resource, manages database workspaces information and metadata, along with the association to the relevant Secret.", ActionVersion: 1, Schemas: actions.TargetActionToTargetSchemaMap,
					},
					StateSchema: &workspacesdbmodels.IdsecSIADBDatabaseTarget{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "create-target", tfactions.ReadOperation: "get-target", tfactions.UpdateOperation: "update-target", tfactions.DeleteOperation: "delete-target"},
				ImportID:            "id",
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-workspaces-db", ActionDescription: "The SIA workspaces database data source, reads database information and metadata, based on the ID of the database.", ActionVersion: 1, Schemas: actions.TargetActionToTargetSchemaMap,
					},
					StateSchema: &workspacesdbmodels.IdsecSIADBDatabaseTarget{},
				},
				DataSourceAction: "get-target",
			},
		},
	})
}
