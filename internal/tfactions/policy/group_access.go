// Copyright (c) CyberArk
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/policy/groupaccess/actions"
	groupaccessmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/policy/groupaccess/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "policy-group-access",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "policy-group-access", ActionDescription: "Group Access Policy resource.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &groupaccessmodels.IdsecPolicyGroupAccessPolicy{},
				},
				ReadSchemaPath:      "metadata",
				DeleteSchemaPath:    "metadata",
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "create-policy", tfactions.ReadOperation: "policy", tfactions.UpdateOperation: "update-policy", tfactions.DeleteOperation: "delete-policy"},
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "policy-group-access", ActionDescription: "Group Access Policy data source.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &groupaccessmodels.IdsecPolicyGroupAccessPolicy{},
				},
				DataSourceAction: "policy",
			},
		},
	})
}
