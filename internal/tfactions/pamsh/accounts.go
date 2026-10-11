// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package pamsh

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/pamsh/pamshaccounts/actions"
	accountsmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/pamsh/pamshaccounts/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "pamsh-accounts",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "pamsh-account",
						ActionDescription: "Manage Privilege Cloud account information, metadata, and credentials",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
						Enabled:           boolPtr(true),
					},
					ExtraRequiredAttributes: []string{},
					ComputedAttributes: []string{
						"id",
						"status",
						"created_time",
						"category_modification_time",
						"secret_management.last_modified_time",
					},
					ImmutableAttributes: []string{
						"safe_name",
						"secret_type",
					},
					StateSchema: &accountsmodels.IdsecPamshAccount{},
					AttributeNotes: map[string][]tfactions.DocNote{
						"secret": {{
							Severity: tfactions.DocNoteWarning,
							Body: "The secret attribute has been migrated to a hashed write-only attribute, " +
								"which means that any value will not be stored in the state file in plain text i.e., " +
								"it will be null, even if set in the configuration. " +
								"Any outputs and locals that reference the state value of secret e.g., " +
								"`idsec_pamsh_account.example_resource.secret` will be null as well." +
								"After `terraform import` you will notice an update on subsequent `terrafrom apply` e.g.: " +
								"`+ secret_write_only_hash               = (sensitive value)`. " +
								"The value configured at import time " +
								"is recorded into Terraform state without being written to the vault. " +
								"Because state then matches your configuration, no later apply will write it either, " +
								"even though the vault still holds the old credential. To actually set the credential, " +
								"change `secret` to a different value on a subsequent apply.",
						}},
					},
				},
				SupportedOperations:       []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:           map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "create", tfactions.ReadOperation: "get", tfactions.UpdateOperation: "update", tfactions.DeleteOperation: "delete"},
				ImportID:                  "id",
				WriteOnlyHashedAttributes: []string{"secret"},
				UpdateOnlyModified:        true,
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "pamsh-account", ActionDescription: "Privilege Cloud account data source, reads account information and metadata, based on the account ID.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap, Enabled: boolPtr(false),
					},
					ExtraRequiredAttributes: []string{"id"},
					StateSchema:             &accountsmodels.IdsecPamshAccount{},
				},
				DataSourceAction: "get",
			},
		},
	})
}

func boolPtr(b bool) *bool {
	return &b
}
