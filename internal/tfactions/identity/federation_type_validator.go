// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"context"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	api "github.com/cyberark/idsec-sdk-golang/pkg"
)

const (
	federationTypeOIDC = "OIDC"
	federationTypeSAML = "SAML 2.0"
)

// samlInboundAttributes are the attributes that tell the server how to read a SAML identity
// provider: a metadata URL, a metadata document, or the manual attributes from saml_sign_in_url on.
var samlInboundAttributes = []string{"saml_metadata_url", "saml_metadata_xml", "saml_sign_in_url", "saml_logout_url", "saml_signing_certificate"}

// federationTypeValidator checks at plan time the federation_type rules that the SDK model's
// validate tags cannot express. The tags, which ValidateConfig runs first, reject attributes of the
// other federation_type, the required OIDC attributes, a metadata URL that is not a discovery
// document, and conflicting ways of reading a SAML identity provider. What is left here needs prior
// state or more than one rule at once: the client secret and the manual signing certificate are
// required on create only, a SAML federation must set one way of reading its identity provider, and
// the URL redirect pattern must be one the server accepts.
//
// It reads the configuration rather than the plan: oidc_client_secret, saml_metadata_xml and
// saml_signing_certificate are write-only, so they are always null in the plan, and the other
// attributes are Optional+Computed, so the plan carries the prior state value for an attribute the
// configuration no longer sets.
type federationTypeValidator struct{}

// ValidatePlan implements tfactions.IdsecPlanValidator.
func (v federationTypeValidator) ValidatePlan(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
	_ *api.IdsecAPI,
) {
	// Destroy plans have no configuration to check.
	if req.Plan.Raw.IsNull() {
		return
	}
	var federationType types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("federation_type"), &federationType)...)
	if resp.Diagnostics.HasError() || federationType.IsNull() || federationType.IsUnknown() {
		return
	}
	switch federationType.ValueString() {
	case federationTypeSAML:
		v.checkSAML(ctx, req, resp)
	case federationTypeOIDC:
		// The secret is only required on create: afterwards the server keeps the stored secret
		// when none is sent, so removing it from the configuration is not an error.
		if req.State.Raw.IsNull() {
			var secret types.String
			resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("oidc_client_secret"), &secret)...)
			if !secret.IsUnknown() && secret.ValueString() == "" {
				resp.Diagnostics.AddAttributeError(
					path.Root("oidc_client_secret"),
					"Missing OIDC Attribute",
					fmt.Sprintf("oidc_client_secret is required when federation_type is %q.", federationTypeOIDC),
				)
			}
		}
	}
}

// checkSAML requires a way of reading the identity provider, and the signing certificate with a
// manually configured one on create.
func (v federationTypeValidator) checkSAML(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Any inbound attribute counts here. An incomplete manual configuration, saml_logout_url or
	// saml_signing_certificate without saml_sign_in_url, is rejected before this runs by the SDK
	// create model's required_with tag on saml_sign_in_url. saml_logout_url itself is optional.
	inboundSet := slices.ContainsFunc(samlInboundAttributes, func(attribute string) bool {
		return configured(ctx, req, resp, attribute)
	})
	switch {
	case !inboundSet && !anyUnknown(ctx, req, resp, samlInboundAttributes...):
		resp.Diagnostics.AddAttributeError(
			path.Root("saml_metadata_url"),
			"Missing SAML Identity Provider",
			"A SAML 2.0 federation needs one of saml_metadata_url, saml_metadata_xml, or saml_sign_in_url with saml_signing_certificate.",
		)
	// As with the OIDC secret, the stored certificate is kept on update when none is sent.
	case configured(ctx, req, resp, "saml_sign_in_url") && req.State.Raw.IsNull() &&
		!configured(ctx, req, resp, "saml_signing_certificate") && !anyUnknown(ctx, req, resp, "saml_signing_certificate"):
		resp.Diagnostics.AddAttributeError(
			path.Root("saml_signing_certificate"),
			"Missing SAML Signing Certificate",
			"saml_signing_certificate is required with saml_sign_in_url.",
		)
	}

	var pattern types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("saml_url_redirect_pattern"), &pattern)...)
	if known(pattern) && !isHTTPSPattern(pattern.ValueString()) {
		resp.Diagnostics.AddAttributeError(
			path.Root("saml_url_redirect_pattern"),
			"Invalid URL Redirect Pattern",
			"saml_url_redirect_pattern must be an https URL, which may contain *, that is still a valid URL "+
				"with every * removed: https://*.example.com is rejected, https://*example.com is not.",
		)
	}
}

// isHTTPSPattern mirrors the server's check of a URL pattern: with every * removed it must parse as
// an absolute https URL, with a host that has no empty label.
func isHTTPSPattern(pattern string) bool {
	parsed, err := url.Parse(strings.ReplaceAll(pattern, "*", ""))
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" {
		return false
	}
	return !slices.Contains(strings.Split(parsed.Hostname(), "."), "")
}

// anyUnknown reports whether the configuration sets any of the string attributes to an unknown
// value, which apply may resolve to anything.
func known(value types.String) bool {
	return !value.IsNull() && !value.IsUnknown()
}

func planString(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse, attribute string) types.String {
	var value types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root(attribute), &value)...)
	return value
}

func anyUnknown(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse, attributes ...string) bool {
	for _, attribute := range attributes {
		var value types.String
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(attribute), &value)...)
		if value.IsUnknown() {
			return true
		}
	}
	return false
}

// configured reports whether the configuration sets a string attribute to a non-null value.
func configured(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse, attribute string) bool {
	var value types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(attribute), &value)...)
	return !value.IsNull()
}
