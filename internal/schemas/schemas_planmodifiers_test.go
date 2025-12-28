// Copyright (c) CyberArk.
// SPDX-License-Identifier: Apache-2.0

package schemas

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// TestImmutableStringModifier tests the ImmutableStringModifier plan modifier.
//
// This test verifies that the modifier correctly:
//   - Allows resource creation (null state)
//   - Allows resource deletion (null plan)
//   - Allows no-change updates (state == plan)
//   - Blocks value changes (state != plan)
//   - Handles unknown values correctly
func TestImmutableStringModifier(t *testing.T) {
	t.Parallel()

	// Helper to create a tfsdk.State with a non-null Raw value
	createNonNullState := func() tfsdk.State {
		return tfsdk.State{
			Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{}),
		}
	}

	// Helper to create a tfsdk.State with a null Raw value (resource creation)
	createNullState := func() tfsdk.State {
		return tfsdk.State{
			Raw: tftypes.NewValue(tftypes.Object{}, nil),
		}
	}

	// Helper to create a tfsdk.Plan with a null Raw value (resource deletion)
	createNullPlan := func() tfsdk.Plan {
		return tfsdk.Plan{
			Raw: tftypes.NewValue(tftypes.Object{}, nil),
		}
	}

	// Helper to create a tfsdk.Plan with a non-null Raw value
	createNonNullPlan := func() tfsdk.Plan {
		return tfsdk.Plan{
			Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{}),
		}
	}

	tests := []struct {
		name             string
		stateValue       types.String
		planValue        types.String
		configValue      types.String
		state            tfsdk.State
		plan             tfsdk.Plan
		expectedError    bool
		expectedErrorMsg string
		validateFunc     func(t *testing.T, resp *planmodifier.StringResponse)
	}{
		{
			name:          "create_operation_state_null_allows_creation",
			stateValue:    types.StringNull(),
			planValue:     types.StringValue("new-value"),
			configValue:   types.StringValue("new-value"),
			state:         createNullState(),
			plan:          createNonNullPlan(),
			expectedError: false,
		},
		{
			name:          "update_operation_no_change_allows_update",
			stateValue:    types.StringValue("same-value"),
			planValue:     types.StringValue("same-value"),
			configValue:   types.StringValue("same-value"),
			state:         createNonNullState(),
			plan:          createNonNullPlan(),
			expectedError: false,
		},
		{
			name:             "update_operation_value_changed_blocks_plan",
			stateValue:       types.StringValue("old-value"),
			planValue:        types.StringValue("new-value"),
			configValue:      types.StringValue("new-value"),
			state:            createNonNullState(),
			plan:             createNonNullPlan(),
			expectedError:    true,
			expectedErrorMsg: "Immutable Attribute Cannot Be Changed",
		},
		{
			name:          "update_operation_plan_value_unknown_allows_update",
			stateValue:    types.StringValue("current-value"),
			planValue:     types.StringUnknown(),
			configValue:   types.StringValue("current-value"),
			state:         createNonNullState(),
			plan:          createNonNullPlan(),
			expectedError: false,
		},
		{
			name:          "delete_operation_plan_null_allows_deletion",
			stateValue:    types.StringValue("current-value"),
			planValue:     types.StringNull(),
			configValue:   types.StringNull(),
			state:         createNonNullState(),
			plan:          createNullPlan(),
			expectedError: false,
		},
		{
			name:          "update_operation_config_value_unknown_allows_update",
			stateValue:    types.StringValue("current-value"),
			planValue:     types.StringValue("new-value"),
			configValue:   types.StringUnknown(),
			state:         createNonNullState(),
			plan:          createNonNullPlan(),
			expectedError: false,
		},
		{
			name:             "update_operation_empty_to_value_blocks_plan",
			stateValue:       types.StringValue(""),
			planValue:        types.StringValue("new-value"),
			configValue:      types.StringValue("new-value"),
			state:            createNonNullState(),
			plan:             createNonNullPlan(),
			expectedError:    true,
			expectedErrorMsg: "Immutable Attribute Cannot Be Changed",
		},
		{
			name:             "update_operation_value_to_empty_blocks_plan",
			stateValue:       types.StringValue("old-value"),
			planValue:        types.StringValue(""),
			configValue:      types.StringValue(""),
			state:            createNonNullState(),
			plan:             createNonNullPlan(),
			expectedError:    true,
			expectedErrorMsg: "Immutable Attribute Cannot Be Changed",
		},
		{
			name:          "update_operation_empty_to_empty_allows_update",
			stateValue:    types.StringValue(""),
			planValue:     types.StringValue(""),
			configValue:   types.StringValue(""),
			state:         createNonNullState(),
			plan:          createNonNullPlan(),
			expectedError: false,
		},
		{
			name:             "error_message_contains_attribute_path",
			stateValue:       types.StringValue("old"),
			planValue:        types.StringValue("new"),
			configValue:      types.StringValue("new"),
			state:            createNonNullState(),
			plan:             createNonNullPlan(),
			expectedError:    true,
			expectedErrorMsg: "Immutable Attribute Cannot Be Changed",
			validateFunc: func(t *testing.T, resp *planmodifier.StringResponse) {
				if len(resp.Diagnostics.Errors()) == 0 {
					t.Error("Expected at least one error diagnostic")
					return
				}
				detail := resp.Diagnostics.Errors()[0].Detail()
				if detail == "" {
					t.Error("Expected error detail to contain attribute information")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			modifier := ImmutableString()
			req := planmodifier.StringRequest{
				StateValue:  tt.stateValue,
				PlanValue:   tt.planValue,
				ConfigValue: tt.configValue,
				State:       tt.state,
				Plan:        tt.plan,
				Path:        path.Root("test_attr"),
			}
			resp := &planmodifier.StringResponse{
				PlanValue: tt.planValue,
			}

			// Execute
			modifier.PlanModifyString(context.Background(), req, resp)

			// Validate error expectation
			if tt.expectedError {
				if !resp.Diagnostics.HasError() {
					t.Errorf("Expected error, got none")
					return
				}
				if tt.expectedErrorMsg != "" {
					found := false
					for _, err := range resp.Diagnostics.Errors() {
						if err.Summary() == tt.expectedErrorMsg {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected error message containing '%s', got: %v",
							tt.expectedErrorMsg, resp.Diagnostics.Errors())
					}
				}
			} else {
				if resp.Diagnostics.HasError() {
					t.Errorf("Expected no error, got: %v", resp.Diagnostics.Errors())
					return
				}
			}

			// Custom validation (if provided)
			if tt.validateFunc != nil {
				tt.validateFunc(t, resp)
			}
		})
	}
}

// TestImmutableInt64Modifier tests the ImmutableInt64Modifier plan modifier.
//
// This test verifies that the modifier correctly handles int64 attributes with
// the same behavior as string attributes.
func TestImmutableInt64Modifier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		stateValue    types.Int64
		planValue     types.Int64
		configValue   types.Int64
		isCreate      bool
		isDelete      bool
		expectedError bool
	}{
		{
			name:          "create_operation_allows_creation",
			stateValue:    types.Int64Null(),
			planValue:     types.Int64Value(42),
			configValue:   types.Int64Value(42),
			isCreate:      true,
			expectedError: false,
		},
		{
			name:          "no_change_allows_update",
			stateValue:    types.Int64Value(42),
			planValue:     types.Int64Value(42),
			configValue:   types.Int64Value(42),
			expectedError: false,
		},
		{
			name:          "value_changed_blocks_plan",
			stateValue:    types.Int64Value(42),
			planValue:     types.Int64Value(100),
			configValue:   types.Int64Value(100),
			expectedError: true,
		},
		{
			name:          "unknown_plan_value_allows_update",
			stateValue:    types.Int64Value(42),
			planValue:     types.Int64Unknown(),
			configValue:   types.Int64Value(42),
			expectedError: false,
		},
		{
			name:          "delete_operation_allows_deletion",
			stateValue:    types.Int64Value(42),
			planValue:     types.Int64Null(),
			configValue:   types.Int64Null(),
			isDelete:      true,
			expectedError: false,
		},
		{
			name:          "zero_to_value_blocks_plan",
			stateValue:    types.Int64Value(0),
			planValue:     types.Int64Value(1),
			configValue:   types.Int64Value(1),
			expectedError: true,
		},
		{
			name:          "value_to_zero_blocks_plan",
			stateValue:    types.Int64Value(1),
			planValue:     types.Int64Value(0),
			configValue:   types.Int64Value(0),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			modifier := ImmutableInt64()

			// Build request with appropriate state/plan Raw values
			var state tfsdk.State
			var plan tfsdk.Plan

			if tt.isCreate {
				state = tfsdk.State{Raw: tftypes.NewValue(tftypes.Object{}, nil)}
			} else {
				state = tfsdk.State{Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{})}
			}

			if tt.isDelete {
				plan = tfsdk.Plan{Raw: tftypes.NewValue(tftypes.Object{}, nil)}
			} else {
				plan = tfsdk.Plan{Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{})}
			}

			req := planmodifier.Int64Request{
				StateValue:  tt.stateValue,
				PlanValue:   tt.planValue,
				ConfigValue: tt.configValue,
				State:       state,
				Plan:        plan,
				Path:        path.Root("test_attr"),
			}
			resp := &planmodifier.Int64Response{
				PlanValue: tt.planValue,
			}

			modifier.PlanModifyInt64(context.Background(), req, resp)

			if tt.expectedError && !resp.Diagnostics.HasError() {
				t.Error("Expected error, got none")
			}
			if !tt.expectedError && resp.Diagnostics.HasError() {
				t.Errorf("Expected no error, got: %v", resp.Diagnostics.Errors())
			}
		})
	}
}

// TestImmutableBoolModifier tests the ImmutableBoolModifier plan modifier.
//
// This test verifies that the modifier correctly handles bool attributes with
// the same behavior as string attributes.
func TestImmutableBoolModifier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		stateValue    types.Bool
		planValue     types.Bool
		configValue   types.Bool
		isCreate      bool
		isDelete      bool
		expectedError bool
	}{
		{
			name:          "create_operation_allows_creation",
			stateValue:    types.BoolNull(),
			planValue:     types.BoolValue(true),
			configValue:   types.BoolValue(true),
			isCreate:      true,
			expectedError: false,
		},
		{
			name:          "no_change_allows_update",
			stateValue:    types.BoolValue(true),
			planValue:     types.BoolValue(true),
			configValue:   types.BoolValue(true),
			expectedError: false,
		},
		{
			name:          "true_to_false_blocks_plan",
			stateValue:    types.BoolValue(true),
			planValue:     types.BoolValue(false),
			configValue:   types.BoolValue(false),
			expectedError: true,
		},
		{
			name:          "false_to_true_blocks_plan",
			stateValue:    types.BoolValue(false),
			planValue:     types.BoolValue(true),
			configValue:   types.BoolValue(true),
			expectedError: true,
		},
		{
			name:          "unknown_plan_value_allows_update",
			stateValue:    types.BoolValue(true),
			planValue:     types.BoolUnknown(),
			configValue:   types.BoolValue(true),
			expectedError: false,
		},
		{
			name:          "delete_operation_allows_deletion",
			stateValue:    types.BoolValue(true),
			planValue:     types.BoolNull(),
			configValue:   types.BoolNull(),
			isDelete:      true,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			modifier := ImmutableBool()

			var state tfsdk.State
			var plan tfsdk.Plan

			if tt.isCreate {
				state = tfsdk.State{Raw: tftypes.NewValue(tftypes.Object{}, nil)}
			} else {
				state = tfsdk.State{Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{})}
			}

			if tt.isDelete {
				plan = tfsdk.Plan{Raw: tftypes.NewValue(tftypes.Object{}, nil)}
			} else {
				plan = tfsdk.Plan{Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{})}
			}

			req := planmodifier.BoolRequest{
				StateValue:  tt.stateValue,
				PlanValue:   tt.planValue,
				ConfigValue: tt.configValue,
				State:       state,
				Plan:        plan,
				Path:        path.Root("test_attr"),
			}
			resp := &planmodifier.BoolResponse{
				PlanValue: tt.planValue,
			}

			modifier.PlanModifyBool(context.Background(), req, resp)

			if tt.expectedError && !resp.Diagnostics.HasError() {
				t.Error("Expected error, got none")
			}
			if !tt.expectedError && resp.Diagnostics.HasError() {
				t.Errorf("Expected no error, got: %v", resp.Diagnostics.Errors())
			}
		})
	}
}

// TestImmutableListModifier tests the ImmutableListModifier plan modifier.
//
// This test verifies that the modifier correctly handles list attributes.
func TestImmutableListModifier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		stateValue    types.List
		planValue     types.List
		configValue   types.List
		isCreate      bool
		isDelete      bool
		expectedError bool
	}{
		{
			name:          "create_operation_allows_creation",
			stateValue:    types.ListNull(types.StringType),
			planValue:     types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a")}),
			configValue:   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a")}),
			isCreate:      true,
			expectedError: false,
		},
		{
			name:          "no_change_allows_update",
			stateValue:    types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a")}),
			planValue:     types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a")}),
			configValue:   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a")}),
			expectedError: false,
		},
		{
			name:          "value_changed_blocks_plan",
			stateValue:    types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a")}),
			planValue:     types.ListValueMust(types.StringType, []attr.Value{types.StringValue("b")}),
			configValue:   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("b")}),
			expectedError: true,
		},
		{
			name:          "unknown_plan_value_allows_update",
			stateValue:    types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a")}),
			planValue:     types.ListUnknown(types.StringType),
			configValue:   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a")}),
			expectedError: false,
		},
		{
			name:          "delete_operation_allows_deletion",
			stateValue:    types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a")}),
			planValue:     types.ListNull(types.StringType),
			configValue:   types.ListNull(types.StringType),
			isDelete:      true,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			modifier := ImmutableList()

			var state tfsdk.State
			var plan tfsdk.Plan

			if tt.isCreate {
				state = tfsdk.State{Raw: tftypes.NewValue(tftypes.Object{}, nil)}
			} else {
				state = tfsdk.State{Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{})}
			}

			if tt.isDelete {
				plan = tfsdk.Plan{Raw: tftypes.NewValue(tftypes.Object{}, nil)}
			} else {
				plan = tfsdk.Plan{Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{})}
			}

			req := planmodifier.ListRequest{
				StateValue:  tt.stateValue,
				PlanValue:   tt.planValue,
				ConfigValue: tt.configValue,
				State:       state,
				Plan:        plan,
				Path:        path.Root("test_attr"),
			}
			resp := &planmodifier.ListResponse{
				PlanValue: tt.planValue,
			}

			modifier.PlanModifyList(context.Background(), req, resp)

			if tt.expectedError && !resp.Diagnostics.HasError() {
				t.Error("Expected error, got none")
			}
			if !tt.expectedError && resp.Diagnostics.HasError() {
				t.Errorf("Expected no error, got: %v", resp.Diagnostics.Errors())
			}
		})
	}
}

// TestImmutableSetModifier tests the ImmutableSetModifier plan modifier.
//
// This test verifies that the modifier correctly handles set attributes.
func TestImmutableSetModifier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		stateValue    types.Set
		planValue     types.Set
		configValue   types.Set
		isCreate      bool
		isDelete      bool
		expectedError bool
	}{
		{
			name:          "create_operation_allows_creation",
			stateValue:    types.SetNull(types.StringType),
			planValue:     types.SetValueMust(types.StringType, []attr.Value{types.StringValue("a")}),
			configValue:   types.SetValueMust(types.StringType, []attr.Value{types.StringValue("a")}),
			isCreate:      true,
			expectedError: false,
		},
		{
			name:          "no_change_allows_update",
			stateValue:    types.SetValueMust(types.StringType, []attr.Value{types.StringValue("a")}),
			planValue:     types.SetValueMust(types.StringType, []attr.Value{types.StringValue("a")}),
			configValue:   types.SetValueMust(types.StringType, []attr.Value{types.StringValue("a")}),
			expectedError: false,
		},
		{
			name:          "value_changed_blocks_plan",
			stateValue:    types.SetValueMust(types.StringType, []attr.Value{types.StringValue("a")}),
			planValue:     types.SetValueMust(types.StringType, []attr.Value{types.StringValue("b")}),
			configValue:   types.SetValueMust(types.StringType, []attr.Value{types.StringValue("b")}),
			expectedError: true,
		},
		{
			name:          "unknown_plan_value_allows_update",
			stateValue:    types.SetValueMust(types.StringType, []attr.Value{types.StringValue("a")}),
			planValue:     types.SetUnknown(types.StringType),
			configValue:   types.SetValueMust(types.StringType, []attr.Value{types.StringValue("a")}),
			expectedError: false,
		},
		{
			name:          "delete_operation_allows_deletion",
			stateValue:    types.SetValueMust(types.StringType, []attr.Value{types.StringValue("a")}),
			planValue:     types.SetNull(types.StringType),
			configValue:   types.SetNull(types.StringType),
			isDelete:      true,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			modifier := ImmutableSet()

			var state tfsdk.State
			var plan tfsdk.Plan

			if tt.isCreate {
				state = tfsdk.State{Raw: tftypes.NewValue(tftypes.Object{}, nil)}
			} else {
				state = tfsdk.State{Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{})}
			}

			if tt.isDelete {
				plan = tfsdk.Plan{Raw: tftypes.NewValue(tftypes.Object{}, nil)}
			} else {
				plan = tfsdk.Plan{Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{})}
			}

			req := planmodifier.SetRequest{
				StateValue:  tt.stateValue,
				PlanValue:   tt.planValue,
				ConfigValue: tt.configValue,
				State:       state,
				Plan:        plan,
				Path:        path.Root("test_attr"),
			}
			resp := &planmodifier.SetResponse{
				PlanValue: tt.planValue,
			}

			modifier.PlanModifySet(context.Background(), req, resp)

			if tt.expectedError && !resp.Diagnostics.HasError() {
				t.Error("Expected error, got none")
			}
			if !tt.expectedError && resp.Diagnostics.HasError() {
				t.Errorf("Expected no error, got: %v", resp.Diagnostics.Errors())
			}
		})
	}
}

// TestImmutableMapModifier tests the ImmutableMapModifier plan modifier.
//
// This test verifies that the modifier correctly handles map attributes.
func TestImmutableMapModifier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		stateValue    types.Map
		planValue     types.Map
		configValue   types.Map
		isCreate      bool
		isDelete      bool
		expectedError bool
	}{
		{
			name:          "create_operation_allows_creation",
			stateValue:    types.MapNull(types.StringType),
			planValue:     types.MapValueMust(types.StringType, map[string]attr.Value{"key": types.StringValue("value")}),
			configValue:   types.MapValueMust(types.StringType, map[string]attr.Value{"key": types.StringValue("value")}),
			isCreate:      true,
			expectedError: false,
		},
		{
			name:          "no_change_allows_update",
			stateValue:    types.MapValueMust(types.StringType, map[string]attr.Value{"key": types.StringValue("value")}),
			planValue:     types.MapValueMust(types.StringType, map[string]attr.Value{"key": types.StringValue("value")}),
			configValue:   types.MapValueMust(types.StringType, map[string]attr.Value{"key": types.StringValue("value")}),
			expectedError: false,
		},
		{
			name:          "value_changed_blocks_plan",
			stateValue:    types.MapValueMust(types.StringType, map[string]attr.Value{"key": types.StringValue("value1")}),
			planValue:     types.MapValueMust(types.StringType, map[string]attr.Value{"key": types.StringValue("value2")}),
			configValue:   types.MapValueMust(types.StringType, map[string]attr.Value{"key": types.StringValue("value2")}),
			expectedError: true,
		},
		{
			name:          "unknown_plan_value_allows_update",
			stateValue:    types.MapValueMust(types.StringType, map[string]attr.Value{"key": types.StringValue("value")}),
			planValue:     types.MapUnknown(types.StringType),
			configValue:   types.MapValueMust(types.StringType, map[string]attr.Value{"key": types.StringValue("value")}),
			expectedError: false,
		},
		{
			name:          "delete_operation_allows_deletion",
			stateValue:    types.MapValueMust(types.StringType, map[string]attr.Value{"key": types.StringValue("value")}),
			planValue:     types.MapNull(types.StringType),
			configValue:   types.MapNull(types.StringType),
			isDelete:      true,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			modifier := ImmutableMap()

			var state tfsdk.State
			var plan tfsdk.Plan

			if tt.isCreate {
				state = tfsdk.State{Raw: tftypes.NewValue(tftypes.Object{}, nil)}
			} else {
				state = tfsdk.State{Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{})}
			}

			if tt.isDelete {
				plan = tfsdk.Plan{Raw: tftypes.NewValue(tftypes.Object{}, nil)}
			} else {
				plan = tfsdk.Plan{Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{})}
			}

			req := planmodifier.MapRequest{
				StateValue:  tt.stateValue,
				PlanValue:   tt.planValue,
				ConfigValue: tt.configValue,
				State:       state,
				Plan:        plan,
				Path:        path.Root("test_attr"),
			}
			resp := &planmodifier.MapResponse{
				PlanValue: tt.planValue,
			}

			modifier.PlanModifyMap(context.Background(), req, resp)

			if tt.expectedError && !resp.Diagnostics.HasError() {
				t.Error("Expected error, got none")
			}
			if !tt.expectedError && resp.Diagnostics.HasError() {
				t.Errorf("Expected no error, got: %v", resp.Diagnostics.Errors())
			}
		})
	}
}

// TestImmutableStringModifier_Description verifies documentation methods.
//
// This test ensures that the Description and MarkdownDescription methods
// return non-empty strings for documentation purposes.
func TestImmutableStringModifier_Description(t *testing.T) {
	t.Parallel()

	modifier := ImmutableString()

	description := modifier.Description(context.Background())
	if description == "" {
		t.Error("Description should not be empty")
	}

	markdownDescription := modifier.MarkdownDescription(context.Background())
	if markdownDescription == "" {
		t.Error("MarkdownDescription should not be empty")
	}
}
