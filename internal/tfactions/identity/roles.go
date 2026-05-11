// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/identity/roles/actions"
	rolesmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/identity/roles/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "identity-roles",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "identity-role", ActionDescription: "The Identity service role resource that is used to manage roles.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema:             &rolesmodels.IdsecIdentityRole{},
					ComputedAsSetAttributes: []string{"admin_rights"},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "create", tfactions.ReadOperation: "get", tfactions.UpdateOperation: "update", tfactions.DeleteOperation: "delete"},
				ImportID:            "role_id",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "identity-role-member", ActionDescription: "The Identity service role member resource that is used to manage role members.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema:               &rolesmodels.IdsecIdentityRoleMember{},
					CaseInsensitiveAttributes: []string{"member_name"},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "add-member", tfactions.ReadOperation: "get-member", tfactions.DeleteOperation: "remove-member"},
				ImportID:            "role_id:member_id",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "identity-role-admin-rights", ActionDescription: "The Identity service role admin rights resource that is used to manage role admin rights.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema:             &rolesmodels.IdsecIdentityRoleAdminRights{},
					ComputedAsSetAttributes: []string{"admin_rights"},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "add-admin-rights", tfactions.ReadOperation: "get-admin-rights", tfactions.DeleteOperation: "remove-admin-rights"},
				ImportID:            "role_id",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "identity-role-attributes-schema", ActionDescription: "The Identity service role attributes schema resource that is used to manage role attributes schema.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &rolesmodels.IdsecIdentityRoleAttributesSchema{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "create-attributes-schema", tfactions.ReadOperation: "attributes-schema", tfactions.UpdateOperation: "update-attributes-schema", tfactions.DeleteOperation: "delete-attributes-schema"},
				ImportID:            tfactions.SingletonResourceImportDummyID,
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "identity-role-attributes", ActionDescription: "The Identity service role attributes resource that is used to manage role attributes.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &rolesmodels.IdsecIdentityRoleAttributes{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "upsert-attributes", tfactions.ReadOperation: "get-attributes", tfactions.UpdateOperation: "upsert-attributes", tfactions.DeleteOperation: "delete-attributes"},
				ImportID:            "role_id",
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "identity-role", ActionDescription: "The Identity service role data source. It reads the role information and metadata and is based on the ID of the role.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema:             &rolesmodels.IdsecIdentityRole{},
					ComputedAsSetAttributes: []string{"admin_rights"},
				},
				DataSourceAction: "get",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "identity-role-member", ActionDescription: "The Identity service role member data source. It reads the role member information and metadata and is based on the ID of the role member.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &rolesmodels.IdsecIdentityRoleMember{},
				},
				DataSourceAction: "get-member",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "identity-role-admin-rights", ActionDescription: "The Identity service role admin rights data source. It reads the role admin rights information and metadata and is based on the role name.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema:             &rolesmodels.IdsecIdentityRoleAdminRights{},
					ComputedAsSetAttributes: []string{"admin_rights"},
				},
				DataSourceAction: "get-admin-rights",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "identity-role-attributes-schema", ActionDescription: "The Identity service role attributes schema data source. It reads the role attributes schema information and metadata and is based on the ID of the role attributes schema.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &rolesmodels.IdsecIdentityRoleAttributesSchema{},
				},
				DataSourceAction: "attributes-schema",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "identity-role-attributes", ActionDescription: "The Identity service role attributes data source. It reads the role attributes information and metadata and is based on the ID of the role attributes.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &rolesmodels.IdsecIdentityRoleAttributes{},
				},
				DataSourceAction: "get-attributes",
			},
		},
	})
}
