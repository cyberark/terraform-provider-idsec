// Copyright CyberArk. 2026
// SPDX-License-Identifier: Apache-2.0

package schemas

import (
	"context"
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

func mustRemovedToNullStringModifier(t *testing.T) removedToNullStringModifier {
	t.Helper()
	m, ok := RemovedToNullString().(removedToNullStringModifier)
	if !ok {
		t.Fatalf("RemovedToNullString(): got %T", RemovedToNullString())
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

func TestRemovedToNullStringModifier(t *testing.T) {
	ctx := context.Background()
	req := planmodifier.StringRequest{
		Path:        path.Root("attr"),
		PlanValue:   types.StringValue("prior"),
		ConfigValue: types.StringNull(),
		StateValue:  types.StringValue("prior"),
	}

	t.Run("nulls_when_in_history", func(t *testing.T) {
		withHistoryLoader(t, map[string]bool{"attr": true})
		resp := &planmodifier.StringResponse{PlanValue: types.StringValue("prior")}
		mustRemovedToNullStringModifier(t).PlanModifyString(ctx, req, resp)
		if !resp.PlanValue.IsNull() {
			t.Errorf("expected null plan, got %v", resp.PlanValue)
		}
	})

	t.Run("noop_without_history", func(t *testing.T) {
		withHistoryLoader(t, map[string]bool{})
		resp := &planmodifier.StringResponse{PlanValue: types.StringValue("prior")}
		mustRemovedToNullStringModifier(t).PlanModifyString(ctx, req, resp)
		if resp.PlanValue.IsNull() {
			t.Error("expected plan preserved")
		}
	})
}

func TestApplyRemovedToNullModifiers(t *testing.T) {
	t.Parallel()

	t.Run("optional_computed_only", func(t *testing.T) {
		t.Parallel()
		attrs := map[string]schema.Attribute{
			"optional_computed": schema.StringAttribute{Optional: true, Computed: true},
			"required":          schema.StringAttribute{Required: true},
			"computed_only":     schema.StringAttribute{Computed: true},
		}
		ApplyRemovedToNullModifiers(attrs)

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
		ApplyRemovedToNullModifiers(attrs, "id")

		if n := stringPlanModifierCount(t, attrs, "id"); n != 0 {
			t.Errorf("id: got %d modifiers, want 0", n)
		}
		if n := stringPlanModifierCount(t, attrs, "name"); n != 2 {
			t.Errorf("name: got %d modifiers, want 2", n)
		}
	})
}
