// Copyright (c) CyberArk
// SPDX-License-Identifier: Apache-2.0

package cmgr

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/cmgr/networks/actions"
	networksmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/cmgr/networks/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "cmgr-networks",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "cmgr-network",
						ActionDescription: "The Connector Management service network resource that is used to manage networks associated with pools.",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
					},
					StateSchema: &networksmodels.IdsecCmgrNetwork{},
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
				ImportID: "network_id",
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "cmgr-network",
						ActionDescription: "The Connector Management service network data source. It reads the network information and metadata and is based on the ID of the network.",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{"network_id"},
					StateSchema:             &networksmodels.IdsecCmgrNetwork{},
				},
				DataSourceAction: "get",
			},
		},
	})
}
