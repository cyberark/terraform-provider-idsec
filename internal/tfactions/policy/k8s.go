// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/policy/k8s/actions"
	policyk8smodels "github.com/cyberark/idsec-sdk-golang/pkg/services/policy/k8s/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "policy-k8s",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "policy-k8s", ActionDescription: "Kubernetes cluster access policy resource.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema:             &policyk8smodels.IdsecPolicyK8sPolicy{},
					ComputedAsSetAttributes: []string{"days_of_the_week", "aws_account_targets", "aws_idc_targets", "azure_targets"},
					ComputedAttributes: []string{
						"metadata.created_by",
						"metadata.updated_on",
					},
					HistoryComputedAttributes: []string{
						"invalid_resources",
						"metadata.created_by",
						"metadata.updated_on",
						"metadata.status.link",
						"metadata.status.status_code",
						"metadata.status.status_description",
						"targets.azure_targets.role_name",
						"targets.azure_targets.workspace_name",
						"targets.azure_targets.cluster_name",
						"targets.azure_targets.namespace_name",
						"targets.azure_targets.fqdn",
						"targets.azure_targets.region",
						"targets.aws_account_targets.role_name",
						"targets.aws_account_targets.workspace_name",
						"targets.aws_account_targets.cluster_name",
						"targets.aws_account_targets.namespace_name",
						"targets.aws_account_targets.fqdn",
						"targets.aws_account_targets.region",
						"targets.aws_idc_targets.role_name",
						"targets.aws_idc_targets.workspace_name",
						"targets.aws_idc_targets.cluster_name",
						"targets.aws_idc_targets.namespace_name",
						"targets.aws_idc_targets.fqdn",
						"targets.aws_idc_targets.region",
					},
				},
				ReadSchemaPath:      "metadata",
				DeleteSchemaPath:    "metadata",
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "create-policy", tfactions.ReadOperation: "policy", tfactions.UpdateOperation: "update-policy", tfactions.DeleteOperation: "delete-policy"},
				ImportID:            "metadata.policy_id",
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "policy-k8s", ActionDescription: "Kubernetes cluster access policy data source.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema:             &policyk8smodels.IdsecPolicyK8sPolicy{},
					ComputedAsSetAttributes: []string{"days_of_the_week", "aws_account_targets", "aws_idc_targets", "azure_targets"},
				},
				DataSourceAction: "policy",
			},
		},
	})
}
