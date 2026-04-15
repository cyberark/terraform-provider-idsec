// Copyright (c) CyberArk
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/policy/db/actions"
	policydbmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/policy/db/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "policy-db",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "policy-db", ActionDescription: "The infrastructure database policy resource.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema:             &policydbmodels.IdsecPolicyDBAccessPolicy{},
					ComputedAsSetAttributes: []string{"days_of_the_week"},
				},
				ReadSchemaPath:      "metadata",
				DeleteSchemaPath:    "metadata",
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "create-policy", tfactions.ReadOperation: "policy", tfactions.UpdateOperation: "update-policy", tfactions.DeleteOperation: "delete-policy"},
				ImportID:            "policy_id",
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "policy-db", ActionDescription: "The infrastructure database policy data source.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &policydbmodels.IdsecPolicyDBAccessPolicy{},
				},
				DataSourceAction: "policy",
			},
		},
	})
}
