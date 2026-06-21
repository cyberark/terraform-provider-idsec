// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/policy/cloudaccess/actions"
	cloudaccessmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/policy/cloudaccess/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "policy-cloud-access",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "policy-cloud-access", ActionDescription: "Cloud Access Policy resource.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema:             &cloudaccessmodels.IdsecPolicyCloudAccessCloudConsoleAccessPolicy{},
					ComputedAsSetAttributes: []string{"days_of_the_week", "aws_account_targets", "aws_organization_targets", "azure_targets", "gcp_targets"},
					HistoryComputedAttributes: []string{
						"invalid_resources",
						"metadata.created_by",
						"metadata.updated_on",
						"metadata.status.link",
						"metadata.status.status_code",
						"metadata.status.status_description",
						"targets.azure_targets.role_name",
						"targets.azure_targets.role_type",
						"targets.azure_targets.workspace_name",
						"targets.aws_account_targets.workspace_name",
						"targets.aws_account_targets.role_name",
						"targets.aws_organization_targets.role_name",
						"targets.aws_organization_targets.workspace_name",
						"targets.gcp_targets.role_name",
						"targets.gcp_targets.workspace_name",
						"targets.gcp_targets.role_type",
						"targets.gcp_targets.role_package",
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
						ActionName: "policy-cloud-access", ActionDescription: "Cloud Access Policy data source.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema:             &cloudaccessmodels.IdsecPolicyCloudAccessCloudConsoleAccessPolicy{},
					ComputedAsSetAttributes: []string{"days_of_the_week", "aws_account_targets", "aws_organization_targets", "azure_targets", "gcp_targets"},
				},
				DataSourceAction: "policy",
			},
		},
	})
}
