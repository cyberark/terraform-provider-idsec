// Copyright (c) CyberArk
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/identity/policies/actions"
	policiesmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/identity/policies/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "identity-policies",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "identity-policy", ActionDescription: "The Identity service policy resource that is used to manage policies.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema:             &policiesmodels.IdsecIdentityPolicy{},
					ComputedAsSetAttributes: []string{"role_names"},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "create", tfactions.ReadOperation: "get", tfactions.UpdateOperation: "update", tfactions.DeleteOperation: "delete"},
				ImportID:            "policy_name",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "identity-policies-order", ActionDescription: "The Identity service policies order resource that is used to manage the order of policies.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &policiesmodels.IdsecIdentityPoliciesOrder{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "set-order", tfactions.ReadOperation: "get-order", tfactions.UpdateOperation: "set-order"},
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "identity-policy", ActionDescription: "The Identity service policy data source. It reads the policy information and metadata and is based on the ID of the policy.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema:             &policiesmodels.IdsecIdentityPolicy{},
					ComputedAsSetAttributes: []string{"role_names"},
				},
				DataSourceAction: "get",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "identity-policies-order", ActionDescription: "The Identity service policies order data source. It reads the order of policies and is based on the ID of the policy order configuration.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &policiesmodels.IdsecIdentityPoliciesOrder{},
				},
				DataSourceAction: "get-order",
			},
		},
	})
}
