// Copyright (c) CyberArk
// SPDX-License-Identifier: Apache-2.0

package sia

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/sia/access/actions"
	accessmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/sia/access/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "sia-access",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-access-connector", ActionDescription: "SIA connector resource, manages SIA connector installation and removal on SIA and target machines.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{"connector_os", "connector_type", "target_machine", "username"},
					SensitiveAttributes:     []string{"password", "private_key_contents"},
					StateSchema:             &accessmodels.IdsecSIAAccessConnectorID{},
				},
				RawStateInference:   true,
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "install-connector", tfactions.DeleteOperation: "uninstall-connector"},
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-access-relay", ActionDescription: "SIA HTTPS relay resource, manages SIA HTTPS relay installation and removal on target machines.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{"https_relay_os", "target_machine", "username", "protocol_port_map"},
					SensitiveAttributes:     []string{"password", "private_key_contents"},
					StateSchema:             &accessmodels.IdsecSIAHTTPSRelay{},
					ComputedAttributes: []string{
						"id", "host_ip", "host_name", "version", "status", "status_code", "os",
						"proxy_settings", "is_latest_version", "version_to_upgrade", "is_upgradable",
						"last_job_status", "last_job_error_code", "last_job_status_description",
						"last_job_info_update_date", "active_sessions_count",
					},
				},
				RawStateInference:   true,
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "install-relay", tfactions.DeleteOperation: "delete-relay"},
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-access-relay", ActionDescription: "The SIA access relay data source, reads HTTPS relay information and metadata based on the relay ID.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema:             &accessmodels.IdsecSIAHTTPSRelay{},
					ExtraRequiredAttributes: []string{"https_relay_id"},
				},
				DataSourceAction: "get-relay",
			},
		},
	})
}
