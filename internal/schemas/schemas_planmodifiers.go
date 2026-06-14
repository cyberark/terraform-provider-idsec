// Copyright CyberArk. 2026
// SPDX-License-Identifier: Apache-2.0

// Package schemas provides schema generation and conversion utilities for the
// Terraform provider. It includes custom plan modifiers for enforcing attribute
// immutability, default value handlers, and validators.
package schemas

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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

// CaseInsensitiveStringModifier compares planned and prior string values with strings.EqualFold.
// When they match under case-folding but differ in exact spelling, the planned value is replaced
// with the state value so Terraform does not show a cosmetic update. Semantic changes are left
// unchanged and never produce diagnostics from this modifier.
type CaseInsensitiveStringModifier struct{}

// CaseInsensitiveString returns a plan modifier that normalizes case-only string differences
// against the value in state. It does not block or validate updates.
func CaseInsensitiveString() planmodifier.String {
	return CaseInsensitiveStringModifier{}
}

// Description returns a human-readable description of the plan modifier.
func (m CaseInsensitiveStringModifier) Description(_ context.Context) string {
	return "When the planned value equals the state value ignoring letter case, the plan uses the state's spelling."
}

// MarkdownDescription returns a markdown-formatted description of the plan modifier.
func (m CaseInsensitiveStringModifier) MarkdownDescription(_ context.Context) string {
	return "If the planned value matches state under **case-insensitive** comparison (`EqualFold`), the plan is updated to match state's exact casing. Other changes are not altered."
}

// PlanModifyString normalizes the plan when state and plan are equal under strings.EqualFold.
func (m CaseInsensitiveStringModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
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

	if req.StateValue.IsNull() || req.PlanValue.IsNull() {
		return
	}

	if req.PlanValue.Equal(req.StateValue) {
		return
	}

	stateStr := req.StateValue.ValueString()
	planStr := req.PlanValue.ValueString()
	if strings.EqualFold(stateStr, planStr) {
		resp.PlanValue = req.StateValue
	}
}

// SetNestedStableModifier suppresses spurious diffs for set-based nested attributes whose
// elements the backend may return in a different order and/or with server-computed fields that
// are unknown at plan time (for example read-only target metadata such as role_type).
//
// The terraform-plugin-framework cannot correlate set elements between prior state and plan,
// so the usual UseStateForUnknown modifier is ineffective inside sets. This modifier compares
// the planned and prior-state element collections directly: when every planned element can be
// uniquely matched to a distinct prior-state element on all of its known (non-unknown) attribute
// values, the entire prior-state value is reused as the planned value. That removes both the
// ordering churn and the "known after apply" churn for stable computed fields.
//
// Real changes are never masked: if an element is added, removed, or has a changed known value,
// no full bijection exists and the planned value is left untouched so Terraform plans normally.
type SetNestedStableModifier struct{}

// SetNestedStable returns a plan modifier that keeps set-based nested attributes stable across
// applies when their contents are semantically unchanged. See SetNestedStableModifier.
func SetNestedStable() planmodifier.Set {
	return SetNestedStableModifier{}
}

// Description returns a human-readable description of the plan modifier.
func (m SetNestedStableModifier) Description(_ context.Context) string {
	return "Reuses the prior state value when the set's elements are semantically unchanged, ignoring element order and server-computed fields that are unknown at plan time."
}

// MarkdownDescription returns a markdown-formatted description of the plan modifier.
func (m SetNestedStableModifier) MarkdownDescription(_ context.Context) string {
	return "Reuses the prior state value when the set's elements are **semantically unchanged**, ignoring element order and server-computed fields that are unknown at plan time."
}

// PlanModifySet implements the plan modification logic for set-based nested attributes.
func (m SetNestedStableModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	// Skip create (no prior state) and destroy (no plan).
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	planElems := req.PlanValue.Elements()
	stateElems := req.StateValue.Elements()
	if len(planElems) != len(stateElems) {
		return
	}

	used := make([]bool, len(stateElems))
	for _, planElem := range planElems {
		matched := -1
		for i, stateElem := range stateElems {
			if used[i] {
				continue
			}
			if attrValueSemanticMatch(planElem, stateElem) {
				matched = i
				break
			}
		}
		if matched == -1 {
			// At least one planned element has no semantically-equal counterpart in
			// state, so this is a real change. Leave the plan untouched.
			return
		}
		used[matched] = true
	}

	// Every planned element matched a distinct prior-state element: the set is
	// semantically unchanged. Reuse the prior state value to avoid a spurious diff.
	resp.PlanValue = req.StateValue
}

// attrValueSemanticMatch reports whether a planned attribute value is compatible with a
// prior-state attribute value, ignoring any planned values that are unknown at plan time.
//
// Unknown planned values always match (they will adopt the prior-state value). Object values
// are compared attribute-by-attribute recursively. All other values must be exactly equal.
func attrValueSemanticMatch(planVal attr.Value, stateVal attr.Value) bool {
	if planVal.IsUnknown() {
		return true
	}

	planObj, planIsObj := planVal.(types.Object)
	stateObj, stateIsObj := stateVal.(types.Object)
	if planIsObj && stateIsObj {
		stateAttrs := stateObj.Attributes()
		for key, planAttr := range planObj.Attributes() {
			stateAttr, ok := stateAttrs[key]
			if !ok {
				return false
			}
			if !attrValueSemanticMatch(planAttr, stateAttr) {
				return false
			}
		}
		return true
	}

	return planVal.Equal(stateVal)
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
func (m ImmutableSetModifier) PlanModifySet(_ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
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

// removedToNullDescription documents the shared behavior of the removed-to-null plan modifiers.
const removedToNullDescription = "Sets the planned value to null when an optional attribute is removed " +
	"from configuration (null in config) but had a value in prior state, so the removal produces an " +
	"explicit change to null (which is then actually removed on apply) instead of silently keeping the " +
	"prior value."

// valueIsAbsent reports whether a prior-state value carries no meaningful content, so that flipping it
// to null would only be a cosmetic shadow change. A value is absent when it is nil, null, an empty
// string, an empty collection, a scalar zero value (bool false, numeric 0), or an object whose
// attributes are all (recursively) absent.
func valueIsAbsent(v attr.Value) bool {
	if v == nil || v.IsNull() {
		return true
	}
	if v.IsUnknown() {
		return false
	}
	switch tv := v.(type) {
	case basetypes.StringValue:
		return tv.ValueString() == ""
	case basetypes.BoolValue:
		return !tv.ValueBool()
	case basetypes.Int64Value:
		return tv.ValueInt64() == 0
	case basetypes.Float64Value:
		return tv.ValueFloat64() == 0
	case basetypes.ListValue:
		return len(tv.Elements()) == 0
	case basetypes.SetValue:
		return len(tv.Elements()) == 0
	case basetypes.MapValue:
		return len(tv.Elements()) == 0
	case basetypes.ObjectValue:
		for _, av := range tv.Attributes() {
			if !valueIsAbsent(av) {
				return false
			}
		}
		return true
	}
	return false
}

// isUserRemoval reports whether an attribute was explicitly removed by the user: its configuration is
// null while the prior state held a meaningful (non-absent) value.
func isUserRemoval(configVal, stateVal attr.Value) bool {
	if configVal == nil || configVal.IsUnknown() || !configVal.IsNull() {
		return false
	}
	return !valueIsAbsent(stateVal)
}

// historyLoader loads the user-set history used to gate plan-time removal. It is a package variable so
// tests can inject history without constructing the framework's internal private-state type (which is
// not importable).
var historyLoader = ReadUserSetPaths

// shouldRemoveToNull is the pure removal decision: a genuine user removal (config null over a
// meaningful prior value) whose path was previously recorded as user-set. Server-defaulted attributes
// the user never set are absent from history and are therefore not nulled.
func shouldRemoveToNull(history map[string]bool, attrPath string, configVal, stateVal attr.Value) bool {
	if !isUserRemoval(configVal, stateVal) {
		return false
	}
	return pathInUserSetHistory(history, attrPath)
}

// isHistoryGatedRemoval reports whether an attribute should be planned as removed-to-null. It loads
// the user-set history from private state and delegates to shouldRemoveToNull. The history gate is
// what distinguishes a real user removal from an Optional+Computed attribute the user never set but
// the backend defaulted: the latter is absent from history, so it is preserved (via
// UseStateForUnknown) instead of perpetually planning value -> null.
func isHistoryGatedRemoval(ctx context.Context, private privateStateReader, attrPath string, configVal, stateVal attr.Value) bool {
	return shouldRemoveToNull(historyLoader(ctx, private), attrPath, configVal, stateVal)
}

// RemovedToNullString returns a plan modifier that nulls a removed optional+computed string attribute.
func RemovedToNullString() planmodifier.String { return removedToNullStringModifier{} }

type removedToNullStringModifier struct{}

func (m removedToNullStringModifier) Description(_ context.Context) string {
	return removedToNullDescription
}
func (m removedToNullStringModifier) MarkdownDescription(_ context.Context) string {
	return removedToNullDescription
}
func (m removedToNullStringModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if isHistoryGatedRemoval(ctx, req.Private, req.Path.String(), req.ConfigValue, req.StateValue) {
		resp.PlanValue = types.StringNull()
	}
}

// RemovedToNullBool returns a plan modifier that nulls a removed optional+computed bool attribute.
func RemovedToNullBool() planmodifier.Bool { return removedToNullBoolModifier{} }

type removedToNullBoolModifier struct{}

func (m removedToNullBoolModifier) Description(_ context.Context) string {
	return removedToNullDescription
}
func (m removedToNullBoolModifier) MarkdownDescription(_ context.Context) string {
	return removedToNullDescription
}
func (m removedToNullBoolModifier) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	if isHistoryGatedRemoval(ctx, req.Private, req.Path.String(), req.ConfigValue, req.StateValue) {
		resp.PlanValue = types.BoolNull()
	}
}

// RemovedToNullInt64 returns a plan modifier that nulls a removed optional+computed int64 attribute.
func RemovedToNullInt64() planmodifier.Int64 { return removedToNullInt64Modifier{} }

type removedToNullInt64Modifier struct{}

func (m removedToNullInt64Modifier) Description(_ context.Context) string {
	return removedToNullDescription
}
func (m removedToNullInt64Modifier) MarkdownDescription(_ context.Context) string {
	return removedToNullDescription
}
func (m removedToNullInt64Modifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if isHistoryGatedRemoval(ctx, req.Private, req.Path.String(), req.ConfigValue, req.StateValue) {
		resp.PlanValue = types.Int64Null()
	}
}

// RemovedToNullList returns a plan modifier that nulls a removed optional+computed list attribute.
func RemovedToNullList() planmodifier.List { return removedToNullListModifier{} }

type removedToNullListModifier struct{}

func (m removedToNullListModifier) Description(_ context.Context) string {
	return removedToNullDescription
}
func (m removedToNullListModifier) MarkdownDescription(_ context.Context) string {
	return removedToNullDescription
}
func (m removedToNullListModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if isHistoryGatedRemoval(ctx, req.Private, req.Path.String(), req.ConfigValue, req.StateValue) {
		resp.PlanValue = types.ListNull(req.StateValue.ElementType(ctx))
	}
}

// RemovedToNullSet returns a plan modifier that nulls a removed optional+computed set attribute.
func RemovedToNullSet() planmodifier.Set { return removedToNullSetModifier{} }

type removedToNullSetModifier struct{}

func (m removedToNullSetModifier) Description(_ context.Context) string {
	return removedToNullDescription
}
func (m removedToNullSetModifier) MarkdownDescription(_ context.Context) string {
	return removedToNullDescription
}
func (m removedToNullSetModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	if isHistoryGatedRemoval(ctx, req.Private, req.Path.String(), req.ConfigValue, req.StateValue) {
		resp.PlanValue = types.SetNull(req.StateValue.ElementType(ctx))
	}
}

// RemovedToNullMap returns a plan modifier that nulls a removed optional+computed map attribute.
func RemovedToNullMap() planmodifier.Map { return removedToNullMapModifier{} }

type removedToNullMapModifier struct{}

func (m removedToNullMapModifier) Description(_ context.Context) string {
	return removedToNullDescription
}
func (m removedToNullMapModifier) MarkdownDescription(_ context.Context) string {
	return removedToNullDescription
}
func (m removedToNullMapModifier) PlanModifyMap(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
	if isHistoryGatedRemoval(ctx, req.Private, req.Path.String(), req.ConfigValue, req.StateValue) {
		resp.PlanValue = types.MapNull(req.StateValue.ElementType(ctx))
	}
}

// RemovedToNullObject returns a plan modifier that nulls a removed optional+computed object attribute.
func RemovedToNullObject() planmodifier.Object { return removedToNullObjectModifier{} }

type removedToNullObjectModifier struct{}

func (m removedToNullObjectModifier) Description(_ context.Context) string {
	return removedToNullDescription
}
func (m removedToNullObjectModifier) MarkdownDescription(_ context.Context) string {
	return removedToNullDescription
}
func (m removedToNullObjectModifier) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	if isHistoryGatedRemoval(ctx, req.Private, req.Path.String(), req.ConfigValue, req.StateValue) {
		resp.PlanValue = types.ObjectNull(req.StateValue.AttributeTypes(ctx))
	}
}

// isComputedOnlyAttr reports whether an attribute is server-managed (computed-only): computed, and
// neither optional nor required. Such attributes (and their nested subtrees) must be skipped: their
// config is always null, so asserting null would conflict with the value the backend supplies.
func isComputedOnlyAttr(optional, required, computed bool) bool {
	return computed && !optional && !required
}

// ApplyRemovedToNullModifiers walks an attribute tree and, for every Optional+Computed attribute
// (recursively into nested objects), appends two plan modifiers in order: UseStateForUnknown followed by
// the matching removed-to-null modifier. It leaves required, default-bearing, and computed-only
// (server-managed) attributes untouched, and does not descend into computed-only objects.
func ApplyRemovedToNullModifiers(attributes map[string]schema.Attribute, skipAttrs ...string) {
	skip := make(map[string]bool, len(skipAttrs))
	for _, name := range skipAttrs {
		skip[name] = true
	}
	for name, attribute := range attributes {
		if skip[name] {
			continue
		}
		switch a := attribute.(type) {
		case schema.StringAttribute:
			if a.Optional && a.Computed && a.Default == nil {
				a.PlanModifiers = append(a.PlanModifiers, stringplanmodifier.UseStateForUnknown(), RemovedToNullString())
				attributes[name] = a
			}
		case schema.BoolAttribute:
			if a.Optional && a.Computed && a.Default == nil {
				a.PlanModifiers = append(a.PlanModifiers, boolplanmodifier.UseStateForUnknown(), RemovedToNullBool())
				attributes[name] = a
			}
		case schema.Int64Attribute:
			if a.Optional && a.Computed && a.Default == nil {
				a.PlanModifiers = append(a.PlanModifiers, int64planmodifier.UseStateForUnknown(), RemovedToNullInt64())
				attributes[name] = a
			}
		case schema.ListAttribute:
			if a.Optional && a.Computed && a.Default == nil {
				a.PlanModifiers = append(a.PlanModifiers, listplanmodifier.UseStateForUnknown(), RemovedToNullList())
				attributes[name] = a
			}
		case schema.SetAttribute:
			if a.Optional && a.Computed && a.Default == nil {
				a.PlanModifiers = append(a.PlanModifiers, setplanmodifier.UseStateForUnknown(), RemovedToNullSet())
				attributes[name] = a
			}
		case schema.MapAttribute:
			if a.Optional && a.Computed && a.Default == nil {
				a.PlanModifiers = append(a.PlanModifiers, mapplanmodifier.UseStateForUnknown(), RemovedToNullMap())
				attributes[name] = a
			}
		case schema.SingleNestedAttribute:
			if isComputedOnlyAttr(a.Optional, a.Required, a.Computed) {
				break
			}
			if a.Optional && a.Computed && a.Default == nil {
				a.PlanModifiers = append(a.PlanModifiers, objectplanmodifier.UseStateForUnknown(), RemovedToNullObject())
			}
			ApplyRemovedToNullModifiers(a.Attributes)
			attributes[name] = a
		case schema.ListNestedAttribute:
			if isComputedOnlyAttr(a.Optional, a.Required, a.Computed) {
				break
			}
			if a.Optional && a.Computed && a.Default == nil {
				a.PlanModifiers = append(a.PlanModifiers, listplanmodifier.UseStateForUnknown(), RemovedToNullList())
			}
			ApplyRemovedToNullModifiers(a.NestedObject.Attributes)
			attributes[name] = a
		case schema.SetNestedAttribute:
			if isComputedOnlyAttr(a.Optional, a.Required, a.Computed) {
				break
			}
			if a.Optional && a.Computed && a.Default == nil {
				a.PlanModifiers = append(a.PlanModifiers, setplanmodifier.UseStateForUnknown(), RemovedToNullSet())
			}
			ApplyRemovedToNullModifiers(a.NestedObject.Attributes)
			attributes[name] = a
		case schema.MapNestedAttribute:
			if isComputedOnlyAttr(a.Optional, a.Required, a.Computed) {
				break
			}
			if a.Optional && a.Computed && a.Default == nil {
				a.PlanModifiers = append(a.PlanModifiers, mapplanmodifier.UseStateForUnknown(), RemovedToNullMap())
			}
			ApplyRemovedToNullModifiers(a.NestedObject.Attributes)
			attributes[name] = a
		}
	}
}
