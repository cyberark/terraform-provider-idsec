// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package sechub

import (
	secretsactions "github.com/cyberark/idsec-sdk-golang/pkg/services/sechub/secrets/actions"
	secretsmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/sechub/secrets/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "sechub-secrets",
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "sechub-onboard-secret",
						ActionDescription: "Secrets Hub secret data source, finds a single discovered secret by its exact name and secret store ID, to be used as input for secret onboarding.",
						ActionVersion:     1,
						Schemas:           secretsactions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{
						"name",
						"store_id",
					},
					StateSchema: &secretsmodels.IdsecSecHubSecret{},
					PageNotes: []tfactions.DocNote{
						{
							Severity: tfactions.DocNoteInfo,
							Body: "The secret must already have been discovered by a Secrets Hub scan of the secret store. " +
								"The lookup fails if no secret exactly matches `name` in the store, if more than one secret matches, " +
								"or if the secret is not a candidate for onboarding.",
						},
					},
				},
				DataSourceAction: "find-by-name-and-store-id",
			},
		},
	})
}
