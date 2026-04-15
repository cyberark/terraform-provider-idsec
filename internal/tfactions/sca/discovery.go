// Copyright (c) CyberArk
// SPDX-License-Identifier: Apache-2.0

package sca

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/sca/actions"
	scamodels "github.com/cyberark/idsec-sdk-golang/pkg/services/sca/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "sca",
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "sca-discovery",
						ActionDescription: "Discover structural updates to an organization/directory that has already been onboarded to CyberArk, and scan for roles and resources",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
					},
					StateSchema: &scamodels.IdsecSCADiscoveryResponse{},
				},
				DataSourceAction: "discovery",
			},
		},
	})
}
