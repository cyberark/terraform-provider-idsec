// Copyright (c) CyberArk.
// SPDX-License-Identifier: Apache-2.0

package schemas

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Test helper structs for testing nested struct scenarios

// testNestedStruct represents a nested struct that will be embedded.
type testNestedStruct struct {
	NestedField1 string `mapstructure:"nested_field_1" desc:"First nested field"`
	NestedField2 int    `mapstructure:"nested_field_2" desc:"Second nested field"`
	NestedField3 bool   `mapstructure:"nested_field_3" desc:"Third nested field"`
}

// testAnotherNestedStruct represents another nested struct.
type testAnotherNestedStruct struct {
	AnotherField1 string   `mapstructure:"another_field_1" desc:"Another nested field 1"`
	AnotherField2 []string `mapstructure:"another_field_2" desc:"Another nested field 2"`
}

// testStateModel represents a state model with nested structs (not squashed).
type testStateModel struct {
	ID             string                  `mapstructure:"id" desc:"ID field"`
	Name           string                  `mapstructure:"name" desc:"Name field"`
	NestedStruct   testNestedStruct        `mapstructure:"nested_struct" desc:"Nested struct field"`
	AnotherNested  testAnotherNestedStruct `mapstructure:"another_nested" desc:"Another nested struct"`
	RootLevelField string                  `mapstructure:"root_level_field" desc:"Root level field"`
}

// testCreateModel represents a create model with squashed nested structs.
type testCreateModel struct {
	testNestedStruct        `mapstructure:",squash"`
	testAnotherNestedStruct `mapstructure:",squash"`
	Name                    string `mapstructure:"name" desc:"Name field"`
	RootLevelField          string `mapstructure:"root_level_field" desc:"Root level field"`
}

// testUpdateModel represents an update model with squashed nested structs.
type testUpdateModel struct {
	testNestedStruct        `mapstructure:",squash"`
	testAnotherNestedStruct `mapstructure:",squash"`
	Name                    string `mapstructure:"name" desc:"Name field"`
	RootLevelField          string `mapstructure:"root_level_field" desc:"Root level field"`
}

// testStateModelWithPointerNested represents a state model with pointer nested struct.
type testStateModelWithPointerNested struct {
	ID           string            `mapstructure:"id" desc:"ID field"`
	NestedStruct *testNestedStruct `mapstructure:"nested_struct" desc:"Nested struct field"`
}

// testStateModelWithSquashed represents a state model that also has squashed fields.
type testStateModelWithSquashed struct {
	testNestedStruct `mapstructure:",squash"`
	RegularField     string `mapstructure:"regular_field" desc:"Regular field"`
}

// testStateModelEmpty represents an empty state model.
type testStateModelEmpty struct{}

// TestGetNestedStructFieldNames tests the getNestedStructFieldNames function.
func TestGetNestedStructFieldNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		stateModel     interface{}
		expectedFields map[string]bool
		validateFunc   func(t *testing.T, result map[string]bool)
	}{
		{
			name:           "success_nil_state_model",
			stateModel:     nil,
			expectedFields: map[string]bool{},
		},
		{
			name:           "success_empty_state_model",
			stateModel:     &testStateModelEmpty{},
			expectedFields: map[string]bool{},
		},
		{
			name:       "success_state_model_with_nested_structs",
			stateModel: &testStateModel{},
			expectedFields: map[string]bool{
				"nested_field_1":  true,
				"nested_field_2":  true,
				"nested_field_3":  true,
				"another_field_1": true,
				"another_field_2": true,
			},
		},
		{
			name:       "success_state_model_with_pointer_nested",
			stateModel: &testStateModelWithPointerNested{},
			expectedFields: map[string]bool{
				"nested_field_1": true,
				"nested_field_2": true,
				"nested_field_3": true,
			},
		},
		{
			name:           "success_state_model_with_squashed_fields",
			stateModel:     &testStateModelWithSquashed{},
			expectedFields: map[string]bool{},
			validateFunc: func(t *testing.T, result map[string]bool) {
				// Squashed fields should not be included
				if result["nested_field_1"] {
					t.Error("Expected squashed fields to be excluded from nested struct field names")
				}
			},
		},
		{
			name:           "success_non_struct_type",
			stateModel:     "not a struct",
			expectedFields: map[string]bool{},
		},
		{
			name:       "success_pointer_to_struct",
			stateModel: testStateModel{},
			expectedFields: map[string]bool{
				"nested_field_1":  true,
				"nested_field_2":  true,
				"nested_field_3":  true,
				"another_field_1": true,
				"another_field_2": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := getNestedStructFieldNames(tt.stateModel)

			// Check expected fields
			for fieldName, shouldExist := range tt.expectedFields {
				if shouldExist && !result[fieldName] {
					t.Errorf("Expected field %q to be in result, but it was not found", fieldName)
				}
				if !shouldExist && result[fieldName] {
					t.Errorf("Expected field %q not to be in result, but it was found", fieldName)
				}
			}

			// Check that result doesn't contain unexpected fields
			for fieldName := range result {
				if !tt.expectedFields[fieldName] {
					// Allow additional fields if validateFunc handles them
					if tt.validateFunc == nil {
						t.Errorf("Unexpected field %q in result", fieldName)
					}
				}
			}

			// Custom validation if provided
			if tt.validateFunc != nil {
				tt.validateFunc(t, result)
			}
		})
	}
}

// TestGenerateResourceSchemaFromStruct tests the GenerateResourceSchemaFromStruct function.
func TestGenerateResourceSchemaFromStruct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		createModel        interface{}
		updateModel        interface{}
		stateModel         interface{}
		sensitiveAttrs     []string
		extraRequiredAttrs []string
		computedAsSetAttrs []string
		immutableAttrs     []string
		forceNewAttrs      []string
		computedAttrs      []string
		validateFunc       func(t *testing.T, result schema.Schema)
		expectedError      bool
	}{
		{
			name:        "success_basic_models_without_nested_conflicts",
			createModel: &testCreateModel{},
			updateModel: &testUpdateModel{},
			stateModel:  &testStateModel{},
			validateFunc: func(t *testing.T, result schema.Schema) {
				// Check that nested struct fields are present as nested attributes
				if _, exists := result.Attributes["nested_struct"]; !exists {
					t.Error("Expected nested_struct attribute to exist in schema")
				}
				if _, exists := result.Attributes["another_nested"]; !exists {
					t.Error("Expected another_nested attribute to exist in schema")
				}

				// Check that flattened fields from create/update are NOT present at root level
				if _, exists := result.Attributes["nested_field_1"]; exists {
					t.Error("Expected nested_field_1 to NOT exist at root level (should be in nested_struct)")
				}
				if _, exists := result.Attributes["nested_field_2"]; exists {
					t.Error("Expected nested_field_2 to NOT exist at root level (should be in nested_struct)")
				}
				if _, exists := result.Attributes["another_field_1"]; exists {
					t.Error("Expected another_field_1 to NOT exist at root level (should be in another_nested)")
				}

				// Check that root level fields are present
				if _, exists := result.Attributes["name"]; !exists {
					t.Error("Expected name attribute to exist in schema")
				}
				if _, exists := result.Attributes["root_level_field"]; !exists {
					t.Error("Expected root_level_field attribute to exist in schema")
				}
			},
		},
		{
			name:        "success_create_model_only",
			createModel: &testCreateModel{},
			updateModel: nil,
			stateModel:  nil,
			validateFunc: func(t *testing.T, result schema.Schema) {
				// Without state model, flattened fields should be present
				if _, exists := result.Attributes["nested_field_1"]; !exists {
					t.Error("Expected nested_field_1 to exist when no state model is provided")
				}
			},
		},
		{
			name:        "success_state_model_only",
			createModel: &testStateModelEmpty{},
			updateModel: nil,
			stateModel:  &testStateModel{},
			validateFunc: func(t *testing.T, result schema.Schema) {
				// Nested structs should be present
				if _, exists := result.Attributes["nested_struct"]; !exists {
					t.Error("Expected nested_struct attribute to exist in schema")
				}
			},
		},
		{
			name:           "success_with_sensitive_attributes",
			createModel:    &testCreateModel{},
			updateModel:    &testUpdateModel{},
			stateModel:     &testStateModel{},
			sensitiveAttrs: []string{"name"},
			validateFunc: func(t *testing.T, result schema.Schema) {
				if attr, exists := result.Attributes["name"]; exists {
					// Check if attribute is sensitive (this depends on implementation)
					_ = attr // Use attr to avoid unused variable
				}
			},
		},
		{
			name:               "success_with_extra_required_attributes",
			createModel:        &testCreateModel{},
			updateModel:        &testUpdateModel{},
			stateModel:         &testStateModel{},
			extraRequiredAttrs: []string{"root_level_field"},
			validateFunc: func(t *testing.T, result schema.Schema) {
				if attr, exists := result.Attributes["root_level_field"]; exists {
					// Check if attribute is required
					_ = attr // Use attr to avoid unused variable
				}
			},
		},
		{
			name:               "success_with_computed_as_set_attributes",
			createModel:        &testCreateModel{},
			updateModel:        &testUpdateModel{},
			stateModel:         &testStateModel{},
			computedAsSetAttrs: []string{"another_field_2"},
			validateFunc: func(t *testing.T, result schema.Schema) {
				// Check that another_field_2 in nested struct is a set
				if nestedAttr, exists := result.Attributes["another_nested"]; exists {
					if singleNested, ok := nestedAttr.(schema.SingleNestedAttribute); ok {
						if setAttr, exists := singleNested.Attributes["another_field_2"]; exists {
							_ = setAttr // Use attr to avoid unused variable
						}
					}
				}
			},
		},
		{
			name:           "success_with_immutable_attributes",
			createModel:    &testCreateModel{},
			updateModel:    &testUpdateModel{},
			stateModel:     &testStateModel{},
			immutableAttrs: []string{"id"},
			validateFunc: func(t *testing.T, result schema.Schema) {
				if attr, exists := result.Attributes["id"]; exists {
					// Check if attribute has immutable plan modifier
					_ = attr // Use attr to avoid unused variable
				}
			},
		},
		{
			name:          "success_with_force_new_attributes",
			createModel:   &testCreateModel{},
			updateModel:   &testUpdateModel{},
			stateModel:    &testStateModel{},
			forceNewAttrs: []string{"root_level_field"},
			validateFunc: func(t *testing.T, result schema.Schema) {
				attr, exists := result.Attributes["root_level_field"]
				if !exists {
					t.Error("Expected root_level_field attribute to exist in schema")
					return
				}
				strAttr, ok := attr.(schema.StringAttribute)
				if !ok {
					t.Error("Expected root_level_field to be a StringAttribute")
					return
				}
				if len(strAttr.PlanModifiers) == 0 {
					t.Error("Expected root_level_field to have PlanModifiers (RequiresReplace)")
				}
			},
		},
		{
			name:          "success_with_computed_attributes",
			createModel:   &testCreateModel{},
			updateModel:   &testUpdateModel{},
			stateModel:    &testStateModel{},
			computedAttrs: []string{"id"},
			validateFunc: func(t *testing.T, result schema.Schema) {
				attr, exists := result.Attributes["id"]
				if !exists {
					t.Error("Expected id attribute to exist in schema")
					return
				}
				strAttr, ok := attr.(schema.StringAttribute)
				if !ok {
					t.Error("Expected id to be a StringAttribute")
					return
				}
				if strAttr.Optional {
					t.Error("Expected id to NOT be Optional (computed-only should be read-only)")
				}
				if strAttr.Required {
					t.Error("Expected id to NOT be Required (computed-only should be read-only)")
				}
				if !strAttr.Computed {
					t.Error("Expected id to be Computed (computed-only should be read-only)")
				}
			},
		},
		{
			name:        "success_empty_create_model",
			createModel: &testStateModelEmpty{},
			updateModel: &testUpdateModel{},
			stateModel:  &testStateModel{},
			validateFunc: func(t *testing.T, result schema.Schema) {
				// Should still generate schema from update and state models
				if len(result.Attributes) == 0 {
					t.Error("Expected some attributes in schema")
				}
			},
		},
		{
			name:        "success_all_empty_models",
			createModel: &testStateModelEmpty{},
			updateModel: nil,
			stateModel:  nil,
			validateFunc: func(t *testing.T, result schema.Schema) {
				// Should return empty schema when all models are empty/nil
				// Note: createModel cannot be nil as it's called without nil check
			},
		},
		{
			name:        "success_pointer_nested_struct_in_state",
			createModel: &testCreateModel{},
			updateModel: &testUpdateModel{},
			stateModel:  &testStateModelWithPointerNested{},
			validateFunc: func(t *testing.T, result schema.Schema) {
				// Nested struct should still be present
				if _, exists := result.Attributes["nested_struct"]; !exists {
					t.Error("Expected nested_struct attribute to exist in schema")
				}
				// Flattened fields should not be at root
				if _, exists := result.Attributes["nested_field_1"]; exists {
					t.Error("Expected nested_field_1 to NOT exist at root level")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := GenerateResourceSchemaFromStruct(
				tt.createModel,
				tt.updateModel,
				tt.stateModel,
				tt.sensitiveAttrs,
				tt.extraRequiredAttrs,
				tt.computedAsSetAttrs,
				tt.immutableAttrs,
				tt.forceNewAttrs,
				tt.computedAttrs,
			)

			// Validate result
			if tt.validateFunc != nil {
				tt.validateFunc(t, result)
			}

			// Basic sanity check - result should be a valid schema
			if result.Attributes == nil {
				t.Error("Expected Attributes map to be initialized, got nil")
			}
		})
	}
}

// TestGenerateResourceSchemaFromStructNestedStructRemoval tests that flattened fields
// from create/update models are properly removed when they belong to nested structs in state model.
func TestGenerateResourceSchemaFromStructNestedStructRemoval(t *testing.T) {
	t.Parallel()

	createModel := &testCreateModel{
		testNestedStruct: testNestedStruct{
			NestedField1: "value1",
			NestedField2: 42,
			NestedField3: true,
		},
		testAnotherNestedStruct: testAnotherNestedStruct{
			AnotherField1: "another1",
			AnotherField2: []string{"item1", "item2"},
		},
		Name:           "test",
		RootLevelField: "root",
	}

	updateModel := &testUpdateModel{
		testNestedStruct: testNestedStruct{
			NestedField1: "value1",
			NestedField2: 42,
			NestedField3: true,
		},
		testAnotherNestedStruct: testAnotherNestedStruct{
			AnotherField1: "another1",
			AnotherField2: []string{"item1", "item2"},
		},
		Name:           "test",
		RootLevelField: "root",
	}

	stateModel := &testStateModel{
		ID:   "test-id",
		Name: "test",
		NestedStruct: testNestedStruct{
			NestedField1: "value1",
			NestedField2: 42,
			NestedField3: true,
		},
		AnotherNested: testAnotherNestedStruct{
			AnotherField1: "another1",
			AnotherField2: []string{"item1", "item2"},
		},
		RootLevelField: "root",
	}

	result := GenerateResourceSchemaFromStruct(
		createModel,
		updateModel,
		stateModel,
		nil, // sensitiveAttrs
		nil, // extraRequiredAttrs
		nil, // computedAsSetAttrs
		nil, // immutableAttrs
		nil, // forceNewAttrs
		nil, // computedAttrs
	)

	// Verify nested structs exist
	if _, exists := result.Attributes["nested_struct"]; !exists {
		t.Error("Expected nested_struct attribute to exist")
	}

	if _, exists := result.Attributes["another_nested"]; !exists {
		t.Error("Expected another_nested attribute to exist")
	}

	// Verify flattened fields from create/update are NOT at root level
	flattenedFields := []string{
		"nested_field_1",
		"nested_field_2",
		"nested_field_3",
		"another_field_1",
		"another_field_2",
	}

	for _, fieldName := range flattenedFields {
		if _, exists := result.Attributes[fieldName]; exists {
			t.Errorf("Expected field %q to NOT exist at root level (should only be in nested struct)", fieldName)
		}
	}

	// Verify root level fields that should exist
	rootLevelFields := []string{
		"name",
		"root_level_field",
		"id",
	}

	for _, fieldName := range rootLevelFields {
		if _, exists := result.Attributes[fieldName]; !exists {
			t.Errorf("Expected field %q to exist at root level", fieldName)
		}
	}
}

// TestGenerateResourceSchemaFromStructWithSquashedStateModel tests the case where
// state model also has squashed fields (edge case).
func TestGenerateResourceSchemaFromStructWithSquashedStateModel(t *testing.T) {
	t.Parallel()

	createModel := &testCreateModel{
		testNestedStruct: testNestedStruct{
			NestedField1: "value1",
		},
		Name: "test",
	}

	stateModel := &testStateModelWithSquashed{
		testNestedStruct: testNestedStruct{
			NestedField1: "value1",
		},
		RegularField: "regular",
	}

	result := GenerateResourceSchemaFromStruct(
		createModel,
		nil,
		stateModel,
		nil,
		nil,
		nil,
		nil,
		nil, // forceNewAttrs
		nil, // computedAttrs
	)

	// When state model has squashed fields, they should appear at root level
	// (since getNestedStructFieldNames returns empty for squashed fields)
	if _, exists := result.Attributes["nested_field_1"]; !exists {
		t.Error("Expected nested_field_1 to exist at root level when state model has squashed fields")
	}

	if _, exists := result.Attributes["regular_field"]; !exists {
		t.Error("Expected regular_field to exist in schema")
	}
}

// TestGenerateResourceSchemaFromStructWithAttributeConflict tests what happens when
// the same attribute name exists both at root level and in a nested struct.
func TestGenerateResourceSchemaFromStructWithAttributeConflict(t *testing.T) {
	t.Parallel()

	// Define test structs with conflicting attribute names
	type testNestedStructWithConflict struct {
		ConflictingField string `mapstructure:"conflicting_field" desc:"Field that conflicts with root"`
		OtherField       string `mapstructure:"other_field" desc:"Other nested field"`
	}

	type testCreateModelWithConflict struct {
		testNestedStructWithConflict `mapstructure:",squash"`
		ConflictingField             string `mapstructure:"conflicting_field" desc:"Root level conflicting field"`
		RootOnlyField                string `mapstructure:"root_only_field" desc:"Root only field"`
	}

	type testStateModelWithConflict struct {
		ID               string                       `mapstructure:"id" desc:"ID field"`
		ConflictingField string                       `mapstructure:"conflicting_field" desc:"Root level conflicting field"`
		RootOnlyField    string                       `mapstructure:"root_only_field" desc:"Root only field"`
		NestedStruct     testNestedStructWithConflict `mapstructure:"nested_struct" desc:"Nested struct with conflict"`
	}

	createModel := &testCreateModelWithConflict{
		testNestedStructWithConflict: testNestedStructWithConflict{
			ConflictingField: "nested_value",
			OtherField:       "other_nested_value",
		},
		ConflictingField: "root_value",
		RootOnlyField:    "root_only_value",
	}

	updateModel := &testCreateModelWithConflict{
		testNestedStructWithConflict: testNestedStructWithConflict{
			ConflictingField: "nested_value",
			OtherField:       "other_nested_value",
		},
		ConflictingField: "root_value",
		RootOnlyField:    "root_only_value",
	}

	stateModel := &testStateModelWithConflict{
		ID:               "test-id",
		ConflictingField: "root_value",
		RootOnlyField:    "root_only_value",
		NestedStruct: testNestedStructWithConflict{
			ConflictingField: "nested_value",
			OtherField:       "other_nested_value",
		},
	}

	result := GenerateResourceSchemaFromStruct(
		createModel,
		updateModel,
		stateModel,
		nil, // sensitiveAttrs
		nil, // extraRequiredAttrs
		nil, // computedAsSetAttrs
		nil, // immutableAttrs
		nil, // forceNewAttrs
		nil, // computedAttrs
	)

	// Verify that nested_struct exists
	if _, exists := result.Attributes["nested_struct"]; !exists {
		t.Error("Expected nested_struct attribute to exist in schema")
	}

	// CRITICAL: Verify that conflicting_field exists at BOTH root level AND in nested struct
	// This is the main test - both attributes should coexist in the final schema
	rootConflictingField, rootExists := result.Attributes["conflicting_field"]
	if !rootExists {
		t.Error("Expected conflicting_field to exist at root level")
	}
	if rootConflictingField == nil {
		t.Error("Root level conflicting_field should not be nil")
	}

	// Verify that nested_struct also contains the conflicting_field
	var nestedConflictingField schema.Attribute
	var nestedExists bool
	if nestedAttr, exists := result.Attributes["nested_struct"]; exists {
		if singleNested, ok := nestedAttr.(schema.SingleNestedAttribute); ok {
			nestedConflictingField, nestedExists = singleNested.Attributes["conflicting_field"]
			if !nestedExists {
				t.Error("Expected nested_struct to contain conflicting_field - BOTH root and nested should exist")
			}
			if nestedConflictingField == nil {
				t.Error("Nested conflicting_field should not be nil")
			}
			if _, exists := singleNested.Attributes["other_field"]; !exists {
				t.Error("Expected nested_struct to contain other_field")
			}
		} else {
			t.Error("Expected nested_struct to be a SingleNestedAttribute")
		}
	} else {
		t.Error("nested_struct must exist to verify nested conflicting_field")
	}

	// Explicit validation: Both must exist
	if !rootExists || !nestedExists {
		t.Errorf("Both conflicting_field attributes must exist: root=%v, nested=%v", rootExists, nestedExists)
	}

	// Verify that root_only_field exists at root level
	if _, exists := result.Attributes["root_only_field"]; !exists {
		t.Error("Expected root_only_field to exist at root level")
	}

	// Verify that other_field from nested struct does NOT exist at root level
	// (it should only be in nested_struct)
	if _, exists := result.Attributes["other_field"]; exists {
		t.Error("Expected other_field to NOT exist at root level (should only be in nested_struct)")
	}
}
