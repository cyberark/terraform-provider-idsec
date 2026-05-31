// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package sia

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/sia/access/actions"
	accessmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/sia/access/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

// siaAccessSchemasMap extends the SDK ActionToSchemaMap with the uninstall-relay action,
// which is supported by the SDK service (IdsecSIAAccessService.UninstallRelay) but not
// registered in the SDK's ActionToSchemaMap.
var siaAccessSchemasMap = func() map[string]interface{} {
	extended := make(map[string]interface{}, len(actions.ActionToSchemaMap)+1)
	for k, v := range actions.ActionToSchemaMap {
		extended[k] = v
	}
	extended["uninstall-relay"] = &accessmodels.IdsecSIAUninstallRelay{}
	return extended
}()

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "sia-access",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-access-connector", ActionDescription: "SIA connector resource, manages SIA connector installation and removal on SIA and target machines.", ActionVersion: 1, Schemas: siaAccessSchemasMap,
					},
					ExtraRequiredAttributes: []string{"connector_os", "connector_type", "target_machine", "username"},
					SensitiveAttributes:     []string{"password", "private_key_contents"},
					StateSchema:             &accessmodels.IdsecSIAAccessConnectorID{},
				},
				RawStateInference:   true,
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				// UpdateOperation is wired to uninstall-connector so that the IdsecSIAUninstallConnector
				// fields (force_delete, destroy-time retry_count/retry_delay, etc.) are merged into
				// the resource schema by the engine's create/update schema merge. Note: any plan-level
				// change to an attribute on this resource will dispatch UpdateOperation, which triggers
				// a remote uninstall on the target machine.
				ActionsMappings: map[tfactions.IdsecServiceActionOperation]string{
					tfactions.CreateOperation: "install-connector",
					tfactions.UpdateOperation: "uninstall-connector",
					tfactions.DeleteOperation: "uninstall-connector",
				},
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-access-relay", ActionDescription: "SIA HTTPS relay resource, manages SIA HTTPS relay installation and removal on target machines.", ActionVersion: 1, Schemas: siaAccessSchemasMap,
					},
					ExtraRequiredAttributes: []string{"https_relay_os", "target_machine", "username", "protocol_port_map"},
					SensitiveAttributes:     []string{"password", "private_key_contents"},
					StateSchema:             &accessmodels.IdsecSIAHTTPSRelay{},
					ComputedAttributes: []string{
						"https_relay_id", "host_ip", "host_name", "version", "status", "status_code", "os",
						"proxy_settings", "is_latest_version", "version_to_upgrade", "is_upgradable",
						"last_job_status", "last_job_error_code", "last_job_status_description",
						"last_job_info_update_date", "active_sessions_count",
					},
				},
				RawStateInference:   true,
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				// UpdateOperation is wired to uninstall-relay so that the IdsecSIAUninstallRelay
				// fields (force_delete, destroy-time retry_count/retry_delay, etc.) are merged into
				// the resource schema by the engine's create/update schema merge. Note: any plan-level
				// change to an attribute on this resource will dispatch UpdateOperation, which triggers
				// a remote uninstall on the target machine.
				ActionsMappings: map[tfactions.IdsecServiceActionOperation]string{
					tfactions.CreateOperation: "install-relay",
					tfactions.UpdateOperation: "uninstall-relay",
					tfactions.DeleteOperation: "uninstall-relay",
				},
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-access-relay", ActionDescription: "The SIA access relay data source, reads HTTPS relay information and metadata based on the relay ID.", ActionVersion: 1, Schemas: siaAccessSchemasMap,
					},
					StateSchema:             &accessmodels.IdsecSIAHTTPSRelay{},
					ExtraRequiredAttributes: []string{"https_relay_id"},
				},
				DataSourceAction: "get-relay",
			},
		},
	})
}
