// Copyright CyberArk. 2026
// SPDX-License-Identifier: Apache-2.0

package schemas

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestReadPendingHashAdoption(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		reader   privateStateReader
		expected map[string]bool
	}{
		{
			name:     "success_reads_marker",
			reader:   fakePrivateReader{data: map[string][]byte{WriteOnlyHashAdoptPrivateKey: []byte(`["api_key","password"]`)}},
			expected: map[string]bool{"api_key": true, "password": true},
		},
		{
			name:     "edge_case_nil_reader",
			reader:   nil,
			expected: nil,
		},
		{
			name:     "edge_case_missing_key",
			reader:   fakePrivateReader{data: map[string][]byte{}},
			expected: nil,
		},
		{
			name:     "error_undecodable_value_treated_as_empty",
			reader:   fakePrivateReader{data: map[string][]byte{WriteOnlyHashAdoptPrivateKey: []byte(`{"not":"a list"}`)}},
			expected: nil,
		},
		{
			name:     "error_reader_diagnostics_treated_as_empty",
			reader:   fakePrivateReader{diags: diag.Diagnostics{diag.NewErrorDiagnostic("boom", "boom")}},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := ReadPendingHashAdoption(context.Background(), tt.reader)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

// adoptedSibling reads the "password_write_only_hash" attribute out of an object value returned by
// AdoptWriteOnlyHashTagsFromPriorState, failing the test if it is absent or not a string.
func adoptedSibling(t *testing.T, obj tftypes.Value) tftypes.Value {
	t.Helper()

	attrs := map[string]tftypes.Value{}
	if err := obj.As(&attrs); err != nil {
		t.Fatalf("reading the returned state as an object: %v", err)
	}
	sibling, ok := attrs["password_write_only_hash"]
	if !ok {
		t.Fatal(`returned state has no "password_write_only_hash" attribute`)
	}
	return sibling
}

// TestAdoptWriteOnlyHashTagsFromPriorState_DerivesTagFromStoredSecret asserts the adopted tag is
// really keyed on the credential the prior state held: its embedded salt, re-derived against that
// same secret, reproduces the tag byte for byte, and against any other secret does not.
func TestAdoptWriteOnlyHashTagsFromPriorState_DerivesTagFromStoredSecret(t *testing.T) {
	t.Parallel()

	const secret = "stored-plain-text-credential"
	prior := hashModifierObjectRaw(hashKnownString(secret), hashNullString())

	got, diags := AdoptWriteOnlyHashTagsFromPriorState(context.Background(), []string{"password"}, prior)
	if diags.HasError() {
		t.Fatalf("expected no error diagnostics, got %v", diags.Errors())
	}
	if len(diags) != 0 {
		t.Errorf("expected zero diagnostics, got %v", diags)
	}

	sibling := adoptedSibling(t, got)
	if sibling.IsNull() {
		t.Fatal("the sibling is still null; the stored credential would be re-sent on the next apply")
	}
	var tag string
	if err := sibling.As(&tag); err != nil {
		t.Fatalf("reading the adopted tag: %v", err)
	}
	if strings.Contains(tag, secret) {
		t.Error("the adopted tag contains the plain text credential")
	}

	_, salt, _, ok := parseTag(tag)
	if !ok {
		t.Fatalf("the adopted tag is not parseable by this version: %q", tag)
	}
	if want := deriveTag(secret, salt); tag != want {
		t.Error("the adopted tag is not the derivation of the stored secret under its own salt")
	}
	if other := deriveTag(secret+"-rotated", salt); tag == other {
		t.Error("the adopted tag is identical for a different secret; it cannot detect rotation")
	}

	// The prior value must not be mutated in place: the caller hands the framework whichever value
	// it gets back, and a hidden mutation of the input would make that choice meaningless.
	if !adoptedSibling(t, prior).IsNull() {
		t.Error("the input state value was mutated in place")
	}
}

// TestAdoptWriteOnlyHashTagsFromPriorState_UnchangedSecretPlansToSameTag is the regression test for
// the bug this adoption exists to fix: after a provider upgrade that made "password" write-only,
// the first apply re-sent the credential (a null sibling plans unknown, which triggerFired reads as
// changed), and a vault can reject that re-send outright against its password policy. With the tag
// adopted, an unchanged secret plans to a byte-identical tag -- no diff, no send -- while an edited
// secret still plans to a different one and is sent.
func TestAdoptWriteOnlyHashTagsFromPriorState_UnchangedSecretPlansToSameTag(t *testing.T) {
	t.Parallel()

	const secret = "stored-plain-text-credential"
	ctx := context.Background()

	upgraded, diags := AdoptWriteOnlyHashTagsFromPriorState(
		ctx, []string{"password"}, hashModifierObjectRaw(hashKnownString(secret), hashNullString()))
	if diags.HasError() {
		t.Fatalf("expected no error diagnostics, got %v", diags.Errors())
	}
	var adoptedTag string
	if err := adoptedSibling(t, upgraded).As(&adoptedTag); err != nil {
		t.Fatalf("reading the adopted tag: %v", err)
	}

	// The framework nulls the write-only source in the state it hands back, so the plan modifier
	// sees exactly this: the adopted tag in state, and the secret only in configuration.
	planFor := func(configSecret string) types.String {
		req := planmodifier.StringRequest{
			Path:       path.Root("password_write_only_hash"),
			StateValue: types.StringValue(adoptedTag),
			State:      tfsdk.State{Raw: hashModifierObjectRaw(hashNullString(), hashKnownString(adoptedTag))},
			Config:     tfsdk.Config{Raw: hashModifierConfigRaw(hashKnownString(configSecret))},
		}
		resp := &planmodifier.StringResponse{PlanValue: types.StringUnknown()}
		newHashPlanModifier("password").PlanModifyString(ctx, req, resp)
		if resp.Diagnostics.HasError() {
			t.Fatalf("plan modifier diagnostics: %v", resp.Diagnostics.Errors())
		}
		return resp.PlanValue
	}

	if planned := planFor(secret); !planned.Equal(types.StringValue(adoptedTag)) {
		t.Errorf("an unchanged secret planned %v, want the adopted tag; the credential would be re-sent", planned)
	}
	if planned := planFor(secret + "-rotated"); planned.IsNull() || planned.IsUnknown() || planned.ValueString() == adoptedTag {
		t.Errorf("an edited secret planned %v, want a different concrete tag; the rotation would not be sent", planned)
	}
}

// TestAdoptWriteOnlyHashTagsFromPriorState_LeavesSiblingNull covers every input that must not
// produce a tag. Each one degrades to the pre-existing behaviour -- the sibling stays null and the
// configured value is sent once -- rather than recording a hash for a value the backend may never
// have received.
func TestAdoptWriteOnlyHashTagsFromPriorState_LeavesSiblingNull(t *testing.T) {
	t.Parallel()

	const existingTag = "v1:_dyUMLoP9jXFKuEQCnp0rA:Zm9vYmFyYmF6cXV1eGNvcmdlZ3JhdWx0Z2FybXBseA"

	tests := []struct {
		name        string
		hashedAttrs []string
		prior       tftypes.Value
		// wantSibling is the exact value the sibling must still hold afterwards.
		wantSibling tftypes.Value
	}{
		{
			name:        "edge_case_no_stored_secret_cannot_adopt",
			hashedAttrs: []string{"password"},
			prior:       hashModifierObjectRaw(hashNullString(), hashNullString()),
			wantSibling: hashNullString(),
		},
		{
			name:        "edge_case_existing_tag_is_the_authoritative_baseline",
			hashedAttrs: []string{"password"},
			prior:       hashModifierObjectRaw(hashKnownString("stored"), hashKnownString(existingTag)),
			wantSibling: hashKnownString(existingTag),
		},
		{
			name:        "edge_case_nested_path_has_no_top_level_sibling",
			hashedAttrs: []string{"nested.password"},
			prior:       hashModifierObjectRaw(hashKnownString("stored"), hashNullString()),
			wantSibling: hashNullString(),
		},
		{
			name:        "edge_case_no_hashed_attributes_declared",
			hashedAttrs: nil,
			prior:       hashModifierObjectRaw(hashKnownString("stored"), hashNullString()),
			wantSibling: hashNullString(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, diags := AdoptWriteOnlyHashTagsFromPriorState(context.Background(), tt.hashedAttrs, tt.prior)
			if diags.HasError() {
				t.Fatalf("expected no error diagnostics, got %v", diags.Errors())
			}
			if sibling := adoptedSibling(t, got); !sibling.Equal(tt.wantSibling) {
				t.Errorf("sibling = %s, want %s", sibling.String(), tt.wantSibling.String())
			}
		})
	}
}

// TestAdoptWriteOnlyHashTagsFromPriorState_UnusableStateReturnedUnchanged asserts a state value
// this function cannot reason about is handed back exactly as received. The upgrade path must not
// fail, and must not lose data, over a hash it merely could not adopt.
func TestAdoptWriteOnlyHashTagsFromPriorState_UnusableStateReturnedUnchanged(t *testing.T) {
	t.Parallel()

	// A source attribute whose sibling the current schema does not declare, a null state, and a
	// non-object state: none can carry an adopted tag.
	sourceOnly := tftypes.NewValue(
		tftypes.Object{AttributeTypes: map[string]tftypes.Type{"password": tftypes.String}},
		map[string]tftypes.Value{"password": hashKnownString("stored")},
	)
	tests := map[string]tftypes.Value{
		"edge_case_sibling_not_in_schema": sourceOnly,
		"edge_case_null_state": tftypes.NewValue(
			tftypes.Object{AttributeTypes: map[string]tftypes.Type{"password": tftypes.String}}, nil),
		"edge_case_non_object_state": hashKnownString("not-an-object"),
	}

	for name, prior := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := AdoptWriteOnlyHashTagsFromPriorState(context.Background(), []string{"password"}, prior)
			if diags.HasError() {
				t.Fatalf("expected no error diagnostics, got %v", diags.Errors())
			}
			if !got.Equal(prior) {
				t.Errorf("state value changed: got %s, want %s", got.String(), prior.String())
			}
		})
	}
}
