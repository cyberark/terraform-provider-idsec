// Copyright CyberArk. 2026
// SPDX-License-Identifier: Apache-2.0

package schemas

import (
	"reflect"
	"testing"

	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type depNested struct {
	OldField string `mapstructure:"old_field" desc:"d" deprecated:"new_field,use new_field"`
}

// depFixture exercises every `deprecated:` tag format across the attribute
// kinds the converters emit (primitive, slice, map, single-/list-nested).
type depFixture struct {
	ID          string            `mapstructure:"id" desc:"d"`
	OldName     string            `mapstructure:"old_name" desc:"d" deprecated:"name,use name"`       // replacement + message
	OldEnabled  bool              `mapstructure:"old_enabled" desc:"d" deprecated:"enabled"`          // replacement only
	OldCount    int64             `mapstructure:"old_count" desc:"d" deprecated:",counted elsewhere"` // message only
	OldTags     []string          `mapstructure:"old_tags" desc:"d" deprecated:""`                    // marker only
	OldLabels   map[string]string `mapstructure:"old_labels" desc:"d" deprecated:"labels"`
	OldChild    depNested         `mapstructure:"old_child" desc:"d" deprecated:"child"`
	OldChildren []depNested       `mapstructure:"old_children" desc:"d" deprecated:"children"`
}

// depStateModel omits the `deprecated:` tag on old_name, so the merge inside
// GenerateResourceSchemaFromStruct has to keep the create model's deprecation.
type depStateModel struct {
	OldName string `mapstructure:"old_name" desc:"d"`
}

func TestGenerateResourceSchemaFromStruct_PropagatesDeprecation(t *testing.T) {
	t.Parallel()
	got := GenerateResourceSchemaFromStruct(depFixture{}, nil, depStateModel{}, nil, nil, nil, nil, nil, nil, nil)

	want := map[string]string{
		"old_name":     `Use "name" instead. use name`,
		"old_enabled":  `Use "enabled" instead.`,
		"old_count":    "counted elsewhere",
		"old_tags":     "Deprecated.",
		"old_labels":   `Use "labels" instead.`,
		"old_child":    `Use "child" instead.`,
		"old_children": `Use "children" instead.`,
	}
	for attr, msg := range want {
		if dm := depMsg(got.Attributes[attr]); dm != msg {
			t.Errorf("%s: got %q, want %q", attr, dm, msg)
		}
	}
	if dm := depMsg(got.Attributes["id"]); dm != "" {
		t.Errorf("non-deprecated id leaked DeprecationMessage: %q", dm)
	}
	nested, ok := got.Attributes["old_child"].(rschema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("old_child: want SingleNestedAttribute, got %T", got.Attributes["old_child"])
	}
	if dm := depMsg(nested.Attributes["old_field"]); dm != `Use "new_field" instead. use new_field` {
		t.Errorf("nested old_field: %q", dm)
	}
	// DeprecationMessage is folded into Description so tfplugindocs picks it
	// up (it ignores DeprecationMessage and only emits a bare marker).
	oldName, ok := got.Attributes["old_name"].(rschema.StringAttribute)
	if !ok {
		t.Fatalf("old_name: want StringAttribute, got %T", got.Attributes["old_name"])
	}
	if d := oldName.Description; d != `d **Deprecated**: Use "name" instead. use name` {
		t.Errorf("old_name Description: %q", d)
	}
	// Marker-only tags (`deprecated:""`) still emit a runtime warning but must
	// NOT fold "**Deprecated**: Deprecated." into Description -- tfplugindocs
	// already renders a bare ", Deprecated" marker from the protocol flag, and
	// the doubled text reads as noise.
	oldTags, ok := got.Attributes["old_tags"].(rschema.ListAttribute)
	if !ok {
		t.Fatalf("old_tags: want ListAttribute, got %T", got.Attributes["old_tags"])
	}
	if d := oldTags.Description; d != "d" {
		t.Errorf("old_tags Description: got %q, want unchanged %q", d, "d")
	}
}

func TestGenerateDataSourceSchemaFromStruct_PropagatesDeprecation(t *testing.T) {
	t.Parallel()
	got := GenerateDataSourceSchemaFromStruct(depFixture{}, depFixture{}, nil, nil, nil)
	if dm := depMsg(got.Attributes["old_name"]); dm != `Use "name" instead. use name` {
		t.Errorf("old_name: %q", dm)
	}
}

// depMsg reads DeprecationMessage off any framework attribute via reflection.
func depMsg(a any) string {
	f := reflect.ValueOf(a).FieldByName("DeprecationMessage")
	if !f.IsValid() || f.Kind() != reflect.String {
		return ""
	}
	return f.String()
}
