// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package pcloud

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ── test helpers ─────────────────────────────────────────────────────────────

// safeMemberObjType is the minimal tftypes object type used to construct test
// plans and states. It mirrors the two attributes the validator reads.
var safeMemberObjType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"safe_id":   tftypes.String,
		"safe_name": tftypes.String,
	},
}

// safeMemberTestSchema satisfies tfsdk.Plan.GetAttribute for the safe_id and
// safe_name attributes. It is intentionally minimal — only what the validator
// actually accesses needs to be present.
var safeMemberTestSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"safe_id":   schema.StringAttribute{Optional: true, Computed: true},
		"safe_name": schema.StringAttribute{Optional: true, Computed: true},
	},
}

// makePlan builds a tfsdk.Plan with the supplied safe_id and safe_name tftypes
// values. Use tftypes.NewValue(tftypes.String, nil) for null,
// tftypes.NewValue(tftypes.String, tftypes.UnknownValue) for unknown, and
// tftypes.NewValue(tftypes.String, "value") for a known string.
func makePlan(safeIDVal, safeNameVal tftypes.Value) tfsdk.Plan {
	return tfsdk.Plan{
		Raw: tftypes.NewValue(safeMemberObjType, map[string]tftypes.Value{
			"safe_id":   safeIDVal,
			"safe_name": safeNameVal,
		}),
		Schema: safeMemberTestSchema,
	}
}

// nullState returns a tfsdk.State with a null Raw value, representing a brand
// new resource that has not been created yet.
func nullState() tfsdk.State {
	return tfsdk.State{Raw: tftypes.NewValue(safeMemberObjType, nil)}
}

// nonNullState returns a tfsdk.State with a non-null Raw value, representing a
// resource that already exists (update / no-op plan). The validator skips any
// resource whose state is not null.
func nonNullState() tfsdk.State {
	return tfsdk.State{Raw: tftypes.NewValue(safeMemberObjType, map[string]tftypes.Value{
		"safe_id":   tftypes.NewValue(tftypes.String, "existing-safe"),
		"safe_name": tftypes.NewValue(tftypes.String, "existing-safe"),
	})}
}

// callValidatePlan invokes ValidatePlan with a nil IdsecAPI. Passing nil is
// safe for any test case in which fetchRemote is never called (i.e. safe_id is
// unknown, null, empty, or the counter's remote cache is already populated for
// the given safe_id).
func callValidatePlan(v safeMemberLimitValidator, plan tfsdk.Plan, state tfsdk.State) resource.ModifyPlanResponse {
	req := resource.ModifyPlanRequest{Plan: plan, State: state}
	resp := resource.ModifyPlanResponse{Plan: plan}
	v.ValidatePlan(context.Background(), req, &resp, nil)
	return resp
}

// ── safeMemberPlanCounter.checkAndReserve ────────────────────────────────────

func TestCheckAndReserve(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		fetchRemote  func() (int, error)
		additions    int
		wantLastOK   bool
		wantLastProj int
		wantErr      bool
	}{
		{
			name:         "first_addition_to_empty_safe",
			fetchRemote:  func() (int, error) { return 0, nil },
			additions:    1,
			wantLastOK:   true,
			wantLastProj: 1,
		},
		{
			name:         "exactly_at_limit",
			fetchRemote:  func() (int, error) { return 63, nil },
			additions:    1,
			wantLastOK:   true,
			wantLastProj: maxSafeMembersLimit,
		},
		{
			name:         "one_over_remote_limit",
			fetchRemote:  func() (int, error) { return 64, nil },
			additions:    1,
			wantLastOK:   false,
			wantLastProj: maxSafeMembersLimit + 1,
		},
		{
			name:         "delta_fills_remaining_slots",
			fetchRemote:  func() (int, error) { return 60, nil },
			additions:    4, // remote=60, delta fills to 64
			wantLastOK:   true,
			wantLastProj: maxSafeMembersLimit,
		},
		{
			name:         "delta_exceeds_limit",
			fetchRemote:  func() (int, error) { return 60, nil },
			additions:    5, // 5th: projected = 60+4+1 = 65 → fail
			wantLastOK:   false,
			wantLastProj: maxSafeMembersLimit + 1,
		},
		{
			name:        "fetch_error_propagates",
			fetchRemote: func() (int, error) { return 0, fmt.Errorf("api error") },
			additions:   1,
			wantErr:     true,
		},
		{
			name:         "nil_fetch_remote_assumes_zero_remote",
			fetchRemote:  nil,
			additions:    1,
			wantLastOK:   true,
			wantLastProj: 1,
		},
		{
			name:         "nil_fetch_remote_exactly_at_limit",
			fetchRemote:  nil,
			additions:    maxSafeMembersLimit,
			wantLastOK:   true,
			wantLastProj: maxSafeMembersLimit,
		},
		{
			name:         "nil_fetch_remote_exceeds_limit",
			fetchRemote:  nil,
			additions:    maxSafeMembersLimit + 1,
			wantLastOK:   false,
			wantLastProj: maxSafeMembersLimit + 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := newSafeMemberPlanCounter()
			var projected int
			var ok bool
			var err error

			for range tt.additions {
				projected, ok, err = c.checkAndReserve("safe-x", tt.fetchRemote)
			}

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ok != tt.wantLastOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantLastOK)
			}
			if projected != tt.wantLastProj {
				t.Errorf("projected = %d, want %d", projected, tt.wantLastProj)
			}
		})
	}
}

func TestCheckAndReserve_RemoteCountCachedAfterFirstCall(t *testing.T) {
	t.Parallel()

	fetchCalls := 0
	fetch := func() (int, error) {
		fetchCalls++
		return 5, nil
	}

	c := newSafeMemberPlanCounter()
	for range 3 {
		_, _, _ = c.checkAndReserve("safe-x", fetch)
	}

	if fetchCalls != 1 {
		t.Errorf("fetchRemote called %d time(s), want exactly 1", fetchCalls)
	}
}

func TestCheckAndReserve_DifferentSafeIDsAreIndependent(t *testing.T) {
	t.Parallel()

	zero := func() (int, error) { return 0, nil }
	c := newSafeMemberPlanCounter()

	// Fill safe-1 to the limit.
	for i := range maxSafeMembersLimit {
		_, ok, err := c.checkAndReserve("safe-1", zero)
		if err != nil || !ok {
			t.Fatalf("safe-1 addition %d: ok=%v err=%v", i+1, ok, err)
		}
	}

	// safe-2 should still accept its first addition regardless of safe-1.
	_, ok, err := c.checkAndReserve("safe-2", zero)
	if err != nil || !ok {
		t.Errorf("safe-2 first addition should succeed; ok=%v err=%v", ok, err)
	}

	// safe-1 is now at the limit; the next addition must be rejected.
	_, ok, err = c.checkAndReserve("safe-1", zero)
	if err != nil {
		t.Fatalf("safe-1 overflow check: unexpected error: %v", err)
	}
	if ok {
		t.Error("safe-1: addition beyond limit should have been rejected")
	}
}

// ── safeMemberLimitValidator.ValidatePlan ────────────────────────────────────

func TestValidatePlan_SkipsExistingResource(t *testing.T) {
	t.Parallel()

	v := safeMemberLimitValidator{counter: newSafeMemberPlanCounter()}
	plan := makePlan(
		tftypes.NewValue(tftypes.String, "safe-123"),
		tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
	)
	resp := callValidatePlan(v, plan, nonNullState())

	if resp.Diagnostics.HasError() {
		t.Errorf("expected no diagnostics for existing resource, got: %v", resp.Diagnostics.Errors())
	}
}

func TestValidatePlan_SkipsNullSafeID(t *testing.T) {
	t.Parallel()

	v := safeMemberLimitValidator{counter: newSafeMemberPlanCounter()}
	plan := makePlan(
		tftypes.NewValue(tftypes.String, nil),
		tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
	)
	resp := callValidatePlan(v, plan, nullState())

	if resp.Diagnostics.HasError() {
		t.Errorf("expected no diagnostics for null safe_id, got: %v", resp.Diagnostics.Errors())
	}
}

func TestValidatePlan_SkipsEmptySafeID(t *testing.T) {
	t.Parallel()

	v := safeMemberLimitValidator{counter: newSafeMemberPlanCounter()}
	plan := makePlan(
		tftypes.NewValue(tftypes.String, ""),
		tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
	)
	resp := callValidatePlan(v, plan, nullState())

	if resp.Diagnostics.HasError() {
		t.Errorf("expected no diagnostics for empty safe_id, got: %v", resp.Diagnostics.Errors())
	}
}

func TestValidatePlan_UnknownSafeID_SingleAdditionAllowed(t *testing.T) {
	t.Parallel()

	v := safeMemberLimitValidator{counter: newSafeMemberPlanCounter()}
	plan := makePlan(
		tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
	)
	resp := callValidatePlan(v, plan, nullState())

	if resp.Diagnostics.HasError() {
		t.Errorf("expected no diagnostics for first addition to new safe, got: %v", resp.Diagnostics.Errors())
	}
}

func TestValidatePlan_UnknownSafeID_AtLimit_NoError(t *testing.T) {
	t.Parallel()

	v := safeMemberLimitValidator{counter: newSafeMemberPlanCounter()}
	plan := makePlan(
		tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
	)

	var resp resource.ModifyPlanResponse
	for range maxSafeMembersLimit {
		resp = callValidatePlan(v, plan, nullState())
	}

	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error at exactly limit=%d, got: %v", maxSafeMembersLimit, resp.Diagnostics.Errors())
	}
}

func TestValidatePlan_UnknownSafeID_ExceedsLimit_ReturnsError(t *testing.T) {
	t.Parallel()

	v := safeMemberLimitValidator{counter: newSafeMemberPlanCounter()}
	plan := makePlan(
		tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
	)

	var resp resource.ModifyPlanResponse
	for range maxSafeMembersLimit + 1 {
		resp = callValidatePlan(v, plan, nullState())
	}

	if !resp.Diagnostics.HasError() {
		t.Errorf("expected error when exceeding limit (%d), got none", maxSafeMembersLimit)
	}
}

func TestValidatePlan_ErrorMessage_UsesNewSafeLabel_WhenSafeIDUnknown(t *testing.T) {
	t.Parallel()

	v := safeMemberLimitValidator{counter: newSafeMemberPlanCounter()}
	plan := makePlan(
		tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
	)

	var resp resource.ModifyPlanResponse
	for range maxSafeMembersLimit + 1 {
		resp = callValidatePlan(v, plan, nullState())
	}

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error, got none")
	}
	detail := resp.Diagnostics.Errors()[0].Detail()
	if !strings.Contains(detail, "(new safe)") {
		t.Errorf("expected error detail to reference \"(new safe)\", got: %s", detail)
	}
}

// TestValidatePlan_ErrorMessage_UsesSafeName verifies that when safe_id is a
// known string and safe_name is populated in the plan, the error message
// reports the name rather than the ID.
//
// The counter is pre-seeded so that checkAndReserve uses its cached remote
// count and never invokes fetchRemote — making it safe to pass a nil IdsecAPI.
func TestValidatePlan_ErrorMessage_UsesSafeName(t *testing.T) {
	t.Parallel()

	const safeID = "known-safe-id"
	const safeName = "my-safe-name"

	c := newSafeMemberPlanCounter()
	for range maxSafeMembersLimit {
		_, _, _ = c.checkAndReserve(safeID, func() (int, error) { return 0, nil })
	}

	v := safeMemberLimitValidator{counter: c}
	plan := makePlan(
		tftypes.NewValue(tftypes.String, safeID),
		tftypes.NewValue(tftypes.String, safeName),
	)
	resp := callValidatePlan(v, plan, nullState())

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error, got none")
	}
	detail := resp.Diagnostics.Errors()[0].Detail()
	if !strings.Contains(detail, safeName) {
		t.Errorf("expected error to reference safe name %q, got: %s", safeName, detail)
	}
}

// TestValidatePlan_ErrorMessage_FallsBackToSafeID verifies that when safe_id is
// known but safe_name is unknown in the plan, the error falls back to the ID.
func TestValidatePlan_ErrorMessage_FallsBackToSafeID(t *testing.T) {
	t.Parallel()

	const safeID = "fallback-safe-id"

	c := newSafeMemberPlanCounter()
	for range maxSafeMembersLimit {
		_, _, _ = c.checkAndReserve(safeID, func() (int, error) { return 0, nil })
	}

	v := safeMemberLimitValidator{counter: c}
	plan := makePlan(
		tftypes.NewValue(tftypes.String, safeID),
		tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
	)
	resp := callValidatePlan(v, plan, nullState())

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error, got none")
	}
	detail := resp.Diagnostics.Errors()[0].Detail()
	if !strings.Contains(detail, safeID) {
		t.Errorf("expected error to reference safe ID %q, got: %s", safeID, detail)
	}
}
