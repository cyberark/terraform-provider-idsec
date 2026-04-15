// Copyright (c) CyberArk
// SPDX-License-Identifier: Apache-2.0

package pcloud

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/pcloud/targetplatforms/actions"
	targetplatformsmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/pcloud/targetplatforms/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "pcloud-target-platforms",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "pcloud-target-platform", ActionDescription: "Privilege Cloud target platform resource, manages the import of Privilege Cloud target platforms.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &targetplatformsmodels.IdsecPCloudTargetPlatform{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "import", tfactions.ReadOperation: "get", tfactions.DeleteOperation: "delete"},
				ImportID:            "target_platform_id",
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "pcloud-target-platform", ActionDescription: "Privilege Cloud target platform data source, reads target platform information and metadata, based on the platform ID.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{"target_platform_id"},
					StateSchema:             &targetplatformsmodels.IdsecPCloudTargetPlatform{},
				},
				DataSourceAction: "get",
			},
		},
	})
}
