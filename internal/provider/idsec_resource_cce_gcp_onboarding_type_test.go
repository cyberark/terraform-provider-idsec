// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestCCEGCPResources_OnboardingTypeIsOptInOwnership covers the contract the SDK update relies on: users may
// set onboarding_type only to terraform_provider, which is how they opt in to converting a programmatic
// onboarding, and leaving it out keeps the value read from CCE.
func TestCCEGCPResources_OnboardingTypeIsOptInOwnership(t *testing.T) {
	for _, actionName := range []string{"cce-gcp-project", "cce-gcp-organization"} {
		t.Run(actionName, func(t *testing.T) {
			attr := gcpOnboardingTypeAttribute(t, actionName)

			if !attr.Optional || !attr.Computed || attr.Required {
				t.Fatalf("onboarding_type must be Optional and Computed, got Optional=%v Computed=%v Required=%v",
					attr.Optional, attr.Computed, attr.Required)
			}
			if len(attr.Validators) == 0 {
				t.Fatal("onboarding_type has no validators, so any onboarding type could be requested")
			}
			for value, wantErr := range map[string]bool{
				"terraform_provider": false,
				"programmatic":       true,
				"standard":           true,
			} {
				if got := validateString(attr.Validators, value); got != wantErr {
					t.Errorf("onboarding_type = %q: got error %v, want %v", value, got, wantErr)
				}
			}
		})
	}
}

func gcpOnboardingTypeAttribute(t *testing.T, actionName string) schema.StringAttribute {
	t.Helper()
	for _, def := range (&IdsecProvider{}).collectTfResources() {
		if def.Second.ActionName != actionName {
			continue
		}
		var resp resource.SchemaResponse
		NewIdsecResource(def.First, def.Second).Schema(context.Background(), resource.SchemaRequest{}, &resp)
		if resp.Diagnostics.HasError() {
			t.Fatalf("schema for %s: %v", actionName, resp.Diagnostics)
		}
		attr, ok := resp.Schema.Attributes["onboarding_type"].(schema.StringAttribute)
		if !ok {
			t.Fatalf("%s has no string onboarding_type attribute", actionName)
		}
		return attr
	}
	t.Fatalf("resource %s is not registered", actionName)
	return schema.StringAttribute{}
}

func validateString(validators []validator.String, value string) bool {
	resp := &validator.StringResponse{}
	for _, v := range validators {
		v.ValidateString(context.Background(), validator.StringRequest{ConfigValue: types.StringValue(value)}, resp)
	}
	return resp.Diagnostics.HasError()
}
