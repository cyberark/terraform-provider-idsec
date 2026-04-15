// Copyright (c) CyberArk
// SPDX-License-Identifier: Apache-2.0

package cmgr

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/cmgr/pools/actions"
	poolsmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/cmgr/pools/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "cmgr-pools",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "cmgr-pool",
						ActionDescription: "The Connector Management service pool resource that manages the pool of Secure Infrastructure Access (SIA) and system connectors.",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{"assigned_network_ids"},
					StateSchema:             &poolsmodels.IdsecCmgrPool{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{
					tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation,
				},
				ActionsMappings: map[tfactions.IdsecServiceActionOperation]string{
					tfactions.CreateOperation: "create", tfactions.ReadOperation: "get", tfactions.UpdateOperation: "update", tfactions.DeleteOperation: "delete",
				},
				ImportID: "pool_id",
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "cmgr-pool",
						ActionDescription: "The Connector Management service pool data source. It reads the pool information and metadata and is based on the ID of the pool.",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{"pool_id"},
					StateSchema:             &poolsmodels.IdsecCmgrPool{},
				},
				DataSourceAction: "get",
			},
		},
	})
}
