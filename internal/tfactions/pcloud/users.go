// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package pcloud

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/pcloud/users/actions"
	usersmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/pcloud/users/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "pcloud-users",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "pcloud-user",
						ActionDescription: "Privilege Cloud User resource, manages AppProvider Vault users.",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
					},
					ComputedAttributes: []string{
						"user_id",
					},
					ImmutableAttributes: []string{
						"username",
						// PVWA may coerce the requested user type (e.g. "AppProvider" is
						// reported back as "EPVUser"), so config must use the value the
						// API returns or the next plan fails this immutability check.
						"user_type",
						"location",
					},
					SensitiveAttributes: []string{"initial_password"},
					StateSchema:         &usersmodels.IdsecPCloudUser{},
					PageNotes: []tfactions.DocNote{
						{
							Severity: tfactions.DocNoteWarning,
							Body: "This resource is deprecated and will be removed in a future release. " +
								"It is no longer the recommended way to manage Vault-native users.",
						},
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
					tfactions.CreateOperation: "create",
					tfactions.ReadOperation:   "get",
					tfactions.UpdateOperation: "update",
					tfactions.DeleteOperation: "delete",
				},
				ImportID:                  "user_id",
				WriteOnlyHashedAttributes: []string{"initial_password"},
				DeprecationMessage:        "idsec_pcloud_user is deprecated and will be removed in a future release. This resource is no longer the recommended way to manage Vault-native users.",
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "pcloud-user",
						ActionDescription: "Privilege Cloud User data source, reads Vault user information based on the user ID.",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{"user_id"},
					StateSchema:             &usersmodels.IdsecPCloudUser{},
				},
				DataSourceAction: "get",
			},
		},
	})
}
