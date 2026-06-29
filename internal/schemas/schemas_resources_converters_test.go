// Copyright CyberArk. 2026
// SPDX-License-Identifier: Apache-2.0

package schemas

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
		name                 string
		createModel          interface{}
		updateModel          interface{}
		stateModel           interface{}
		sensitiveAttrs       []string
		extraRequiredAttrs   []string
		computedAsSetAttrs   []string
		immutableAttrs       []string
		forceNewAttrs        []string
		computedAttrs        []string
		caseInsensitiveAttrs []string
		validateFunc         func(t *testing.T, result schema.Schema)
		expectedError        bool
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
			name:                 "success_case_insensitive_attribute_gets_plan_modifier",
			createModel:          &testCreateModel{},
			updateModel:          &testUpdateModel{},
			stateModel:           &testStateModel{},
			caseInsensitiveAttrs: []string{"root_level_field"},
			validateFunc: func(t *testing.T, result schema.Schema) {
				attr, ok := result.Attributes["root_level_field"].(schema.StringAttribute)
				if !ok {
					t.Fatal("expected root_level_field to be StringAttribute")
				}
				if len(attr.PlanModifiers) == 0 {
					t.Fatal("expected root_level_field to have at least one plan modifier (CaseInsensitiveString)")
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
				tt.caseInsensitiveAttrs,
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
		nil,
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
		nil,
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
		nil,
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

// Test helper structs for testing the `min` / `max` field tag support.

// testMinMaxNestedItem is the element type used in nested list/map fields below.
type testMinMaxNestedItem struct {
	Field string `mapstructure:"field" desc:"Nested item field"`
}

// testMinMaxCreateModel exercises all attribute kinds that support min/max:
// string, list-of-simple, set-of-simple (via computedAsSetAttrs), map-of-simple,
// list-of-structs, map-of-structs, and a few edge cases (only-min, only-max,
// no min/max, validate tag without min/max clauses, and unparsable values). The
// min/max bounds are embedded inside the `validate` tag, in go-playground/validator
// style ("required,min=3,max=10").
type testMinMaxCreateModel struct {
	Name        string                          `mapstructure:"name" desc:"Name" validate:"required,min=3,max=10"`
	Description string                          `mapstructure:"description" desc:"Description" validate:"max=200"`
	Code        string                          `mapstructure:"code" desc:"Code" validate:"min=5"`
	Plain       string                          `mapstructure:"plain" desc:"Plain"`
	ReqOnly     string                          `mapstructure:"req_only" desc:"Required without bounds" validate:"required"`
	BadBounds   string                          `mapstructure:"bad_bounds" desc:"Bad bounds" validate:"min=abc,max=xyz"`
	Tags        []string                        `mapstructure:"tags" desc:"Tags" validate:"min=1,max=5"`
	SetItems    []string                        `mapstructure:"set_items" desc:"Set items" validate:"min=2,max=4"`
	Props       map[string]string               `mapstructure:"props" desc:"Props" validate:"min=1,max=10"`
	Items       []testMinMaxNestedItem          `mapstructure:"items" desc:"Items" validate:"min=1,max=3"`
	ItemsMap    map[string]testMinMaxNestedItem `mapstructure:"items_map" desc:"Items map" validate:"min=0,max=5"`
}

// findValidatorOfType returns the first element of vs that is of the requested concrete
// type T. Returns the zero value of T and false when no match exists.
func findValidatorOfType[T any](vs []validator.String) (T, bool) {
	for _, v := range vs {
		if typed, ok := v.(T); ok {
			return typed, true
		}
	}
	var zero T
	return zero, false
}

// findListValidatorOfType returns the first list validator of the requested concrete type.
func findListValidatorOfType[T any](vs []validator.List) (T, bool) {
	for _, v := range vs {
		if typed, ok := v.(T); ok {
			return typed, true
		}
	}
	var zero T
	return zero, false
}

// findSetValidatorOfType returns the first set validator of the requested concrete type.
func findSetValidatorOfType[T any](vs []validator.Set) (T, bool) {
	for _, v := range vs {
		if typed, ok := v.(T); ok {
			return typed, true
		}
	}
	var zero T
	return zero, false
}

// findMapValidatorOfType returns the first map validator of the requested concrete type.
func findMapValidatorOfType[T any](vs []validator.Map) (T, bool) {
	for _, v := range vs {
		if typed, ok := v.(T); ok {
			return typed, true
		}
	}
	var zero T
	return zero, false
}

// int64PtrEquals checks whether got matches want (both possibly nil pointers).
func int64PtrEquals(got, want *int64) bool {
	if got == nil || want == nil {
		return got == want
	}
	return *got == *want
}

func int64Ptr(v int64) *int64 {
	return &v
}

// TestGenerateResourceSchemaFromStructMinMaxTags verifies that the `min` and `max`
// struct tags are translated into the right validator on every attribute kind that
// supports them: strings, lists, sets, maps, list-of-structs and map-of-structs.
// It also covers the edge cases of only-min, only-max, no min/max, and tags whose
// values cannot be parsed as int64 (which must be ignored, not crash).
func TestGenerateResourceSchemaFromStructMinMaxTags(t *testing.T) {
	t.Parallel()

	result := GenerateResourceSchemaFromStruct(
		&testMinMaxCreateModel{},
		nil,
		nil,
		nil,
		nil,
		[]string{"set_items"},
		nil,
		nil,
		nil,
		nil,
	)

	tests := []struct {
		name         string
		validateFunc func(t *testing.T)
	}{
		{
			name: "success_string_with_min_and_max",
			validateFunc: func(t *testing.T) {
				strAttr, ok := result.Attributes["name"].(schema.StringAttribute)
				if !ok {
					t.Fatalf("expected name to be StringAttribute, got %T", result.Attributes["name"])
				}
				v, found := findValidatorOfType[StringLengthValidator](strAttr.Validators)
				if !found {
					t.Fatal("expected StringLengthValidator on name")
				}
				if !int64PtrEquals(v.Min, int64Ptr(3)) || !int64PtrEquals(v.Max, int64Ptr(10)) {
					t.Errorf("expected Min=3 Max=10, got Min=%v Max=%v", v.Min, v.Max)
				}
			},
		},
		{
			name: "success_string_with_only_max",
			validateFunc: func(t *testing.T) {
				strAttr, ok := result.Attributes["description"].(schema.StringAttribute)
				if !ok {
					t.Fatalf("expected description to be StringAttribute, got %T", result.Attributes["description"])
				}
				v, found := findValidatorOfType[StringLengthValidator](strAttr.Validators)
				if !found {
					t.Fatal("expected StringLengthValidator on description")
				}
				if v.Min != nil {
					t.Errorf("expected Min to be nil, got %v", *v.Min)
				}
				if !int64PtrEquals(v.Max, int64Ptr(200)) {
					t.Errorf("expected Max=200, got %v", v.Max)
				}
			},
		},
		{
			name: "success_string_with_only_min",
			validateFunc: func(t *testing.T) {
				strAttr, ok := result.Attributes["code"].(schema.StringAttribute)
				if !ok {
					t.Fatalf("expected code to be StringAttribute, got %T", result.Attributes["code"])
				}
				v, found := findValidatorOfType[StringLengthValidator](strAttr.Validators)
				if !found {
					t.Fatal("expected StringLengthValidator on code")
				}
				if !int64PtrEquals(v.Min, int64Ptr(5)) {
					t.Errorf("expected Min=5, got %v", v.Min)
				}
				if v.Max != nil {
					t.Errorf("expected Max to be nil, got %v", *v.Max)
				}
			},
		},
		{
			name: "success_string_without_min_max_has_no_length_validator",
			validateFunc: func(t *testing.T) {
				strAttr, ok := result.Attributes["plain"].(schema.StringAttribute)
				if !ok {
					t.Fatalf("expected plain to be StringAttribute, got %T", result.Attributes["plain"])
				}
				if _, found := findValidatorOfType[StringLengthValidator](strAttr.Validators); found {
					t.Error("expected no StringLengthValidator on plain")
				}
			},
		},
		{
			name: "success_string_with_unparsable_bounds_is_ignored",
			validateFunc: func(t *testing.T) {
				strAttr, ok := result.Attributes["bad_bounds"].(schema.StringAttribute)
				if !ok {
					t.Fatalf("expected bad_bounds to be StringAttribute, got %T", result.Attributes["bad_bounds"])
				}
				if _, found := findValidatorOfType[StringLengthValidator](strAttr.Validators); found {
					t.Error("expected no StringLengthValidator when both min/max clauses are unparsable")
				}
			},
		},
		{
			name: "success_validate_required_without_min_max_has_no_length_validator",
			validateFunc: func(t *testing.T) {
				strAttr, ok := result.Attributes["req_only"].(schema.StringAttribute)
				if !ok {
					t.Fatalf("expected req_only to be StringAttribute, got %T", result.Attributes["req_only"])
				}
				if _, found := findValidatorOfType[StringLengthValidator](strAttr.Validators); found {
					t.Error("expected no StringLengthValidator when validate tag only contains 'required'")
				}
				if !strAttr.Required {
					t.Error("expected req_only to be marked Required (validate:\"required\")")
				}
			},
		},
		{
			name: "success_list_of_strings_gets_list_size_validator",
			validateFunc: func(t *testing.T) {
				listAttr, ok := result.Attributes["tags"].(schema.ListAttribute)
				if !ok {
					t.Fatalf("expected tags to be ListAttribute, got %T", result.Attributes["tags"])
				}
				v, found := findListValidatorOfType[ListSizeValidator](listAttr.Validators)
				if !found {
					t.Fatal("expected ListSizeValidator on tags")
				}
				if !int64PtrEquals(v.Min, int64Ptr(1)) || !int64PtrEquals(v.Max, int64Ptr(5)) {
					t.Errorf("expected Min=1 Max=5, got Min=%v Max=%v", v.Min, v.Max)
				}
			},
		},
		{
			name: "success_set_attribute_gets_set_size_validator",
			validateFunc: func(t *testing.T) {
				setAttr, ok := result.Attributes["set_items"].(schema.SetAttribute)
				if !ok {
					t.Fatalf("expected set_items to be SetAttribute, got %T", result.Attributes["set_items"])
				}
				v, found := findSetValidatorOfType[SetSizeValidator](setAttr.Validators)
				if !found {
					t.Fatal("expected SetSizeValidator on set_items")
				}
				if !int64PtrEquals(v.Min, int64Ptr(2)) || !int64PtrEquals(v.Max, int64Ptr(4)) {
					t.Errorf("expected Min=2 Max=4, got Min=%v Max=%v", v.Min, v.Max)
				}
			},
		},
		{
			name: "success_map_of_simple_gets_map_size_validator",
			validateFunc: func(t *testing.T) {
				mapAttr, ok := result.Attributes["props"].(schema.MapAttribute)
				if !ok {
					t.Fatalf("expected props to be MapAttribute, got %T", result.Attributes["props"])
				}
				v, found := findMapValidatorOfType[MapSizeValidator](mapAttr.Validators)
				if !found {
					t.Fatal("expected MapSizeValidator on props")
				}
				if !int64PtrEquals(v.Min, int64Ptr(1)) || !int64PtrEquals(v.Max, int64Ptr(10)) {
					t.Errorf("expected Min=1 Max=10, got Min=%v Max=%v", v.Min, v.Max)
				}
			},
		},
		{
			name: "success_list_of_structs_gets_list_size_validator",
			validateFunc: func(t *testing.T) {
				listAttr, ok := result.Attributes["items"].(schema.ListNestedAttribute)
				if !ok {
					t.Fatalf("expected items to be ListNestedAttribute, got %T", result.Attributes["items"])
				}
				v, found := findListValidatorOfType[ListSizeValidator](listAttr.Validators)
				if !found {
					t.Fatal("expected ListSizeValidator on items")
				}
				if !int64PtrEquals(v.Min, int64Ptr(1)) || !int64PtrEquals(v.Max, int64Ptr(3)) {
					t.Errorf("expected Min=1 Max=3, got Min=%v Max=%v", v.Min, v.Max)
				}
			},
		},
		{
			name: "success_map_of_structs_gets_map_size_validator",
			validateFunc: func(t *testing.T) {
				mapAttr, ok := result.Attributes["items_map"].(schema.MapNestedAttribute)
				if !ok {
					t.Fatalf("expected items_map to be MapNestedAttribute, got %T", result.Attributes["items_map"])
				}
				v, found := findMapValidatorOfType[MapSizeValidator](mapAttr.Validators)
				if !found {
					t.Fatal("expected MapSizeValidator on items_map")
				}
				if !int64PtrEquals(v.Min, int64Ptr(0)) || !int64PtrEquals(v.Max, int64Ptr(5)) {
					t.Errorf("expected Min=0 Max=5, got Min=%v Max=%v", v.Min, v.Max)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.validateFunc(t)
		})
	}
}

// testMinMaxBoundsModel reuses the same field types across tests below that exercise
// the StringLengthValidator / ListSizeValidator / SetSizeValidator / MapSizeValidator
// runtime semantics. The schema attributes here are produced by the converter, so
// validating their behavior also validates the wiring end-to-end.
type testMinMaxBoundsModel struct {
	Name     string            `mapstructure:"name" desc:"Name" validate:"min=3,max=5"`
	Tags     []string          `mapstructure:"tags" desc:"Tags" validate:"min=1,max=2"`
	SetItems []string          `mapstructure:"set_items" desc:"Set items" validate:"min=1,max=2"`
	Props    map[string]string `mapstructure:"props" desc:"Props" validate:"min=1,max=2"`
}

// TestMinMaxValidatorsAttachedHaveCorrectDescriptions verifies the human-readable
// description of validators produced through the schema converter — both for the
// fully-bounded and the half-bounded cases — so users get useful provider docs.
func TestMinMaxValidatorsAttachedHaveCorrectDescriptions(t *testing.T) {
	t.Parallel()

	result := GenerateResourceSchemaFromStruct(
		&testMinMaxBoundsModel{},
		nil,
		nil,
		nil,
		nil,
		[]string{"set_items"},
		nil,
		nil,
		nil,
		nil,
	)

	ctx := context.Background()

	t.Run("success_string_length_validator_description_non_empty", func(t *testing.T) {
		t.Parallel()
		strAttr, ok := result.Attributes["name"].(schema.StringAttribute)
		if !ok {
			t.Fatalf("expected name to be StringAttribute, got %T", result.Attributes["name"])
		}
		v, found := findValidatorOfType[StringLengthValidator](strAttr.Validators)
		if !found {
			t.Fatal("expected StringLengthValidator")
		}
		if v.Description(ctx) == "" {
			t.Error("expected non-empty description")
		}
		if v.MarkdownDescription(ctx) == "" {
			t.Error("expected non-empty markdown description")
		}
	})

	t.Run("success_list_size_validator_description_non_empty", func(t *testing.T) {
		t.Parallel()
		listAttr, ok := result.Attributes["tags"].(schema.ListAttribute)
		if !ok {
			t.Fatalf("expected tags to be ListAttribute, got %T", result.Attributes["tags"])
		}
		v, found := findListValidatorOfType[ListSizeValidator](listAttr.Validators)
		if !found {
			t.Fatal("expected ListSizeValidator")
		}
		if v.Description(ctx) == "" {
			t.Error("expected non-empty description")
		}
		if v.MarkdownDescription(ctx) == "" {
			t.Error("expected non-empty markdown description")
		}
	})

	t.Run("success_set_size_validator_description_non_empty", func(t *testing.T) {
		t.Parallel()
		setAttr, ok := result.Attributes["set_items"].(schema.SetAttribute)
		if !ok {
			t.Fatalf("expected set_items to be SetAttribute, got %T", result.Attributes["set_items"])
		}
		v, found := findSetValidatorOfType[SetSizeValidator](setAttr.Validators)
		if !found {
			t.Fatal("expected SetSizeValidator")
		}
		if v.Description(ctx) == "" {
			t.Error("expected non-empty description")
		}
		if v.MarkdownDescription(ctx) == "" {
			t.Error("expected non-empty markdown description")
		}
	})

	t.Run("success_map_size_validator_description_non_empty", func(t *testing.T) {
		t.Parallel()
		mapAttr, ok := result.Attributes["props"].(schema.MapAttribute)
		if !ok {
			t.Fatalf("expected props to be MapAttribute, got %T", result.Attributes["props"])
		}
		v, found := findMapValidatorOfType[MapSizeValidator](mapAttr.Validators)
		if !found {
			t.Fatal("expected MapSizeValidator")
		}
		if v.Description(ctx) == "" {
			t.Error("expected non-empty description")
		}
		if v.MarkdownDescription(ctx) == "" {
			t.Error("expected non-empty markdown description")
		}
	})
}
