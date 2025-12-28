// Copyright (c) CyberArk.
// SPDX-License-Identifier: Apache-2.0

// Package schemas provides schema generation and conversion utilities for the
// Terraform provider. It includes custom plan modifiers for enforcing attribute
// immutability, default value handlers, and validators.
package schemas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

const (
	// errImmutableAttributeSummary is the error summary for immutable attribute modification attempts.
	errImmutableAttributeSummary = "Immutable Attribute Cannot Be Changed"

	// errImmutableAttributeDetailWithValues is the error detail template for immutable attributes
	// with displayable values (string, int64, bool). The format string expects:
	//   - %s: attribute path
	//   - %v: current value (works for string, int64, bool)
	//   - %v: attempted new value (works for string, int64, bool)
	errImmutableAttributeDetailWithValues = "The attribute '%s' is immutable and cannot be changed after resource creation.\n\n" +
		"Current value: %v\n" +
		"Attempted new value: %v\n\n" +
		"To use a different value, you must create a new resource."

	// errImmutableAttributeDetailSimple is the error detail template for immutable attributes
	// without displayable values (list, set, map). The format string expects:
	//   - %s: attribute path
	errImmutableAttributeDetailSimple = "The attribute '%s' is immutable and cannot be changed after resource creation.\n\n" +
		"To use a different value, you must create a new resource."
)

// ImmutableStringModifier prevents changes to string attributes after resource creation.
//
// This plan modifier implements the planmodifier.String interface and blocks any
// plan that attempts to modify an immutable attribute value. It allows resource
// creation, deletion, and no-op updates while preventing value changes.
//
// The modifier follows Terraform Plugin Framework best practices as documented at:
// https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification
//
// Example usage:
//
//	schema.StringAttribute{
//	    Description: "Immutable identifier",
//	    PlanModifiers: []planmodifier.String{
//	        ImmutableString(),
//	    },
//	}
type ImmutableStringModifier struct{}

// ImmutableString returns a plan modifier that prevents changes to string attributes
// after resource creation. Use this for identity fields that should never change.
//
// Returns a plan modifier implementing planmodifier.String interface.
func ImmutableString() planmodifier.String {
	return ImmutableStringModifier{}
}

// Description returns a human-readable description of the plan modifier.
//
// Parameters:
//   - ctx: Context for the operation (unused but required by interface)
//
// Returns a description string for use in Terraform documentation.
func (m ImmutableStringModifier) Description(_ context.Context) string {
	return "Prevents changes to this attribute after initial creation. Any attempt to modify will result in an error."
}

// MarkdownDescription returns a markdown-formatted description of the plan modifier.
//
// Parameters:
//   - ctx: Context for the operation (unused but required by interface)
//
// Returns a markdown description string for use in Terraform documentation.
func (m ImmutableStringModifier) MarkdownDescription(_ context.Context) string {
	return "**Immutable attribute** - Cannot be changed after initial creation. Any modification attempt will result in an error."
}

// PlanModifyString implements the plan modification logic for string attributes.
//
// This method is called by Terraform during the planning phase to determine if
// changes to the attribute should be allowed. It checks whether the resource is
// being created, deleted, or updated, and blocks updates that would change the
// value of an immutable attribute.
//
// Parameters:
//   - ctx: Context for the operation, can be used for cancellation
//   - req: The plan modification request containing state, plan, and config values
//   - resp: The response where diagnostics or plan modifications are written
//
// The method adds an attribute error to resp.Diagnostics if a change is detected,
// which will cause the Terraform plan to fail.
func (m ImmutableStringModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Check if the resource is being created (no prior state exists).
	// Per Terraform docs: https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification#checking-resource-change-operations
	if req.State.Raw.IsNull() {
		return
	}

	// Allow unknown plan values - these occur during interpolation and computed values.
	// We cannot validate unknown values, so we must allow them through.
	if req.PlanValue.IsUnknown() {
		return
	}

	// Allow unknown configuration values to prevent interpolation issues.
	// Per Terraform docs example in UseStateForUnknown modifier.
	if req.ConfigValue.IsUnknown() {
		return
	}

	// Check if the resource is being destroyed (plan is null).
	// Per Terraform docs: https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification#checking-resource-change-operations
	if req.Plan.Raw.IsNull() {
		return
	}

	// Allow if the attribute value is not changing
	if req.PlanValue.Equal(req.StateValue) {
		return
	}

	// BLOCK: Values differ - this is an attempt to modify an immutable attribute.
	// Add an error diagnostic to prevent the plan from succeeding.
	resp.Diagnostics.AddAttributeError(
		req.Path,
		errImmutableAttributeSummary,
		fmt.Sprintf(
			errImmutableAttributeDetailWithValues,
			req.Path.String(),
			req.StateValue.ValueString(),
			req.PlanValue.ValueString(),
		),
	)
}

// ImmutableInt64Modifier prevents changes to int64 attributes after resource creation.
//
// This plan modifier implements the planmodifier.Int64 interface with the same
// behavior as ImmutableStringModifier but for numeric attributes.
type ImmutableInt64Modifier struct{}

// ImmutableInt64 returns a plan modifier that prevents changes to int64 attributes
// after resource creation.
//
// Returns a plan modifier implementing planmodifier.Int64 interface.
func ImmutableInt64() planmodifier.Int64 {
	return ImmutableInt64Modifier{}
}

// Description returns a human-readable description of the plan modifier.
//
// Parameters:
//   - ctx: Context for the operation (unused but required by interface)
//
// Returns a description string for use in Terraform documentation.
func (m ImmutableInt64Modifier) Description(_ context.Context) string {
	return "Prevents changes to this attribute after initial creation. Any attempt to modify will result in an error."
}

// MarkdownDescription returns a markdown-formatted description of the plan modifier.
//
// Parameters:
//   - ctx: Context for the operation (unused but required by interface)
//
// Returns a markdown description string for use in Terraform documentation.
func (m ImmutableInt64Modifier) MarkdownDescription(_ context.Context) string {
	return "**Immutable attribute** - Cannot be changed after initial creation. Any modification attempt will result in an error."
}

// PlanModifyInt64 implements the plan modification logic for int64 attributes.
//
// Parameters:
//   - ctx: Context for the operation
//   - req: The plan modification request containing state, plan, and config values
//   - resp: The response where diagnostics or plan modifications are written
func (m ImmutableInt64Modifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if req.State.Raw.IsNull() {
		return
	}
	if req.PlanValue.IsUnknown() {
		return
	}
	if req.ConfigValue.IsUnknown() {
		return
	}
	if req.Plan.Raw.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		errImmutableAttributeSummary,
		fmt.Sprintf(
			errImmutableAttributeDetailWithValues,
			req.Path.String(),
			req.StateValue.ValueInt64(),
			req.PlanValue.ValueInt64(),
		),
	)
}

// ImmutableBoolModifier prevents changes to bool attributes after resource creation.
//
// This plan modifier implements the planmodifier.Bool interface with the same
// behavior as ImmutableStringModifier but for boolean attributes.
type ImmutableBoolModifier struct{}

// ImmutableBool returns a plan modifier that prevents changes to bool attributes
// after resource creation.
//
// Returns a plan modifier implementing planmodifier.Bool interface.
func ImmutableBool() planmodifier.Bool {
	return ImmutableBoolModifier{}
}

// Description returns a human-readable description of the plan modifier.
//
// Parameters:
//   - ctx: Context for the operation (unused but required by interface)
//
// Returns a description string for use in Terraform documentation.
func (m ImmutableBoolModifier) Description(_ context.Context) string {
	return "Prevents changes to this attribute after initial creation. Any attempt to modify will result in an error."
}

// MarkdownDescription returns a markdown-formatted description of the plan modifier.
//
// Parameters:
//   - ctx: Context for the operation (unused but required by interface)
//
// Returns a markdown description string for use in Terraform documentation.
func (m ImmutableBoolModifier) MarkdownDescription(_ context.Context) string {
	return "**Immutable attribute** - Cannot be changed after initial creation. Any modification attempt will result in an error."
}

// PlanModifyBool implements the plan modification logic for bool attributes.
//
// Parameters:
//   - ctx: Context for the operation
//   - req: The plan modification request containing state, plan, and config values
//   - resp: The response where diagnostics or plan modifications are written
func (m ImmutableBoolModifier) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	if req.State.Raw.IsNull() {
		return
	}
	if req.PlanValue.IsUnknown() {
		return
	}
	if req.ConfigValue.IsUnknown() {
		return
	}
	if req.Plan.Raw.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		errImmutableAttributeSummary,
		fmt.Sprintf(
			errImmutableAttributeDetailWithValues,
			req.Path.String(),
			req.StateValue.ValueBool(),
			req.PlanValue.ValueBool(),
		),
	)
}

// ImmutableListModifier prevents changes to list attributes after resource creation.
//
// This plan modifier implements the planmodifier.List interface with the same
// behavior as ImmutableStringModifier but for list attributes.
type ImmutableListModifier struct{}

// ImmutableList returns a plan modifier that prevents changes to list attributes
// after resource creation.
//
// Returns a plan modifier implementing planmodifier.List interface.
func ImmutableList() planmodifier.List {
	return ImmutableListModifier{}
}

// Description returns a human-readable description of the plan modifier.
//
// Parameters:
//   - ctx: Context for the operation (unused but required by interface)
//
// Returns a description string for use in Terraform documentation.
func (m ImmutableListModifier) Description(_ context.Context) string {
	return "Prevents changes to this attribute after initial creation. Any attempt to modify will result in an error."
}

// MarkdownDescription returns a markdown-formatted description of the plan modifier.
//
// Parameters:
//   - ctx: Context for the operation (unused but required by interface)
//
// Returns a markdown description string for use in Terraform documentation.
func (m ImmutableListModifier) MarkdownDescription(_ context.Context) string {
	return "**Immutable attribute** - Cannot be changed after initial creation. Any modification attempt will result in an error."
}

// PlanModifyList implements the plan modification logic for list attributes.
//
// Parameters:
//   - ctx: Context for the operation
//   - req: The plan modification request containing state, plan, and config values
//   - resp: The response where diagnostics or plan modifications are written
func (m ImmutableListModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if req.State.Raw.IsNull() {
		return
	}
	if req.PlanValue.IsUnknown() {
		return
	}
	if req.ConfigValue.IsUnknown() {
		return
	}
	if req.Plan.Raw.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		errImmutableAttributeSummary,
		fmt.Sprintf(
			errImmutableAttributeDetailSimple,
			req.Path.String(),
		),
	)
}

// ImmutableSetModifier prevents changes to set attributes after resource creation.
//
// This plan modifier implements the planmodifier.Set interface with the same
// behavior as ImmutableStringModifier but for set attributes.
type ImmutableSetModifier struct{}

// ImmutableSet returns a plan modifier that prevents changes to set attributes
// after resource creation.
//
// Returns a plan modifier implementing planmodifier.Set interface.
func ImmutableSet() planmodifier.Set {
	return ImmutableSetModifier{}
}

// Description returns a human-readable description of the plan modifier.
//
// Parameters:
//   - ctx: Context for the operation (unused but required by interface)
//
// Returns a description string for use in Terraform documentation.
func (m ImmutableSetModifier) Description(_ context.Context) string {
	return "Prevents changes to this attribute after initial creation. Any attempt to modify will result in an error."
}

// MarkdownDescription returns a markdown-formatted description of the plan modifier.
//
// Parameters:
//   - ctx: Context for the operation (unused but required by interface)
//
// Returns a markdown description string for use in Terraform documentation.
func (m ImmutableSetModifier) MarkdownDescription(_ context.Context) string {
	return "**Immutable attribute** - Cannot be changed after initial creation. Any modification attempt will result in an error."
}

// PlanModifySet implements the plan modification logic for set attributes.
//
// Parameters:
//   - ctx: Context for the operation
//   - req: The plan modification request containing state, plan, and config values
//   - resp: The response where diagnostics or plan modifications are written
func (m ImmutableSetModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	if req.State.Raw.IsNull() {
		return
	}
	if req.PlanValue.IsUnknown() {
		return
	}
	if req.ConfigValue.IsUnknown() {
		return
	}
	if req.Plan.Raw.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		errImmutableAttributeSummary,
		fmt.Sprintf(
			errImmutableAttributeDetailSimple,
			req.Path.String(),
		),
	)
}

// ImmutableMapModifier prevents changes to map attributes after resource creation.
//
// This plan modifier implements the planmodifier.Map interface with the same
// behavior as ImmutableStringModifier but for map attributes.
type ImmutableMapModifier struct{}

// ImmutableMap returns a plan modifier that prevents changes to map attributes
// after resource creation.
//
// Returns a plan modifier implementing planmodifier.Map interface.
func ImmutableMap() planmodifier.Map {
	return ImmutableMapModifier{}
}

// Description returns a human-readable description of the plan modifier.
//
// Parameters:
//   - ctx: Context for the operation (unused but required by interface)
//
// Returns a description string for use in Terraform documentation.
func (m ImmutableMapModifier) Description(_ context.Context) string {
	return "Prevents changes to this attribute after initial creation. Any attempt to modify will result in an error."
}

// MarkdownDescription returns a markdown-formatted description of the plan modifier.
//
// Parameters:
//   - ctx: Context for the operation (unused but required by interface)
//
// Returns a markdown description string for use in Terraform documentation.
func (m ImmutableMapModifier) MarkdownDescription(_ context.Context) string {
	return "**Immutable attribute** - Cannot be changed after initial creation. Any modification attempt will result in an error."
}

// PlanModifyMap implements the plan modification logic for map attributes.
//
// Parameters:
//   - ctx: Context for the operation
//   - req: The plan modification request containing state, plan, and config values
//   - resp: The response where diagnostics or plan modifications are written
func (m ImmutableMapModifier) PlanModifyMap(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
	if req.State.Raw.IsNull() {
		return
	}
	if req.PlanValue.IsUnknown() {
		return
	}
	if req.ConfigValue.IsUnknown() {
		return
	}
	if req.Plan.Raw.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		errImmutableAttributeSummary,
		fmt.Sprintf(
			errImmutableAttributeDetailSimple,
			req.Path.String(),
		),
	)
}
