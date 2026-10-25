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
					// connector_type and connector_os are always required. target_machine and username
					// are required only for machine-based connector_os values (Linux, windows); when
					// connector_os is k8s-ephemeral, k8s_details (with a non-empty k8s_namespace) is
					// required instead. This conditional-attribute validation is enforced on the SDK
					// side: IdsecSIAInstallConnector.K8SDetails has validate:"required_if=...", and
					// IdsecSIAK8SDetails.K8SNamespace itself has validate:"required", so
					// ValidateConfig (which validates HCL config against the create schema) rejects a
					// missing/incomplete k8s_details block for k8s-ephemeral connectors at plan time.
					ExtraRequiredAttributes: []string{"connector_type", "connector_os"},
					StateSchema:             &accessmodels.IdsecSIAAccessConnectorID{},
					PageNotes: []tfactions.DocNote{
						{
							Severity:       tfactions.DocNoteWarning,
							BreakingChange: true,
							Body:           "The `k_8_s_details` nested block and the `k_8_s_namespace` attribute (previously mangled by Terraform's attribute-name conversion, which silently dropped any value assigned to them) have been renamed to `k8s_details` and `k8s_namespace`; the attributes inside `k8s_details` are likewise renamed from `k_8_s_*` to `k8s_*`. Update your Terraform configuration to use the new names.",
						},
						{
							Severity:       tfactions.DocNoteWarning,
							BreakingChange: true,
							Body:           "The separate, top-level `k8s_namespace` attribute (used only at connector removal time) has been removed. The connector's namespace is now tracked solely via `k8s_details.k8s_namespace`, which is also used to uninstall it. If your configuration set a top-level `k8s_namespace`, remove it; `k8s_details.k8s_namespace` is required for `connector_os = \"k8s-ephemeral\"` and is validated at `terraform plan` time.",
						},
					},
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
