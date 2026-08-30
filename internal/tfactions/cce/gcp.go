// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package cce

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/cce/gcp/actions"
	gcpmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/cce/gcp/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func boolPtr(b bool) *bool { return &b }

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "cce-gcp",
		Resources:   []*tfactions.IdsecServiceTerraformResourceActionDefinition{},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						// TODO: GCP TF is still in development. Enabled flag is set to false to prevent
						// this data source from being published in new releases until the feature is complete.
						// Remove boolPtr(false) once the feature is ready.
						ActionName: "cce-gcp-workspaces", Enabled: boolPtr(false), ActionDescription: "CCE GCP workspaces data source, retrieves GCP workspaces with optional filtering.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{},
					StateSchema:             &gcpmodels.TfIdsecCCEGCPWorkspaces{},
				},
				DataSourceAction: "tf-workspaces",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						// TODO: GCP TF is still in development. Enabled flag is set to false to prevent
						// this data source from being published in new releases until the feature is complete.
						// Remove boolPtr(false) once the feature is ready.
						ActionName: "cce-gcp-identity-params", Enabled: boolPtr(false), ActionDescription: "CCE GCP Identity Params data source, retrieves GCP workload identity federation parameters for active services.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{},
					StateSchema:             &gcpmodels.TfIdsecCCEGCPIdentityParams{},
				},
				DataSourceAction: "tf-identity-params",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						// TODO: GCP TF is still in development. Enabled flag is set to false to prevent
						// this data source from being published in new releases until the feature is complete.
						// Remove boolPtr(false) once the feature is ready.
						ActionName: "cce-gcp-project", Enabled: boolPtr(false), ActionDescription: "CCE GCP Project data source, reads project details based on the CCE project onboarding ID.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{"id"},
					StateSchema:             &gcpmodels.TfIdsecCCEGCPProject{},
				},
				DataSourceAction: "tf-project",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						// TODO: GCP TF is still in development. Enabled flag is set to false to prevent
						// this data source from being published in new releases until the feature is complete.
						// Remove boolPtr(false) once the feature is ready.
						ActionName: "cce-gcp-organization", Enabled: boolPtr(false), ActionDescription: "CCE GCP organization data source, reads organization details based on the CCE organization onboarding ID.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{"id"},
					StateSchema:             &gcpmodels.TfIdsecCCEGCPOrganization{},
				},
				DataSourceAction: "tf-organization",
			},
		},
	})
}
