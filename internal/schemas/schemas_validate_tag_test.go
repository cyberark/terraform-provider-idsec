// Copyright CyberArk. 2026
// SPDX-License-Identifier: Apache-2.0

package schemas

import (
	"testing"

	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// validateTagFixture represents a model covering every validate tag shape that affects the Required flag.
type validateTagFixture struct {
	Plain       string `mapstructure:"plain" desc:"d" validate:"required"`
	WithOthers  string `mapstructure:"with_others" desc:"d" validate:"required,max=10"`
	CondIf      string `mapstructure:"cond_if" desc:"d" validate:"required_if=Plain foo"`
	CondUnless  string `mapstructure:"cond_unless" desc:"d" validate:"required_unless=Plain foo"`
	CondWith    string `mapstructure:"cond_with" desc:"d" validate:"required_with=Plain"`
	CondWithout string `mapstructure:"cond_without" desc:"d" validate:"required_without=Plain"`
	CondThenReq string `mapstructure:"cond_then_req" desc:"d" validate:"required_if=Plain foo,required"`
	NoTag       string `mapstructure:"no_tag" desc:"d"`
	Unrelated   string `mapstructure:"unrelated" desc:"d" validate:"omitempty,min=2,max=8"`
}

// TestIsRequiredTag tests that isRequiredTag matches a whole "required" entry and not the required_* family.
func TestIsRequiredTag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		tag  string
		want bool
	}{
		{name: "bare_required", tag: "required", want: true},
		{name: "required_first_of_many", tag: "required,max=10", want: true},
		{name: "required_last_of_many", tag: "omitempty,required", want: true},
		{name: "required_if_is_not_required", tag: "required_if=Field val", want: false},
		{name: "required_unless_is_not_required", tag: "required_unless=Field val", want: false},
		{name: "required_with_is_not_required", tag: "required_with=Field", want: false},
		{name: "required_without_is_not_required", tag: "required_without=Field", want: false},
		{name: "required_if_alongside_required", tag: "required_if=Field val,required", want: true},
		{name: "empty_tag", tag: "", want: false},
		{name: "unrelated_rules_only", tag: "omitempty,min=2,max=8", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := isRequiredTag(tc.tag); got != tc.want {
				t.Errorf("isRequiredTag(%q) = %v, want %v", tc.tag, got, tc.want)
			}
		})
	}
}

// TestGenerateResourceSchemaFromStruct_RequiredFromValidateTag tests that Required comes from a bare
// "required" entry, not from a conditional required_* variant.
func TestGenerateResourceSchemaFromStruct_RequiredFromValidateTag(t *testing.T) {
	t.Parallel()

	model := &validateTagFixture{}
	schema, diags := GenerateResourceSchemaFromStruct(model, model, model, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	tests := []struct {
		attrName     string
		wantRequired bool
	}{
		{attrName: "plain", wantRequired: true},
		{attrName: "with_others", wantRequired: true},
		{attrName: "cond_if", wantRequired: false},
		{attrName: "cond_unless", wantRequired: false},
		{attrName: "cond_with", wantRequired: false},
		{attrName: "cond_without", wantRequired: false},
		{attrName: "cond_then_req", wantRequired: true},
		{attrName: "no_tag", wantRequired: false},
		{attrName: "unrelated", wantRequired: false},
	}

	for _, tc := range tests {
		t.Run(tc.attrName, func(t *testing.T) {
			t.Parallel()

			attr, found := schema.Attributes[tc.attrName]
			if !found {
				t.Fatalf("attribute %q missing from schema", tc.attrName)
			}
			strAttr, ok := attr.(rschema.StringAttribute)
			if !ok {
				t.Fatalf("attribute %q is %T, want StringAttribute", tc.attrName, attr)
			}
			if strAttr.Required != tc.wantRequired {
				t.Errorf("attribute %q Required = %v, want %v", tc.attrName, strAttr.Required, tc.wantRequired)
			}
			if !tc.wantRequired && !strAttr.Optional {
				t.Errorf("attribute %q is neither Required nor Optional", tc.attrName)
			}
		})
	}
}
