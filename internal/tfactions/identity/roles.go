// Copyright (c) CyberArk
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
					StateSchema: &rolesmodels.IdsecIdentityRoleMember{},
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
		},
	})
}
