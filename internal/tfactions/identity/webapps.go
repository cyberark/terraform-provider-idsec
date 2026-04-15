// Copyright (c) CyberArk
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/identity/webapps/actions"
	webappsmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/identity/webapps/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "identity-webapps",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "identity-webapp",
						ActionDescription: "The Identity service webapp resource that is used to manage webapps.",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
					},
					StateSchema: &webappsmodels.IdsecIdentityWebapp{},
					ComputedAttributes: []string{
						"generic",
						"webapp_type",
						"state",
						"app_type_display_name",
						"category",
					},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{
					tfactions.CreateOperation,
					tfactions.ReadOperation,
					tfactions.UpdateOperation,
					tfactions.DeleteOperation,
					tfactions.StateOperation,
				},
				ActionsMappings: map[tfactions.IdsecServiceActionOperation]string{
					tfactions.CreateOperation: "import",
					tfactions.ReadOperation:   "get",
					tfactions.UpdateOperation: "update",
					tfactions.DeleteOperation: "delete",
				},
				ImportID: "webapp_id",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "identity-webapp-permission",
						ActionDescription: "The Identity service webapp permission resource that is used to manage webapp permissions.",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
					},
					StateSchema: &webappsmodels.IdsecIdentityWebappPermission{},
					ComputedAsSetAttributes: []string{
						"rights",
					},
					ComputedAttributes: []string{
						"principal_id",
					},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{
					tfactions.CreateOperation,
					tfactions.ReadOperation,
					tfactions.UpdateOperation,
					tfactions.StateOperation,
				},
				ActionsMappings: map[tfactions.IdsecServiceActionOperation]string{
					tfactions.CreateOperation: "set-permission",
					tfactions.ReadOperation:   "get-permission",
					tfactions.UpdateOperation: "set-permission",
				},
				ImportID: "webapp_id:principal_id:principal_type",
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "identity-webapp",
						ActionDescription: "The Identity service webapp data source. It reads the webapp information and metadata and is based on the ID of the webapp or its name.",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
					},
					StateSchema: &webappsmodels.IdsecIdentityWebapp{},
				},
				DataSourceAction: "get",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "identity-webapp-permissions",
						ActionDescription: "The Identity service webapp permissions data source. It reads the webapp permissions information and metadata and is based on the ID of the webapp or its name.",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
					},
					StateSchema: &webappsmodels.IdsecIdentityWebappPermissions{},
					ComputedAsSetAttributes: []string{
						"grants",
					},
				},
				DataSourceAction: "get-permissions",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "identity-webapp-permission",
						ActionDescription: "The Identity service webapp permission data source. It reads the webapp permission information and metadata and is based on the ID of the webapp or its name.",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
					},
					StateSchema: &webappsmodels.IdsecIdentityWebappPermission{},
					ComputedAsSetAttributes: []string{
						"rights",
					},
				},
				DataSourceAction: "get-permission",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "identity-webapp-template",
						ActionDescription: "The Identity service webapp template data source. It reads the webapp template information and metadata and is based on the ID of the webapp template or its name.",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
					},
					StateSchema: &webappsmodels.IdsecIdentityWebappTemplate{},
				},
				DataSourceAction: "get-template",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "identity-webapp-custom-template",
						ActionDescription: "The Identity service webapp custom template data source. It reads the webapp custom template information and metadata and is based on the ID of the webapp custom template or its name.",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
					},
					StateSchema: &webappsmodels.IdsecIdentityWebappTemplate{},
				},
				DataSourceAction: "get-custom-template",
			},
		},
	})
}
