// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package sia

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/sia/k8s/actions"
	k8smodels "github.com/cyberark/idsec-sdk-golang/pkg/services/sia/k8s/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "sia-k8s",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-k8s-generate-cluster-config", ActionDescription: "The SIA K8s generate cluster config resource, generates a kubeconfig file for a specific set of K8s clusters and access reasons.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					// GenerateClusterConfig returns a dedicated result struct (just the generated
					// kubeconfig path), not the request model: StateSchema/ComputedAttributes only
					// need to describe that one computed field. The engine fills in every other
					// (input) attribute on the resource from the plan via its create/update merge.
					StateSchema:        &k8smodels.IdsecSIAK8SGenerateClusterConfigResult{},
					ComputedAttributes: []string{"kubeconfig_path"},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "generate-cluster-config"},
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-k8s-list-clusters", ActionDescription: "The SIA K8s list clusters data source, lists the K8s clusters available for the current user.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema:        &k8smodels.IdsecSIAK8SClusters{},
					ComputedAttributes: []string{"clusters"},
				},
				DataSourceAction: "list-clusters",
			},
		},
	})
}
