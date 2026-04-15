// Copyright (c) CyberArk.
// SPDX-License-Identifier: Apache-2.0

package schemas

import (
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Test helper structs for data source testing

// testDataSourceInputModel represents a simple input model with only required fields.
type testDataSourceInputModel struct {
	ID string `mapstructure:"id" desc:"ID field" validate:"required"`
}

// testDataSourceStateModel represents a state model with nested attributes.
type testDataSourceStateModel struct {
	ID               string                         `mapstructure:"id" desc:"ID field"`
	Name             string                         `mapstructure:"name" desc:"Name field"`
	Status           string                         `mapstructure:"status" desc:"Status field"`
	SecretManagement testDataSourceSecretManagement `mapstructure:"secret_management" desc:"Secret management"`
	RemoteMachines   testDataSourceRemoteMachines   `mapstructure:"remote_machines_access" desc:"Remote machines access"`
}

// testDataSourceSecretManagement represents nested secret management attributes.
type testDataSourceSecretManagement struct {
	AutomaticManagementEnabled bool   `mapstructure:"automatic_management_enabled" desc:"Whether automatic management is enabled"`
	ManualManagementReason     string `mapstructure:"manual_management_reason" desc:"Reason for disabling automatic management"`
	LastModifiedTime           int    `mapstructure:"last_modified_time" desc:"Last modified time"`
}

// testDataSourceRemoteMachines represents nested remote machines attributes.
type testDataSourceRemoteMachines struct {
	AccessRestrictedToRemoteMachines bool     `mapstructure:"access_restricted_to_remote_machines" desc:"Whether access is restricted"`
	RemoteMachines                   []string `mapstructure:"remote_machines" desc:"List of remote machines"`
}

// testDataSourceStateModelWithNestedOnly represents a state model where nested attributes exist only in state.
type testDataSourceStateModelWithNestedOnly struct {
	ID               string                         `mapstructure:"id" desc:"ID field"`
	SecretManagement testDataSourceSecretManagement `mapstructure:"secret_management" desc:"Secret management"`
}

// TestCollectAllNestedAttributePaths tests the collectAllNestedAttributePaths function.
func TestCollectAllNestedAttributePaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		attrs         map[string]schema.Attribute
		prefix        string
		expectedPaths []string
		validateFunc  func(t *testing.T, result []string)
	}{
		{
			name:          "success_empty_attributes",
			attrs:         map[string]schema.Attribute{},
			prefix:        "",
			expectedPaths: []string{},
		},
		{
			name: "success_simple_attributes_no_prefix",
			attrs: map[string]schema.Attribute{
				"field1": schema.StringAttribute{Description: "Field 1"},
				"field2": schema.BoolAttribute{Description: "Field 2"},
			},
			prefix:        "",
			expectedPaths: []string{"field1", "field2"},
		},
		{
			name: "success_simple_attributes_with_prefix",
			attrs: map[string]schema.Attribute{
				"field1": schema.StringAttribute{Description: "Field 1"},
				"field2": schema.BoolAttribute{Description: "Field 2"},
			},
			prefix:        "parent",
			expectedPaths: []string{"parent.field1", "parent.field2"},
		},
		{
			name: "success_nested_single_nested_attribute",
			attrs: map[string]schema.Attribute{
				"nested": schema.SingleNestedAttribute{
					Attributes: map[string]schema.Attribute{
						"nested_field1": schema.StringAttribute{Description: "Nested field 1"},
						"nested_field2": schema.Int64Attribute{Description: "Nested field 2"},
					},
				},
			},
			prefix:        "",
			expectedPaths: []string{"nested", "nested.nested_field1", "nested.nested_field2"},
		},
		{
			name: "success_nested_list_nested_attribute",
			attrs: map[string]schema.Attribute{
				"list_nested": schema.ListNestedAttribute{
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"item_field1": schema.StringAttribute{Description: "Item field 1"},
							"item_field2": schema.BoolAttribute{Description: "Item field 2"},
						},
					},
				},
			},
			prefix:        "",
			expectedPaths: []string{"list_nested", "list_nested.item_field1", "list_nested.item_field2"},
		},
		{
			name: "success_deeply_nested_attributes",
			attrs: map[string]schema.Attribute{
				"level1": schema.SingleNestedAttribute{
					Attributes: map[string]schema.Attribute{
						"level2": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"level3_field": schema.StringAttribute{Description: "Level 3 field"},
							},
						},
					},
				},
			},
			prefix:        "",
			expectedPaths: []string{"level1", "level1.level2", "level1.level2.level3_field"},
		},
		{
			name: "success_mixed_nested_types",
			attrs: map[string]schema.Attribute{
				"single": schema.SingleNestedAttribute{
					Attributes: map[string]schema.Attribute{
						"field1": schema.StringAttribute{Description: "Field 1"},
					},
				},
				"list": schema.ListNestedAttribute{
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"field2": schema.BoolAttribute{Description: "Field 2"},
						},
					},
				},
				"map": schema.MapNestedAttribute{
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"field3": schema.Int64Attribute{Description: "Field 3"},
						},
					},
				},
			},
			prefix: "parent",
			expectedPaths: []string{
				"parent.single",
				"parent.single.field1",
				"parent.list",
				"parent.list.field2",
				"parent.map",
				"parent.map.field3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := collectAllNestedAttributePaths(tt.attrs, tt.prefix)

			// Validate expected paths
			if len(tt.expectedPaths) > 0 {
				for _, expectedPath := range tt.expectedPaths {
					if !slices.Contains(result, expectedPath) {
						t.Errorf("Expected path %q not found in result: %v", expectedPath, result)
					}
				}
			}

			// Validate result length matches expected
			if len(result) != len(tt.expectedPaths) {
				t.Errorf("Expected %d paths, got %d. Expected: %v, Got: %v", len(tt.expectedPaths), len(result), tt.expectedPaths, result)
			}

			// Custom validation if provided
			if tt.validateFunc != nil {
				tt.validateFunc(t, result)
			}
		})
	}
}

// TestMergeNestedAttributesAndFindReadOnly tests the mergeNestedAttributesAndFindReadOnly function.
func TestMergeNestedAttributesAndFindReadOnly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		inputAttrs       map[string]schema.Attribute
		stateAttrs       map[string]schema.Attribute
		prefix           string
		expectedReadOnly []string
		validateFunc     func(t *testing.T, result []string, mergedAttrs map[string]schema.Attribute)
	}{
		{
			name:       "success_no_overlap",
			inputAttrs: map[string]schema.Attribute{},
			stateAttrs: map[string]schema.Attribute{
				"state_only_field": schema.StringAttribute{Description: "State only"},
			},
			prefix:           "",
			expectedReadOnly: []string{"state_only_field"},
		},
		{
			name:       "success_nested_attribute_only_in_state",
			inputAttrs: map[string]schema.Attribute{},
			stateAttrs: map[string]schema.Attribute{
				"secret_management": schema.SingleNestedAttribute{
					Attributes: map[string]schema.Attribute{
						"last_modified_time":           schema.Int64Attribute{Description: "Last modified time"},
						"automatic_management_enabled": schema.BoolAttribute{Description: "Automatic management enabled"},
					},
				},
			},
			prefix: "",
			expectedReadOnly: []string{
				"secret_management",
				"secret_management.last_modified_time",
				"secret_management.automatic_management_enabled",
			},
			validateFunc: func(t *testing.T, result []string, mergedAttrs map[string]schema.Attribute) {
				// Verify the nested attribute was added
				if _, exists := mergedAttrs["secret_management"]; !exists {
					t.Error("Expected secret_management to be added to inputAttrs")
				}
			},
		},
		{
			name: "success_nested_attribute_in_both_merge_nested",
			inputAttrs: map[string]schema.Attribute{
				"secret_management": schema.SingleNestedAttribute{
					Attributes: map[string]schema.Attribute{
						"automatic_management_enabled": schema.BoolAttribute{Description: "Automatic management enabled"},
					},
				},
			},
			stateAttrs: map[string]schema.Attribute{
				"secret_management": schema.SingleNestedAttribute{
					Attributes: map[string]schema.Attribute{
						"automatic_management_enabled": schema.BoolAttribute{Description: "Automatic management enabled"},
						"last_modified_time":           schema.Int64Attribute{Description: "Last modified time"},
					},
				},
			},
			prefix:           "",
			expectedReadOnly: []string{"secret_management.last_modified_time"},
			validateFunc: func(t *testing.T, result []string, mergedAttrs map[string]schema.Attribute) {
				// Verify the nested attribute exists
				attr, exists := mergedAttrs["secret_management"]
				if !exists {
					t.Error("Expected secret_management to exist in mergedAttrs")
					return
				}
				nestedAttr, ok := attr.(schema.SingleNestedAttribute)
				if !ok {
					t.Error("Expected secret_management to be SingleNestedAttribute")
					return
				}
				// Verify both nested fields exist
				if _, exists := nestedAttr.Attributes["automatic_management_enabled"]; !exists {
					t.Error("Expected automatic_management_enabled to exist in nested attributes")
				}
				if _, exists := nestedAttr.Attributes["last_modified_time"]; !exists {
					t.Error("Expected last_modified_time to exist in nested attributes")
				}
			},
		},
		{
			name:       "success_list_nested_attribute_only_in_state",
			inputAttrs: map[string]schema.Attribute{},
			stateAttrs: map[string]schema.Attribute{
				"items": schema.ListNestedAttribute{
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"item_field": schema.StringAttribute{Description: "Item field"},
						},
					},
				},
			},
			prefix: "",
			expectedReadOnly: []string{
				"items",
				"items.item_field",
			},
		},
		{
			name:       "success_map_nested_attribute_only_in_state",
			inputAttrs: map[string]schema.Attribute{},
			stateAttrs: map[string]schema.Attribute{
				"mapping": schema.MapNestedAttribute{
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"map_field": schema.Int64Attribute{Description: "Map field"},
						},
					},
				},
			},
			prefix: "",
			expectedReadOnly: []string{
				"mapping",
				"mapping.map_field",
			},
		},
		{
			name:       "success_deeply_nested_attributes",
			inputAttrs: map[string]schema.Attribute{},
			stateAttrs: map[string]schema.Attribute{
				"level1": schema.SingleNestedAttribute{
					Attributes: map[string]schema.Attribute{
						"level2": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"level3_field": schema.StringAttribute{Description: "Level 3 field"},
							},
						},
					},
				},
			},
			prefix: "",
			expectedReadOnly: []string{
				"level1",
				"level1.level2",
				"level1.level2.level3_field",
			},
		},
		{
			name:       "success_with_prefix",
			inputAttrs: map[string]schema.Attribute{},
			stateAttrs: map[string]schema.Attribute{
				"nested": schema.SingleNestedAttribute{
					Attributes: map[string]schema.Attribute{
						"field": schema.StringAttribute{Description: "Field"},
					},
				},
			},
			prefix: "parent",
			expectedReadOnly: []string{
				"parent.nested",
				"parent.nested.field",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a copy of inputAttrs to avoid mutation across tests
			inputAttrsCopy := make(map[string]schema.Attribute)
			for k, v := range tt.inputAttrs {
				inputAttrsCopy[k] = v
			}

			result := mergeNestedAttributesAndFindReadOnly(inputAttrsCopy, tt.stateAttrs, tt.prefix)

			// Validate expected read-only paths
			for _, expectedPath := range tt.expectedReadOnly {
				if !slices.Contains(result, expectedPath) {
					t.Errorf("Expected read-only path %q not found in result: %v", expectedPath, result)
				}
			}

			// Validate result length matches expected (allowing for additional paths)
			if len(result) < len(tt.expectedReadOnly) {
				t.Errorf("Expected at least %d read-only paths, got %d. Expected: %v, Got: %v", len(tt.expectedReadOnly), len(result), tt.expectedReadOnly, result)
			}

			// Custom validation if provided
			if tt.validateFunc != nil {
				tt.validateFunc(t, result, inputAttrsCopy)
			}
		})
	}
}

// TestGenerateDataSourceSchemaFromStructNestedAttributes tests that nested attributes
// that exist only in state model are marked as read-only.
func TestGenerateDataSourceSchemaFromStructNestedAttributes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		inputModel         interface{}
		stateModel         interface{}
		sensitiveAttrs     []string
		extraRequiredAttrs []string
		computedAsSetAttrs []string
		validateFunc       func(t *testing.T, result schema.Schema)
	}{
		{
			name:       "success_nested_attribute_only_in_state",
			inputModel: &testDataSourceInputModel{},
			stateModel: &testDataSourceStateModelWithNestedOnly{},
			validateFunc: func(t *testing.T, result schema.Schema) {
				// Verify ID is required (from input model)
				idAttr, exists := result.Attributes["id"]
				if !exists {
					t.Error("Expected id attribute to exist")
					return
				}
				if strAttr, ok := idAttr.(schema.StringAttribute); ok {
					if !strAttr.Required {
						t.Error("Expected id to be Required")
					}
					if strAttr.Computed {
						t.Error("Expected id to NOT be Computed")
					}
				}

				// Verify secret_management exists and is read-only
				secretMgmtAttr, exists := result.Attributes["secret_management"]
				if !exists {
					t.Error("Expected secret_management attribute to exist")
					return
				}
				nestedAttr, ok := secretMgmtAttr.(schema.SingleNestedAttribute)
				if !ok {
					t.Error("Expected secret_management to be SingleNestedAttribute")
					return
				}
				if nestedAttr.Optional {
					t.Error("Expected secret_management to NOT be Optional (should be read-only)")
				}
				if nestedAttr.Required {
					t.Error("Expected secret_management to NOT be Required (should be read-only)")
				}
				if !nestedAttr.Computed {
					t.Error("Expected secret_management to be Computed (read-only)")
				}

				// Verify nested attributes within secret_management are read-only
				if nestedAttr.Attributes == nil {
					t.Error("Expected secret_management to have nested attributes")
					return
				}
				lastModifiedAttr, exists := nestedAttr.Attributes["last_modified_time"]
				if !exists {
					t.Error("Expected last_modified_time to exist in nested attributes")
					return
				}
				if intAttr, ok := lastModifiedAttr.(schema.Int64Attribute); ok {
					if intAttr.Optional {
						t.Error("Expected last_modified_time to NOT be Optional (should be read-only)")
					}
					if intAttr.Required {
						t.Error("Expected last_modified_time to NOT be Required (should be read-only)")
					}
					if !intAttr.Computed {
						t.Error("Expected last_modified_time to be Computed (read-only)")
					}
				}
			},
		},
		{
			name:       "success_multiple_nested_attributes",
			inputModel: &testDataSourceInputModel{},
			stateModel: &testDataSourceStateModel{},
			validateFunc: func(t *testing.T, result schema.Schema) {
				// Verify all state-only attributes are read-only
				readOnlyAttrs := []string{"name", "status", "secret_management", "remote_machines_access"}
				for _, attrName := range readOnlyAttrs {
					attr, exists := result.Attributes[attrName]
					if !exists {
						t.Errorf("Expected %s attribute to exist", attrName)
						continue
					}

					// Check if it's a nested attribute
					switch a := attr.(type) {
					case schema.StringAttribute:
						if a.Optional {
							t.Errorf("Expected %s to NOT be Optional (should be read-only)", attrName)
						}
						if a.Required {
							t.Errorf("Expected %s to NOT be Required (should be read-only)", attrName)
						}
						if !a.Computed {
							t.Errorf("Expected %s to be Computed (read-only)", attrName)
						}
					case schema.SingleNestedAttribute:
						if a.Optional {
							t.Errorf("Expected %s to NOT be Optional (should be read-only)", attrName)
						}
						if a.Required {
							t.Errorf("Expected %s to NOT be Required (should be read-only)", attrName)
						}
						if !a.Computed {
							t.Errorf("Expected %s to be Computed (read-only)", attrName)
						}
					}
				}

				// Verify nested attributes within secret_management are read-only
				secretMgmtAttr, exists := result.Attributes["secret_management"]
				if !exists {
					t.Error("Expected secret_management to exist")
					return
				}
				if nestedAttr, ok := secretMgmtAttr.(schema.SingleNestedAttribute); ok && nestedAttr.Attributes != nil {
					for nestedKey, nestedAttr := range nestedAttr.Attributes {
						switch a := nestedAttr.(type) {
						case schema.BoolAttribute:
							if a.Optional {
								t.Errorf("Expected nested attribute %s to NOT be Optional (should be read-only)", nestedKey)
							}
							if !a.Computed {
								t.Errorf("Expected nested attribute %s to be Computed (read-only)", nestedKey)
							}
						case schema.StringAttribute:
							if a.Optional {
								t.Errorf("Expected nested attribute %s to NOT be Optional (should be read-only)", nestedKey)
							}
							if !a.Computed {
								t.Errorf("Expected nested attribute %s to be Computed (read-only)", nestedKey)
							}
						case schema.Int64Attribute:
							if a.Optional {
								t.Errorf("Expected nested attribute %s to NOT be Optional (should be read-only)", nestedKey)
							}
							if !a.Computed {
								t.Errorf("Expected nested attribute %s to be Computed (read-only)", nestedKey)
							}
						}
					}
				}
			},
		},
		{
			name:       "success_nil_input_model",
			inputModel: nil,
			stateModel: &testDataSourceStateModel{},
			validateFunc: func(t *testing.T, result schema.Schema) {
				// All attributes from state model should be read-only
				if len(result.Attributes) == 0 {
					t.Error("Expected attributes from state model to be present")
				}
				for attrName, attr := range result.Attributes {
					switch a := attr.(type) {
					case schema.StringAttribute:
						if a.Optional {
							t.Errorf("Expected %s to NOT be Optional (should be read-only)", attrName)
						}
						if !a.Computed {
							t.Errorf("Expected %s to be Computed (read-only)", attrName)
						}
					case schema.SingleNestedAttribute:
						if a.Optional {
							t.Errorf("Expected %s to NOT be Optional (should be read-only)", attrName)
						}
						if !a.Computed {
							t.Errorf("Expected %s to be Computed (read-only)", attrName)
						}
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := GenerateDataSourceSchemaFromStruct(
				tt.inputModel,
				tt.stateModel,
				tt.sensitiveAttrs,
				tt.extraRequiredAttrs,
				tt.computedAsSetAttrs,
			)

			// Validate result
			if tt.validateFunc != nil {
				tt.validateFunc(t, result)
			}

			// Basic sanity check
			if result.Attributes == nil {
				t.Error("Expected Attributes map to be initialized, got nil")
			}
		})
	}
}
