// Copyright (c) CyberArk
// SPDX-License-Identifier: Apache-2.0

package sia

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/sia/workspacestargetsets/actions"
	targetsetsmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/sia/workspacestargetsets/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "sia-workspaces-target-sets",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-workspaces-target-set", ActionDescription: "The SIA workspaces target set resource, manages target set information about one or more targets and how they are represented, along with the association to the relevant Secret.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &targetsetsmodels.IdsecSIATargetSet{},
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
						ActionName: "sia-workspaces-target-set", ActionDescription: "The SIA workspaces target set data source, reads target set information and metadata, based on the ID of the target set.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema:             &targetsetsmodels.IdsecSIATargetSet{},
					ExtraRequiredAttributes: []string{"id"},
				},
				DataSourceAction: "get",
			},
		},
	})
}
