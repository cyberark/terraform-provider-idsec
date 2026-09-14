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

	s, _ := GenerateResourceSchemaFromStruct(
		&computedAttrsModel{},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		[]string{"id"},
		nil, // semanticEqualityAttrs
		nil, // writeOnlyAttrs
		nil, // writeOnlyHashedAttrs
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

	s, _ := GenerateResourceSchemaFromStruct(
		&computedAttrsModel{},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		[]string{"source.id"},
		nil, // semanticEqualityAttrs
		nil, // writeOnlyAttrs
		nil, // writeOnlyHashedAttrs
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

// roleCreateModel mirrors the shape of IdsecIdentityCreateRole — deliberately
// without RoleAttributes to reflect the real SDK contract (DVP-11593).
type roleCreateModel struct {
	RoleName          string   `mapstructure:"role_name"`
	Description       string   `mapstructure:"description"`
	AdminRights       []string `mapstructure:"admin_rights"`
	RoleType          string   `mapstructure:"role_type"`
	DynamicRoleScript string   `mapstructure:"dynamic_role_script"`
}

// roleUpdateModel mirrors the shape of IdsecIdentityUpdateRole — also without RoleAttributes.
// Note: no RoleType field — IdsecIdentityUpdateRole does not expose it.
type roleUpdateModel struct {
	RoleID            string   `mapstructure:"role_id"`
	RoleName          string   `mapstructure:"role_name"`
	Description       string   `mapstructure:"description"`
	AdminRights       []string `mapstructure:"admin_rights"`
	DynamicRoleScript string   `mapstructure:"dynamic_role_script"`
}

// roleStateModel mirrors the shape of IdsecIdentityRole — the full state including
// RoleAttributes which is populated by a separate API call in Get().
type roleStateModel struct {
	RoleID         string            `mapstructure:"role_id"`
	RoleName       string            `mapstructure:"role_name"`
	Description    string            `mapstructure:"description"`
	AdminRights    []string          `mapstructure:"admin_rights"`
	RoleType       string            `mapstructure:"role_type"`
	RoleAttributes map[string]string `mapstructure:"role_attributes"`
}

// TestGenerateResourceSchema_RoleAttributesIsReadOnly verifies that marking
// role_attributes in ComputedAttributes makes it Computed-only (not settable).
// This is the schema-level guard for DVP-11593: role_attributes cannot be written
// via IdsecIdentityCreateRole/IdsecIdentityUpdateRole (those structs have no such
// field), so the schema must not advertise it as Optional.
func TestGenerateResourceSchema_RoleAttributesIsReadOnly(t *testing.T) {
	t.Parallel()

	s, _ := GenerateResourceSchemaFromStruct(
		&roleCreateModel{},
		&roleUpdateModel{},
		&roleStateModel{},
		nil,
		nil,
		nil,
		nil,
		nil,
		[]string{"role_attributes"},
		nil, // semanticEqualityAttrs
		nil, // writeOnlyAttrs
		nil, // writeOnlyHashedAttrs
	)

	if !attrIsReadOnly(s.Attributes["role_attributes"]) {
		t.Errorf("expected role_attributes to be read-only (Computed=true, Optional=false), got %+v", s.Attributes["role_attributes"])
	}
}

// TestGenerateResourceSchema_RoleAttributesIsSettableWithoutComputedAttr is the
// negative counterpart: without ComputedAttributes the field appears as Optional+Computed
// (the broken state before DVP-11593).
func TestGenerateResourceSchema_RoleAttributesIsSettableWithoutComputedAttr(t *testing.T) {
	t.Parallel()

	s, _ := GenerateResourceSchemaFromStruct(
		&roleCreateModel{},
		&roleUpdateModel{},
		&roleStateModel{},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil, // no computedAttrs — role_attributes would be Optional+Computed
		nil, // semanticEqualityAttrs
		nil, // writeOnlyAttrs
		nil, // writeOnlyHashedAttrs
	)

	if !attrIsSettable(s.Attributes["role_attributes"]) {
		t.Errorf("expected role_attributes to be settable when not in ComputedAttributes, got %+v", s.Attributes["role_attributes"])
	}
}
