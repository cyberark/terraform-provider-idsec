// Copyright (c) CyberArk
// SPDX-License-Identifier: Apache-2.0

package sia

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/sia/dbstrongaccounts/actions"
	dbstrongaccountsmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/sia/dbstrongaccounts/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "sia-db-strong-accounts",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-db-strong-accounts", ActionDescription: "The SIA strong accounts resource, manages strong account information and metadata.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					SensitiveAttributes: []string{"password", "secret_access_key"},
					StateSchema:         &dbstrongaccountsmodels.IdsecSIADBStrongAccount{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "create", tfactions.ReadOperation: "get", tfactions.UpdateOperation: "update", tfactions.DeleteOperation: "delete"},
				ImportID:            "strong_account_id",
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-db-strong-accounts", ActionDescription: "The SIA strong accounts data source, reads strong account information and metadata, based on the ID of the account.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					SensitiveAttributes: []string{"password", "secret_access_key"},
					StateSchema:         &dbstrongaccountsmodels.IdsecSIADBStrongAccount{},
				},
				DataSourceAction: "get",
			},
		},
	})
}
