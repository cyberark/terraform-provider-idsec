// Copyright (c) CyberArk
// SPDX-License-Identifier: Apache-2.0

package sia

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/sia/secretsvm/actions"
	secretsvmmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/sia/secretsvm/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "sia-secrets-vm",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-secrets-vm", ActionDescription: "The SIA Secrets VM resource, manages VM Secrets information and metadata, based on the type of Secret.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					SensitiveAttributes: []string{"provisioner_password", "secret_data"},
					StateSchema:         &secretsvmmodels.IdsecSIAVMSecret{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "create", tfactions.ReadOperation: "get", tfactions.UpdateOperation: "change", tfactions.DeleteOperation: "delete"},
				ImportID:            "secret_id",
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-secrets-vm", ActionDescription: "The SIA Secrets VM data source, reads VM Secrets information and metadata, based on the ID of the Secret.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &secretsvmmodels.IdsecSIAVMSecret{},
				},
				DataSourceAction: "get",
			},
		},
	})
}
