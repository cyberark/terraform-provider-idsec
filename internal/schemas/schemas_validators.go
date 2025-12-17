// Copyright (c) CyberArk.
// SPDX-License-Identifier: Apache-2.0

package schemas

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StringDefault is a default value for string attributes.
type StringDefault struct {
	Value string
}

// Description returns a description of the default value.
func (d StringDefault) Description(ctx context.Context) string {
	return "Default value for string attribute"
}

// MarkdownDescription returns a markdown description of the default value.
func (d StringDefault) MarkdownDescription(ctx context.Context) string {
	return "Default value for **string** attribute"
}

// DefaultString sets the default value for string attributes.
func (d StringDefault) DefaultString(ctx context.Context, req defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(d.Value)
}

// BoolDefault is a default value for boolean attributes.
type BoolDefault struct {
	Value bool
}

// Description returns a description of the default value.
func (d BoolDefault) Description(ctx context.Context) string {
	return "Default value for boolean attribute"
}

// MarkdownDescription returns a markdown description of the default value.
func (d BoolDefault) MarkdownDescription(ctx context.Context) string {
	return "Default value for **boolean** attribute"
}

// DefaultBool sets the default value for boolean attributes.
func (d BoolDefault) DefaultBool(ctx context.Context, req defaults.BoolRequest, resp *defaults.BoolResponse) {
	resp.PlanValue = types.BoolValue(d.Value)
}

// Int64Default is a default value for int64 attributes.
type Int64Default struct {
	Value int64
}

// Description returns a description of the default value.
func (d Int64Default) Description(ctx context.Context) string {
	return "Default value for int64 attribute"
}

// MarkdownDescription returns a markdown description of the default value.
func (d Int64Default) MarkdownDescription(ctx context.Context) string {
	return "Default value for **int64** attribute"
}

// DefaultInt64 sets the default value for int64 attributes.
func (d Int64Default) DefaultInt64(ctx context.Context, req defaults.Int64Request, resp *defaults.Int64Response) {
	resp.PlanValue = types.Int64Value(d.Value)
}

// SetStringDefault is a default value for set of strings attributes.
type SetStringDefault struct {
	Values []string
}

// Description returns a description of the default value.
func (d SetStringDefault) Description(ctx context.Context) string {
	return "Default value for set of strings attribute"
}

// MarkdownDescription returns a markdown description of the default value.
func (d SetStringDefault) MarkdownDescription(ctx context.Context) string {
	return "Default value for **set of strings** attribute"
}

// DefaultSet sets the default value for set attributes.
func (d SetStringDefault) DefaultSet(ctx context.Context, req defaults.SetRequest, resp *defaults.SetResponse) {
	values := make([]attr.Value, len(d.Values))
	for i, v := range d.Values {
		values[i] = types.StringValue(v)
	}
	resp.PlanValue = types.SetValueMust(types.StringType, values)
}

// SetNumericDefault is a default value for set of numerics attributes.
type SetNumericDefault struct {
	Values []int64
}

// Description returns a description of the default value.
func (d SetNumericDefault) Description(ctx context.Context) string {
	return "Default value for set of numerics attribute"
}

// MarkdownDescription returns a markdown description of the default value.
func (d SetNumericDefault) MarkdownDescription(ctx context.Context) string {
	return "Default value for **set of numerics** attribute"
}

// DefaultSet sets the default value for set attributes.
func (d SetNumericDefault) DefaultSet(ctx context.Context, req defaults.SetRequest, resp *defaults.SetResponse) {
	values := make([]attr.Value, len(d.Values))
	for i, v := range d.Values {
		values[i] = types.Int64Value(v)
	}
	resp.PlanValue = types.SetValueMust(types.Int64Type, values)
}

// SetBoolDefault is a default value for set of bools attributes.
type SetBoolDefault struct {
	Values []bool
}

// Description returns a description of the default value.
func (d SetBoolDefault) Description(ctx context.Context) string {
	return "Default value for set of bools attribute"
}

// MarkdownDescription returns a markdown description of the default value.
func (d SetBoolDefault) MarkdownDescription(ctx context.Context) string {
	return "Default value for **set of bools** attribute"
}

// DefaultSet sets the default value for set attributes.
func (d SetBoolDefault) DefaultSet(ctx context.Context, req defaults.SetRequest, resp *defaults.SetResponse) {
	values := make([]attr.Value, len(d.Values))
	for i, v := range d.Values {
		values[i] = types.BoolValue(v)
	}
	resp.PlanValue = types.SetValueMust(types.BoolType, values)
}

// ListStringDefault is a default value for list of strings attributes.
type ListStringDefault struct {
	Values []string
}

// Description returns a description of the default value.
func (d ListStringDefault) Description(ctx context.Context) string {
	return "Default value for list of strings attribute"
}

// MarkdownDescription returns a markdown description of the default value.
func (d ListStringDefault) MarkdownDescription(ctx context.Context) string {
	return "Default value for **list of strings** attribute"
}

// DefaultList sets the default value for list attributes.
func (d ListStringDefault) DefaultList(ctx context.Context, req defaults.ListRequest, resp *defaults.ListResponse) {
	values := make([]attr.Value, len(d.Values))
	for i, v := range d.Values {
		values[i] = types.StringValue(v)
	}
	resp.PlanValue = types.ListValueMust(types.StringType, values)
}

// ListNumericDefault is a default value for list of numerics attributes.
type ListNumericDefault struct {
	Values []int64
}

// Description returns a description of the default value.
func (d ListNumericDefault) Description(ctx context.Context) string {
	return "Default value for list of numerics attribute"
}

// MarkdownDescription returns a markdown description of the default value.
func (d ListNumericDefault) MarkdownDescription(ctx context.Context) string {
	return "Default value for **list of numerics** attribute"
}

// DefaultList sets the default value for list attributes.
func (d ListNumericDefault) DefaultList(ctx context.Context, req defaults.ListRequest, resp *defaults.ListResponse) {
	values := make([]attr.Value, len(d.Values))
	for i, v := range d.Values {
		values[i] = types.Int64Value(v)
	}
	resp.PlanValue = types.ListValueMust(types.Int64Type, values)
}

// ListBoolDefault is a default value for list of bools attributes.
type ListBoolDefault struct {
	Values []bool
}

// Description returns a description of the default value.
func (d ListBoolDefault) Description(ctx context.Context) string {
	return "Default value for list of bools attribute"
}

// MarkdownDescription returns a markdown description of the default value.
func (d ListBoolDefault) MarkdownDescription(ctx context.Context) string {
	return "Default value for **list of bools** attribute"
}

// DefaultList sets the default value for list attributes.
func (d ListBoolDefault) DefaultList(ctx context.Context, req defaults.ListRequest, resp *defaults.ListResponse) {
	values := make([]attr.Value, len(d.Values))
	for i, v := range d.Values {
		values[i] = types.BoolValue(v)
	}
	resp.PlanValue = types.ListValueMust(types.BoolType, values)
}

// StringInChoicesValidator ensures a string is in the allowed choices.
type StringInChoicesValidator struct {
	Choices []string
}

// Description returns a description of the validator.
func (v StringInChoicesValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("Value must be one of: %s", strings.Join(v.Choices, ", "))
}

// MarkdownDescription returns a markdown description of the validator.
func (v StringInChoicesValidator) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("Value must be one of: `%s`", strings.Join(v.Choices, "`, `"))
}

// ValidateString checks if the string is in the allowed choices.
func (v StringInChoicesValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if slices.Contains(v.Choices, value) {
		return
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid Value",
		fmt.Sprintf("Value must be one of: %s", strings.Join(v.Choices, ", ")),
	)
}

// SliceInChoicesValidator ensures all strings in a slice are in the allowed choices.
type SliceInChoicesValidator struct {
	Choices []string
}

// Description returns a description of the validator.
func (v SliceInChoicesValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("All values must be one of: %s", strings.Join(v.Choices, ", "))
}

// MarkdownDescription returns a markdown description of the validator.
func (v SliceInChoicesValidator) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("All values must be one of: `%s`", strings.Join(v.Choices, "`, `"))
}

// ValidateList checks if all strings in the list are in the allowed choices.
func (v SliceInChoicesValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var values []string
	diags := req.ConfigValue.ElementsAs(ctx, &values, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, value := range values {
		valid := slices.Contains(v.Choices, value)
		if !valid {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid Value in List",
				fmt.Sprintf("All values must be one of: %s", strings.Join(v.Choices, ", ")),
			)
			return
		}
	}
}

// SliceInSetValidator ensures all strings in a slice are in the allowed choices.
type SliceInSetValidator struct {
	Choices []string
}

// Description returns a description of the validator.
func (v SliceInSetValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("All values must be one of: %s", strings.Join(v.Choices, ", "))
}

// MarkdownDescription returns a markdown description of the validator.
func (v SliceInSetValidator) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("All values must be one of: `%s`", strings.Join(v.Choices, "`, `"))
}

// ValidateSet checks if all strings in the set are in the allowed choices.
func (v SliceInSetValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var values []string
	diags := req.ConfigValue.ElementsAs(ctx, &values, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, value := range values {
		valid := slices.Contains(v.Choices, value)
		if !valid {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid Value in Set",
				fmt.Sprintf("All values must be one of: %s", strings.Join(v.Choices, ", ")),
			)
			return
		}
	}
}
