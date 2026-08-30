// Copyright CyberArk. 2026
// SPDX-License-Identifier: Apache-2.0

package schemas

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
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

// TestCaseInsensitiveStringModifier tests CaseInsensitiveStringModifier.
func TestCaseInsensitiveStringModifier(t *testing.T) {
	t.Parallel()

	createNonNullState := func() tfsdk.State {
		return tfsdk.State{
			Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{}),
		}
	}
	createNullState := func() tfsdk.State {
		return tfsdk.State{
			Raw: tftypes.NewValue(tftypes.Object{}, nil),
		}
	}
	createNullPlan := func() tfsdk.Plan {
		return tfsdk.Plan{
			Raw: tftypes.NewValue(tftypes.Object{}, nil),
		}
	}
	createNonNullPlan := func() tfsdk.Plan {
		return tfsdk.Plan{
			Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{}),
		}
	}

	tests := []struct {
		name         string
		stateValue   types.String
		planValue    types.String
		configValue  types.String
		state        tfsdk.State
		plan         tfsdk.Plan
		validateFunc func(t *testing.T, req planmodifier.StringRequest, resp *planmodifier.StringResponse)
	}{
		{
			name:        "create_operation_state_null_noop",
			stateValue:  types.StringNull(),
			planValue:   types.StringValue("NEW"),
			configValue: types.StringValue("NEW"),
			state:       createNullState(),
			plan:        createNonNullPlan(),
		},
		{
			name:        "update_exact_match_noop",
			stateValue:  types.StringValue("stable"),
			planValue:   types.StringValue("stable"),
			configValue: types.StringValue("stable"),
			state:       createNonNullState(),
			plan:        createNonNullPlan(),
		},
		{
			name:        "update_case_only_normalizes_plan_to_state",
			stateValue:  types.StringValue("MyValue"),
			planValue:   types.StringValue("myvalue"),
			configValue: types.StringValue("myvalue"),
			state:       createNonNullState(),
			plan:        createNonNullPlan(),
			validateFunc: func(t *testing.T, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
				if !resp.PlanValue.Equal(req.StateValue) {
					t.Errorf("expected plan normalized to state, state=%v plan=%v", req.StateValue, resp.PlanValue)
				}
			},
		},
		{
			name:        "update_semantic_change_leaves_plan_unchanged",
			stateValue:  types.StringValue("alpha"),
			planValue:   types.StringValue("beta"),
			configValue: types.StringValue("beta"),
			state:       createNonNullState(),
			plan:        createNonNullPlan(),
			validateFunc: func(t *testing.T, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
				if !resp.PlanValue.Equal(types.StringValue("beta")) {
					t.Errorf("expected plan unchanged, got %v", resp.PlanValue)
				}
			},
		},
		{
			name:        "delete_operation_plan_null_noop",
			stateValue:  types.StringValue("x"),
			planValue:   types.StringNull(),
			configValue: types.StringNull(),
			state:       createNonNullState(),
			plan:        createNullPlan(),
		},
		{
			name:        "plan_unknown_noop",
			stateValue:  types.StringValue("a"),
			planValue:   types.StringUnknown(),
			configValue: types.StringValue("a"),
			state:       createNonNullState(),
			plan:        createNonNullPlan(),
		},
		{
			name:        "config_unknown_noop",
			stateValue:  types.StringValue("a"),
			planValue:   types.StringValue("b"),
			configValue: types.StringUnknown(),
			state:       createNonNullState(),
			plan:        createNonNullPlan(),
		},
		{
			name:        "null_state_known_plan_noop",
			stateValue:  types.StringNull(),
			planValue:   types.StringValue("v"),
			configValue: types.StringValue("v"),
			state:       createNonNullState(),
			plan:        createNonNullPlan(),
			validateFunc: func(t *testing.T, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
				if !resp.PlanValue.Equal(types.StringValue("v")) {
					t.Errorf("expected plan unchanged, got %v", resp.PlanValue)
				}
			},
		},
		{
			name:        "known_state_null_plan_noop",
			stateValue:  types.StringValue("v"),
			planValue:   types.StringNull(),
			configValue: types.StringNull(),
			state:       createNonNullState(),
			plan:        createNonNullPlan(),
			validateFunc: func(t *testing.T, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
				if !resp.PlanValue.IsNull() {
					t.Errorf("expected plan still null, got %v", resp.PlanValue)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			modifier := CaseInsensitiveString()
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

			modifier.PlanModifyString(context.Background(), req, resp)

			if resp.Diagnostics.HasError() {
				t.Fatalf("expected no diagnostics, got %v", resp.Diagnostics.Errors())
			}

			if tt.validateFunc != nil {
				tt.validateFunc(t, req, resp)
			}
		})
	}
}

// TestTrailingSlashEqualStringModifier tests TrailingSlashEqualStringModifier.
func TestTrailingSlashEqualStringModifier(t *testing.T) {
	t.Parallel()

	nullState := tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{}}, nil)
	nonNullState := tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{}}, map[string]tftypes.Value{})
	nonNullPlan := tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{}}, map[string]tftypes.Value{})

	tests := []struct {
		name         string
		stateValue   types.String
		planValue    types.String
		configValue  types.String
		state        tfsdk.State
		plan         tfsdk.Plan
		validateFunc func(t *testing.T, req planmodifier.StringRequest, resp *planmodifier.StringResponse)
	}{
		{
			name:        "update_slash_only_diff_normalizes_plan_to_state",
			stateValue:  types.StringValue("secret/"),
			planValue:   types.StringValue("secret"),
			configValue: types.StringValue("secret"),
			state:       tfsdk.State{Raw: nonNullState, Schema: schema.Schema{}},
			plan:        tfsdk.Plan{Raw: nonNullPlan, Schema: schema.Schema{}},
			validateFunc: func(t *testing.T, _ planmodifier.StringRequest, resp *planmodifier.StringResponse) {
				if resp.PlanValue.ValueString() != "secret/" {
					t.Errorf("expected plan normalized to state value 'secret/', got %q", resp.PlanValue.ValueString())
				}
			},
		},
		{
			name:        "update_real_change_not_suppressed",
			stateValue:  types.StringValue("secret/"),
			planValue:   types.StringValue("other"),
			configValue: types.StringValue("other"),
			state:       tfsdk.State{Raw: nonNullState, Schema: schema.Schema{}},
			plan:        tfsdk.Plan{Raw: nonNullPlan, Schema: schema.Schema{}},
			validateFunc: func(t *testing.T, _ planmodifier.StringRequest, resp *planmodifier.StringResponse) {
				if resp.PlanValue.ValueString() != "other" {
					t.Errorf("expected plan unchanged as 'other', got %q", resp.PlanValue.ValueString())
				}
			},
		},
		{
			name:        "create_state_null_skips",
			stateValue:  types.StringValue("secret/"),
			planValue:   types.StringValue("secret"),
			configValue: types.StringValue("secret"),
			state:       tfsdk.State{Raw: nullState, Schema: schema.Schema{}},
			plan:        tfsdk.Plan{Raw: nonNullPlan, Schema: schema.Schema{}},
			validateFunc: func(t *testing.T, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
				if resp.PlanValue.ValueString() != req.PlanValue.ValueString() {
					t.Errorf("expected plan unchanged during create, got %q", resp.PlanValue.ValueString())
				}
			},
		},
		{
			name:        "already_equal_noop",
			stateValue:  types.StringValue("secret/"),
			planValue:   types.StringValue("secret/"),
			configValue: types.StringValue("secret/"),
			state:       tfsdk.State{Raw: nonNullState, Schema: schema.Schema{}},
			plan:        tfsdk.Plan{Raw: nonNullPlan, Schema: schema.Schema{}},
			validateFunc: func(t *testing.T, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
				if resp.PlanValue.ValueString() != "secret/" {
					t.Errorf("expected no change when already equal, got %q", resp.PlanValue.ValueString())
				}
			},
		},
		{
			// TrimSuffix strips exactly one slash; multiple trailing slashes are a real change.
			name:        "multiple_trailing_slashes_not_suppressed",
			stateValue:  types.StringValue("secret"),
			planValue:   types.StringValue("secret///"),
			configValue: types.StringValue("secret///"),
			state:       tfsdk.State{Raw: nonNullState, Schema: schema.Schema{}},
			plan:        tfsdk.Plan{Raw: nonNullPlan, Schema: schema.Schema{}},
			validateFunc: func(t *testing.T, _ planmodifier.StringRequest, resp *planmodifier.StringResponse) {
				if resp.PlanValue.ValueString() != "secret///" {
					t.Errorf("expected plan unchanged for multiple-slash diff, got %q", resp.PlanValue.ValueString())
				}
			},
		},
		{
			// TrimSuffix("", "/") == TrimSuffix("/", "/") == "", so a lone slash is semantically
			// equal to empty string and the plan is normalized to the state value.
			name:        "slash_equal_to_empty_normalizes_plan_to_state",
			stateValue:  types.StringValue(""),
			planValue:   types.StringValue("/"),
			configValue: types.StringValue("/"),
			state:       tfsdk.State{Raw: nonNullState, Schema: schema.Schema{}},
			plan:        tfsdk.Plan{Raw: nonNullPlan, Schema: schema.Schema{}},
			validateFunc: func(t *testing.T, _ planmodifier.StringRequest, resp *planmodifier.StringResponse) {
				if resp.PlanValue.ValueString() != "" {
					t.Errorf("expected plan normalized to empty state, got %q", resp.PlanValue.ValueString())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			modifier := TrailingSlashEqualString()
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

			modifier.PlanModifyString(context.Background(), req, resp)

			if resp.Diagnostics.HasError() {
				t.Fatalf("expected no diagnostics, got %v", resp.Diagnostics.Errors())
			}
			if tt.validateFunc != nil {
				tt.validateFunc(t, req, resp)
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

// TestDefaultBearingOptionalComputedPaths verifies that only Optional+Computed attributes with a
// non-nil Default are reported, mirroring the a.Default == nil guard in
// ApplyRemovedToUnknownModifiers. Nested defaults are reported via the child's dotted path.
func TestDefaultBearingOptionalComputedPaths(t *testing.T) {
	t.Parallel()

	attrs := map[string]schema.Attribute{
		"state": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString("ENABLED"),
		},
		"name": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"id": schema.StringAttribute{
			Computed: true,
		},
		"metadata": schema.SingleNestedAttribute{
			Optional: true,
			Computed: true,
			Attributes: map[string]schema.Attribute{
				"status": schema.StringAttribute{
					Optional: true,
					Computed: true,
					Default:  stringdefault.StaticString("ACTIVE"),
				},
			},
		},
	}

	got := DefaultBearingOptionalComputedPaths(attrs)
	want := []string{
		"metadata.status",
		"state",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DefaultBearingOptionalComputedPaths = %v, want %v", got, want)
	}
}
