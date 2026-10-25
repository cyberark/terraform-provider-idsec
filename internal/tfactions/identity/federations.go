// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/identity/federations/actions"
	federationsmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/identity/federations/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
	"github.com/cyberark/terraform-provider-idsec/internal/schemas"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "identity-federations",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "identity-federation",
						ActionDescription: "The Identity service federation resource, which manages an external SAML 2.0 or OIDC identity provider.",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
					},
					StateSchema:             &federationsmodels.IdsecIdentityFederation{},
					ComputedAsSetAttributes: []string{"domains", "oidc_scopes", "device_types", "user_agents"},
					ImmutableAttributes:     []string{"federation_type"},
					ComputedAttributes: []string{
						"federation_id", "oidc_redirect_url", "oidc_initiate_login_url",
						"saml_idp_certificate_thumbprint", "saml_idp_certificate_subject", "saml_idp_certificate_expires",
						"saml_sp_metadata_url", "saml_sp_acs_url", "saml_sp_signing_certificate_url",
					},
					AttributeNotes: map[string][]tfactions.DocNote{
						"attribute_mappings": {{
							Severity: tfactions.DocNoteWarning,
							Body: "Attribute mapping for external identity providers must be enabled on the tenant, which it is " +
								"by default. When it is not, the server saves the federation but discards attribute_mappings " +
								"without an error, so the apply fails with \"Provider produced inconsistent result\" and every " +
								"later apply tries again. On such a tenant, leave attribute_mappings unset.",
						}},
						"domains": {{
							Severity: tfactions.DocNoteInfo,
							Body: "With routing rules enabled on the tenant, the server rejects a federation that has no " +
								"routing rule configured. Domains are the only routing rule this resource manages today, " +
								"so set at least one.",
						}},
						"saml_metadata_url": {{
							Severity: tfactions.DocNoteInfo,
							Body: "The server fetches the metadata again on every create and update, so the URL must stay " +
								"reachable from the tenant, and an apply picks up a rotated identity provider certificate.",
						}},
						"saml_metadata_xml": {{
							Severity: tfactions.DocNoteInfo,
							Body: "A federation created from a metadata document reads back like one configured manually: " +
								"saml_sign_in_url, saml_logout_url and the certificate attributes hold the values read from it. " +
								"The document is only sent again when it changes.",
						}},
						"saml_signing_certificate": {{
							Severity: tfactions.DocNoteWarning,
							Body: "Two SAML federations on a tenant cannot share an identity provider signing certificate, " +
								"whichever way it was supplied.",
						}},
					},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings: map[tfactions.IdsecServiceActionOperation]string{
					tfactions.CreateOperation: "create",
					tfactions.ReadOperation:   "get",
					tfactions.UpdateOperation: "update",
					tfactions.DeleteOperation: "delete",
				},
				PlanValidators:            []tfactions.IdsecPlanValidator{federationTypeValidator{}, federationMappingValidator{}, federationSAMLPlanModifier{}},
				ImportID:                  "federation_id",
				WriteOnlyHashedAttributes: []string{"oidc_client_secret", "saml_metadata_xml", "saml_signing_certificate", "saml_oauth_client_secret"},
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "identity-federation-global-group-mapping",
						ActionDescription: "The Identity service federation global group mapping resource, which maps a group attribute value from any external identity provider to an Identity group.",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
					},
					StateSchema:        &federationsmodels.IdsecIdentityFederationGlobalGroupMapping{},
					ForceNewAttributes: []string{"attribute_value", "group_name"},
					// The server matches global mappings without regard to case and keeps the
					// spelling it stored first.
					SemanticEqualityAttributes: map[string]schemas.SemanticEqualityKind{
						"attribute_value": schemas.SemanticEqualityCaseInsensitive,
						"group_name":      schemas.SemanticEqualityCaseInsensitive,
					},
					AttributeNotes: map[string][]tfactions.DocNote{
						"group_name": {{
							Severity: tfactions.DocNoteInfo,
							Body:     "The server creates the federated group when no group of this name exists.",
						}},
					},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings: map[tfactions.IdsecServiceActionOperation]string{
					tfactions.CreateOperation: "add-global-group-mapping",
					tfactions.ReadOperation:   "get-global-group-mapping",
					tfactions.DeleteOperation: "remove-global-group-mapping",
				},
				ImportID: "attribute_value:group_name",
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName:        "identity-federation",
						ActionDescription: "The Identity service federation data source, which reads an external SAML 2.0 or OIDC identity provider.",
						ActionVersion:     1,
						Schemas:           actions.ActionToSchemaMap,
					},
					StateSchema:             &federationsmodels.IdsecIdentityFederation{},
					ExtraRequiredAttributes: []string{"federation_id"},
					ComputedAsSetAttributes: []string{"domains", "oidc_scopes", "device_types", "user_agents"},
				},
				DataSourceAction: "get",
			},
		},
	})
}
