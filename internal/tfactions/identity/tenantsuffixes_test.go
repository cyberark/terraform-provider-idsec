// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var tenantSuffixObjType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"id":     tftypes.String,
		"suffix": tftypes.String,
	},
}

var tenantSuffixTestSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"id":     schema.StringAttribute{Computed: true},
		"suffix": schema.StringAttribute{Required: true},
	},
}

func tenantSuffixValue(id, suffix interface{}) tftypes.Value {
	return tftypes.NewValue(tenantSuffixObjType, map[string]tftypes.Value{
		"id":     tftypes.NewValue(tftypes.String, id),
		"suffix": tftypes.NewValue(tftypes.String, suffix),
	})
}

func planRenameID(t *testing.T, state, plan tftypes.Value) types.String {
	t.Helper()
	req := resource.ModifyPlanRequest{
		State: tfsdk.State{Raw: state, Schema: tenantSuffixTestSchema},
		Plan:  tfsdk.Plan{Raw: plan, Schema: tenantSuffixTestSchema},
	}
	resp := resource.ModifyPlanResponse{Plan: req.Plan}
	tenantSuffixRenameIDValidator{}.ValidatePlan(context.Background(), req, &resp, nil)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	if resp.Plan.Raw.IsNull() {
		return types.StringNull()
	}
	var id types.String
	if diags := resp.Plan.GetAttribute(context.Background(), path.Root("id"), &id); diags.HasError() {
		t.Fatalf("reading planned id: %v", diags)
	}
	return id
}

func TestTenantSuffixRenamePlansTheIDAsUnknown(t *testing.T) {
	id := planRenameID(t,
		tenantSuffixValue("aek7723.a.example.com", "a.example.com"),
		tenantSuffixValue("aek7723.a.example.com", "b.example.com"))
	if !id.IsUnknown() {
		t.Fatalf("id = %v, want unknown so the apply records the renamed key", id)
	}
}

func TestTenantSuffixCaseOnlyChangeKeepsTheID(t *testing.T) {
	id := planRenameID(t,
		tenantSuffixValue("aek7723.a.example.com", "a.example.com"),
		tenantSuffixValue("aek7723.a.example.com", "A.Example.com"))
	if id.ValueString() != "aek7723.a.example.com" {
		t.Fatalf("id = %v, want the prior key", id)
	}
}

func TestTenantSuffixUnchangedSuffixKeepsTheID(t *testing.T) {
	id := planRenameID(t,
		tenantSuffixValue("aek7723.a.example.com", "a.example.com"),
		tenantSuffixValue("aek7723.a.example.com", "a.example.com"))
	if id.ValueString() != "aek7723.a.example.com" {
		t.Fatalf("id = %v, want the prior key", id)
	}
}

func TestTenantSuffixCreateAndDestroyAreLeftAlone(t *testing.T) {
	null := tftypes.NewValue(tenantSuffixObjType, nil)
	if id := planRenameID(t, null, tenantSuffixValue(tftypes.UnknownValue, "a.example.com")); !id.IsUnknown() {
		t.Fatalf("create: id = %v, want it left unknown", id)
	}
	if id := planRenameID(t, tenantSuffixValue("aek7723.a.example.com", "a.example.com"), null); !id.IsNull() {
		t.Fatalf("destroy: plan = %v, want it left null", id)
	}
}

func TestTenantSuffixUnknownPlannedSuffixIsLeftAlone(t *testing.T) {
	id := planRenameID(t,
		tenantSuffixValue("aek7723.a.example.com", "a.example.com"),
		tenantSuffixValue("aek7723.a.example.com", tftypes.UnknownValue))
	if id.ValueString() != "aek7723.a.example.com" {
		t.Fatalf("id = %v, want the prior key", id)
	}
}
