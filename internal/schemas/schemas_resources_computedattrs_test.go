// Copyright CyberArk. 2026
// SPDX-License-Identifier: Apache-2.0

package schemas

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// computedAttrsNestedRef mirrors the sechub sync policy source/target objects,
// where the nested "id" is a required user input (NOT server-assigned).
type computedAttrsNestedRef struct {
	ID   string `mapstructure:"id"`
	Type string `mapstructure:"type"`
}

// computedAttrsModel has a top-level server-assigned "id" plus nested
// source/target objects that each ALSO carry an "id" user input. This is the
// exact shape that triggered the bare-name leak (a bare "id" computed attribute
// wrongly marking source.id / target.id read-only).
type computedAttrsModel struct {
	ID     string                 `mapstructure:"id"`
	Name   string                 `mapstructure:"name"`
	Source computedAttrsNestedRef `mapstructure:"source"`
	Target computedAttrsNestedRef `mapstructure:"target"`
}

func attrIsReadOnly(a schema.Attribute) bool {
	if a == nil {
		return false
	}
	return a.IsComputed() && !a.IsOptional() && !a.IsRequired()
}

func attrIsSettable(a schema.Attribute) bool {
	if a == nil {
		return false
	}
	return a.IsOptional() || a.IsRequired()
}

func nestedIDAttr(t *testing.T, attrs map[string]schema.Attribute, parent string) schema.Attribute {
	t.Helper()
	p, ok := attrs[parent]
	if !ok {
		t.Fatalf("expected parent attribute %q to exist", parent)
	}
	single, ok := p.(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("expected %q to be a SingleNestedAttribute, got %T", parent, p)
	}
	return single.Attributes["id"]
}

// TestGenerateResourceSchema_BareComputedNameIsTopLevelOnly is the guardrail for the
// bare-name leak across the FULL schema pipeline (generator + post-processor). A bare
// "id" computed attribute must mark ONLY the top-level id read-only and must leave the
// nested source.id / target.id settable. Regressing this re-breaks resources like the
// sechub sync policy with "Invalid Configuration for Read-Only Attribute" on source.id.
func TestGenerateResourceSchema_BareComputedNameIsTopLevelOnly(t *testing.T) {
	t.Parallel()

	s := GenerateResourceSchemaFromStruct(
		&computedAttrsModel{},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		[]string{"id"},
		nil,
	)

	if !attrIsReadOnly(s.Attributes["id"]) {
		t.Errorf("expected top-level id to be read-only, got %+v", s.Attributes["id"])
	}
	if got := nestedIDAttr(t, s.Attributes, "source"); !attrIsSettable(got) {
		t.Errorf("expected source.id to remain settable, got %+v", got)
	}
	if got := nestedIDAttr(t, s.Attributes, "target"); !attrIsSettable(got) {
		t.Errorf("expected target.id to remain settable, got %+v", got)
	}
}

// TestGenerateResourceSchema_DottedComputedPathTargetsNestedOnly verifies a dotted path
// marks the exact nested attribute read-only without affecting the same-named top-level
// attribute or the same-named attribute under a different parent.
func TestGenerateResourceSchema_DottedComputedPathTargetsNestedOnly(t *testing.T) {
	t.Parallel()

	s := GenerateResourceSchemaFromStruct(
		&computedAttrsModel{},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		[]string{"source.id"},
		nil,
	)

	if !attrIsSettable(s.Attributes["id"]) {
		t.Errorf("expected top-level id to remain settable when only source.id is computed, got %+v", s.Attributes["id"])
	}
	if got := nestedIDAttr(t, s.Attributes, "source"); !attrIsReadOnly(got) {
		t.Errorf("expected source.id to be read-only, got %+v", got)
	}
	if got := nestedIDAttr(t, s.Attributes, "target"); !attrIsSettable(got) {
		t.Errorf("expected target.id to remain settable, got %+v", got)
	}
}
