// Copyright (c) CyberArk
// SPDX-License-Identifier: Apache-2.0

package sechub

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/sechub/secretstores/actions"
	secretstoresmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/sechub/secretstores/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "sechub-secretstores",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "sechub-secret-store",
						ActionDescription: "Manage Secrets Hub secret store resource that represent secret management systems, including their configuration and metadata",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
						Enabled:           boolPtr(false),
					},
					ExtraRequiredAttributes: []string{},
					ComputedAttributes: []string{
						"id",
						"created_at",
						"created_by",
						"updated_at",
						"updated_by",
						"creation_details",
						"organization_id",
						"engine_type",
						"engine_api_version",
					},
					ImmutableAttributes: []string{
						"id",
						"behaviors",
						"account_id",
						"region_id",
						"azure_vault_url",
					},
					SensitiveAttributes: []string{
						"password",
					},
					StateSchema: &secretstoresmodels.IdsecSecHubSecretStore{},
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
					tfactions.UpdateOperation: "update-tf",
					tfactions.DeleteOperation: "delete",
				},
				ImportID: "id",
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "sechub-secret-store",
						ActionDescription: "Secrets Hub secret store data source, reads secret store information and metadata, based on the Secret Store ID.",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
						Enabled:           boolPtr(false),
					},
					ExtraRequiredAttributes: []string{
						"id",
					},
					SensitiveAttributes: []string{
						"data.password",
					},
					StateSchema: &secretstoresmodels.IdsecSecHubSecretStore{},
				},
				DataSourceAction: "get",
			},
		},
	})
}

func boolPtr(b bool) *bool {
	return &b
}
