// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	api "github.com/cyberark/idsec-sdk-golang/pkg"
	"github.com/cyberark/terraform-provider-idsec/internal/schemas"
)

// samlCertificateAttributes describe the identity provider signing certificate the server stores.
var samlCertificateAttributes = []string{"saml_idp_certificate_thumbprint", "saml_idp_certificate_subject", "saml_idp_certificate_expires"}

// samlDerivedURLAttributes are read from the metadata when the federation is configured with it.
var samlDerivedURLAttributes = []string{"saml_sign_in_url", "saml_logout_url"}

// federationSAMLPlanModifier plans the attributes the server derives from a SAML identity provider
// as unknown when what they are derived from may change, instead of carrying the prior state
// value, which the apply would otherwise keep until the next refresh.
//
// They may change when the metadata URL, the metadata document or the signing certificate changes,
// and on any update of a federation configured with a metadata URL, since the server fetches the
// metadata again on every update.
type federationSAMLPlanModifier struct{}

// ValidatePlan implements tfactions.IdsecPlanValidator.
func (m federationSAMLPlanModifier) ValidatePlan(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
	_ *api.IdsecAPI,
) {
	// Create plans them unknown already, and destroy has nothing to plan.
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() || req.Plan.Raw.Equal(req.State.Raw) {
		return
	}
	if planString(ctx, req, resp, "federation_type").ValueString() != federationTypeSAML {
		return
	}
	// A planned metadata URL means the server fetches the metadata again. One removed from the
	// configuration is planned unknown, so it counts too.
	metadataURL := planString(ctx, req, resp, "saml_metadata_url")
	changed := metadataURL.IsUnknown() || metadataURL.ValueString() != ""
	for _, attribute := range []string{
		schemas.WriteOnlyHashAttributeName("saml_metadata_xml"),
		schemas.WriteOnlyHashAttributeName("saml_signing_certificate"),
	} {
		var prior types.String
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root(attribute), &prior)...)
		if planned := planString(ctx, req, resp, attribute); planned.IsUnknown() || !planned.Equal(prior) {
			changed = true
		}
	}
	if !changed || resp.Diagnostics.HasError() {
		return
	}
	unknown := append([]string{}, samlCertificateAttributes...)
	for _, attribute := range samlDerivedURLAttributes {
		var configuredValue types.String
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(attribute), &configuredValue)...)
		if configuredValue.IsNull() {
			unknown = append(unknown, attribute)
		}
	}
	for _, attribute := range unknown {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(attribute), types.StringUnknown())...)
	}
}
