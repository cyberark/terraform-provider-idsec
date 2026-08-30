// Copyright CyberArk. 2026
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/cyberark/terraform-provider-idsec/internal/actions"
)

// stringObjectValue builds an all-string tftypes.Object from the exact values supplied, so a
// caller can pass a null or unknown value for a key rather than only a known one.
func stringObjectValue(fields map[string]tftypes.Value) tftypes.Value {
	attrTypes := make(map[string]tftypes.Type, len(fields))
	for k := range fields {
		attrTypes[k] = tftypes.String
	}
	return tftypes.NewValue(tftypes.Object{AttributeTypes: attrTypes}, fields)
}

func knownString(s string) tftypes.Value { return tftypes.NewValue(tftypes.String, s) }
func nullString() tftypes.Value          { return tftypes.NewValue(tftypes.String, nil) }
func unknownString() tftypes.Value       { return tftypes.NewValue(tftypes.String, tftypes.UnknownValue) }

func writeOnlyTestActionDef(writeOnlyAttrs map[string]string) *actions.IdsecServiceTerraformResourceActionDefinition {
	return &actions.IdsecServiceTerraformResourceActionDefinition{
		WriteOnlyAttributes: writeOnlyAttrs,
	}
}

// TestIdsecResource_triggerFired covers create, changed, unchanged, unknown, and unwalkable
// triggers. triggerFired reads only its plan and state arguments, so a zero-value IdsecResource is
// enough to drive it.
func TestIdsecResource_triggerFired(t *testing.T) {
	t.Parallel()

	nestedTypes := map[string]tftypes.Type{
		"meta": tftypes.Object{AttributeTypes: map[string]tftypes.Type{"name": tftypes.String}},
	}
	nestedObj := func(name string) tftypes.Value {
		return tftypes.NewValue(tftypes.Object{AttributeTypes: nestedTypes}, map[string]tftypes.Value{
			"meta": tftypes.NewValue(nestedTypes["meta"], map[string]tftypes.Value{"name": knownString(name)}),
		})
	}

	listTypes := map[string]tftypes.Type{"tags": tftypes.List{ElementType: tftypes.String}}
	listObj := func(tags ...string) tftypes.Value {
		elems := make([]tftypes.Value, len(tags))
		for i, tag := range tags {
			elems[i] = knownString(tag)
		}
		return tftypes.NewValue(tftypes.Object{AttributeTypes: listTypes}, map[string]tftypes.Value{
			"tags": tftypes.NewValue(listTypes["tags"], elems),
		})
	}
	trigger := func(v tftypes.Value) tftypes.Value {
		return stringObjectValue(map[string]tftypes.Value{"trigger": v})
	}

	tests := []struct {
		name        string
		triggerPath string
		plan        *tfsdk.Plan
		state       *tfsdk.State
		wantFired   bool
		wantErr     bool
	}{
		{
			name:        "success_create_fires_with_nil_state",
			triggerPath: "trigger",
			plan:        &tfsdk.Plan{Raw: trigger(knownString("v"))},
			wantFired:   true,
		},
		{
			name:        "success_create_fires_with_null_state_raw",
			triggerPath: "trigger",
			plan:        &tfsdk.Plan{Raw: trigger(knownString("v"))},
			state:       &tfsdk.State{Raw: tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"trigger": tftypes.String}}, nil)},
			wantFired:   true,
		},
		{
			name:        "success_changed_trigger_fires",
			triggerPath: "trigger",
			plan:        &tfsdk.Plan{Raw: trigger(knownString("new"))},
			state:       &tfsdk.State{Raw: trigger(knownString("old"))},
			wantFired:   true,
		},
		{
			name:        "success_nested_trigger_fires",
			triggerPath: "meta.name",
			plan:        &tfsdk.Plan{Raw: nestedObj("new")},
			state:       &tfsdk.State{Raw: nestedObj("old")},
			wantFired:   true,
		},
		{
			name:        "edge_case_reordered_list_trigger_fires",
			triggerPath: "tags",
			plan:        &tfsdk.Plan{Raw: listObj("b", "a")},
			state:       &tfsdk.State{Raw: listObj("a", "b")},
			wantFired:   true,
		},
		{
			name:        "edge_case_unknown_plan_value_fires",
			triggerPath: "trigger",
			plan:        &tfsdk.Plan{Raw: trigger(unknownString())},
			state:       &tfsdk.State{Raw: trigger(knownString("anything"))},
			wantFired:   true,
		},
		{
			name:        "edge_case_unchanged_does_not_fire",
			triggerPath: "trigger",
			plan:        &tfsdk.Plan{Raw: trigger(knownString("same"))},
			state:       &tfsdk.State{Raw: trigger(knownString("same"))},
			wantFired:   false,
		},
		{
			name:        "edge_case_null_on_both_sides_does_not_fire",
			triggerPath: "trigger",
			plan:        &tfsdk.Plan{Raw: trigger(nullString())},
			state:       &tfsdk.State{Raw: trigger(nullString())},
			wantFired:   false,
		},
		{
			name:        "error_trigger_path_not_addressable",
			triggerPath: "does_not_exist",
			plan:        &tfsdk.Plan{Raw: trigger(knownString("v"))},
			state:       &tfsdk.State{Raw: trigger(knownString("v"))},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &IdsecResource{}
			fired, err := s.triggerFired(context.Background(), "write_only_attr", tt.triggerPath, tt.plan, tt.state)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got nil (fired=%v)", fired)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if fired != tt.wantFired {
				t.Errorf("expected fired=%v, got %v", tt.wantFired, fired)
			}
		})
	}
}

// These mirror the request-model shapes FieldByAttributePath has to resolve through: a field
// promoted from a squashed embed, and a direct field shadowing an identically-named embedded one.
type woEmbedded struct {
	Secret string `mapstructure:"secret"`
}

type woSquashModel struct {
	woEmbedded `mapstructure:",squash"`
	Username   string `mapstructure:"username"`
}

type woShadowModel struct {
	woEmbedded `mapstructure:",squash"`
	Secret     string `mapstructure:"secret"`
	Token      string `mapstructure:"token"`
}

// TestIdsecResource_applyWriteOnlyValues covers injection of a config-only value into the request
// model, including through a squashed embed and a shadowing direct field, and the trigger gate.
func TestIdsecResource_applyWriteOnlyValues(t *testing.T) {
	t.Parallel()

	t.Run("success_create_injects_into_squashed_embed_field", func(t *testing.T) {
		t.Parallel()
		s := &IdsecResource{actionDefinition: writeOnlyTestActionDef(map[string]string{"secret": "trigger"})}
		target := &woSquashModel{}
		config := &tfsdk.Config{Raw: stringObjectValue(map[string]tftypes.Value{"secret": knownString("s3cr3t")})}

		if err := s.applyWriteOnlyValues(context.Background(), target, config, nil, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if target.Secret != "s3cr3t" {
			t.Errorf("expected Secret to be %q, got %q", "s3cr3t", target.Secret)
		}
	})

	t.Run("success_update_injects_into_shadowing_field_when_trigger_changed", func(t *testing.T) {
		t.Parallel()
		s := &IdsecResource{actionDefinition: writeOnlyTestActionDef(map[string]string{"secret": "trigger"})}
		target := &woShadowModel{}
		config := &tfsdk.Config{Raw: stringObjectValue(map[string]tftypes.Value{"secret": knownString("new-secret")})}
		plan := &tfsdk.Plan{Raw: stringObjectValue(map[string]tftypes.Value{"trigger": knownString("new")})}
		state := &tfsdk.State{Raw: stringObjectValue(map[string]tftypes.Value{"trigger": knownString("old")})}

		if err := s.applyWriteOnlyValues(context.Background(), target, config, plan, state); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if target.Secret != "new-secret" {
			t.Errorf("expected the shadowing Secret to be %q, got %q", "new-secret", target.Secret)
		}
		if target.woEmbedded.Secret != "" {
			t.Errorf("expected the embedded Secret to stay untouched, got %q", target.woEmbedded.Secret)
		}
	})

	t.Run("success_update_skips_when_trigger_unchanged", func(t *testing.T) {
		t.Parallel()
		s := &IdsecResource{actionDefinition: writeOnlyTestActionDef(map[string]string{"secret": "trigger"})}
		target := &woShadowModel{}
		config := &tfsdk.Config{Raw: stringObjectValue(map[string]tftypes.Value{"secret": knownString("unchanged")})}
		plan := &tfsdk.Plan{Raw: stringObjectValue(map[string]tftypes.Value{"trigger": knownString("same")})}
		state := &tfsdk.State{Raw: stringObjectValue(map[string]tftypes.Value{"trigger": knownString("same")})}

		if err := s.applyWriteOnlyValues(context.Background(), target, config, plan, state); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if target.Secret != "" {
			t.Errorf("expected Secret to stay unset when the trigger did not fire, got %q", target.Secret)
		}
	})

	t.Run("success_two_keys_both_injected", func(t *testing.T) {
		t.Parallel()
		s := &IdsecResource{actionDefinition: writeOnlyTestActionDef(map[string]string{
			"secret": "secret_trigger",
			"token":  "token_trigger",
		})}
		target := &woShadowModel{}
		config := &tfsdk.Config{Raw: stringObjectValue(map[string]tftypes.Value{
			"secret": knownString("multi-secret"),
			"token":  knownString("multi-token"),
		})}

		if err := s.applyWriteOnlyValues(context.Background(), target, config, nil, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if target.Secret != "multi-secret" || target.Token != "multi-token" {
			t.Errorf("expected both values injected, got Secret=%q Token=%q", target.Secret, target.Token)
		}
	})

	t.Run("error_declared_path_has_no_request_model_field", func(t *testing.T) {
		t.Parallel()
		s := &IdsecResource{actionDefinition: writeOnlyTestActionDef(map[string]string{"does_not_exist": "trigger"})}
		config := &tfsdk.Config{Raw: stringObjectValue(map[string]tftypes.Value{"does_not_exist": knownString("v")})}

		if err := s.applyWriteOnlyValues(context.Background(), &woSquashModel{}, config, nil, nil); err == nil {
			t.Fatal("expected an error for a write-only path with no backing request-model field")
		}
	})

	t.Run("error_declared_path_absent_from_config", func(t *testing.T) {
		t.Parallel()
		s := &IdsecResource{actionDefinition: writeOnlyTestActionDef(map[string]string{"secret": "trigger"})}
		config := &tfsdk.Config{Raw: stringObjectValue(map[string]tftypes.Value{"other_field": knownString("v")})}

		err := s.applyWriteOnlyValues(context.Background(), &woSquashModel{}, config, nil, nil)
		if err == nil {
			t.Fatal("expected an error for a write-only path not present in configuration")
		}
		if !strings.Contains(err.Error(), "secret") {
			t.Errorf("expected the error to name the attribute path, got: %v", err)
		}
	})
}

// TestIdsecResource_applyWriteOnlyValues_NoOps covers every condition under which nothing is
// injected and no error is raised.
func TestIdsecResource_applyWriteOnlyValues_NoOps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		writeOnly map[string]string
		config    *tfsdk.Config
		nilTarget bool
	}{
		{
			name:      "null_config_value",
			writeOnly: map[string]string{"secret": "trigger"},
			config:    &tfsdk.Config{Raw: stringObjectValue(map[string]tftypes.Value{"secret": nullString()})},
		},
		{
			name:      "unknown_config_value",
			writeOnly: map[string]string{"secret": "trigger"},
			config:    &tfsdk.Config{Raw: stringObjectValue(map[string]tftypes.Value{"secret": unknownString()})},
		},
		{
			name:      "no_write_only_attributes_declared",
			writeOnly: nil,
			config:    &tfsdk.Config{Raw: stringObjectValue(map[string]tftypes.Value{"secret": knownString("never-read")})},
		},
		{
			name:      "nil_config",
			writeOnly: map[string]string{"secret": "trigger"},
			config:    nil,
		},
		{
			name:      "nil_target",
			writeOnly: map[string]string{"secret": "trigger"},
			config:    &tfsdk.Config{Raw: stringObjectValue(map[string]tftypes.Value{"secret": knownString("s3cr3t")})},
			nilTarget: true,
		},
	}

	for _, tt := range tests {
		t.Run("edge_case_"+tt.name+"_is_noop", func(t *testing.T) {
			t.Parallel()
			s := &IdsecResource{actionDefinition: writeOnlyTestActionDef(tt.writeOnly)}

			if tt.nilTarget {
				if err := s.applyWriteOnlyValues(context.Background(), nil, tt.config, nil, nil); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			target := &woSquashModel{}
			if err := s.applyWriteOnlyValues(context.Background(), target, tt.config, nil, nil); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if target.Secret != "" {
				t.Errorf("expected Secret to stay unset, got %q", target.Secret)
			}
		})
	}
}

// TestIdsecResource_parsePlanAndState_createConfigViewIsInert is the evidence that passing a real
// config into triggerOperation on Create cannot change behavior for a resource that declares no
// write-only attribute: the request model comes out byte-identical either way.
func TestIdsecResource_parsePlanAndState_createConfigViewIsInert(t *testing.T) {
	t.Parallel()

	type inertCreateModel struct {
		Name string `json:"name,omitempty" mapstructure:"name"`
	}

	actionDef := &actions.IdsecServiceTerraformResourceActionDefinition{
		IdsecServiceBaseTerraformActionDefinition: actions.IdsecServiceBaseTerraformActionDefinition{
			IdsecServiceBaseActionDefinition: actions.IdsecServiceBaseActionDefinition{
				ActionName: "test-action",
				Schemas: map[string]interface{}{
					"create-action": inertCreateModel{},
				},
			},
		},
		SupportedOperations: []actions.IdsecServiceActionOperation{actions.CreateOperation},
		ActionsMappings: map[actions.IdsecServiceActionOperation]string{
			actions.CreateOperation: "create-action",
		},
		WriteOnlyAttributes: map[string]string{},
	}
	s := &IdsecResource{actionDefinition: actionDef}

	plan := &tfsdk.Plan{
		Schema: schema.Schema{Attributes: map[string]schema.Attribute{"name": schema.StringAttribute{Optional: true}}},
		Raw:    stringObjectValue(map[string]tftypes.Value{"name": knownString("policy-1")}),
	}
	populatedConfig := &tfsdk.Config{
		Schema: schema.Schema{Attributes: map[string]schema.Attribute{
			"name":   schema.StringAttribute{Optional: true},
			"secret": schema.StringAttribute{Optional: true, WriteOnly: true},
		}},
		Raw: stringObjectValue(map[string]tftypes.Value{
			"name":   nullString(),
			"secret": knownString("this-should-never-be-read"),
		}),
	}

	ctx := context.Background()
	marshalled := make([][]byte, 0, 2)
	for _, config := range []*tfsdk.Config{nil, populatedConfig} {
		var diags diag.Diagnostics
		result, err := s.parsePlanAndState(ctx, actions.CreateOperation, &diags, plan, nil, config, nil)
		if err != nil || diags.HasError() {
			t.Fatalf("unexpected failure (config != nil: %v): err=%v diags=%v", config != nil, err, diags.Errors())
		}
		encoded, err := json.Marshal(result)
		if err != nil {
			t.Fatalf("failed to marshal request model: %v", err)
		}
		marshalled = append(marshalled, encoded)
	}

	if !bytes.Equal(marshalled[0], marshalled[1]) {
		t.Errorf("expected byte-identical request models, got %s vs %s", marshalled[0], marshalled[1])
	}
}
