// Copyright CyberArk. 2026
// SPDX-License-Identifier: Apache-2.0

package schemas

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// writeOnlyTriggerConfig builds a {secret, trigger} config from the raw values given. rawNull makes
// the whole config Raw null, simulating a null resource configuration block.
func writeOnlyTriggerConfig(secret, trigger tftypes.Value, rawNull bool) tfsdk.Config {
	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"secret":  schema.StringAttribute{Optional: true, WriteOnly: true},
			"trigger": schema.StringAttribute{Optional: true},
		},
	}
	objType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"secret":  tftypes.String,
		"trigger": tftypes.String,
	}}

	if rawNull {
		return tfsdk.Config{Schema: s, Raw: tftypes.NewValue(objType, nil)}
	}
	return tfsdk.Config{Schema: s, Raw: tftypes.NewValue(objType, map[string]tftypes.Value{
		"secret":  secret,
		"trigger": trigger,
	})}
}

// TestWriteOnlyTriggerValidator_ValidateString covers when the validator warns and when it stays
// quiet. Warning and error counts are both asserted: a validator that emits nothing, or an error
// instead of a warning, would slip past a looser check.
func TestWriteOnlyTriggerValidator_ValidateString(t *testing.T) {
	t.Parallel()

	str := func(s string) tftypes.Value { return tftypes.NewValue(tftypes.String, s) }
	null := tftypes.NewValue(tftypes.String, nil)
	unknown := tftypes.NewValue(tftypes.String, tftypes.UnknownValue)

	tests := []struct {
		name         string
		secret       tftypes.Value
		trigger      tftypes.Value
		rawNull      bool
		triggerPath  string
		wantWarnings int
	}{
		{
			name:    "success_trigger_present_is_quiet",
			secret:  str("s3cr3t"),
			trigger: str("v2"),
		},
		{
			name:         "warning_trigger_absent",
			secret:       str("s3cr3t"),
			trigger:      null,
			wantWarnings: 1,
		},
		{
			name:    "edge_case_write_only_value_null_is_quiet",
			secret:  null,
			trigger: null,
		},
		{
			name:    "edge_case_write_only_value_unknown_is_quiet",
			secret:  unknown,
			trigger: null,
		},
		{
			name:    "edge_case_null_config_is_quiet",
			secret:  null,
			trigger: null,
			rawNull: true,
		},
		{
			name:        "edge_case_unresolvable_trigger_path_is_quiet",
			secret:      str("s3cr3t"),
			trigger:     str("v2"),
			triggerPath: "trigger.nested.does.not.exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			triggerPath := tt.triggerPath
			if triggerPath == "" {
				triggerPath = "trigger"
			}

			// The framework derives ConfigValue from the raw value, so mirror it here.
			configValue := types.StringNull()
			switch {
			case !tt.secret.IsFullyKnown():
				configValue = types.StringUnknown()
			case !tt.secret.IsNull():
				var s string
				if err := tt.secret.As(&s); err != nil {
					t.Fatalf("failed to read the secret fixture: %v", err)
				}
				configValue = types.StringValue(s)
			}

			resp := &validator.StringResponse{}
			WriteOnlyTriggerValidator{TriggerPath: triggerPath}.ValidateString(
				context.Background(),
				validator.StringRequest{
					Path:        path.Root("secret"),
					ConfigValue: configValue,
					Config:      writeOnlyTriggerConfig(tt.secret, tt.trigger, tt.rawNull),
				},
				resp,
			)

			if got := len(resp.Diagnostics.Warnings()); got != tt.wantWarnings {
				t.Errorf("warnings = %d, want %d (diags: %v)", got, tt.wantWarnings, resp.Diagnostics)
			}
			if got := len(resp.Diagnostics.Errors()); got != 0 {
				t.Errorf("errors = %d, want 0 (diags: %v)", got, resp.Diagnostics)
			}
		})
	}
}

// TestWriteOnlyTriggerValidator_DescriptionMethods checks both descriptions name the configured
// trigger, which is the validator's only practitioner-visible surface absent a warning.
func TestWriteOnlyTriggerValidator_DescriptionMethods(t *testing.T) {
	t.Parallel()

	v := WriteOnlyTriggerValidator{TriggerPath: "secret_rotation_trigger"}
	ctx := context.Background()

	if got := v.Description(ctx); !strings.Contains(got, v.TriggerPath) {
		t.Errorf("Description() does not name the trigger path: %q", got)
	}
	if got := v.MarkdownDescription(ctx); !strings.Contains(got, v.TriggerPath) {
		t.Errorf("MarkdownDescription() does not name the trigger path: %q", got)
	}
}
