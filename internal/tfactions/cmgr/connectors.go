// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package cmgr

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/cmgr/connectors/actions"
	connectorsmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/cmgr/connectors/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "cmgr-connectors",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "cmgr-connector",
						ActionDescription: "Connector Management connector resource, manages connector management agent installation and removal on Connector Management and target machines.",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
					},
					StateSchema:         &connectorsmodels.IdsecCmgrConnectorID{},
					SensitiveAttributes: []string{"password", "private_key_contents"},
					ComputedAttributes:  []string{"connector_id"},
				},
				RawStateInference: true,
				SupportedOperations: []tfactions.IdsecServiceActionOperation{
					tfactions.CreateOperation,
					tfactions.UpdateOperation,
					tfactions.DeleteOperation,
					tfactions.ReadOperation,
					tfactions.StateOperation,
				},
				ActionsMappings: map[tfactions.IdsecServiceActionOperation]string{
					tfactions.CreateOperation: "install",
					tfactions.UpdateOperation: "update",
					tfactions.DeleteOperation: "uninstall",
					tfactions.ReadOperation:   "get",
				},
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "cmgr-connector",
						ActionDescription: "The Connector Management service connector data source. It reads the connector information and metadata based on the connector ID.",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{"connector_id"},
					StateSchema:             &connectorsmodels.IdsecCmgrConnector{},
				},
				DataSourceAction: "get",
			},
		},
	})
}
