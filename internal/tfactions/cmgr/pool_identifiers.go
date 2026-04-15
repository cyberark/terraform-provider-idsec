// Copyright (c) CyberArk
// SPDX-License-Identifier: Apache-2.0

package cmgr

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/cmgr/poolidentifiers/actions"
	identifiersmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/cmgr/poolidentifiers/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "cmgr-pool-identifiers",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "cmgr-pool-identifier",
						ActionDescription: "The Connector Management service pool identifier resource that is associated with a pool and is used to identify the pool in a simplified manner. It is not identified using only the network name",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{"pool_id", "type", "value"},
					StateSchema:             &identifiersmodels.IdsecCmgrPoolIdentifier{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{
					tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation,
				},
				ActionsMappings: map[tfactions.IdsecServiceActionOperation]string{
					tfactions.CreateOperation: "create", tfactions.ReadOperation: "get", tfactions.UpdateOperation: "update", tfactions.DeleteOperation: "delete",
				},
				ImportID: "pool_id:identifier_id",
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "cmgr-pool-identifier",
						ActionDescription: "The Connector Management service pool data source. It reads the pool information and metadata and is based on the ID of the pool.",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{"pool_id", "identifier_id"},
					StateSchema:             &identifiersmodels.IdsecCmgrPoolIdentifier{},
				},
				DataSourceAction: "get",
			},
		},
	})
}
