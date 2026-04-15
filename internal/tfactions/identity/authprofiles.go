// Copyright (c) CyberArk
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/identity/authprofiles/actions"
	authprofilesmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/identity/authprofiles/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "identity-auth-profiles",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "identity-auth-profile", ActionDescription: "The Identity service auth profile resource that is used to manage auth profiles.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &authprofilesmodels.IdsecIdentityAuthProfile{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "create", tfactions.ReadOperation: "get", tfactions.UpdateOperation: "update", tfactions.DeleteOperation: "delete"},
				ImportID:            "auth_profile_id",
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "identity-auth-profile", ActionDescription: "The Identity service auth profile data source. It reads the auth profile information and metadata and is based on the ID of the auth profile.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &authprofilesmodels.IdsecIdentityAuthProfile{},
				},
				DataSourceAction: "get",
			},
		},
	})
}
