// Copyright (c) CyberArk
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/identity/users/actions"
	usersmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/identity/users/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "identity-users",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "identity-user", ActionDescription: "The Identity service user resource that is used to manage users.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					SensitiveAttributes: []string{"password"},
					ComputedAttributes:  []string{"user_attributes"},
					StateSchema:         &usersmodels.IdsecIdentityUser{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "create", tfactions.ReadOperation: "get", tfactions.UpdateOperation: "update", tfactions.DeleteOperation: "delete"},
				ImportID:            "user_id",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "identity-user-attributes-schema", ActionDescription: "The Identity service user attributes schema resource that is used to manage user attributes schema.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &usersmodels.IdsecIdentityUserAttributesSchema{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "upsert-attributes-schema", tfactions.ReadOperation: "attributes-schema", tfactions.UpdateOperation: "upsert-attributes-schema", tfactions.DeleteOperation: "delete-attributes-schema"},
				ImportID:            tfactions.SingletonResourceImportDummyID,
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "identity-user-attributes", ActionDescription: "The Identity service user attributes resource that is used to manage user attributes.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &usersmodels.IdsecIdentityUserAttributes{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "upsert-attributes", tfactions.ReadOperation: "get-attributes", tfactions.UpdateOperation: "upsert-attributes", tfactions.DeleteOperation: "delete-attributes"},
				ImportID:            "user_id",
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "identity-user", ActionDescription: "The Identity service user data source. It reads the user information and metadata and is based on the ID of the user.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &usersmodels.IdsecIdentityUser{},
				},
				DataSourceAction: "get",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "identity-user-attributes-schema", ActionDescription: "The Identity service user attributes schema data source. It reads the user attributes schema information and metadata and is based on the ID of the user attributes schema.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &usersmodels.IdsecIdentityUserAttributesSchema{},
				},
				DataSourceAction: "attributes-schema",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "identity-user-attributes", ActionDescription: "The Identity service user attributes data source. It reads the user attributes information and metadata and is based on the ID of the user attributes.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &usersmodels.IdsecIdentityUserAttributes{},
				},
				DataSourceAction: "get-attributes",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "identity-user-info", ActionDescription: "The Identity service user info data source. It reads the user info information and metadata.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &usersmodels.IdsecIdentityUserInfo{},
				},
				DataSourceAction: "info",
			},
		},
	})
}
