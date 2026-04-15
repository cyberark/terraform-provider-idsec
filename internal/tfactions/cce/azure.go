// Copyright (c) CyberArk
// SPDX-License-Identifier: Apache-2.0

package cce

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/cce/azure/actions"
	azuremodels "github.com/cyberark/idsec-sdk-golang/pkg/services/cce/azure/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "cce-azure",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "cce-azure-entra", ActionDescription: "CCE Microsoft Entra tenant resource, manages Microsoft Entra tenant manual onboarding.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{},
					StateSchema:             &azuremodels.TfIdsecCCEAzureEntra{},
				},
				RawStateInference:   true,
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "tf-add-entra", tfactions.ReadOperation: "tf-entra", tfactions.UpdateOperation: "tf-update-entra", tfactions.DeleteOperation: "tf-delete-entra"},
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "cce-azure-management-group", ActionDescription: "CCE Azure management group resource, manages Azure management group manual onboarding.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{},
					StateSchema:             &azuremodels.TfIdsecCCEAzureManagementGroup{},
				},
				RawStateInference:   true,
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "tf-add-management-group", tfactions.ReadOperation: "tf-management-group", tfactions.UpdateOperation: "tf-update-management-group", tfactions.DeleteOperation: "tf-delete-management-group"},
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "cce-azure-subscription", ActionDescription: "CCE Azure subscription resource, manages Azure subscription manual onboarding.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{},
					StateSchema:             &azuremodels.TfIdsecCCEAzureSubscription{},
				},
				RawStateInference:   true,
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "tf-add-subscription", tfactions.ReadOperation: "tf-subscription", tfactions.UpdateOperation: "tf-update-subscription", tfactions.DeleteOperation: "tf-delete-subscription"},
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "cce-azure-entra", ActionDescription: "CCE Microsoft Entra tenant data source, reads Microsoft Entra tenant details based on the CCE Microsoft Entra tenant onboarding ID.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{},
					StateSchema:             &azuremodels.TfIdsecCCEAzureEntra{},
				},
				DataSourceAction: "tf-entra",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "cce-azure-management-group", ActionDescription: "CCE Azure management group data source, reads management group details based on the CCE management group onboarding ID.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{},
					StateSchema:             &azuremodels.TfIdsecCCEAzureManagementGroup{},
				},
				DataSourceAction: "tf-management-group",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "cce-azure-subscription", ActionDescription: "CCE Azure Subscription data source, reads subscription details based on the CCE subscription onboarding ID.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{},
					StateSchema:             &azuremodels.TfIdsecCCEAzureSubscription{},
				},
				DataSourceAction: "tf-subscription",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "cce-azure-workspaces", ActionDescription: "CCE Azure workspaces data source, retrieves Azure workspaces with optional filtering.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{},
					StateSchema:             &azuremodels.TfIdsecCCEAzureWorkspaces{},
				},
				DataSourceAction: "tf-workspaces",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "cce-azure-identity-params", ActionDescription: "CCE Azure Identity Params data source, retrieves Azure identity federation parameters for active services.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{},
					StateSchema:             &azuremodels.TfIdsecCCEAzureIdentityParams{},
				},
				DataSourceAction: "tf-identity-params",
			},
		},
	})
}
