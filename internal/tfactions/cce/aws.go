// Copyright (c) CyberArk
// SPDX-License-Identifier: Apache-2.0

package cce

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/cce/aws/actions"
	awsmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/cce/aws/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "cce-aws",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "cce-aws-organization", ActionDescription: "CCE AWS organization resource, manages AWS organization programmatic onboarding.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{"organization_root_id", "management_account_id", "organization_id", "services", "scan_organization_role_arn", "cross_account_role_external_id"},
					StateSchema:             &awsmodels.TfIdsecCCEAWSOrganization{},
				},
				RawStateInference:   true,
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "tf-add-organization", tfactions.ReadOperation: "tf-organization", tfactions.UpdateOperation: "tf-update-organization", tfactions.DeleteOperation: "tf-delete-organization"},
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "cce-aws-account", ActionDescription: "CCE AWS account resource, manages AWS account programmatic onboarding.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{"account_id", "services"},
					StateSchema:             &awsmodels.TfIdsecCCEAWSAccount{},
				},
				RawStateInference:   true,
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "tf-add-account", tfactions.ReadOperation: "tf-account", tfactions.UpdateOperation: "tf-update-account", tfactions.DeleteOperation: "tf-delete-account"},
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "cce-aws-organization-account", ActionDescription: "CCE AWS organization account resource, adds AWS accounts to an organization.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{"parent_organization_id", "account_id", "services"},
					StateSchema:             &awsmodels.TfIdsecCCEAWSAccount{},
				},
				RawStateInference:   true,
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "tf-add-organization-account-sync", tfactions.ReadOperation: "tf-account", tfactions.UpdateOperation: "tf-update-organization-account"},
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "cce-aws-organization", ActionDescription: "CCE AWS organization datasource, reads organization details including added services based on the organization's management account ID.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{"id"},
					StateSchema:             &awsmodels.TfIdsecCCEAWSOrganizationDatasource{},
				},
				DataSourceAction: "tf-organization-datasource",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "cce-aws-workspaces", ActionDescription: "CCE AWS workspaces data source, retrieves AWS organizations and accounts with filtering.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{},
					StateSchema:             &awsmodels.TfIdsecCCEAWSWorkspaces{},
				},
				DataSourceAction: "tf-workspaces",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "cce-aws-account", ActionDescription: "CCE AWS account data source, reads account details based on the CCE account onboarding ID.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{"id"},
					StateSchema:             &awsmodels.TfIdsecCCEAWSAccount{},
				},
				DataSourceAction: "tf-account",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "cce-aws-tenant-service-details", ActionDescription: "CCE AWS tenant service details data source, retrieves tenant service details.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{},
					StateSchema:             &awsmodels.TfIdsecCCEAWSTenantServiceDetails{},
				},
				DataSourceAction: "tf-tenant-service-details",
			},
		},
	})
}
