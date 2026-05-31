// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package schemas

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	policydbmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/policy/db/models"
)

func TestParseImportAttributePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		expectError bool
		expected    string
	}{
		{
			name:     "top_level_attribute",
			input:    "safe_id",
			expected: "safe_id",
		},
		{
			name:     "nested_attribute",
			input:    "metadata.policy_id",
			expected: "metadata.policy_id",
		},
		{
			name:        "empty_path",
			input:       "",
			expectError: true,
		},
		{
			name:        "empty_segment",
			input:       "metadata..policy_id",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseImportAttributePath(tt.input)
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error for input %q", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.String() != tt.expected {
				t.Fatalf("expected path %q, got %q", tt.expected, got.String())
			}
		})
	}
}

func TestSplitImportIDAttributes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		importID string
		expected []string
	}{
		{
			name:     "single_attribute",
			importID: "safe_id",
			expected: []string{"safe_id"},
		},
		{
			name:     "multiple_top_level",
			importID: "safe_id:member_name",
			expected: []string{"safe_id", "member_name"},
		},
		{
			name:     "nested_and_top_level",
			importID: "metadata.policy_id:other_id",
			expected: []string{"metadata.policy_id", "other_id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := SplitImportIDAttributes(tt.importID)
			if len(got) != len(tt.expected) {
				t.Fatalf("expected %d attributes, got %d: %v", len(tt.expected), len(got), got)
			}
			for i := range tt.expected {
				if got[i] != tt.expected[i] {
					t.Fatalf("attribute[%d]: expected %q, got %q", i, tt.expected[i], got[i])
				}
			}
		})
	}
}

func TestValidateStateSchemaImportAttribute_policy_metadata(t *testing.T) {
	t.Parallel()

	stateSchema := &policydbmodels.IdsecPolicyDBAccessPolicy{}
	if err := ValidateStateSchemaImportAttribute(stateSchema, "metadata.policy_id"); err != nil {
		t.Fatalf("expected metadata.policy_id to be valid: %v", err)
	}
}

func TestValidateStateSchemaImportAttribute_invalid_path(t *testing.T) {
	t.Parallel()

	stateSchema := &policydbmodels.IdsecPolicyDBAccessPolicy{}
	if err := ValidateStateSchemaImportAttribute(stateSchema, "metadata.missing_field"); err == nil {
		t.Fatal("expected error for missing nested field")
	}
}

func TestParseImportAttributePath_sets_nested_path(t *testing.T) {
	t.Parallel()

	attrPath, err := ParseImportAttributePath("metadata.policy_id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := path.Root("metadata").AtName("policy_id")
	if attrPath.String() != expected.String() {
		t.Fatalf("expected %q, got %q", expected.String(), attrPath.String())
	}
}
