// Copyright CyberArk. 2026
// SPDX-License-Identifier: Apache-2.0

package schemas

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Every model pairs a candidate write-only field with a plain "rotate" string field acting as an
// ordinary, always-valid trigger, unless the test points the trigger somewhere else.

type woStringModel struct {
	Secret string `mapstructure:"secret" desc:"a secret"`
	Rotate string `mapstructure:"rotate" desc:"rotate trigger"`
}

type woBoolModel struct {
	Flag   bool   `mapstructure:"flag" desc:"a flag"`
	Rotate string `mapstructure:"rotate" desc:"rotate trigger"`
}

type woInt64Model struct {
	Num    int64  `mapstructure:"num" desc:"a number"`
	Rotate string `mapstructure:"rotate" desc:"rotate trigger"`
}

type woListModel struct {
	Items  []string `mapstructure:"items" desc:"items"`
	Rotate string   `mapstructure:"rotate" desc:"rotate trigger"`
}

type woMapModel struct {
	Values map[string]string `mapstructure:"values" desc:"values"`
	Rotate string            `mapstructure:"rotate" desc:"rotate trigger"`
}

type woDynamicModel struct {
	Anything interface{} `mapstructure:"anything" desc:"anything"`
	Rotate   string      `mapstructure:"rotate" desc:"rotate trigger"`
}

type woRequiredModel struct {
	Secret string `mapstructure:"secret" required:"true" desc:"a secret"`
	Rotate string `mapstructure:"rotate" desc:"rotate trigger"`
}

type woNestedChild struct {
	A string `mapstructure:"a" desc:"a"`
	B string `mapstructure:"b" desc:"b"`
}

type woNestedSingleModel struct {
	Creds  woNestedChild `mapstructure:"creds" desc:"creds"`
	Rotate string        `mapstructure:"rotate" desc:"rotate trigger"`
}

type woNestedListModel struct {
	Creds  []woNestedChild `mapstructure:"creds" desc:"creds"`
	Rotate string          `mapstructure:"rotate" desc:"rotate trigger"`
}

type woNestedMapModel struct {
	Creds  map[string]woNestedChild `mapstructure:"creds" desc:"creds"`
	Rotate string                   `mapstructure:"rotate" desc:"rotate trigger"`
}

type woDefaultTagModel struct {
	Secret string `mapstructure:"secret" default:"x" desc:"a secret"`
	Rotate string `mapstructure:"rotate" desc:"rotate trigger"`
}

type woNestedWithSetChild struct {
	Items []string `mapstructure:"items" desc:"items"`
}

type woNestedDescendantSetModel struct {
	Creds  woNestedWithSetChild `mapstructure:"creds" desc:"creds"`
	Rotate string               `mapstructure:"rotate" desc:"rotate trigger"`
}

type woNestedWithDefaultChild struct {
	Mode string `mapstructure:"mode" default:"x" desc:"mode"`
}

type woNestedDescendantDefaultModel struct {
	Creds  woNestedWithDefaultChild `mapstructure:"creds" desc:"creds"`
	Rotate string                   `mapstructure:"rotate" desc:"rotate trigger"`
}

type woAncestorChild struct {
	Secret string `mapstructure:"secret" desc:"a secret"`
}

type woAncestorModel struct {
	Container woAncestorChild `mapstructure:"container" desc:"container"`
	Rotate    string          `mapstructure:"rotate" desc:"rotate trigger"`
}

type woValidatorsModel struct {
	Secret string `mapstructure:"secret" desc:"a secret" choices:"a,b,c"`
	Rotate string `mapstructure:"rotate" desc:"rotate trigger"`
}

type woTwoSecretsModel struct {
	SecretA string `mapstructure:"secret_a" desc:"secret a"`
	SecretB string `mapstructure:"secret_b" desc:"secret b"`
	Rotate  string `mapstructure:"rotate" desc:"rotate trigger"`
}

type woTriggerDefaultModel struct {
	Secret string `mapstructure:"secret" desc:"a secret"`
	Rotate string `mapstructure:"rotate" default:"v1" desc:"rotate trigger"`
}

type woTriggerSetModel struct {
	Secret string   `mapstructure:"secret" desc:"a secret"`
	Rotate []string `mapstructure:"rotate" desc:"rotate trigger"`
}

type woSharedTriggerModel struct {
	SecretA string `mapstructure:"secret_a" desc:"secret a"`
	SecretB string `mapstructure:"secret_b" desc:"secret b"`
}

type woCaseCollisionModel struct {
	Secret     string `mapstructure:"secret" desc:"a secret"`
	SecretType string `mapstructure:"secret_type" desc:"secret type"`
}

type woDroppedFieldModel struct {
	Secret  string      `mapstructure:"secret" desc:"a secret"`
	Dropped chan string `mapstructure:"dropped_field" desc:"dropped"`
}

// ---- Helpers ----

func stringAttr(t *testing.T, attrs map[string]schema.Attribute, name string) schema.StringAttribute {
	t.Helper()
	a, ok := attrs[name]
	if !ok {
		t.Fatalf("expected attribute %q to exist", name)
	}
	sa, ok := a.(schema.StringAttribute)
	if !ok {
		t.Fatalf("expected attribute %q to be a StringAttribute, got %T", name, a)
	}
	return sa
}

func requireNoErrors(t *testing.T, diags diag.Diagnostics) {
	t.Helper()
	if diags.HasError() {
		t.Fatalf("expected no error diagnostics, got: %v", diags.Errors())
	}
}

// requireSingleError fails the test unless diags contains exactly one error whose detail contains
// substr. Only a short, stable fragment is matched: asserting on the full prose would make every
// wording change a test failure.
func requireSingleError(t *testing.T, diags diag.Diagnostics, substr string) {
	t.Helper()
	errs := diags.Errors()
	if len(errs) != 1 {
		t.Fatalf("expected exactly one error diagnostic, got %d: %v", len(errs), errs)
	}
	if errs[0].Summary() != errWriteOnlySummary {
		t.Errorf("expected summary %q, got %q", errWriteOnlySummary, errs[0].Summary())
	}
	if !strings.Contains(errs[0].Detail(), substr) {
		t.Errorf("expected error detail to contain %q, got: %s", substr, errs[0].Detail())
	}
}

// assertAttributeUnchanged compares attributes by their query methods rather than
// reflect.DeepEqual, which is unsafe here: plan modifiers embed func literals, and Go considers
// any two non-nil func values unequal even when behaviorally identical.
func assertAttributeUnchanged(t *testing.T, name string, baseline, got schema.Attribute) {
	t.Helper()
	if !baseline.GetType().Equal(got.GetType()) {
		t.Errorf("%s: type changed, baseline=%v got=%v", name, baseline.GetType(), got.GetType())
	}
	if baseline.IsRequired() != got.IsRequired() || baseline.IsOptional() != got.IsOptional() ||
		baseline.IsComputed() != got.IsComputed() || baseline.IsSensitive() != got.IsSensitive() ||
		baseline.IsWriteOnly() != got.IsWriteOnly() {
		t.Errorf("%s: flags changed, baseline=%+v got=%+v", name, baseline, got)
	}
	if baseline.GetDescription() != got.GetDescription() {
		t.Errorf("%s: description changed, baseline=%q got=%q", name, baseline.GetDescription(), got.GetDescription())
	}
}

// TestApplyWriteOnlyAttributes_TypeCoverage marks one attribute of every type the generator emits
// and checks it came back write-only and no longer computed, including every descendant of a
// nested container.
func TestApplyWriteOnlyAttributes_TypeCoverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		model  interface{}
		attr   string
		nested bool
	}{
		{name: "string", model: &woStringModel{}, attr: "secret"},
		{name: "bool", model: &woBoolModel{}, attr: "flag"},
		{name: "int64", model: &woInt64Model{}, attr: "num"},
		{name: "list", model: &woListModel{}, attr: "items"},
		{name: "map", model: &woMapModel{}, attr: "values"},
		{name: "dynamic", model: &woDynamicModel{}, attr: "anything"},
		{name: "nested_single", model: &woNestedSingleModel{}, attr: "creds", nested: true},
		{name: "nested_list", model: &woNestedListModel{}, attr: "creds", nested: true},
		{name: "nested_map", model: &woNestedMapModel{}, attr: "creds", nested: true},
	}

	for _, tt := range tests {
		t.Run("success_"+tt.name+"_marked_write_only", func(t *testing.T) {
			t.Parallel()
			s, diags := generateResourceSchema(resourceSchemaOptions{
				CreateModel:         tt.model,
				WriteOnlyAttributes: map[string]string{tt.attr: "rotate"},
			})
			requireNoErrors(t, diags)

			a, ok := s.Attributes[tt.attr]
			if !ok {
				t.Fatalf("expected attribute %q to exist", tt.attr)
			}
			if !a.IsWriteOnly() || a.IsComputed() {
				t.Errorf("expected %s to be WriteOnly and not Computed, got %+v", tt.attr, a)
			}

			if !tt.nested {
				return
			}
			children, ok := writeOnlyNestedChildren(a)
			if !ok {
				t.Fatalf("expected %s to be a nested container, got %T", tt.attr, a)
			}
			for name, child := range children {
				if !child.IsWriteOnly() || child.IsComputed() {
					t.Errorf("expected %s.%s to be WriteOnly and not Computed, got %+v", tt.attr, name, child)
				}
			}
		})
	}
}

// TestApplyWriteOnlyAttributes_Rejections covers every declaration this pass refuses, on both the
// key and its trigger.
func TestApplyWriteOnlyAttributes_Rejections(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		opts  resourceSchemaOptions
		error string
	}{
		{
			name:  "key_not_in_schema",
			opts:  resourceSchemaOptions{CreateModel: &woStringModel{}, WriteOnlyAttributes: map[string]string{"nope": "rotate"}},
			error: "does not exist in the generated schema",
		},
		{
			name:  "key_computed_only",
			opts:  resourceSchemaOptions{CreateModel: &woStringModel{}, ComputedAttributes: []string{"secret"}, WriteOnlyAttributes: map[string]string{"secret": "rotate"}},
			error: "computed-only",
		},
		{
			name:  "key_under_computed_ancestor",
			opts:  resourceSchemaOptions{CreateModel: &woAncestorModel{}, WriteOnlyAttributes: map[string]string{"container.secret": "rotate"}},
			error: "is Computed",
		},
		{
			name:  "key_set_typed",
			opts:  resourceSchemaOptions{CreateModel: &woListModel{}, ComputedAsSetAttributes: []string{"items"}, WriteOnlyAttributes: map[string]string{"items": "rotate"}},
			error: "is set-typed",
		},
		{
			name:  "key_carries_default",
			opts:  resourceSchemaOptions{CreateModel: &woDefaultTagModel{}, WriteOnlyAttributes: map[string]string{"secret": "rotate"}},
			error: "carries a Default",
		},
		{
			name:  "key_in_immutable_attributes",
			opts:  resourceSchemaOptions{CreateModel: &woStringModel{}, ImmutableAttributes: []string{"secret"}, WriteOnlyAttributes: map[string]string{"secret": "rotate"}},
			error: "also listed in ImmutableAttributes",
		},
		{
			name:  "key_in_force_new_attributes",
			opts:  resourceSchemaOptions{CreateModel: &woStringModel{}, ForceNewAttributes: []string{"secret"}, WriteOnlyAttributes: map[string]string{"secret": "rotate"}},
			error: "also listed in ForceNewAttributes",
		},
		{
			name:  "descendant_set_typed",
			opts:  resourceSchemaOptions{CreateModel: &woNestedDescendantSetModel{}, ComputedAsSetAttributes: []string{"items"}, WriteOnlyAttributes: map[string]string{"creds": "rotate"}},
			error: "creds.items",
		},
		{
			name:  "descendant_carries_default",
			opts:  resourceSchemaOptions{CreateModel: &woNestedDescendantDefaultModel{}, WriteOnlyAttributes: map[string]string{"creds": "rotate"}},
			error: "creds.mode",
		},
		{
			// Marking a computed-only descendant would clear Computed without setting Optional,
			// leaving an attribute with none of Required, Optional or Computed, which the
			// framework rejects on every plan.
			name:  "descendant_computed_only",
			opts:  resourceSchemaOptions{CreateModel: &woNestedSingleModel{}, ComputedAttributes: []string{"creds.b"}, WriteOnlyAttributes: map[string]string{"creds": "rotate"}},
			error: "creds.b",
		},
		{
			name:  "trigger_empty",
			opts:  resourceSchemaOptions{CreateModel: &woStringModel{}, WriteOnlyAttributes: map[string]string{"secret": ""}},
			error: "empty trigger",
		},
		{
			name:  "trigger_is_the_key_itself",
			opts:  resourceSchemaOptions{CreateModel: &woStringModel{}, WriteOnlyAttributes: map[string]string{"secret": "secret"}},
			error: "its own trigger",
		},
		{
			name:  "trigger_computed_only",
			opts:  resourceSchemaOptions{CreateModel: &woStringModel{}, ComputedAttributes: []string{"rotate"}, WriteOnlyAttributes: map[string]string{"secret": "rotate"}},
			error: "is computed-only",
		},
		{
			name:  "trigger_is_another_write_only_key",
			opts:  resourceSchemaOptions{CreateModel: &woTwoSecretsModel{}, WriteOnlyAttributes: map[string]string{"secret_a": "secret_b", "secret_b": "rotate"}},
			error: "itself declared write-only",
		},
		{
			name:  "trigger_immutable",
			opts:  resourceSchemaOptions{CreateModel: &woStringModel{}, ImmutableAttributes: []string{"rotate"}, WriteOnlyAttributes: map[string]string{"secret": "rotate"}},
			error: "could never be rotated",
		},
		{
			name:  "trigger_set_typed",
			opts:  resourceSchemaOptions{CreateModel: &woTriggerSetModel{}, ComputedAsSetAttributes: []string{"rotate"}, WriteOnlyAttributes: map[string]string{"secret": "rotate"}},
			error: "is set-typed",
		},
		{
			name:  "trigger_unresolvable_and_dotted",
			opts:  resourceSchemaOptions{CreateModel: &woStringModel{}, WriteOnlyAttributes: map[string]string{"secret": "does.not.exist"}},
			error: "contains a dot",
		},
		{
			name:  "trigger_collides_with_existing_name",
			opts:  resourceSchemaOptions{CreateModel: &woCaseCollisionModel{}, WriteOnlyAttributes: map[string]string{"secret": "SecretType"}},
			error: "almost certainly a typo",
		},
		{
			name:  "trigger_matches_model_field_absent_from_schema",
			opts:  resourceSchemaOptions{CreateModel: &woDroppedFieldModel{}, WriteOnlyAttributes: map[string]string{"secret": "dropped_field"}},
			error: "absent from the generated schema",
		},
	}

	for _, tt := range tests {
		t.Run("error_"+tt.name, func(t *testing.T) {
			t.Parallel()
			_, diags := generateResourceSchema(tt.opts)
			requireSingleError(t, diags, tt.error)
		})
	}
}

// TestApplyWriteOnlyAttributes_UnmodifiedOnError checks that a rejected declaration leaves the
// schema exactly as it would have been without any write-only entry. This matters most for the
// set-typed case: a partial rewrite there would silently yield a valid, ordinary *persisted* set
// attribute holding a credential instead of an error.
func TestApplyWriteOnlyAttributes_UnmodifiedOnError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		base resourceSchemaOptions
		attr string
	}{
		{
			name: "set_typed_key",
			base: resourceSchemaOptions{CreateModel: &woListModel{}, ComputedAsSetAttributes: []string{"items"}},
			attr: "items",
		},
		{
			name: "immutable_conflict",
			base: resourceSchemaOptions{CreateModel: &woStringModel{}, ImmutableAttributes: []string{"secret"}},
			attr: "secret",
		},
		{
			name: "default_carrying_key",
			base: resourceSchemaOptions{CreateModel: &woDefaultTagModel{}},
			attr: "secret",
		},
	}

	for _, tt := range tests {
		t.Run("error_"+tt.name+"_leaves_attribute_untouched", func(t *testing.T) {
			t.Parallel()
			baseline, baseDiags := generateResourceSchema(tt.base)
			requireNoErrors(t, baseDiags)

			withWriteOnly := tt.base
			withWriteOnly.WriteOnlyAttributes = map[string]string{tt.attr: "rotate"}
			got, diags := generateResourceSchema(withWriteOnly)
			if !diags.HasError() {
				t.Fatal("expected an error diagnostic")
			}
			assertAttributeUnchanged(t, tt.attr, baseline.Attributes[tt.attr], got.Attributes[tt.attr])
		})
	}
}

// TestApplyWriteOnlyAttributes_MarkedAttributeShape covers what marking preserves and what it
// discards on the attribute a write-only key names.
func TestApplyWriteOnlyAttributes_MarkedAttributeShape(t *testing.T) {
	t.Parallel()

	t.Run("success_required_stays_required", func(t *testing.T) {
		t.Parallel()
		s, diags := generateResourceSchema(resourceSchemaOptions{
			CreateModel:         &woRequiredModel{},
			WriteOnlyAttributes: map[string]string{"secret": "rotate"},
		})
		requireNoErrors(t, diags)
		a := stringAttr(t, s.Attributes, "secret")
		if !a.Required || a.Optional || !a.WriteOnly {
			t.Errorf("expected secret to stay Required and become WriteOnly, got %+v", a)
		}
	})

	t.Run("success_existing_validators_preserved_and_trigger_validator_added", func(t *testing.T) {
		t.Parallel()
		s, diags := generateResourceSchema(resourceSchemaOptions{
			CreateModel:         &woValidatorsModel{},
			WriteOnlyAttributes: map[string]string{"secret": "rotate"},
		})
		requireNoErrors(t, diags)
		a := stringAttr(t, s.Attributes, "secret")
		if _, ok := findValidatorOfType[StringInChoicesValidator](a.Validators); !ok {
			t.Error("expected original StringInChoicesValidator to be preserved")
		}
		if _, ok := findValidatorOfType[WriteOnlyTriggerValidator](a.Validators); !ok {
			t.Error("expected WriteOnlyTriggerValidator to be appended")
		}
	})

	t.Run("success_plan_modifiers_dropped", func(t *testing.T) {
		t.Parallel()
		s, diags := generateResourceSchema(resourceSchemaOptions{
			CreateModel:                &woStringModel{},
			SemanticEqualityAttributes: map[string]SemanticEqualityKind{"secret": SemanticEqualityCaseInsensitive},
			WriteOnlyAttributes:        map[string]string{"secret": "rotate"},
		})
		requireNoErrors(t, diags)
		if a := stringAttr(t, s.Attributes, "secret"); len(a.PlanModifiers) != 0 {
			t.Errorf("expected plan modifiers to be dropped, got %+v", a.PlanModifiers)
		}
	})

	t.Run("success_sensitivity_kept_and_description_extended", func(t *testing.T) {
		t.Parallel()
		s, diags := generateResourceSchema(resourceSchemaOptions{
			CreateModel:         &woStringModel{},
			SensitiveAttributes: []string{"secret"},
			WriteOnlyAttributes: map[string]string{"secret": "rotate"},
		})
		requireNoErrors(t, diags)
		a := stringAttr(t, s.Attributes, "secret")
		if !a.Sensitive {
			t.Error("expected secret to remain Sensitive")
		}
		if !strings.Contains(a.Description, "a secret") || !strings.Contains(a.Description, "Write-only:") {
			t.Errorf("expected description to keep its original text and gain the write-only suffix, got %q", a.Description)
		}
	})

	t.Run("error_unrecognized_attribute_type", func(t *testing.T) {
		t.Parallel()
		attrs := map[string]schema.Attribute{
			"secret": schema.Float64Attribute{Optional: true, Description: "d"},
			"rotate": schema.StringAttribute{Optional: true},
		}
		diags := applyWriteOnlyAttributes(attrs, resourceSchemaOptions{
			WriteOnlyAttributes: map[string]string{"secret": "rotate"},
		})
		requireSingleError(t, diags, "Float64Attribute")
	})
}

// TestApplyWriteOnlyAttributes_Triggers covers trigger handling that is accepted: an existing
// attribute keeps its shape and gains a note, and an unresolved name is synthesized.
func TestApplyWriteOnlyAttributes_Triggers(t *testing.T) {
	t.Parallel()

	t.Run("success_existing_trigger_keeps_its_shape_apart_from_the_description", func(t *testing.T) {
		t.Parallel()
		baseline, baseDiags := generateResourceSchema(resourceSchemaOptions{CreateModel: &woStringModel{}})
		requireNoErrors(t, baseDiags)

		s, diags := generateResourceSchema(resourceSchemaOptions{
			CreateModel:         &woStringModel{},
			WriteOnlyAttributes: map[string]string{"secret": "rotate"},
		})
		requireNoErrors(t, diags)

		got := stringAttr(t, s.Attributes, "rotate")
		if !strings.Contains(got.Description, "secret") {
			t.Errorf("expected trigger description to name the write-only key, got %q", got.Description)
		}
		got.Description = baseline.Attributes["rotate"].GetDescription()
		assertAttributeUnchanged(t, "rotate", baseline.Attributes["rotate"], got)
	})

	t.Run("success_existing_trigger_description_is_appended_not_replaced", func(t *testing.T) {
		t.Parallel()
		attrs := map[string]schema.Attribute{
			"secret": schema.StringAttribute{Optional: true},
			"rotate": schema.StringAttribute{Optional: true, Description: "Original text."},
		}
		diags := applyWriteOnlyAttributes(attrs, resourceSchemaOptions{
			WriteOnlyAttributes: map[string]string{"secret": "rotate"},
		})
		requireNoErrors(t, diags)

		got := attrs["rotate"].GetDescription()
		if !strings.HasPrefix(got, "Original text.") || !strings.Contains(got, "secret") {
			t.Errorf("expected the original description to be kept and extended, got %q", got)
		}
	})

	t.Run("success_trigger_keeps_its_default", func(t *testing.T) {
		t.Parallel()
		s, diags := generateResourceSchema(resourceSchemaOptions{
			CreateModel:         &woTriggerDefaultModel{},
			WriteOnlyAttributes: map[string]string{"secret": "rotate"},
		})
		requireNoErrors(t, diags)
		if stringAttr(t, s.Attributes, "rotate").Default == nil {
			t.Error("expected trigger to retain its Default")
		}
	})

	t.Run("success_unresolved_trigger_synthesized_as_plain_optional_string", func(t *testing.T) {
		t.Parallel()
		s, diags := generateResourceSchema(resourceSchemaOptions{
			CreateModel:         &woStringModel{},
			WriteOnlyAttributes: map[string]string{"secret": "credential_rotation_trigger"},
		})
		requireNoErrors(t, diags)
		trigger := stringAttr(t, s.Attributes, "credential_rotation_trigger")
		if !trigger.Optional || trigger.Computed || trigger.Required || trigger.WriteOnly {
			t.Errorf("expected synthesized trigger to be plain Optional, got %+v", trigger)
		}
		if len(trigger.PlanModifiers) != 0 || len(trigger.Validators) != 0 {
			t.Errorf("expected synthesized trigger to carry no plan modifiers or validators, got %+v", trigger)
		}
		if !strings.Contains(trigger.Description, "secret") {
			t.Errorf("expected description to name the write-only key, got %q", trigger.Description)
		}
	})

	t.Run("success_trigger_shared_by_two_keys_synthesized_once", func(t *testing.T) {
		t.Parallel()
		s, diags := generateResourceSchema(resourceSchemaOptions{
			CreateModel: &woSharedTriggerModel{},
			WriteOnlyAttributes: map[string]string{
				"secret_a": "shared_trigger",
				"secret_b": "shared_trigger",
			},
		})
		requireNoErrors(t, diags)
		trigger := stringAttr(t, s.Attributes, "shared_trigger")
		if !strings.Contains(trigger.Description, "secret_a") || !strings.Contains(trigger.Description, "secret_b") {
			t.Errorf("expected shared trigger description to name both keys, got %q", trigger.Description)
		}
		if !stringAttr(t, s.Attributes, "secret_a").WriteOnly || !stringAttr(t, s.Attributes, "secret_b").WriteOnly {
			t.Error("expected both keys to be write-only")
		}
	})
}

// TestApplyWriteOnlyAttributes_Determinism guards against a regression to raw map iteration. A
// single run would very likely pass even with unsorted iteration, since Go randomizes map order
// per range rather than per process, so the pass has to be run repeatedly to have any power.
func TestApplyWriteOnlyAttributes_Determinism(t *testing.T) {
	t.Parallel()

	buildAttrs := func() map[string]schema.Attribute {
		return map[string]schema.Attribute{
			"alpha":    schema.SetAttribute{Optional: true, ElementType: types.StringType},
			"beta":     schema.StringAttribute{Optional: true, Computed: true, Default: StringDefault{Value: "x"}},
			"gamma":    schema.StringAttribute{Optional: true},
			"secret_a": schema.StringAttribute{Optional: true},
			"secret_b": schema.StringAttribute{Optional: true},
		}
	}
	opts := resourceSchemaOptions{
		WriteOnlyAttributes: map[string]string{
			"alpha":    "gamma",
			"beta":     "gamma",
			"missing":  "gamma",
			"secret_a": "shared",
			"secret_b": "shared",
		},
	}

	var wantDiags string
	for i := range 25 {
		attrs := buildAttrs()
		diags := applyWriteOnlyAttributes(attrs, opts)

		var sb strings.Builder
		for _, d := range diags.Errors() {
			sb.WriteString(d.Detail() + "\n")
		}
		got := sb.String()
		if got == "" {
			t.Fatal("expected at least one diagnostic from this fixture")
		}
		if i == 0 {
			wantDiags = got
			continue
		}
		if got != wantDiags {
			t.Fatalf("diagnostic order changed on run %d:\nwant:\n%s\ngot:\n%s", i, wantDiags, got)
		}
	}

	// Shared-trigger descriptions name their keys in sorted order, so the rendered description
	// must not vary either. Checked separately because the fixture above stops at validation.
	var wantDesc string
	for i := range 25 {
		s, diags := generateResourceSchema(resourceSchemaOptions{
			CreateModel: &woSharedTriggerModel{},
			WriteOnlyAttributes: map[string]string{
				"secret_a": "shared_trigger",
				"secret_b": "shared_trigger",
			},
		})
		requireNoErrors(t, diags)
		got := stringAttr(t, s.Attributes, "shared_trigger").Description
		if i == 0 {
			wantDesc = got
			continue
		}
		if got != wantDesc {
			t.Fatalf("shared trigger description changed on run %d: want %q, got %q", i, wantDesc, got)
		}
	}
}
