// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"context"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/cyberark/idsec-sdk-golang/pkg/services/identity/federations"
)

// federationTestSchema mirrors the attributes the validator reads.
var federationTestSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"federation_type":                          schema.StringAttribute{Required: true},
		"oidc_metadata_url":                        schema.StringAttribute{Optional: true, Computed: true},
		"oidc_client_id":                           schema.StringAttribute{Optional: true, Computed: true},
		"oidc_client_secret":                       schema.StringAttribute{Optional: true, WriteOnly: true},
		"oidc_grant_type":                          schema.StringAttribute{Optional: true, Computed: true},
		"oidc_pkce_enabled":                        schema.BoolAttribute{Optional: true, Computed: true},
		"oidc_scopes":                              schema.SetAttribute{ElementType: types.StringType, Optional: true, Computed: true},
		"oidc_post_logout_redirect_uri":            schema.StringAttribute{Optional: true, Computed: true},
		"saml_metadata_url":                        schema.StringAttribute{Optional: true, Computed: true},
		"saml_metadata_xml":                        schema.StringAttribute{Optional: true, WriteOnly: true},
		"saml_sign_in_url":                         schema.StringAttribute{Optional: true, Computed: true},
		"saml_logout_url":                          schema.StringAttribute{Optional: true, Computed: true},
		"saml_signing_certificate":                 schema.StringAttribute{Optional: true, WriteOnly: true},
		"saml_sign_algorithm":                      schema.StringAttribute{Optional: true, Computed: true},
		"saml_url_redirect_enabled":                schema.BoolAttribute{Optional: true, Computed: true},
		"saml_url_redirect_pattern":                schema.StringAttribute{Optional: true, Computed: true},
		"saml_metadata_xml_write_only_hash":        schema.StringAttribute{Computed: true},
		"saml_signing_certificate_write_only_hash": schema.StringAttribute{Computed: true},
		"saml_idp_certificate_thumbprint":          schema.StringAttribute{Computed: true},
		"saml_idp_certificate_subject":             schema.StringAttribute{Computed: true},
		"saml_idp_certificate_expires":             schema.StringAttribute{Computed: true},
		"map_users":                                schema.StringAttribute{Optional: true, Computed: true},
		"map_user_directory_service":               schema.StringAttribute{Optional: true, Computed: true},
		"map_user_create_if_not_found":             schema.BoolAttribute{Optional: true, Computed: true},
		"map_user_sync_with_cloud_user":            schema.BoolAttribute{Optional: true, Computed: true},
		"map_user_sync_with_federated_user":        schema.BoolAttribute{Optional: true, Computed: true},
		"device_types":                             schema.SetAttribute{ElementType: types.StringType, Optional: true, Computed: true},
		"user_agents":                              schema.SetAttribute{ElementType: types.StringType, Optional: true, Computed: true},
		"group_mappings": schema.ListNestedAttribute{Optional: true, Computed: true, NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"attribute_value": schema.StringAttribute{Required: true},
				"group_name":      schema.StringAttribute{Required: true},
			},
		}},
		"attribute_mappings": schema.ListNestedAttribute{Optional: true, Computed: true, NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"target": schema.StringAttribute{Required: true},
				"source": schema.StringAttribute{Required: true},
			},
		}},
	},
}

var (
	groupMappingType     = tftypes.Object{AttributeTypes: map[string]tftypes.Type{"attribute_value": tftypes.String, "group_name": tftypes.String}}
	attributeMappingType = tftypes.Object{AttributeTypes: map[string]tftypes.Type{"target": tftypes.String, "source": tftypes.String}}
	stringSetType        = tftypes.Set{ElementType: tftypes.String}
)

var federationObjType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"federation_type":                          tftypes.String,
		"oidc_metadata_url":                        tftypes.String,
		"oidc_client_id":                           tftypes.String,
		"oidc_client_secret":                       tftypes.String,
		"oidc_grant_type":                          tftypes.String,
		"oidc_pkce_enabled":                        tftypes.Bool,
		"oidc_scopes":                              tftypes.Set{ElementType: tftypes.String},
		"oidc_post_logout_redirect_uri":            tftypes.String,
		"saml_metadata_url":                        tftypes.String,
		"saml_metadata_xml":                        tftypes.String,
		"saml_sign_in_url":                         tftypes.String,
		"saml_logout_url":                          tftypes.String,
		"saml_signing_certificate":                 tftypes.String,
		"saml_sign_algorithm":                      tftypes.String,
		"saml_url_redirect_enabled":                tftypes.Bool,
		"saml_url_redirect_pattern":                tftypes.String,
		"saml_metadata_xml_write_only_hash":        tftypes.String,
		"saml_signing_certificate_write_only_hash": tftypes.String,
		"saml_idp_certificate_thumbprint":          tftypes.String,
		"saml_idp_certificate_subject":             tftypes.String,
		"saml_idp_certificate_expires":             tftypes.String,
		"map_users":                                tftypes.String,
		"map_user_directory_service":               tftypes.String,
		"map_user_create_if_not_found":             tftypes.Bool,
		"map_user_sync_with_cloud_user":            tftypes.Bool,
		"map_user_sync_with_federated_user":        tftypes.Bool,
		"device_types":                             stringSetType,
		"user_agents":                              stringSetType,
		"group_mappings":                           tftypes.List{ElementType: groupMappingType},
		"attribute_mappings":                       tftypes.List{ElementType: attributeMappingType},
	},
}

// federationValues builds a raw object, with every attribute not in set left null.
func federationValues(set map[string]tftypes.Value) tftypes.Value {
	values := map[string]tftypes.Value{}
	for name, typ := range federationObjType.AttributeTypes {
		values[name] = tftypes.NewValue(typ, nil)
	}
	for name, value := range set {
		values[name] = value
	}
	return tftypes.NewValue(federationObjType, values)
}

func str(value string) tftypes.Value {
	return tftypes.NewValue(tftypes.String, value)
}

func strSet(values ...string) tftypes.Value {
	elements := []tftypes.Value{}
	for _, value := range values {
		elements = append(elements, str(value))
	}
	return tftypes.NewValue(stringSetType, elements)
}

// strSetWithUnknown is a set of the given strings plus one element that is unknown until apply.
func strSetWithUnknown(values ...string) tftypes.Value {
	elements := []tftypes.Value{tftypes.NewValue(tftypes.String, tftypes.UnknownValue)}
	for _, value := range values {
		elements = append(elements, str(value))
	}
	return tftypes.NewValue(stringSetType, elements)
}

// pairs builds a list of two-string objects, such as group_mappings, from alternating values.
func pairs(objectType tftypes.Object, first, second string, values ...string) tftypes.Value {
	elements := []tftypes.Value{}
	for i := 0; i+1 < len(values); i += 2 {
		elements = append(elements, tftypes.NewValue(objectType, map[string]tftypes.Value{first: str(values[i]), second: str(values[i+1])}))
	}
	return tftypes.NewValue(tftypes.List{ElementType: objectType}, elements)
}

// plannedValues is what Terraform plans for a configuration. On create, an Optional+Computed
// attribute the configuration leaves unset is planned unknown, not null: the plan hooks must not read
// it as unset, which is why the tests do not pass the configuration as the plan.
func plannedValues(config map[string]tftypes.Value, existing bool) tftypes.Value {
	if existing {
		return federationValues(config)
	}
	planned := map[string]tftypes.Value{}
	for name, attribute := range federationTestSchema.Attributes {
		if _, set := config[name]; !set && attribute.IsComputed() && attribute.IsOptional() {
			planned[name] = tftypes.NewValue(federationObjType.AttributeTypes[name], tftypes.UnknownValue)
		}
	}
	for name, value := range config {
		planned[name] = value
	}
	return federationValues(planned)
}

func validateFederation(t *testing.T, config map[string]tftypes.Value, existing bool) resource.ModifyPlanResponse {
	t.Helper()
	raw := federationValues(config)
	plan := tfsdk.Plan{Raw: plannedValues(config, existing), Schema: federationTestSchema}
	state := tfsdk.State{Raw: tftypes.NewValue(federationObjType, nil), Schema: federationTestSchema}
	if existing {
		state.Raw = raw
	}
	req := resource.ModifyPlanRequest{
		Config: tfsdk.Config{Raw: raw, Schema: federationTestSchema},
		Plan:   plan,
		State:  state,
	}
	resp := resource.ModifyPlanResponse{Plan: plan}
	federationTypeValidator{}.ValidatePlan(context.Background(), req, &resp, nil)
	federationMappingValidator{}.ValidatePlan(context.Background(), req, &resp, nil)
	return resp
}

// errorList renders each error as "<attribute path>: <summary>", sorted, so a test pins which
// check fired and on which attribute, not just how many errors there were.
func errorList(diags diag.Diagnostics) []string {
	got := []string{}
	for _, d := range diags.Errors() {
		where := ""
		if withPath, ok := d.(diag.DiagnosticWithPath); ok {
			where = withPath.Path().String()
		}
		got = append(got, where+": "+d.Summary())
	}
	slices.Sort(got)
	return got
}

func TestFederationTypeValidator(t *testing.T) {
	completeOIDC := map[string]tftypes.Value{
		"federation_type":    str("OIDC"),
		"oidc_metadata_url":  str("https://idp.example.com/.well-known/openid-configuration"),
		"oidc_client_id":     str("client"),
		"oidc_client_secret": str("secret"),
	}
	withoutSecret := map[string]tftypes.Value{
		"federation_type":   str("OIDC"),
		"oidc_metadata_url": str("https://idp.example.com/.well-known/openid-configuration"),
		"oidc_client_id":    str("client"),
	}
	tests := []struct {
		name     string
		config   map[string]tftypes.Value
		existing bool
		want     []string
	}{
		{name: "complete OIDC create", config: completeOIDC},
		{name: "OIDC create without secret", config: withoutSecret, want: []string{"oidc_client_secret: Missing OIDC Attribute"}},
		{name: "OIDC update without secret keeps the stored one", config: withoutSecret, existing: true},
		{name: "OIDC with unknown client ID", config: map[string]tftypes.Value{
			"federation_type":    str("OIDC"),
			"oidc_metadata_url":  str("https://idp.example.com/.well-known/openid-configuration"),
			"oidc_client_id":     tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			"oidc_client_secret": str("secret"),
		}},
		{name: "SAML without identity provider", config: map[string]tftypes.Value{"federation_type": str("SAML 2.0")}, want: []string{"saml_metadata_url: Missing SAML Identity Provider"}},
		{name: "SAML metadata URL", config: map[string]tftypes.Value{
			"federation_type":           str("SAML 2.0"),
			"saml_metadata_url":         str("https://idp.example.com/metadata"),
			"saml_sign_algorithm":       str("SHA256"),
			"saml_url_redirect_enabled": tftypes.NewValue(tftypes.Bool, true),
			"saml_url_redirect_pattern": str("https://app*.example.com/*"),
		}},
		{name: "SAML metadata document", config: map[string]tftypes.Value{
			"federation_type":   str("SAML 2.0"),
			"saml_metadata_xml": str("<EntityDescriptor/>"),
		}},
		{name: "SAML manual", config: map[string]tftypes.Value{
			"federation_type":          str("SAML 2.0"),
			"saml_sign_in_url":         str("https://idp.example.com/sso"),
			"saml_logout_url":          str("https://idp.example.com/slo"),
			"saml_signing_certificate": str("cert"),
		}},
		{name: "SAML manual create without certificate", config: map[string]tftypes.Value{
			"federation_type":  str("SAML 2.0"),
			"saml_sign_in_url": str("https://idp.example.com/sso"),
		}, want: []string{"saml_signing_certificate: Missing SAML Signing Certificate"}},
		{name: "SAML manual update without certificate keeps the stored one", config: map[string]tftypes.Value{
			"federation_type":  str("SAML 2.0"),
			"saml_sign_in_url": str("https://idp.example.com/sso"),
		}, existing: true},
		{name: "SAML unknown metadata URL", config: map[string]tftypes.Value{
			"federation_type":   str("SAML 2.0"),
			"saml_metadata_url": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		}},
		{name: "SAML http redirect pattern", config: map[string]tftypes.Value{
			"federation_type":           str("SAML 2.0"),
			"saml_metadata_url":         str("https://idp.example.com/metadata"),
			"saml_url_redirect_pattern": str("http://*.example.com"),
		}, want: []string{"saml_url_redirect_pattern: Invalid URL Redirect Pattern"}},
		{name: "SAML wildcard before a dot", config: map[string]tftypes.Value{
			"federation_type":           str("SAML 2.0"),
			"saml_metadata_url":         str("https://idp.example.com/metadata"),
			"saml_url_redirect_pattern": str("https://*.example.com/*"),
		}, want: []string{"saml_url_redirect_pattern: Invalid URL Redirect Pattern"}},
		{name: "SAML wildcard inside a label", config: map[string]tftypes.Value{
			"federation_type":           str("SAML 2.0"),
			"saml_metadata_url":         str("https://idp.example.com/metadata"),
			"saml_url_redirect_pattern": str("https://app*.example.com/*"),
		}},
		{name: "unknown federation type", config: map[string]tftypes.Value{
			"federation_type": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := validateFederation(t, tt.config, tt.existing)
			if got := errorList(resp.Diagnostics); !slices.Equal(got, tt.want) {
				t.Errorf("errors = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFederationTypeValidatorSkipsDestroy(t *testing.T) {
	plan := tfsdk.Plan{Raw: tftypes.NewValue(federationObjType, nil), Schema: federationTestSchema}
	req := resource.ModifyPlanRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(federationObjType, nil), Schema: federationTestSchema},
		Plan:   plan,
		State:  tfsdk.State{Raw: federationValues(map[string]tftypes.Value{"federation_type": str("OIDC")}), Schema: federationTestSchema},
	}
	resp := resource.ModifyPlanResponse{Plan: plan}
	federationTypeValidator{}.ValidatePlan(context.Background(), req, &resp, nil)
	if resp.Diagnostics.HasError() {
		t.Errorf("destroy plan reported errors: %v", resp.Diagnostics)
	}
}

func TestFederationMappingValidator(t *testing.T) {
	saml := func(set map[string]tftypes.Value) map[string]tftypes.Value {
		set["federation_type"] = str("SAML 2.0")
		set["saml_metadata_url"] = str("https://idp.example.com/metadata")
		return set
	}
	tests := []struct {
		name   string
		config map[string]tftypes.Value
		want   []string
	}{
		{name: "ordered group mappings", config: saml(map[string]tftypes.Value{
			"group_mappings": pairs(groupMappingType, "attribute_value", "group_name", "b", "B", "a", "A", "a", "B"),
		})},
		{name: "duplicate group mapping", config: saml(map[string]tftypes.Value{
			"group_mappings": pairs(groupMappingType, "attribute_value", "group_name", "a", "A", "a", "A"),
		}), want: []string{"group_mappings[1]: Duplicate Group Mapping"}},
		{name: "create users without Required mapping", config: saml(map[string]tftypes.Value{
			"map_users":                    str("Optional"),
			"map_user_create_if_not_found": tftypes.NewValue(tftypes.Bool, true),
		}), want: []string{"map_user_create_if_not_found: User Creation Needs Required Mapping"}},
		{name: "OIDC create users without email scope", config: map[string]tftypes.Value{
			"federation_type":              str("OIDC"),
			"oidc_metadata_url":            str("https://idp.example.com/.well-known/openid-configuration"),
			"oidc_client_id":               str("client"),
			"oidc_client_secret":           str("secret"),
			"oidc_scopes":                  strSet("openid", "profile"),
			"map_users":                    str("Required"),
			"map_user_create_if_not_found": tftypes.NewValue(tftypes.Bool, true),
		}, want: []string{"oidc_scopes: Missing Email Scope"}},
		{name: "OIDC create users with a scope unknown until apply", config: map[string]tftypes.Value{
			"federation_type":              str("OIDC"),
			"oidc_metadata_url":            str("https://idp.example.com/.well-known/openid-configuration"),
			"oidc_client_id":               str("client"),
			"oidc_client_secret":           str("secret"),
			"oidc_scopes":                  strSetWithUnknown("openid", "profile"),
			"map_users":                    str("Required"),
			"map_user_create_if_not_found": tftypes.NewValue(tftypes.Bool, true),
		}},
		{name: "attribute mappings on the Cloud Directory", config: saml(map[string]tftypes.Value{
			"map_users":                  str("Required"),
			"map_user_directory_service": str(federations.CloudDirectoryServiceUUID),
			"attribute_mappings":         pairs(attributeMappingType, "target", "source", "DisplayName", "name"),
		})},
		{name: "attribute mappings on any directory", config: saml(map[string]tftypes.Value{
			"map_users":                  str("Required"),
			"map_user_directory_service": str(""),
			"attribute_mappings":         pairs(attributeMappingType, "target", "source", "DisplayName", "name"),
		}), want: []string{"map_user_directory_service: Attribute Mappings Need The Cloud Directory"}},
		{name: "attribute mappings without a sync option", config: saml(map[string]tftypes.Value{
			"map_users":                         str("Required"),
			"map_user_directory_service":        str(federations.CloudDirectoryServiceUUID),
			"map_user_sync_with_federated_user": tftypes.NewValue(tftypes.Bool, false),
			"map_user_sync_with_cloud_user":     tftypes.NewValue(tftypes.Bool, false),
			"map_user_create_if_not_found":      tftypes.NewValue(tftypes.Bool, false),
			"attribute_mappings":                pairs(attributeMappingType, "target", "source", "DisplayName", "name"),
		}), want: []string{"attribute_mappings: Attribute Mappings Are Not Applied"}},
		{name: "duplicate target", config: saml(map[string]tftypes.Value{
			"map_users":                  str("Optional"),
			"map_user_directory_service": str(federations.CloudDirectoryServiceUUID),
			"attribute_mappings":         pairs(attributeMappingType, "target", "source", "DisplayName", "name", "displayname", "nickname"),
		}), want: []string{"attribute_mappings[1].target: Duplicate Attribute Mapping"}},
		{name: "partial device types with Required mapping", config: saml(map[string]tftypes.Value{
			"map_users":    str("Required"),
			"device_types": strSet("iOS"),
			"user_agents":  strSet("Browser"),
		})},
		{name: "partial device types without Required mapping", config: saml(map[string]tftypes.Value{
			"map_users":    str("Disabled"),
			"device_types": strSet("Others"),
		}), want: []string{"device_types: Device Routing Needs Required Mapping"}},
		{name: "mobile devices with any user agent", config: saml(map[string]tftypes.Value{
			"map_users":    str("Required"),
			"device_types": strSet("Android"),
			"user_agents":  strSet("Any"),
		}), want: []string{"user_agents: Unsupported User Agent For Mobile Devices"}},
		{name: "empty user agents", config: saml(map[string]tftypes.Value{
			"user_agents": strSet(),
		}), want: []string{"user_agents: No User Agent Selected"}},
		{name: "user agents with a value unknown until apply", config: saml(map[string]tftypes.Value{
			"user_agents": strSetWithUnknown(),
		})},
		{name: "device types with a value unknown until apply", config: saml(map[string]tftypes.Value{
			"map_users":    str("Optional"),
			"device_types": strSetWithUnknown("Android"),
		})},
		// Checked against the settings the server stores, not only the configuration: an unset
		// user_agents is stored as Any, which the server refuses for a mobile federation.
		{name: "mobile devices with user agents left unset", config: saml(map[string]tftypes.Value{
			"map_users":    str("Required"),
			"device_types": strSet("iOS"),
		}), want: []string{"user_agents: Unsupported User Agent For Mobile Devices"}},
		// And with the server's defaults for what the configuration leaves unset: no directory.
		{name: "attribute mappings with the user mapping left unset", config: saml(map[string]tftypes.Value{
			"attribute_mappings": pairs(attributeMappingType, "target", "source", "DisplayName", "name"),
		}), want: []string{"map_user_directory_service: Attribute Mappings Need The Cloud Directory"}},
		{name: "attribute mappings with a directory unknown until apply", config: saml(map[string]tftypes.Value{
			"map_user_directory_service": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			"attribute_mappings":         pairs(attributeMappingType, "target", "source", "DisplayName", "name"),
		})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := validateFederation(t, tt.config, false)
			if got := errorList(resp.Diagnostics); !slices.Equal(got, tt.want) {
				t.Errorf("errors = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestFederationMappingValidatorOnUpdate checks an update against the prior state, the way the
// SDK's Update checks it against the stored federation: a problem counts only when the plan changes
// an attribute it depends on, so a federation edited in the portal can still be updated.
func TestFederationMappingValidatorOnUpdate(t *testing.T) {
	stored := map[string]tftypes.Value{
		"federation_type":              str("SAML 2.0"),
		"saml_metadata_url":            str("https://idp.example.com/metadata"),
		"map_users":                    str("Optional"),
		"map_user_create_if_not_found": tftypes.NewValue(tftypes.Bool, false),
	}
	update := func(changes map[string]tftypes.Value) resource.ModifyPlanResponse {
		planned := map[string]tftypes.Value{}
		for name, value := range stored {
			planned[name] = value
		}
		for name, value := range changes {
			planned[name] = value
		}
		raw := federationValues(planned)
		req := resource.ModifyPlanRequest{
			Config: tfsdk.Config{Raw: raw, Schema: federationTestSchema},
			Plan:   tfsdk.Plan{Raw: raw, Schema: federationTestSchema},
			State:  tfsdk.State{Raw: federationValues(stored), Schema: federationTestSchema},
		}
		resp := resource.ModifyPlanResponse{Plan: req.Plan}
		federationMappingValidator{}.ValidatePlan(context.Background(), req, &resp, nil)
		return resp
	}

	if got := errorList(update(map[string]tftypes.Value{"saml_sign_algorithm": str("SHA256")}).Diagnostics); len(got) != 0 {
		t.Errorf("an unrelated change: errors = %q, want none", got)
	}
	got := errorList(update(map[string]tftypes.Value{"map_user_create_if_not_found": tftypes.NewValue(tftypes.Bool, true)}).Diagnostics)
	if want := []string{"map_user_create_if_not_found: User Creation Needs Required Mapping"}; !slices.Equal(got, want) {
		t.Errorf("turning on user creation: errors = %q, want %q", got, want)
	}

	// Leaving Required mode with user creation on: the server would silently stop creating users.
	stored["map_users"] = str("Required")
	stored["map_user_create_if_not_found"] = tftypes.NewValue(tftypes.Bool, true)
	got = errorList(update(map[string]tftypes.Value{"map_users": str("Optional")}).Diagnostics)
	if want := []string{"map_user_create_if_not_found: User Creation Needs Required Mapping"}; !slices.Equal(got, want) {
		t.Errorf("leaving Required mode with user creation on: errors = %q, want %q", got, want)
	}
	stored["map_users"] = str("Optional")

	// Stored by the portal with the option on in Optional mode, as an import reads it.
	stored["map_user_create_if_not_found"] = tftypes.NewValue(tftypes.Bool, true)
	if got := errorList(update(map[string]tftypes.Value{"saml_sign_algorithm": str("SHA256")}).Diagnostics); len(got) != 0 {
		t.Errorf("an update of a portal-made federation: errors = %q, want none", got)
	}
}

func TestFederationSAMLPlanModifier(t *testing.T) {
	stored := map[string]tftypes.Value{
		"federation_type":                   str("SAML 2.0"),
		"saml_sign_in_url":                  str("https://idp.example.com/sso"),
		"saml_logout_url":                   str("https://idp.example.com/slo"),
		"saml_metadata_xml_write_only_hash": str("hash-1"),
		"saml_idp_certificate_thumbprint":   str("AB12"),
		"saml_idp_certificate_subject":      str("CN=idp"),
		"saml_idp_certificate_expires":      str("2030-01-01"),
	}
	with := func(base map[string]tftypes.Value, set map[string]tftypes.Value) map[string]tftypes.Value {
		out := map[string]tftypes.Value{}
		for name, value := range base {
			out[name] = value
		}
		for name, value := range set {
			out[name] = value
		}
		return out
	}
	unknown := tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	derived := append(append([]string{}, samlCertificateAttributes...), samlDerivedURLAttributes...)
	tests := []struct {
		name        string
		state       map[string]tftypes.Value
		plan        map[string]tftypes.Value
		config      map[string]tftypes.Value
		wantUnknown []string
	}{
		{name: "no change", state: stored, plan: stored},
		{name: "rename keeps the document's values", state: stored,
			plan: with(stored, map[string]tftypes.Value{"saml_sign_algorithm": str("SHA1")})},
		{name: "new document", state: stored,
			plan: with(stored, map[string]tftypes.Value{"saml_metadata_xml_write_only_hash": unknown}), wantUnknown: derived},
		{name: "new certificate keeps the configured URLs", state: stored,
			plan:        with(stored, map[string]tftypes.Value{"saml_signing_certificate_write_only_hash": str("hash-2")}),
			config:      map[string]tftypes.Value{"saml_sign_in_url": str("https://idp.example.com/sso"), "saml_logout_url": str("https://idp.example.com/slo")},
			wantUnknown: samlCertificateAttributes},
		{name: "any update refetches the metadata URL",
			state:       with(stored, map[string]tftypes.Value{"saml_metadata_url": str("https://idp.example.com/metadata")}),
			plan:        with(stored, map[string]tftypes.Value{"saml_metadata_url": str("https://idp.example.com/metadata"), "saml_sign_algorithm": str("SHA1")}),
			wantUnknown: derived},
		{name: "OIDC is untouched",
			state: with(stored, map[string]tftypes.Value{"federation_type": str("OIDC")}),
			plan:  with(stored, map[string]tftypes.Value{"federation_type": str("OIDC"), "saml_metadata_xml_write_only_hash": unknown})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := tfsdk.Plan{Raw: federationValues(tt.plan), Schema: federationTestSchema}
			req := resource.ModifyPlanRequest{
				Config: tfsdk.Config{Raw: federationValues(tt.config), Schema: federationTestSchema},
				Plan:   plan,
				State:  tfsdk.State{Raw: federationValues(tt.state), Schema: federationTestSchema},
			}
			resp := resource.ModifyPlanResponse{Plan: plan}
			federationSAMLPlanModifier{}.ValidatePlan(context.Background(), req, &resp, nil)
			if resp.Diagnostics.HasError() {
				t.Fatalf("errors: %v", resp.Diagnostics)
			}
			for _, attribute := range derived {
				var value types.String
				resp.Diagnostics.Append(resp.Plan.GetAttribute(context.Background(), path.Root(attribute), &value)...)
				if want := slices.Contains(tt.wantUnknown, attribute); value.IsUnknown() != want {
					t.Errorf("%s unknown = %v, want %v", attribute, value.IsUnknown(), want)
				}
			}
		})
	}
}
