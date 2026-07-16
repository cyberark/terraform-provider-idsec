// Copyright CyberArk. 2026
// SPDX-License-Identifier: Apache-2.0

package schemas

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func withHistoryLoader(t *testing.T, history map[string]bool) {
	t.Helper()
	prev := historyLoader
	historyLoader = func(context.Context, privateStateReader) map[string]bool { return history }
	t.Cleanup(func() { historyLoader = prev })
}

func mustRemovedToUnknownStringModifier(t *testing.T) removedToUnknownStringModifier {
	t.Helper()
	m, ok := RemovedToUnknownString().(removedToUnknownStringModifier)
	if !ok {
		t.Fatalf("RemovedToUnknownString(): got %T", RemovedToUnknownString())
	}
	return m
}

func stringPlanModifierCount(t *testing.T, attrs map[string]schema.Attribute, name string) int {
	t.Helper()
	a, ok := attrs[name].(schema.StringAttribute)
	if !ok {
		t.Fatalf("%s: expected StringAttribute, got %T", name, attrs[name])
	}
	return len(a.PlanModifiers)
}

func TestRemovalPredicates(t *testing.T) {
	t.Parallel()

	null := types.StringNull()
	set := types.StringValue("v")

	for _, tt := range []struct {
		name string
		fn   func() bool
		want bool
	}{
		{"valueIsAbsent_null", func() bool { return valueIsAbsent(null) }, true},
		{"valueIsAbsent_set", func() bool { return valueIsAbsent(set) }, false},
		{"isUserRemoval", func() bool { return isUserRemoval(null, set) }, true},
		{"isUserRemoval_empty_state", func() bool { return isUserRemoval(null, types.StringValue("")) }, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.fn(); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldRemoveToNull(t *testing.T) {
	t.Parallel()

	null := types.StringNull()
	state := types.StringValue("password")

	for _, tt := range []struct {
		name    string
		history map[string]bool
		path    string
		want    bool
	}{
		{"in_history", map[string]bool{"secret_type": true}, "secret_type", true},
		{"not_in_history", map[string]bool{}, "secret_type", false},
		{"indexed_path", map[string]bool{"targets.name": true}, "targets[0].name", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := shouldRemoveToNull(tt.history, tt.path, null, state); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemovedToUnknownStringModifier(t *testing.T) {
	ctx := context.Background()
	req := planmodifier.StringRequest{
		Path:        path.Root("attr"),
		PlanValue:   types.StringValue("prior"),
		ConfigValue: types.StringNull(),
		StateValue:  types.StringValue("prior"),
	}

	t.Run("unknown_when_in_history", func(t *testing.T) {
		withHistoryLoader(t, map[string]bool{"attr": true})
		resp := &planmodifier.StringResponse{PlanValue: types.StringValue("prior")}
		mustRemovedToUnknownStringModifier(t).PlanModifyString(ctx, req, resp)
		if !resp.PlanValue.IsUnknown() {
			t.Errorf("expected unknown plan, got %v", resp.PlanValue)
		}
	})

	t.Run("noop_without_history", func(t *testing.T) {
		withHistoryLoader(t, map[string]bool{})
		resp := &planmodifier.StringResponse{PlanValue: types.StringValue("prior")}
		mustRemovedToUnknownStringModifier(t).PlanModifyString(ctx, req, resp)
		if resp.PlanValue.IsUnknown() {
			t.Error("expected plan preserved")
		}
	})
}

func TestApplyRemovedToUnknownModifiers(t *testing.T) {
	t.Parallel()

	t.Run("optional_computed_only", func(t *testing.T) {
		t.Parallel()
		attrs := map[string]schema.Attribute{
			"optional_computed": schema.StringAttribute{Optional: true, Computed: true},
			"required":          schema.StringAttribute{Required: true},
			"computed_only":     schema.StringAttribute{Computed: true},
		}
		ApplyRemovedToUnknownModifiers(attrs, nil, nil)

		if n := stringPlanModifierCount(t, attrs, "optional_computed"); n != 2 {
			t.Fatalf("optional_computed: got %d modifiers, want 2", n)
		}
		for _, name := range []string{"required", "computed_only"} {
			if n := stringPlanModifierCount(t, attrs, name); n != 0 {
				t.Errorf("%s: got %d modifiers, want 0", name, n)
			}
		}
	})

	t.Run("skips_read_key", func(t *testing.T) {
		t.Parallel()
		attrs := map[string]schema.Attribute{
			"id":   schema.StringAttribute{Optional: true, Computed: true},
			"name": schema.StringAttribute{Optional: true, Computed: true},
		}
		ApplyRemovedToUnknownModifiers(attrs, []string{"id"}, nil)

		if n := stringPlanModifierCount(t, attrs, "id"); n != 0 {
			t.Errorf("id: got %d modifiers, want 0", n)
		}
		if n := stringPlanModifierCount(t, attrs, "name"); n != 2 {
			t.Errorf("name: got %d modifiers, want 2", n)
		}
	})

	t.Run("immutable_keeps_use_state_only", func(t *testing.T) {
		t.Parallel()
		attrs := map[string]schema.Attribute{
			"secret_type": schema.StringAttribute{Optional: true, Computed: true},
			"name":        schema.StringAttribute{Optional: true, Computed: true},
		}
		ApplyRemovedToUnknownModifiers(attrs, nil, []string{"secret_type"})

		// Immutable attribute gets UseStateForUnknown only (no removed-to-null), so omitting
		// it from config keeps the prior value instead of planning it to null.
		if n := stringPlanModifierCount(t, attrs, "secret_type"); n != 1 {
			t.Errorf("secret_type: got %d modifiers, want 1 (UseStateForUnknown only)", n)
		}
		if n := stringPlanModifierCount(t, attrs, "name"); n != 2 {
			t.Errorf("name: got %d modifiers, want 2", n)
		}
	})
}

func TestApplyRemovedToUnknownModifiersDynamic(t *testing.T) {
	t.Parallel()

	attrs := map[string]schema.Attribute{
		"optional_computed": schema.DynamicAttribute{Optional: true, Computed: true},
		"required":          schema.DynamicAttribute{Required: true},
		"computed_only":     schema.DynamicAttribute{Computed: true},
	}
	ApplyRemovedToUnknownModifiers(attrs, nil, nil)

	dynamicPlanModifierCount := func(name string) int {
		t.Helper()
		a, ok := attrs[name].(schema.DynamicAttribute)
		if !ok {
			t.Fatalf("%s: expected DynamicAttribute, got %T", name, attrs[name])
		}
		return len(a.PlanModifiers)
	}

	// Optional+Computed dynamic attributes get exactly one modifier (UseStateForUnknown); no
	// removed-to-null modifier is attached for dynamic values.
	if n := dynamicPlanModifierCount("optional_computed"); n != 1 {
		t.Fatalf("optional_computed: got %d modifiers, want 1", n)
	}
	for _, name := range []string{"required", "computed_only"} {
		if n := dynamicPlanModifierCount(name); n != 0 {
			t.Errorf("%s: got %d modifiers, want 0", name, n)
		}
	}
}

func TestComputedOnlyAttributePaths(t *testing.T) {
	t.Parallel()

	attrs := map[string]schema.Attribute{
		"id":   schema.StringAttribute{Computed: true},
		"name": schema.StringAttribute{Optional: true, Computed: true},
		"metadata": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"status": schema.StringAttribute{Computed: true},
				"note":   schema.StringAttribute{Optional: true},
			},
		},
		"targets": schema.SetNestedAttribute{
			Optional: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"role_name": schema.StringAttribute{Computed: true},
					"role_id":   schema.StringAttribute{Optional: true},
				},
			},
		},
		"server_only": schema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]schema.Attribute{
				"x": schema.StringAttribute{Computed: true},
			},
		},
	}

	got := ComputedOnlyAttributePaths(attrs)
	want := []string{
		"id",
		"metadata.status",
		"server_only",
		"targets.role_name",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ComputedOnlyAttributePaths = %v, want %v", got, want)
	}
}
