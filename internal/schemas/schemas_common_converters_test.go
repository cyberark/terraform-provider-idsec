// Copyright (c) CyberArk
// SPDX-License-Identifier: Apache-2.0

package schemas

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDeepCopy(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		Name  string
		Value int
	}

	tests := []struct {
		name           string
		input          interface{}
		validateFunc   func(t *testing.T, original, copied interface{})
		shouldBeSame   bool
		expectedResult interface{}
	}{
		{
			name:           "success_nil_value",
			input:          nil,
			expectedResult: nil,
			shouldBeSame:   true,
		},
		{
			name:           "success_primitive_string",
			input:          "test string",
			expectedResult: "test string",
			shouldBeSame:   true,
		},
		{
			name:           "success_primitive_int",
			input:          42,
			expectedResult: 42,
			shouldBeSame:   true,
		},
		{
			name:           "success_primitive_bool",
			input:          true,
			expectedResult: true,
			shouldBeSame:   true,
		},
		{
			name:           "success_slice_of_ints",
			input:          []int{1, 2, 3, 4, 5},
			expectedResult: []int{1, 2, 3, 4, 5},
			validateFunc: func(t *testing.T, original, copied interface{}) {
				origSlice, ok := original.([]int)
				if !ok {
					t.Fatal("Failed to cast original to []int")
				}
				copiedSlice, ok := copied.([]int)
				if !ok {
					t.Fatal("Failed to cast copied to []int")
				}
				if &origSlice[0] == &copiedSlice[0] {
					t.Error("Expected different slice backing arrays, got same")
				}
				origSlice[0] = 999
				if copiedSlice[0] == 999 {
					t.Error("Modifying original affected copy")
				}
			},
		},
		{
			name:           "success_slice_of_strings",
			input:          []string{"a", "b", "c"},
			expectedResult: []string{"a", "b", "c"},
			validateFunc: func(t *testing.T, original, copied interface{}) {
				origSlice, ok := original.([]string)
				if !ok {
					t.Fatal("Failed to cast original to []string")
				}
				copiedSlice, ok := copied.([]string)
				if !ok {
					t.Fatal("Failed to cast copied to []string")
				}
				origSlice[0] = "modified"
				if copiedSlice[0] == "modified" {
					t.Error("Modifying original affected copy")
				}
			},
		},
		{
			name:           "success_nil_slice",
			input:          []int(nil),
			expectedResult: []int(nil),
			shouldBeSame:   true,
		},
		{
			name:           "success_empty_slice",
			input:          []int{},
			expectedResult: []int{},
		},
		{
			name:           "success_array_of_ints",
			input:          [3]int{1, 2, 3},
			expectedResult: [3]int{1, 2, 3},
		},
		{
			name:           "success_map_string_to_int",
			input:          map[string]int{"a": 1, "b": 2, "c": 3},
			expectedResult: map[string]int{"a": 1, "b": 2, "c": 3},
			validateFunc: func(t *testing.T, original, copied interface{}) {
				origMap, ok := original.(map[string]int)
				if !ok {
					t.Fatal("Failed to cast original to map[string]int")
				}
				copiedMap, ok := copied.(map[string]int)
				if !ok {
					t.Fatal("Failed to cast copied to map[string]int")
				}
				origMap["a"] = 999
				if copiedMap["a"] == 999 {
					t.Error("Modifying original map affected copy")
				}
			},
		},
		{
			name:           "success_nil_map",
			input:          map[string]int(nil),
			expectedResult: map[string]int(nil),
			shouldBeSame:   true,
		},
		{
			name:           "success_empty_map",
			input:          map[string]int{},
			expectedResult: map[string]int{},
		},
		{
			name:           "success_struct_simple",
			input:          testStruct{Name: "test", Value: 42},
			expectedResult: testStruct{Name: "test", Value: 42},
		},
		{
			name:           "success_pointer_to_int",
			input:          intPtr(42),
			expectedResult: intPtr(42),
			validateFunc: func(t *testing.T, original, copied interface{}) {
				origPtr, ok := original.(*int)
				if !ok {
					t.Fatal("Failed to cast original to *int")
				}
				copiedPtr, ok := copied.(*int)
				if !ok {
					t.Fatal("Failed to cast copied to *int")
				}
				if origPtr == copiedPtr {
					t.Error("Expected different pointers, got same")
				}
				*origPtr = 999
				if *copiedPtr == 999 {
					t.Error("Modifying original pointer affected copy")
				}
			},
		},
		{
			name:           "success_nil_pointer",
			input:          (*int)(nil),
			expectedResult: (*int)(nil),
			shouldBeSame:   true,
		},
		{
			name:           "success_pointer_to_struct",
			input:          &testStruct{Name: "test", Value: 42},
			expectedResult: &testStruct{Name: "test", Value: 42},
			validateFunc: func(t *testing.T, original, copied interface{}) {
				origPtr, ok := original.(*testStruct)
				if !ok {
					t.Fatal("Failed to cast original to *testStruct")
				}
				copiedPtr, ok := copied.(*testStruct)
				if !ok {
					t.Fatal("Failed to cast copied to *testStruct")
				}
				if origPtr == copiedPtr {
					t.Error("Expected different pointers, got same")
				}
				origPtr.Name = "modified"
				if copiedPtr.Name == "modified" {
					t.Error("Modifying original struct affected copy")
				}
			},
		},
		{
			name: "success_nested_slice_of_slices",
			input: [][]int{
				{1, 2, 3},
				{4, 5, 6},
			},
			expectedResult: [][]int{
				{1, 2, 3},
				{4, 5, 6},
			},
			validateFunc: func(t *testing.T, original, copied interface{}) {
				origSlice, ok := original.([][]int)
				if !ok {
					t.Fatal("Failed to cast original to [][]int")
				}
				copiedSlice, ok := copied.([][]int)
				if !ok {
					t.Fatal("Failed to cast copied to [][]int")
				}
				origSlice[0][0] = 999
				if copiedSlice[0][0] == 999 {
					t.Error("Modifying nested original affected copy")
				}
			},
		},
		{
			name: "success_map_with_slice_values",
			input: map[string][]int{
				"a": {1, 2, 3},
				"b": {4, 5, 6},
			},
			expectedResult: map[string][]int{
				"a": {1, 2, 3},
				"b": {4, 5, 6},
			},
			validateFunc: func(t *testing.T, original, copied interface{}) {
				origMap, ok := original.(map[string][]int)
				if !ok {
					t.Fatal("Failed to cast original to map[string][]int")
				}
				copiedMap, ok := copied.(map[string][]int)
				if !ok {
					t.Fatal("Failed to cast copied to map[string][]int")
				}
				origMap["a"][0] = 999
				if copiedMap["a"][0] == 999 {
					t.Error("Modifying nested original affected copy")
				}
			},
		},
		{
			name: "success_struct_with_nested_fields",
			input: struct {
				Inner *testStruct
				Slice []int
			}{
				Inner: &testStruct{Name: "inner", Value: 10},
				Slice: []int{1, 2, 3},
			},
			expectedResult: struct {
				Inner *testStruct
				Slice []int
			}{
				Inner: &testStruct{Name: "inner", Value: 10},
				Slice: []int{1, 2, 3},
			},
		},
		{
			name:           "success_interface_with_string",
			input:          interface{}("test"),
			expectedResult: interface{}("test"),
		},
		{
			name:           "success_interface_with_struct",
			input:          interface{}(testStruct{Name: "test", Value: 42}),
			expectedResult: interface{}(testStruct{Name: "test", Value: 42}),
		},
		{
			name:           "success_nil_interface",
			input:          (interface{})(nil),
			expectedResult: (interface{})(nil),
			shouldBeSame:   true,
		},
		{
			name: "success_complex_struct_with_nested_lists",
			input: struct {
				Name     string
				Tags     []string
				Metadata map[string][]int
				Config   *struct {
					Values []string
					Counts []int
				}
			}{
				Name: "complex",
				Tags: []string{"tag1", "tag2", "tag3"},
				Metadata: map[string][]int{
					"key1": {1, 2, 3},
					"key2": {4, 5, 6},
				},
				Config: &struct {
					Values []string
					Counts []int
				}{
					Values: []string{"a", "b", "c"},
					Counts: []int{10, 20, 30},
				},
			},
			expectedResult: struct {
				Name     string
				Tags     []string
				Metadata map[string][]int
				Config   *struct {
					Values []string
					Counts []int
				}
			}{
				Name: "complex",
				Tags: []string{"tag1", "tag2", "tag3"},
				Metadata: map[string][]int{
					"key1": {1, 2, 3},
					"key2": {4, 5, 6},
				},
				Config: &struct {
					Values []string
					Counts []int
				}{
					Values: []string{"a", "b", "c"},
					Counts: []int{10, 20, 30},
				},
			},
			validateFunc: func(t *testing.T, original, copied interface{}) {
				origVal := reflect.ValueOf(original)
				copiedVal := reflect.ValueOf(copied)

				origTags := origVal.FieldByName("Tags")
				copiedTags := copiedVal.FieldByName("Tags")
				if origTags.Len() > 0 && copiedTags.Len() > 0 {
					origSlice, ok := origTags.Interface().([]string)
					if !ok {
						t.Fatal("Failed to cast original Tags to []string")
					}
					copiedSlice, ok := copiedTags.Interface().([]string)
					if !ok {
						t.Fatal("Failed to cast copied Tags to []string")
					}
					origSlice[0] = "modified_tag"
					if copiedSlice[0] == "modified_tag" {
						t.Error("Modifying original Tags affected copy")
					}
				}

				origMetadata := origVal.FieldByName("Metadata")
				copiedMetadata := copiedVal.FieldByName("Metadata")
				if origMetadata.Len() > 0 && copiedMetadata.Len() > 0 {
					origMap, ok := origMetadata.Interface().(map[string][]int)
					if !ok {
						t.Fatal("Failed to cast original Metadata to map[string][]int")
					}
					copiedMap, ok := copiedMetadata.Interface().(map[string][]int)
					if !ok {
						t.Fatal("Failed to cast copied Metadata to map[string][]int")
					}
					origMap["key1"][0] = 999
					if copiedMap["key1"][0] == 999 {
						t.Error("Modifying original Metadata slice affected copy")
					}
					origMap["new_key"] = []int{100}
					if _, exists := copiedMap["new_key"]; exists {
						t.Error("Adding to original map affected copy")
					}
				}

				origConfig := origVal.FieldByName("Config")
				copiedConfig := copiedVal.FieldByName("Config")
				if !origConfig.IsNil() && !copiedConfig.IsNil() {
					if origConfig.Pointer() == copiedConfig.Pointer() {
						t.Error("Config pointers are the same")
					}
					origValues, ok := origConfig.Elem().FieldByName("Values").Interface().([]string)
					if !ok {
						t.Fatal("Failed to cast original Config.Values to []string")
					}
					copiedValues, ok := copiedConfig.Elem().FieldByName("Values").Interface().([]string)
					if !ok {
						t.Fatal("Failed to cast copied Config.Values to []string")
					}
					origValues[0] = "modified_value"
					if copiedValues[0] == "modified_value" {
						t.Error("Modifying original Config.Values affected copy")
					}
					origCounts, ok := origConfig.Elem().FieldByName("Counts").Interface().([]int)
					if !ok {
						t.Fatal("Failed to cast original Config.Counts to []int")
					}
					copiedCounts, ok := copiedConfig.Elem().FieldByName("Counts").Interface().([]int)
					if !ok {
						t.Fatal("Failed to cast copied Config.Counts to []int")
					}
					origCounts[0] = 999
					if copiedCounts[0] == 999 {
						t.Error("Modifying original Config.Counts affected copy")
					}
				}
			},
		},
		{
			name: "success_deeply_nested_struct_with_lists",
			input: struct {
				Level1 struct {
					Data   []string
					Level2 struct {
						Items  []int
						Level3 struct {
							Values [][]string
						}
					}
				}
			}{
				Level1: struct {
					Data   []string
					Level2 struct {
						Items  []int
						Level3 struct {
							Values [][]string
						}
					}
				}{
					Data: []string{"l1a", "l1b"},
					Level2: struct {
						Items  []int
						Level3 struct {
							Values [][]string
						}
					}{
						Items: []int{1, 2, 3},
						Level3: struct {
							Values [][]string
						}{
							Values: [][]string{{"a", "b"}, {"c", "d"}},
						},
					},
				},
			},
			expectedResult: struct {
				Level1 struct {
					Data   []string
					Level2 struct {
						Items  []int
						Level3 struct {
							Values [][]string
						}
					}
				}
			}{
				Level1: struct {
					Data   []string
					Level2 struct {
						Items  []int
						Level3 struct {
							Values [][]string
						}
					}
				}{
					Data: []string{"l1a", "l1b"},
					Level2: struct {
						Items  []int
						Level3 struct {
							Values [][]string
						}
					}{
						Items: []int{1, 2, 3},
						Level3: struct {
							Values [][]string
						}{
							Values: [][]string{{"a", "b"}, {"c", "d"}},
						},
					},
				},
			},
			validateFunc: func(t *testing.T, original, copied interface{}) {
				origVal := reflect.ValueOf(original)
				copiedVal := reflect.ValueOf(copied)

				origL1 := origVal.FieldByName("Level1")
				copiedL1 := copiedVal.FieldByName("Level1")

				origData, ok := origL1.FieldByName("Data").Interface().([]string)
				if !ok {
					t.Fatal("Failed to cast original Level1.Data to []string")
				}
				copiedData, ok := copiedL1.FieldByName("Data").Interface().([]string)
				if !ok {
					t.Fatal("Failed to cast copied Level1.Data to []string")
				}
				origData[0] = "modified"
				if copiedData[0] == "modified" {
					t.Error("Modifying original Level1.Data affected copy")
				}

				origL2 := origL1.FieldByName("Level2")
				copiedL2 := copiedL1.FieldByName("Level2")

				origItems, ok := origL2.FieldByName("Items").Interface().([]int)
				if !ok {
					t.Fatal("Failed to cast original Level2.Items to []int")
				}
				copiedItems, ok := copiedL2.FieldByName("Items").Interface().([]int)
				if !ok {
					t.Fatal("Failed to cast copied Level2.Items to []int")
				}
				origItems[0] = 999
				if copiedItems[0] == 999 {
					t.Error("Modifying original Level2.Items affected copy")
				}

				origL3 := origL2.FieldByName("Level3")
				copiedL3 := copiedL2.FieldByName("Level3")

				origValues, ok := origL3.FieldByName("Values").Interface().([][]string)
				if !ok {
					t.Fatal("Failed to cast original Level3.Values to [][]string")
				}
				copiedValues, ok := copiedL3.FieldByName("Values").Interface().([][]string)
				if !ok {
					t.Fatal("Failed to cast copied Level3.Values to [][]string")
				}
				origValues[0][0] = "deeply_modified"
				if copiedValues[0][0] == "deeply_modified" {
					t.Error("Modifying original Level3.Values affected copy")
				}
			},
		},
		{
			name: "success_struct_with_slice_of_pointers",
			input: struct {
				Name string
				Refs []*int
			}{
				Name: "pointer_slice",
				Refs: []*int{intPtr(1), intPtr(2), intPtr(3)},
			},
			expectedResult: struct {
				Name string
				Refs []*int
			}{
				Name: "pointer_slice",
				Refs: []*int{intPtr(1), intPtr(2), intPtr(3)},
			},
			validateFunc: func(t *testing.T, original, copied interface{}) {
				origVal := reflect.ValueOf(original)
				copiedVal := reflect.ValueOf(copied)

				origRefs, ok := origVal.FieldByName("Refs").Interface().([]*int)
				if !ok {
					t.Fatal("Failed to cast original Refs to []*int")
				}
				copiedRefs, ok := copiedVal.FieldByName("Refs").Interface().([]*int)
				if !ok {
					t.Fatal("Failed to cast copied Refs to []*int")
				}

				if len(origRefs) > 0 && len(copiedRefs) > 0 {
					if origRefs[0] == copiedRefs[0] {
						t.Error("Pointer elements are the same")
					}
					*origRefs[0] = 999
					if *copiedRefs[0] == 999 {
						t.Error("Modifying original pointer element affected copy")
					}
				}
			},
		},
		{
			name: "success_struct_with_map_of_struct_slices",
			input: struct {
				Data map[string][]testStruct
			}{
				Data: map[string][]testStruct{
					"group1": {{Name: "a", Value: 1}, {Name: "b", Value: 2}},
					"group2": {{Name: "c", Value: 3}},
				},
			},
			expectedResult: struct {
				Data map[string][]testStruct
			}{
				Data: map[string][]testStruct{
					"group1": {{Name: "a", Value: 1}, {Name: "b", Value: 2}},
					"group2": {{Name: "c", Value: 3}},
				},
			},
			validateFunc: func(t *testing.T, original, copied interface{}) {
				origVal := reflect.ValueOf(original)
				copiedVal := reflect.ValueOf(copied)

				origData, ok := origVal.FieldByName("Data").Interface().(map[string][]testStruct)
				if !ok {
					t.Fatal("Failed to cast original Data to map[string][]testStruct")
				}
				copiedData, ok := copiedVal.FieldByName("Data").Interface().(map[string][]testStruct)
				if !ok {
					t.Fatal("Failed to cast copied Data to map[string][]testStruct")
				}

				origData["group1"][0].Name = "modified"
				if copiedData["group1"][0].Name == "modified" {
					t.Error("Modifying original struct in slice affected copy")
				}

				origData["new_group"] = []testStruct{{Name: "new", Value: 99}}
				if _, exists := copiedData["new_group"]; exists {
					t.Error("Adding to original map affected copy")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := DeepCopy(tt.input)

			if !reflect.DeepEqual(result, tt.expectedResult) {
				t.Errorf("Expected result %+v, got %+v", tt.expectedResult, result)
			}

			if tt.validateFunc != nil {
				tt.validateFunc(t, tt.input, result)
			}

			if tt.shouldBeSame {
				if !reflect.DeepEqual(reflect.ValueOf(tt.input), reflect.ValueOf(result)) {
					t.Errorf("Expected same values for primitives/nil")
				}
			}
		})
	}
}

func TestMergePlanAndStateMap(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name           string
		existingAttrs  map[string]attr.Value
		attrsToMerge   map[string]attr.Value
		expectedResult map[string]attr.Value
		validateFunc   func(t *testing.T, result map[string]attr.Value)
	}{
		{
			name: "success_merge_nested_objects",
			existingAttrs: map[string]attr.Value{
				"config": types.ObjectValueMust(
					map[string]attr.Type{
						"name":  types.StringType,
						"value": types.Int64Type,
					},
					map[string]attr.Value{
						"name":  types.StringValue("existing"),
						"value": types.Int64Value(10),
					},
				),
			},
			attrsToMerge: map[string]attr.Value{
				"config": types.ObjectValueMust(
					map[string]attr.Type{
						"name":  types.StringType,
						"value": types.Int64Type,
					},
					map[string]attr.Value{
						"name":  types.StringValue("updated"),
						"value": types.Int64Value(20),
					},
				),
			},
			validateFunc: func(t *testing.T, result map[string]attr.Value) {
				obj, ok := result["config"].(types.Object)
				if !ok {
					t.Fatalf("Expected types.Object for 'config', got %T", result["config"])
				}
				name, ok := obj.Attributes()["name"].(types.String)
				if !ok {
					t.Fatalf("Expected types.String for 'name', got %T", obj.Attributes()["name"])
				}
				value, ok := obj.Attributes()["value"].(types.Int64)
				if !ok {
					t.Fatalf("Expected types.Int64 for 'value', got %T", obj.Attributes()["value"])
				}
				if name.ValueString() != "updated" {
					t.Errorf("Expected name 'updated', got '%s'", name.ValueString())
				}
				if value.ValueInt64() != 20 {
					t.Errorf("Expected value 20, got %d", value.ValueInt64())
				}
			},
		},
		{
			name: "success_merge_map_with_object_values",
			existingAttrs: map[string]attr.Value{
				"instances": types.MapValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"id":     types.StringType,
							"status": types.StringType,
						},
					},
					map[string]attr.Value{
						"instance1": types.ObjectValueMust(
							map[string]attr.Type{
								"id":     types.StringType,
								"status": types.StringType,
							},
							map[string]attr.Value{
								"id":     types.StringValue("i-123"),
								"status": types.StringValue("running"),
							},
						),
					},
				),
			},
			attrsToMerge: map[string]attr.Value{
				"instances": types.MapValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"id":     types.StringType,
							"status": types.StringType,
						},
					},
					map[string]attr.Value{
						"instance1": types.ObjectValueMust(
							map[string]attr.Type{
								"id":     types.StringType,
								"status": types.StringType,
							},
							map[string]attr.Value{
								"id":     types.StringValue("i-123"),
								"status": types.StringValue("stopped"),
							},
						),
						"instance2": types.ObjectValueMust(
							map[string]attr.Type{
								"id":     types.StringType,
								"status": types.StringType,
							},
							map[string]attr.Value{
								"id":     types.StringValue("i-456"),
								"status": types.StringValue("running"),
							},
						),
					},
				),
			},
			validateFunc: func(t *testing.T, result map[string]attr.Value) {
				mapVal, ok := result["instances"].(types.Map)
				if !ok {
					t.Fatalf("Expected types.Map for 'instances', got %T", result["instances"])
				}
				elems := mapVal.Elements()
				if len(elems) != 2 {
					t.Errorf("Expected 2 map elements, got %d", len(elems))
				}
				inst1, ok := elems["instance1"].(types.Object)
				if !ok {
					t.Fatalf("Expected types.Object for 'instance1', got %T", elems["instance1"])
				}
				status1, ok := inst1.Attributes()["status"].(types.String)
				if !ok {
					t.Fatalf("Expected types.String for 'status', got %T", inst1.Attributes()["status"])
				}
				if status1.ValueString() != "stopped" {
					t.Errorf("Expected instance1 status 'stopped', got '%s'", status1.ValueString())
				}
				inst2, ok := elems["instance2"].(types.Object)
				if !ok {
					t.Fatalf("Expected types.Object for 'instance2', got %T", elems["instance2"])
				}
				id2, ok := inst2.Attributes()["id"].(types.String)
				if !ok {
					t.Fatalf("Expected types.String for 'id', got %T", inst2.Attributes()["id"])
				}
				if id2.ValueString() != "i-456" {
					t.Errorf("Expected instance2 id 'i-456', got '%s'", id2.ValueString())
				}
			},
		},
		{
			name: "success_merge_map_with_nested_objects",
			existingAttrs: map[string]attr.Value{
				"configs": types.MapValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"name": types.StringType,
							"settings": types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"enabled": types.BoolType,
									"count":   types.Int64Type,
								},
							},
						},
					},
					map[string]attr.Value{
						"config1": types.ObjectValueMust(
							map[string]attr.Type{
								"name": types.StringType,
								"settings": types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"enabled": types.BoolType,
										"count":   types.Int64Type,
									},
								},
							},
							map[string]attr.Value{
								"name": types.StringValue("test"),
								"settings": types.ObjectValueMust(
									map[string]attr.Type{
										"enabled": types.BoolType,
										"count":   types.Int64Type,
									},
									map[string]attr.Value{
										"enabled": types.BoolValue(true),
										"count":   types.Int64Value(5),
									},
								),
							},
						),
					},
				),
			},
			attrsToMerge: map[string]attr.Value{
				"configs": types.MapValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"name": types.StringType,
							"settings": types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"enabled": types.BoolType,
									"count":   types.Int64Type,
								},
							},
						},
					},
					map[string]attr.Value{
						"config1": types.ObjectValueMust(
							map[string]attr.Type{
								"name": types.StringType,
								"settings": types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"enabled": types.BoolType,
										"count":   types.Int64Type,
									},
								},
							},
							map[string]attr.Value{
								"name": types.StringValue("test"),
								"settings": types.ObjectValueMust(
									map[string]attr.Type{
										"enabled": types.BoolType,
										"count":   types.Int64Type,
									},
									map[string]attr.Value{
										"enabled": types.BoolValue(false),
										"count":   types.Int64Value(10),
									},
								),
							},
						),
					},
				),
			},
			validateFunc: func(t *testing.T, result map[string]attr.Value) {
				mapVal, ok := result["configs"].(types.Map)
				if !ok {
					t.Fatalf("Expected types.Map for 'configs', got %T", result["configs"])
				}
				config1, ok := mapVal.Elements()["config1"].(types.Object)
				if !ok {
					t.Fatalf("Expected types.Object for 'config1', got %T", mapVal.Elements()["config1"])
				}
				settings, ok := config1.Attributes()["settings"].(types.Object)
				if !ok {
					t.Fatalf("Expected types.Object for 'settings', got %T", config1.Attributes()["settings"])
				}
				enabled, ok := settings.Attributes()["enabled"].(types.Bool)
				if !ok {
					t.Fatalf("Expected types.Bool for 'enabled', got %T", settings.Attributes()["enabled"])
				}
				count, ok := settings.Attributes()["count"].(types.Int64)
				if !ok {
					t.Fatalf("Expected types.Int64 for 'count', got %T", settings.Attributes()["count"])
				}
				if enabled.ValueBool() != false {
					t.Errorf("Expected enabled false, got %v", enabled.ValueBool())
				}
				if count.ValueInt64() != 10 {
					t.Errorf("Expected count 10, got %d", count.ValueInt64())
				}
			},
		},
		{
			name: "success_merge_list_with_object_elements",
			existingAttrs: map[string]attr.Value{
				"items": types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"id":   types.StringType,
							"name": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"id":   types.StringType,
								"name": types.StringType,
							},
							map[string]attr.Value{
								"id":   types.StringValue("1"),
								"name": types.StringValue("item1"),
							},
						),
						types.ObjectValueMust(
							map[string]attr.Type{
								"id":   types.StringType,
								"name": types.StringType,
							},
							map[string]attr.Value{
								"id":   types.StringValue("2"),
								"name": types.StringValue("item2"),
							},
						),
					},
				),
			},
			attrsToMerge: map[string]attr.Value{
				"items": types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"id":   types.StringType,
							"name": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"id":   types.StringType,
								"name": types.StringType,
							},
							map[string]attr.Value{
								"id":   types.StringValue("1"),
								"name": types.StringValue("updated1"),
							},
						),
						types.ObjectValueMust(
							map[string]attr.Type{
								"id":   types.StringType,
								"name": types.StringType,
							},
							map[string]attr.Value{
								"id":   types.StringValue("2"),
								"name": types.StringValue("updated2"),
							},
						),
					},
				),
			},
			validateFunc: func(t *testing.T, result map[string]attr.Value) {
				listVal, ok := result["items"].(types.List)
				if !ok {
					t.Fatalf("Expected types.List for 'items', got %T", result["items"])
				}
				elems := listVal.Elements()
				if len(elems) != 2 {
					t.Errorf("Expected 2 list elements, got %d", len(elems))
				}
				item1, ok := elems[0].(types.Object)
				if !ok {
					t.Fatalf("Expected types.Object for first item, got %T", elems[0])
				}
				name1, ok := item1.Attributes()["name"].(types.String)
				if !ok {
					t.Fatalf("Expected types.String for 'name', got %T", item1.Attributes()["name"])
				}
				if name1.ValueString() != "updated1" {
					t.Errorf("Expected name 'updated1', got '%s'", name1.ValueString())
				}
				item2, ok := elems[1].(types.Object)
				if !ok {
					t.Fatalf("Expected types.Object for second item, got %T", elems[1])
				}
				name2, ok := item2.Attributes()["name"].(types.String)
				if !ok {
					t.Fatalf("Expected types.String for 'name', got %T", item2.Attributes()["name"])
				}
				if name2.ValueString() != "updated2" {
					t.Errorf("Expected name 'updated2', got '%s'", name2.ValueString())
				}
			},
		},
		{
			name: "success_merge_list_with_unknown_values_preserves_state",
			existingAttrs: map[string]attr.Value{
				"items": types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"id":   types.StringType,
							"name": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"id":   types.StringType,
								"name": types.StringType,
							},
							map[string]attr.Value{
								"id":   types.StringValue("1"),
								"name": types.StringValue("existing"),
							},
						),
					},
				),
			},
			attrsToMerge: map[string]attr.Value{
				"items": types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"id":   types.StringType,
							"name": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectUnknown(
							map[string]attr.Type{
								"id":   types.StringType,
								"name": types.StringType,
							},
						),
					},
				),
			},
			validateFunc: func(t *testing.T, result map[string]attr.Value) {
				listVal, ok := result["items"].(types.List)
				if !ok {
					t.Fatalf("Expected types.List for 'items', got %T", result["items"])
				}
				elems := listVal.Elements()
				if len(elems) != 1 {
					t.Errorf("Expected 1 list element, got %d", len(elems))
				}
				item, ok := elems[0].(types.Object)
				if !ok {
					t.Fatalf("Expected types.Object for item, got %T", elems[0])
				}
				name, ok := item.Attributes()["name"].(types.String)
				if !ok {
					t.Fatalf("Expected types.String for 'name', got %T", item.Attributes()["name"])
				}
				if name.ValueString() != "existing" {
					t.Errorf("Expected existing value to be preserved, got '%s'", name.ValueString())
				}
			},
		},
		{
			name: "success_merge_set_with_object_elements_filters_unknown",
			existingAttrs: map[string]attr.Value{
				"tags": types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"key":   types.StringType,
							"value": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"key":   types.StringType,
								"value": types.StringType,
							},
							map[string]attr.Value{
								"key":   types.StringValue("env"),
								"value": types.StringValue("prod"),
							},
						),
					},
				),
			},
			attrsToMerge: map[string]attr.Value{
				"tags": types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"key":   types.StringType,
							"value": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"key":   types.StringType,
								"value": types.StringType,
							},
							map[string]attr.Value{
								"key":   types.StringValue("env"),
								"value": types.StringValue("dev"),
							},
						),
						types.ObjectUnknown(
							map[string]attr.Type{
								"key":   types.StringType,
								"value": types.StringType,
							},
						),
						types.ObjectValueMust(
							map[string]attr.Type{
								"key":   types.StringType,
								"value": types.StringType,
							},
							map[string]attr.Value{
								"key":   types.StringValue("team"),
								"value": types.StringValue("platform"),
							},
						),
					},
				),
			},
			validateFunc: func(t *testing.T, result map[string]attr.Value) {
				setVal, ok := result["tags"].(types.Set)
				if !ok {
					t.Fatalf("Expected types.Set for 'tags', got %T", result["tags"])
				}
				elems := setVal.Elements()
				if len(elems) != 2 {
					t.Errorf("Expected 2 set elements (unknown filtered), got %d", len(elems))
				}
			},
		},
		{
			name: "success_merge_list_non_object_elements_replaces_entirely",
			existingAttrs: map[string]attr.Value{
				"numbers": types.ListValueMust(
					types.Int64Type,
					[]attr.Value{
						types.Int64Value(1),
						types.Int64Value(2),
						types.Int64Value(3),
					},
				),
			},
			attrsToMerge: map[string]attr.Value{
				"numbers": types.ListValueMust(
					types.Int64Type,
					[]attr.Value{
						types.Int64Value(4),
						types.Int64Value(5),
					},
				),
			},
			validateFunc: func(t *testing.T, result map[string]attr.Value) {
				listVal, ok := result["numbers"].(types.List)
				if !ok {
					t.Fatalf("Expected types.List for 'numbers', got %T", result["numbers"])
				}
				elems := listVal.Elements()
				if len(elems) != 2 {
					t.Errorf("Expected 2 elements, got %d", len(elems))
				}
				val0, ok := elems[0].(types.Int64)
				if !ok {
					t.Fatalf("Expected types.Int64 for first element, got %T", elems[0])
				}
				if val0.ValueInt64() != 4 {
					t.Errorf("Expected first element 4, got %d", val0.ValueInt64())
				}
			},
		},
		{
			name: "success_merge_map_non_object_values_replaces_entirely",
			existingAttrs: map[string]attr.Value{
				"settings": types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"key1": types.StringValue("value1"),
						"key2": types.StringValue("value2"),
					},
				),
			},
			attrsToMerge: map[string]attr.Value{
				"settings": types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"key1": types.StringValue("updated1"),
						"key3": types.StringValue("value3"),
					},
				),
			},
			validateFunc: func(t *testing.T, result map[string]attr.Value) {
				mapVal, ok := result["settings"].(types.Map)
				if !ok {
					t.Fatalf("Expected types.Map for 'settings', got %T", result["settings"])
				}
				elems := mapVal.Elements()
				if len(elems) != 2 {
					t.Errorf("Expected 2 elements, got %d", len(elems))
				}
				valKey1, ok := elems["key1"].(types.String)
				if !ok {
					t.Fatalf("Expected types.String for 'key1', got %T", elems["key1"])
				}
				if valKey1.ValueString() != "updated1" {
					t.Errorf("Expected key1 'updated1', got '%s'", valKey1.ValueString())
				}
				if _, exists := elems["key2"]; exists {
					t.Error("Expected key2 to be replaced, but it still exists")
				}
			},
		},
		{
			name: "success_skip_null_plan_values",
			existingAttrs: map[string]attr.Value{
				"name":  types.StringValue("existing"),
				"count": types.Int64Value(10),
			},
			attrsToMerge: map[string]attr.Value{
				"name":  types.StringNull(),
				"count": types.Int64Value(20),
			},
			validateFunc: func(t *testing.T, result map[string]attr.Value) {
				valName, ok := result["name"].(types.String)
				if !ok {
					t.Fatalf("Expected types.String for 'name', got %T", result["name"])
				}
				if valName.ValueString() != "existing" {
					t.Error("Null plan value should not override existing state")
				}
				valCount, ok := result["count"].(types.Int64)
				if !ok {
					t.Fatalf("Expected types.Int64 for 'count', got %T", result["count"])
				}
				if valCount.ValueInt64() != 20 {
					t.Error("Non-null plan value should override existing state")
				}
			},
		},
		{
			name: "success_skip_unknown_plan_values",
			existingAttrs: map[string]attr.Value{
				"name": types.StringValue("existing"),
			},
			attrsToMerge: map[string]attr.Value{
				"name": types.StringUnknown(),
			},
			validateFunc: func(t *testing.T, result map[string]attr.Value) {
				valName, ok := result["name"].(types.String)
				if !ok {
					t.Fatalf("Expected types.String for 'name', got %T", result["name"])
				}
				if valName.ValueString() != "existing" {
					t.Error("Unknown plan value should not override existing state")
				}
			},
		},
		{
			name: "success_merge_deeply_nested_list_objects",
			existingAttrs: map[string]attr.Value{
				"groups": types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"name": types.StringType,
							"members": types.ListType{
								ElemType: types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"id":   types.StringType,
										"role": types.StringType,
									},
								},
							},
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"name": types.StringType,
								"members": types.ListType{
									ElemType: types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"id":   types.StringType,
											"role": types.StringType,
										},
									},
								},
							},
							map[string]attr.Value{
								"name": types.StringValue("admins"),
								"members": types.ListValueMust(
									types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"id":   types.StringType,
											"role": types.StringType,
										},
									},
									[]attr.Value{
										types.ObjectValueMust(
											map[string]attr.Type{
												"id":   types.StringType,
												"role": types.StringType,
											},
											map[string]attr.Value{
												"id":   types.StringValue("u1"),
												"role": types.StringValue("admin"),
											},
										),
									},
								),
							},
						),
					},
				),
			},
			attrsToMerge: map[string]attr.Value{
				"groups": types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"name": types.StringType,
							"members": types.ListType{
								ElemType: types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"id":   types.StringType,
										"role": types.StringType,
									},
								},
							},
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"name": types.StringType,
								"members": types.ListType{
									ElemType: types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"id":   types.StringType,
											"role": types.StringType,
										},
									},
								},
							},
							map[string]attr.Value{
								"name": types.StringValue("admins"),
								"members": types.ListValueMust(
									types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"id":   types.StringType,
											"role": types.StringType,
										},
									},
									[]attr.Value{
										types.ObjectValueMust(
											map[string]attr.Type{
												"id":   types.StringType,
												"role": types.StringType,
											},
											map[string]attr.Value{
												"id":   types.StringValue("u1"),
												"role": types.StringValue("superadmin"),
											},
										),
									},
								),
							},
						),
					},
				),
			},
			validateFunc: func(t *testing.T, result map[string]attr.Value) {
				listVal, ok := result["groups"].(types.List)
				if !ok {
					t.Fatalf("Expected types.List for 'groups', got %T", result["groups"])
				}
				group, ok := listVal.Elements()[0].(types.Object)
				if !ok {
					t.Fatalf("Expected types.Object for first group, got %T", listVal.Elements()[0])
				}
				members, ok := group.Attributes()["members"].(types.List)
				if !ok {
					t.Fatalf("Expected types.List for 'members', got %T", group.Attributes()["members"])
				}
				member, ok := members.Elements()[0].(types.Object)
				if !ok {
					t.Fatalf("Expected types.Object for first member, got %T", members.Elements()[0])
				}
				role, ok := member.Attributes()["role"].(types.String)
				if !ok {
					t.Fatalf("Expected types.String for 'role', got %T", member.Attributes()["role"])
				}
				if role.ValueString() != "superadmin" {
					t.Errorf("Expected deeply nested role 'superadmin', got '%s'", role.ValueString())
				}
			},
		},
		{
			name:          "success_empty_existing_attrs",
			existingAttrs: map[string]attr.Value{},
			attrsToMerge: map[string]attr.Value{
				"name": types.StringValue("new"),
			},
			validateFunc: func(t *testing.T, result map[string]attr.Value) {
				valName, ok := result["name"].(types.String)
				if !ok {
					t.Fatalf("Expected types.String for 'name', got %T", result["name"])
				}
				if valName.ValueString() != "new" {
					t.Error("Expected new value to be added")
				}
			},
		},
		{
			name: "success_empty_attrs_to_merge",
			existingAttrs: map[string]attr.Value{
				"name": types.StringValue("existing"),
			},
			attrsToMerge: map[string]attr.Value{},
			validateFunc: func(t *testing.T, result map[string]attr.Value) {
				valName, ok := result["name"].(types.String)
				if !ok {
					t.Fatalf("Expected types.String for 'name', got %T", result["name"])
				}
				if valName.ValueString() != "existing" {
					t.Error("Expected existing value to be preserved")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a copy of existing attrs to avoid mutation
			existingCopy := make(map[string]attr.Value)
			for k, v := range tt.existingAttrs {
				existingCopy[k] = v
			}

			// Execute the merge
			mergePlanAndStateMap(ctx, existingCopy, tt.attrsToMerge)

			// Validate using custom validation function if provided
			if tt.validateFunc != nil {
				tt.validateFunc(t, existingCopy)
			}

			// If expectedResult is provided, do deep equality check
			if tt.expectedResult != nil {
				if !reflect.DeepEqual(existingCopy, tt.expectedResult) {
					t.Errorf("Result mismatch.\nExpected: %+v\nGot: %+v", tt.expectedResult, existingCopy)
				}
			}
		})
	}
}

func TestObjectToMap(t *testing.T) {
	t.Parallel()

	// Test prototype structs
	type SimpleStruct struct {
		Name  string `mapstructure:"name"`
		Count int64  `mapstructure:"count"`
	}

	type NestedStruct struct {
		ID      string       `mapstructure:"id"`
		Config  SimpleStruct `mapstructure:"config"`
		Enabled bool         `mapstructure:"enabled"`
	}

	type PointerFieldStruct struct {
		Name  *string `mapstructure:"name"`
		Value *int64  `mapstructure:"value"`
	}

	type ListStruct struct {
		Items   []string `mapstructure:"items"`
		Numbers []int64  `mapstructure:"numbers"`
	}

	type MapStruct struct {
		Settings map[string]string       `mapstructure:"settings"`
		Configs  map[string]SimpleStruct `mapstructure:"configs"`
	}

	tests := []struct {
		name          string
		input         types.Object
		prototype     interface{}
		expectedError bool
		validateFunc  func(t *testing.T, result map[string]interface{})
	}{
		{
			name: "success_simple_object_with_basic_types",
			input: types.ObjectValueMust(
				map[string]attr.Type{
					"name":  types.StringType,
					"count": types.Int64Type,
				},
				map[string]attr.Value{
					"name":  types.StringValue("test"),
					"count": types.Int64Value(42),
				},
			),
			prototype:     &SimpleStruct{},
			expectedError: false,
			validateFunc: func(t *testing.T, result map[string]interface{}) {
				if len(result) != 2 {
					t.Errorf("Expected 2 attributes, got %d", len(result))
				}
				if result["name"] != "test" {
					t.Errorf("Expected name 'test', got %v", result["name"])
				}
				if result["count"] != int64(42) {
					t.Errorf("Expected count 42, got %v", result["count"])
				}
			},
		},
		{
			name: "success_nested_object",
			input: types.ObjectValueMust(
				map[string]attr.Type{
					"id":      types.StringType,
					"enabled": types.BoolType,
					"config": types.ObjectType{AttrTypes: map[string]attr.Type{
						"name":  types.StringType,
						"count": types.Int64Type,
					}},
				},
				map[string]attr.Value{
					"id":      types.StringValue("obj-123"),
					"enabled": types.BoolValue(true),
					"config": types.ObjectValueMust(
						map[string]attr.Type{
							"name":  types.StringType,
							"count": types.Int64Type,
						},
						map[string]attr.Value{
							"name":  types.StringValue("nested"),
							"count": types.Int64Value(10),
						},
					),
				},
			),
			prototype:     &NestedStruct{},
			expectedError: false,
			validateFunc: func(t *testing.T, result map[string]interface{}) {
				if result["id"] != "obj-123" {
					t.Errorf("Expected id 'obj-123', got %v", result["id"])
				}
				if result["enabled"] != true {
					t.Errorf("Expected enabled true, got %v", result["enabled"])
				}
				configMap, ok := result["config"].(map[string]interface{})
				if !ok {
					t.Fatal("Expected config to be map[string]interface{}")
				}
				if configMap["name"] != "nested" {
					t.Errorf("Expected config name 'nested', got %v", configMap["name"])
				}
				if configMap["count"] != int64(10) {
					t.Errorf("Expected config count 10, got %v", configMap["count"])
				}
			},
		},
		{
			name: "success_null_values_converted_to_nil",
			input: types.ObjectValueMust(
				map[string]attr.Type{
					"name":  types.StringType,
					"count": types.Int64Type,
				},
				map[string]attr.Value{
					"name":  types.StringNull(),
					"count": types.Int64Value(5),
				},
			),
			prototype:     &SimpleStruct{},
			expectedError: false,
			validateFunc: func(t *testing.T, result map[string]interface{}) {
				if result["name"] != nil {
					t.Errorf("Expected name to be nil, got %v", result["name"])
				}
				if result["count"] != int64(5) {
					t.Errorf("Expected count 5, got %v", result["count"])
				}
			},
		},
		{
			name: "success_unknown_values_skipped",
			input: types.ObjectValueMust(
				map[string]attr.Type{
					"name":  types.StringType,
					"count": types.Int64Type,
				},
				map[string]attr.Value{
					"name":  types.StringUnknown(),
					"count": types.Int64Value(7),
				},
			),
			prototype:     &SimpleStruct{},
			expectedError: false,
			validateFunc: func(t *testing.T, result map[string]interface{}) {
				if _, exists := result["name"]; exists {
					t.Error("Expected unknown name to be skipped")
				}
				if result["count"] != int64(7) {
					t.Errorf("Expected count 7, got %v", result["count"])
				}
			},
		},
		{
			name: "success_zero_values_skipped",
			input: types.ObjectValueMust(
				map[string]attr.Type{
					"name":  types.StringType,
					"count": types.Int64Type,
				},
				map[string]attr.Value{
					"name":  types.StringValue(""),
					"count": types.Int64Value(0),
				},
			),
			prototype:     &SimpleStruct{},
			expectedError: false,
			validateFunc: func(t *testing.T, result map[string]interface{}) {
				if len(result) != 0 {
					t.Errorf("Expected empty result for zero values, got %d items", len(result))
				}
			},
		},
		{
			name: "success_pointer_fields_wrapped",
			input: types.ObjectValueMust(
				map[string]attr.Type{
					"name":  types.StringType,
					"value": types.Int64Type,
				},
				map[string]attr.Value{
					"name":  types.StringValue("pointer-test"),
					"value": types.Int64Value(99),
				},
			),
			prototype:     &PointerFieldStruct{},
			expectedError: false,
			validateFunc: func(t *testing.T, result map[string]interface{}) {
				nameVal, ok := result["name"].(*string)
				if !ok {
					t.Fatal("Expected name to be *string")
				}
				if *nameVal != "pointer-test" {
					t.Errorf("Expected name 'pointer-test', got %v", *nameVal)
				}
				valueVal, ok := result["value"].(*int64)
				if !ok {
					t.Fatal("Expected value to be *int64")
				}
				if *valueVal != int64(99) {
					t.Errorf("Expected value 99, got %v", *valueVal)
				}
			},
		},
		{
			name: "success_non_pointer_fields_unwrapped",
			input: types.ObjectValueMust(
				map[string]attr.Type{
					"name":  types.StringType,
					"count": types.Int64Type,
				},
				map[string]attr.Value{
					"name":  types.StringValue("unwrap-test"),
					"count": types.Int64Value(123),
				},
			),
			prototype:     &SimpleStruct{},
			expectedError: false,
			validateFunc: func(t *testing.T, result map[string]interface{}) {
				nameVal, ok := result["name"].(string)
				if !ok {
					t.Fatal("Expected name to be string, not pointer")
				}
				if nameVal != "unwrap-test" {
					t.Errorf("Expected name 'unwrap-test', got %v", nameVal)
				}
			},
		},
		{
			name: "success_list_attribute",
			input: types.ObjectValueMust(
				map[string]attr.Type{
					"items":   types.ListType{ElemType: types.StringType},
					"numbers": types.ListType{ElemType: types.Int64Type},
				},
				map[string]attr.Value{
					"items": types.ListValueMust(types.StringType, []attr.Value{
						types.StringValue("item1"),
						types.StringValue("item2"),
					}),
					"numbers": types.ListValueMust(types.Int64Type, []attr.Value{
						types.Int64Value(1),
						types.Int64Value(2),
					}),
				},
			),
			prototype:     &ListStruct{},
			expectedError: false,
			validateFunc: func(t *testing.T, result map[string]interface{}) {
				items, ok := result["items"].([]interface{})
				if !ok {
					t.Fatal("Expected items to be []interface{}")
				}
				if len(items) != 2 {
					t.Errorf("Expected 2 items, got %d", len(items))
				}
				if items[0] != "item1" {
					t.Errorf("Expected first item 'item1', got %v", items[0])
				}
			},
		},
		{
			name: "success_map_attribute",
			input: types.ObjectValueMust(
				map[string]attr.Type{
					"settings": types.MapType{ElemType: types.StringType},
				},
				map[string]attr.Value{
					"settings": types.MapValueMust(types.StringType, map[string]attr.Value{
						"key1": types.StringValue("value1"),
						"key2": types.StringValue("value2"),
					}),
				},
			),
			prototype:     &MapStruct{},
			expectedError: false,
			validateFunc: func(t *testing.T, result map[string]interface{}) {
				settings, ok := result["settings"].(map[string]interface{})
				if !ok {
					t.Fatal("Expected settings to be map[string]interface{}")
				}
				if settings["key1"] != "value1" {
					t.Errorf("Expected key1 'value1', got %v", settings["key1"])
				}
				if settings["key2"] != "value2" {
					t.Errorf("Expected key2 'value2', got %v", settings["key2"])
				}
			},
		},
		{
			name: "success_map_with_object_values",
			input: types.ObjectValueMust(
				map[string]attr.Type{
					"configs": types.MapType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
						"name":  types.StringType,
						"count": types.Int64Type,
					}}},
				},
				map[string]attr.Value{
					"configs": types.MapValueMust(
						types.ObjectType{AttrTypes: map[string]attr.Type{
							"name":  types.StringType,
							"count": types.Int64Type,
						}},
						map[string]attr.Value{
							"config1": types.ObjectValueMust(
								map[string]attr.Type{
									"name":  types.StringType,
									"count": types.Int64Type,
								},
								map[string]attr.Value{
									"name":  types.StringValue("cfg1"),
									"count": types.Int64Value(1),
								},
							),
						},
					),
				},
			),
			prototype:     &MapStruct{},
			expectedError: false,
			validateFunc: func(t *testing.T, result map[string]interface{}) {
				configs, ok := result["configs"].(map[string]interface{})
				if !ok {
					t.Fatal("Expected configs to be map[string]interface{}")
				}
				config1, ok := configs["config1"].(map[string]interface{})
				if !ok {
					t.Fatal("Expected config1 to be map[string]interface{}")
				}
				if config1["name"] != "cfg1" {
					t.Errorf("Expected config1 name 'cfg1', got %v", config1["name"])
				}
			},
		},
		{
			name: "success_boolean_true_included",
			input: types.ObjectValueMust(
				map[string]attr.Type{
					"enabled": types.BoolType,
				},
				map[string]attr.Value{
					"enabled": types.BoolValue(true),
				},
			),
			prototype: &struct {
				Enabled bool `mapstructure:"enabled"`
			}{},
			expectedError: false,
			validateFunc: func(t *testing.T, result map[string]interface{}) {
				// true is non-zero and should be included
				if result["enabled"] != true {
					t.Errorf("Expected enabled true, got %v", result["enabled"])
				}
			},
		},
		{
			name: "success_float64_type",
			input: types.ObjectValueMust(
				map[string]attr.Type{
					"price": types.Float64Type,
				},
				map[string]attr.Value{
					"price": types.Float64Value(19.99),
				},
			),
			prototype: &struct {
				Price float64 `mapstructure:"price"`
			}{},
			expectedError: false,
			validateFunc: func(t *testing.T, result map[string]interface{}) {
				if result["price"] != float64(19.99) {
					t.Errorf("Expected price 19.99, got %v", result["price"])
				}
			},
		},
		{
			name: "success_empty_object",
			input: types.ObjectValueMust(
				map[string]attr.Type{},
				map[string]attr.Value{},
			),
			prototype:     &struct{}{},
			expectedError: false,
			validateFunc: func(t *testing.T, result map[string]interface{}) {
				if len(result) != 0 {
					t.Errorf("Expected empty result, got %d items", len(result))
				}
			},
		},
		{
			name: "error_null_object",
			input: types.ObjectNull(map[string]attr.Type{
				"name": types.StringType,
			}),
			prototype:     &SimpleStruct{},
			expectedError: true,
		},
		{
			name: "error_unknown_object",
			input: types.ObjectUnknown(map[string]attr.Type{
				"name": types.StringType,
			}),
			prototype:     &SimpleStruct{},
			expectedError: true,
		},
		{
			name: "success_deeply_nested_objects",
			input: types.ObjectValueMust(
				map[string]attr.Type{
					"level1": types.ObjectType{AttrTypes: map[string]attr.Type{
						"level2": types.ObjectType{AttrTypes: map[string]attr.Type{
							"level3": types.StringType,
						}},
					}},
				},
				map[string]attr.Value{
					"level1": types.ObjectValueMust(
						map[string]attr.Type{
							"level2": types.ObjectType{AttrTypes: map[string]attr.Type{
								"level3": types.StringType,
							}},
						},
						map[string]attr.Value{
							"level2": types.ObjectValueMust(
								map[string]attr.Type{
									"level3": types.StringType,
								},
								map[string]attr.Value{
									"level3": types.StringValue("deep"),
								},
							),
						},
					),
				},
			),
			prototype: &struct {
				Level1 struct {
					Level2 struct {
						Level3 string `mapstructure:"level3"`
					} `mapstructure:"level2"`
				} `mapstructure:"level1"`
			}{},
			expectedError: false,
			validateFunc: func(t *testing.T, result map[string]interface{}) {
				level1, ok := result["level1"].(map[string]interface{})
				if !ok {
					t.Fatal("Expected level1 to be map")
				}
				level2, ok := level1["level2"].(map[string]interface{})
				if !ok {
					t.Fatal("Expected level2 to be map")
				}
				if level2["level3"] != "deep" {
					t.Errorf("Expected level3 'deep', got %v", level2["level3"])
				}
			},
		},
		{
			name: "success_mixed_null_and_valid_attributes",
			input: types.ObjectValueMust(
				map[string]attr.Type{
					"name":    types.StringType,
					"count":   types.Int64Type,
					"enabled": types.BoolType,
				},
				map[string]attr.Value{
					"name":    types.StringValue("test"),
					"count":   types.Int64Null(),
					"enabled": types.BoolValue(true),
				},
			),
			prototype: &struct {
				Name    string `mapstructure:"name"`
				Count   int64  `mapstructure:"count"`
				Enabled bool   `mapstructure:"enabled"`
			}{},
			expectedError: false,
			validateFunc: func(t *testing.T, result map[string]interface{}) {
				if result["name"] != "test" {
					t.Errorf("Expected name 'test', got %v", result["name"])
				}
				if result["count"] != nil {
					t.Errorf("Expected count to be nil, got %v", result["count"])
				}
				if result["enabled"] != true {
					t.Errorf("Expected enabled true, got %v", result["enabled"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := objectToMap(tt.input, tt.prototype)

			if tt.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
				return
			}

			if tt.validateFunc != nil {
				tt.validateFunc(t, result)
			}
		})
	}
}

func TestSetTargetValueFromPlanAndState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		planVal      interface{}
		stateVal     interface{}
		targetType   reflect.Type
		validateFunc func(t *testing.T, target reflect.Value)
	}{
		{
			name:       "success_nil_pointer_plan_value",
			planVal:    (*int)(nil),
			stateVal:   42,
			targetType: reflect.TypeOf((*int)(nil)),
			validateFunc: func(t *testing.T, target reflect.Value) {
				if !target.IsNil() {
					t.Error("Expected target to remain zero/nil when plan is nil pointer")
				}
			},
		},
		{
			name:       "success_pointer_int_to_pointer_target",
			planVal:    intPtr(42),
			stateVal:   10,
			targetType: reflect.TypeOf((*int)(nil)),
			validateFunc: func(t *testing.T, target reflect.Value) {
				if target.IsNil() {
					t.Fatal("Expected target to be set, got nil")
				}
				val, ok := target.Interface().(*int)
				if !ok {
					t.Fatal("Type assertion to *int failed")
				}
				if *val != 42 {
					t.Errorf("Expected target value 42, got %d", *val)
				}
			},
		},
		{
			name:       "success_value_int_to_pointer_target_with_addr",
			planVal:    42,
			stateVal:   10,
			targetType: reflect.TypeOf((*int)(nil)),
			validateFunc: func(t *testing.T, target reflect.Value) {
				if target.IsNil() {
					t.Fatal("Expected target to be set, got nil")
				}
				val, ok := target.Interface().(*int)
				if !ok {
					t.Fatal("Type assertion to *int failed")
				}
				if *val != 42 {
					t.Errorf("Expected target value 42, got %d", *val)
				}
			},
		},
		{
			name:       "success_bool_true_to_pointer_target",
			planVal:    true,
			stateVal:   false,
			targetType: reflect.TypeOf((*bool)(nil)),
			validateFunc: func(t *testing.T, target reflect.Value) {
				if target.IsNil() {
					t.Fatal("Expected target to be set, got nil")
				}
				val, ok := target.Interface().(*bool)
				if !ok {
					t.Fatal("Type assertion to *bool failed")
				}
				if !*val {
					t.Error("Expected target value true, got false")
				}
			},
		},
		{
			name:       "success_bool_false_to_pointer_target",
			planVal:    false,
			stateVal:   true,
			targetType: reflect.TypeOf((*bool)(nil)),
			validateFunc: func(t *testing.T, target reflect.Value) {
				if target.IsNil() {
					t.Fatal("Expected target to be set, got nil")
				}
				val, ok := target.Interface().(*bool)
				if !ok {
					t.Fatal("Type assertion to *bool failed")
				}
				if *val {
					t.Error("Expected target value false, got true")
				}
			},
		},
		{
			name:       "success_string_to_pointer_target",
			planVal:    "hello",
			stateVal:   "world",
			targetType: reflect.TypeOf((*string)(nil)),
			validateFunc: func(t *testing.T, target reflect.Value) {
				if target.IsNil() {
					t.Fatal("Expected target to be set, got nil")
				}
				val, ok := target.Interface().(*string)
				if !ok {
					t.Fatal("Type assertion to *string failed")
				}
				if *val != "hello" {
					t.Errorf("Expected target value 'hello', got '%s'", *val)
				}
			},
		},
		{
			name:       "success_empty_string_not_set_to_pointer",
			planVal:    "",
			stateVal:   "existing",
			targetType: reflect.TypeOf((*string)(nil)),
			validateFunc: func(t *testing.T, target reflect.Value) {
				if !target.IsNil() {
					t.Error("Expected target to remain nil when plan value is empty string")
				}
			},
		},
		{
			name:       "success_non_zero_int_updates_pointer",
			planVal:    100,
			stateVal:   50,
			targetType: reflect.TypeOf((*int)(nil)),
			validateFunc: func(t *testing.T, target reflect.Value) {
				if target.IsNil() {
					t.Fatal("Expected target to be set, got nil")
				}
				val, ok := target.Interface().(*int)
				if !ok {
					t.Fatal("Type assertion to *int failed")
				}
				if *val != 100 {
					t.Errorf("Expected target value 100, got %d", *val)
				}
			},
		},
		{
			name:       "success_pointer_string_to_pointer_target",
			planVal:    stringPtr("test"),
			stateVal:   "old",
			targetType: reflect.TypeOf((*string)(nil)),
			validateFunc: func(t *testing.T, target reflect.Value) {
				if target.IsNil() {
					t.Fatal("Expected target to be set, got nil")
				}
				val, ok := target.Interface().(*string)
				if !ok {
					t.Fatal("Type assertion to *string failed")
				}
				if *val != "test" {
					t.Errorf("Expected target value 'test', got '%s'", *val)
				}
			},
		},
		{
			name:       "success_value_to_value_target_simple_int",
			planVal:    42,
			stateVal:   10,
			targetType: reflect.TypeOf(0),
			validateFunc: func(t *testing.T, target reflect.Value) {
				if target.Int() != 42 {
					t.Errorf("Expected target value 42, got %d", target.Int())
				}
			},
		},
		{
			name:       "success_value_to_value_target_bool_true",
			planVal:    true,
			stateVal:   false,
			targetType: reflect.TypeOf(false),
			validateFunc: func(t *testing.T, target reflect.Value) {
				if !target.Bool() {
					t.Error("Expected target value true, got false")
				}
			},
		},
		{
			name:       "success_slice_to_pointer_target",
			planVal:    []int{1, 2, 3},
			stateVal:   []int{4, 5, 6},
			targetType: reflect.TypeOf((*[]int)(nil)),
			validateFunc: func(t *testing.T, target reflect.Value) {
				if target.IsNil() {
					t.Fatal("Expected target to be set, got nil")
				}
				slice, ok := target.Interface().(*[]int)
				if !ok {
					t.Fatal("Type assertion to *[]int failed")
				}
				if !reflect.DeepEqual(*slice, []int{1, 2, 3}) {
					t.Errorf("Expected target slice [1 2 3], got %v", *slice)
				}
			},
		},
		{
			name:       "success_map_to_pointer_target",
			planVal:    map[string]int{"a": 1, "b": 2},
			stateVal:   map[string]int{"c": 3},
			targetType: reflect.TypeOf((*map[string]int)(nil)),
			validateFunc: func(t *testing.T, target reflect.Value) {
				if target.IsNil() {
					t.Fatal("Expected target to be set, got nil")
				}
				m, ok := target.Interface().(*map[string]int)
				if !ok {
					t.Fatal("Type assertion to *map[string]int failed")
				}
				expected := map[string]int{"a": 1, "b": 2}
				if !reflect.DeepEqual(*m, expected) {
					t.Errorf("Expected target map %v, got %v", expected, *m)
				}
			},
		},
		{
			name:       "success_zero_state_sets_plan_to_value_target",
			planVal:    42,
			stateVal:   0, // zero state
			targetType: reflect.TypeOf(0),
			validateFunc: func(t *testing.T, target reflect.Value) {
				if target.Int() != 42 {
					t.Errorf("Expected target value 42, got %d", target.Int())
				}
			},
		},
		{
			name:       "success_nil_slice_not_set",
			planVal:    []int(nil),
			stateVal:   []int{1, 2, 3},
			targetType: reflect.TypeOf((*[]int)(nil)),
			validateFunc: func(t *testing.T, target reflect.Value) {
				if !target.IsNil() {
					t.Error("Expected target to remain nil when plan slice is nil")
				}
			},
		},
		{
			name:       "success_nil_map_not_set",
			planVal:    map[string]int(nil),
			stateVal:   map[string]int{"a": 1},
			targetType: reflect.TypeOf((*map[string]int)(nil)),
			validateFunc: func(t *testing.T, target reflect.Value) {
				if !target.IsNil() {
					t.Error("Expected target to remain nil when plan map is nil")
				}
			},
		},
		{
			name:       "success_chan_to_pointer",
			planVal:    make(chan int, 1),
			stateVal:   make(chan int, 1),
			targetType: reflect.TypeOf((*chan int)(nil)),
			validateFunc: func(t *testing.T, target reflect.Value) {
				if target.IsNil() {
					t.Fatal("Expected target to be set, got nil")
				}
			},
		},
		{
			name:       "success_pointer_to_pointer_different_values",
			planVal:    intPtr(100),
			stateVal:   intPtr(50),
			targetType: reflect.TypeOf((*int)(nil)),
			validateFunc: func(t *testing.T, target reflect.Value) {
				if target.IsNil() {
					t.Fatal("Expected target to be set, got nil")
				}
				val, ok := target.Interface().(*int)
				if !ok {
					t.Fatal("Type assertion to *int failed")
				}
				if *val != 100 {
					t.Errorf("Expected target value 100, got %d", *val)
				}
			},
		},
		{
			name: "success_struct_with_pointer_state_value",
			planVal: struct {
				Name  string
				Value int
			}{
				Name:  "updated",
				Value: 100,
			},
			stateVal: &struct {
				Name  string
				Value int
			}{
				Name:  "existing",
				Value: 50,
			},
			targetType: reflect.TypeOf(struct {
				Name  string
				Value int
			}{}),
			validateFunc: func(t *testing.T, target reflect.Value) {
				result, ok := target.Interface().(struct {
					Name  string
					Value int
				})
				if !ok {
					t.Fatal("Type assertion to struct failed")
				}
				if result.Name != "updated" {
					t.Errorf("Expected Name 'updated', got '%s'", result.Name)
				}
				if result.Value != 100 {
					t.Errorf("Expected Value 100, got %d", result.Value)
				}
			},
		},
		{
			name: "success_nested_struct_with_pointer_state",
			planVal: struct {
				Outer struct {
					Inner struct {
						Value string
					}
				}
			}{
				Outer: struct {
					Inner struct {
						Value string
					}
				}{
					Inner: struct {
						Value string
					}{
						Value: "plan-value",
					},
				},
			},
			stateVal: &struct {
				Outer struct {
					Inner struct {
						Value string
					}
				}
			}{
				Outer: struct {
					Inner struct {
						Value string
					}
				}{
					Inner: struct {
						Value string
					}{
						Value: "state-value",
					},
				},
			},
			targetType: reflect.TypeOf(struct {
				Outer struct {
					Inner struct {
						Value string
					}
				}
			}{}),
			validateFunc: func(t *testing.T, target reflect.Value) {
				result, ok := target.Interface().(struct {
					Outer struct {
						Inner struct {
							Value string
						}
					}
				})
				if !ok {
					t.Fatal("Type assertion to struct failed")
				}
				if result.Outer.Inner.Value != "plan-value" {
					t.Errorf("Expected nested Value 'plan-value', got '%s'", result.Outer.Inner.Value)
				}
			},
		},
		{
			name: "success_struct_mixed_field_types_with_pointer_state",
			planVal: struct {
				Name    string
				Count   int
				Enabled *bool
				Tags    []string
			}{
				Name:    "test",
				Count:   42,
				Enabled: boolPtr(true),
				Tags:    []string{"tag1", "tag2"},
			},
			stateVal: &struct {
				Name    string
				Count   int
				Enabled *bool
				Tags    []string
			}{
				Name:    "old",
				Count:   10,
				Enabled: boolPtr(false),
				Tags:    []string{"old-tag"},
			},
			targetType: reflect.TypeOf(struct {
				Name    string
				Count   int
				Enabled *bool
				Tags    []string
			}{}),
			validateFunc: func(t *testing.T, target reflect.Value) {
				result, ok := target.Interface().(struct {
					Name    string
					Count   int
					Enabled *bool
					Tags    []string
				})
				if !ok {
					t.Fatal("Type assertion to struct failed")
				}
				if result.Name != "test" {
					t.Errorf("Expected Name 'test', got '%s'", result.Name)
				}
				if result.Count != 42 {
					t.Errorf("Expected Count 42, got %d", result.Count)
				}
				if result.Enabled == nil {
					t.Fatal("Expected Enabled to be set, got nil")
				}
				if !*result.Enabled {
					t.Error("Expected Enabled true, got false")
				}
				if !reflect.DeepEqual(result.Tags, []string{"tag1", "tag2"}) {
					t.Errorf("Expected Tags [tag1 tag2], got %v", result.Tags)
				}
			},
		},
		{
			name: "success_deeply_nested_struct_with_pointer_state",
			planVal: struct {
				Level1 struct {
					Level2 struct {
						Level3 struct {
							Value string
						}
					}
				}
			}{
				Level1: struct {
					Level2 struct {
						Level3 struct {
							Value string
						}
					}
				}{
					Level2: struct {
						Level3 struct {
							Value string
						}
					}{
						Level3: struct {
							Value string
						}{
							Value: "deep-plan",
						},
					},
				},
			},
			stateVal: &struct {
				Level1 struct {
					Level2 struct {
						Level3 struct {
							Value string
						}
					}
				}
			}{
				Level1: struct {
					Level2 struct {
						Level3 struct {
							Value string
						}
					}
				}{
					Level2: struct {
						Level3 struct {
							Value string
						}
					}{
						Level3: struct {
							Value string
						}{
							Value: "deep-state",
						},
					},
				},
			},
			targetType: reflect.TypeOf(struct {
				Level1 struct {
					Level2 struct {
						Level3 struct {
							Value string
						}
					}
				}
			}{}),
			validateFunc: func(t *testing.T, target reflect.Value) {
				result, ok := target.Interface().(struct {
					Level1 struct {
						Level2 struct {
							Level3 struct {
								Value string
							}
						}
					}
				})
				if !ok {
					t.Fatal("Type assertion to struct failed")
				}
				if result.Level1.Level2.Level3.Value != "deep-plan" {
					t.Errorf("Expected deeply nested Value 'deep-plan', got '%s'", result.Level1.Level2.Level3.Value)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create target value based on type
			target := reflect.New(tt.targetType).Elem()

			// Get plan and state reflect values
			var planVal reflect.Value
			if tt.planVal == nil {
				planVal = reflect.ValueOf((*int)(nil))
			} else {
				planVal = reflect.ValueOf(tt.planVal)
			}

			var stateVal reflect.Value
			if sv, ok := tt.stateVal.(reflect.Value); ok {
				stateVal = sv
			} else {
				stateVal = reflect.ValueOf(tt.stateVal)
			}

			// Execute function
			setTargetValueFromPlanAndState(planVal, stateVal, target)

			// Validate result
			if tt.validateFunc != nil {
				tt.validateFunc(t, target)
			}
		})
	}
}

// Helper function for creating bool pointers in tests.
func boolPtr(b bool) *bool {
	return &b
}

// Helper function for creating string pointers in tests.
func stringPtr(s string) *string {
	return &s
}

// Helper function for tests.
func intPtr(i int) *int {
	return &i
}
