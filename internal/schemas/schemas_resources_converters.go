// Copyright (c) CyberArk.
// SPDX-License-Identifier: Apache-2.0

package schemas

import (
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var intTypes = []reflect.Kind{
	reflect.Int,
	reflect.Int8,
	reflect.Int16,
	reflect.Int32,
	reflect.Int64,
	reflect.Uint,
	reflect.Uint8,
	reflect.Uint16,
	reflect.Uint32,
	reflect.Uint64,
	reflect.Float32,
	reflect.Float64,
}

var simpleTypes = []reflect.Kind{
	reflect.String,
	reflect.Bool,
	reflect.Int,
	reflect.Int8,
	reflect.Int16,
	reflect.Int32,
	reflect.Int64,
	reflect.Uint,
	reflect.Uint8,
	reflect.Uint16,
	reflect.Uint32,
	reflect.Uint64,
	reflect.Float32,
	reflect.Float64,
}

func hasInterfaceInnerType(fieldType reflect.Type) bool {
	if fieldType.Kind() == reflect.Interface {
		return true
	}
	if fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Array ||
		fieldType.Kind() == reflect.Map {
		if fieldType.Elem().Kind() == reflect.Interface {
			return true
		}
		if fieldType.Elem().Kind() == reflect.Struct || fieldType.Elem().Kind() == reflect.Map {
			return hasInterfaceInnerType(fieldType.Elem())
		}
	}
	if fieldType.Kind() == reflect.Struct {
		actualFields := resolveFieldsSquashed(fieldType)
		for i := range actualFields {
			hasType := hasInterfaceInnerType(actualFields[i].Type)
			if hasType {
				return true
			}
		}
	}
	return false
}

func resourceSchemaAttrsFromStruct(inputModel interface{}, setAsComputed bool, sensitiveAttrs []string, extraRequiredAttrs []string, computedAsSetAttrs []string, immutableAttrs []string, forceNewAttrs []string, computedAttrs []string) map[string]schema.Attribute {
	modelType := reflect.TypeOf(inputModel)
	if modelType.Kind() == reflect.Pointer {
		modelType = modelType.Elem()
	}
	attributes := map[string]schema.Attribute{}
	actualFields := resolveFieldsSquashed(modelType)
	for i := range actualFields {
		field := actualFields[i]
		fieldType := field.Type
		desc := field.Tag.Get("desc")
		required := field.Tag.Get("required")
		validate := field.Tag.Get("validate")
		choices := field.Tag.Get("choices")
		defaultValue := field.Tag.Get("default")
		fieldName := resolveFieldName(field)
		isRequired := strings.Contains(required, "true") || strings.Contains(validate, "required") || slices.Contains(extraRequiredAttrs, fieldName)
		isSensitive := slices.Contains(sensitiveAttrs, fieldName)
		isImmutable := slices.Contains(immutableAttrs, fieldName)
		isForceNew := slices.Contains(forceNewAttrs, fieldName)
		isComputedOnly := slices.Contains(computedAttrs, fieldName)
		if fieldType.Kind() == reflect.Pointer {
			fieldType = fieldType.Elem()
		}
		switch fieldType.Kind() {
		case reflect.String:
			if setAsComputed || isComputedOnly {
				strAttr := schema.StringAttribute{
					Description: desc,
					Optional:    !isComputedOnly,
					Computed:    true,
					Sensitive:   isSensitive,
				}
				attributes[fieldName] = strAttr
				continue
			}
			strAttr := schema.StringAttribute{
				Description: desc,
				Optional:    !isRequired,
				Required:    isRequired,
				Computed:    !isRequired || isComputedOnly,
				Sensitive:   isSensitive,
			}
			if isComputedOnly {
				strAttr.Optional = false
				strAttr.Required = false
				strAttr.Computed = true
			}
			if defaultValue != "" {
				strAttr.Default = StringDefault{Value: defaultValue}
				strAttr.Required = false
				strAttr.Optional = true
				strAttr.Computed = true
			}
			if choices != "" {
				strAttr.Validators = append(strAttr.Validators, StringInChoicesValidator{Choices: strings.Split(choices, ",")})
			}
			if isImmutable {
				strAttr.PlanModifiers = []planmodifier.String{
					ImmutableString(),
				}
			} else if isForceNew {
				strAttr.PlanModifiers = []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				}
			}
			attributes[fieldName] = strAttr
		case reflect.Bool:
			if setAsComputed || isComputedOnly {
				boolAttr := schema.BoolAttribute{
					Description: desc,
					Optional:    !isComputedOnly,
					Computed:    true,
					Sensitive:   isSensitive,
				}
				attributes[fieldName] = boolAttr
				continue
			}
			boolAttr := schema.BoolAttribute{
				Description: desc,
				Optional:    !isRequired,
				Required:    isRequired,
				Computed:    !isRequired || isComputedOnly,
				Sensitive:   isSensitive,
			}
			if isComputedOnly {
				boolAttr.Optional = false
				boolAttr.Required = false
				boolAttr.Computed = true
			}
			if defaultValue != "" {
				boolValue, _ := strconv.ParseBool(defaultValue)
				boolAttr.Default = BoolDefault{Value: boolValue}
				boolAttr.Required = false
				boolAttr.Optional = true
				boolAttr.Computed = true
			}
			if isImmutable {
				boolAttr.PlanModifiers = []planmodifier.Bool{
					ImmutableBool(),
				}
			} else if isForceNew {
				boolAttr.PlanModifiers = []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				}
			}
			attributes[fieldName] = boolAttr
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if setAsComputed || isComputedOnly {
				intAttr := schema.Int64Attribute{
					Description: desc,
					Optional:    !isComputedOnly,
					Computed:    true,
					Sensitive:   isSensitive,
				}
				attributes[fieldName] = intAttr
				continue
			}
			int64Attr := schema.Int64Attribute{
				Description: desc,
				Optional:    !isRequired,
				Required:    isRequired,
				Computed:    !isRequired || isComputedOnly,
				Sensitive:   isSensitive,
			}
			if isComputedOnly {
				int64Attr.Optional = false
				int64Attr.Required = false
				int64Attr.Computed = true
			}
			if defaultValue != "" {
				intValue, _ := strconv.ParseInt(defaultValue, 10, 64)
				int64Attr.Default = Int64Default{Value: intValue}
				int64Attr.Required = false
				int64Attr.Optional = true
				int64Attr.Computed = true
			}
			if isImmutable {
				int64Attr.PlanModifiers = []planmodifier.Int64{
					ImmutableInt64(),
				}
			} else if isForceNew {
				int64Attr.PlanModifiers = []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				}
			}
			attributes[fieldName] = int64Attr
		case reflect.Slice, reflect.Array:
			// Inner dynamic types are not supported in terraform
			if hasInterfaceInnerType(fieldType) {
				if setAsComputed {
					attributes[fieldName] = schema.DynamicAttribute{
						Description: desc,
						Optional:    true,
						Computed:    true,
						Sensitive:   isSensitive,
					}
					continue
				}
				attributes[fieldName] = schema.DynamicAttribute{
					Description: desc,
					Optional:    !isRequired,
					Required:    isRequired,
					Computed:    !isRequired,
					Sensitive:   isSensitive,
				}
				continue
			}
			if slices.Contains(simpleTypes, fieldType.Elem().Kind()) {
				terraType, err := reflectTypeToTerraformType(fieldType.Elem())
				if err != nil {
					continue
				}
				if slices.Contains(computedAsSetAttrs, fieldName) {
					if setAsComputed || isComputedOnly {
						sliceAttr := schema.SetAttribute{
							ElementType: terraType,
							Description: desc,
							Optional:    !isComputedOnly,
							Computed:    true,
							Sensitive:   isSensitive,
						}
						attributes[fieldName] = sliceAttr
						continue
					}
					sliceAttr := schema.SetAttribute{
						ElementType: terraType,
						Description: desc,
						Optional:    !isRequired,
						Required:    isRequired,
						Computed:    !isRequired || isComputedOnly,
						Sensitive:   isSensitive,
					}
					if isComputedOnly {
						sliceAttr.Optional = false
						sliceAttr.Required = false
						sliceAttr.Computed = true
					}
					if defaultValue != "" {
						if fieldType.Elem().Kind() == reflect.String {
							sliceAttr.Default = SetStringDefault{Values: strings.Split(defaultValue, ",")}
						} else if slices.Contains(intTypes, fieldType.Elem().Kind()) {
							intValues := strings.Split(defaultValue, ",")
							int64Values := make([]int64, 0)
							for _, v := range intValues {
								if v != "" {
									int64Value, err := strconv.ParseInt(v, 10, 64)
									if err == nil {
										int64Values = append(int64Values, int64Value)
									}
								}
							}
							sliceAttr.Default = SetNumericDefault{Values: int64Values}
						} else if fieldType.Elem().Kind() == reflect.Bool {
							boolValues := strings.Split(defaultValue, ",")
							boolSlice := make([]bool, 0)
							for _, v := range boolValues {
								if v != "" {
									boolValue, err := strconv.ParseBool(v)
									if err == nil {
										boolSlice = append(boolSlice, boolValue)
									}
								}
							}
							sliceAttr.Default = SetBoolDefault{Values: boolSlice}
						}
						sliceAttr.Required = false
						sliceAttr.Optional = true
						sliceAttr.Computed = true
					}
					if choices != "" {
						sliceAttr.Validators = append(sliceAttr.Validators, SliceInSetValidator{Choices: strings.Split(choices, ",")})
					}
					if isImmutable {
						sliceAttr.PlanModifiers = []planmodifier.Set{
							ImmutableSet(),
						}
					} else if isForceNew {
						sliceAttr.PlanModifiers = []planmodifier.Set{
							setplanmodifier.RequiresReplace(),
						}
					}
					attributes[fieldName] = sliceAttr
				} else {
					if setAsComputed || isComputedOnly {
						sliceAttr := schema.ListAttribute{
							ElementType: terraType,
							Description: desc,
							Optional:    !isComputedOnly,
							Computed:    true,
							Sensitive:   isSensitive,
						}
						attributes[fieldName] = sliceAttr
						continue
					}
					sliceAttr := schema.ListAttribute{
						ElementType: terraType,
						Description: desc,
						Optional:    !isRequired,
						Required:    isRequired,
						Computed:    !isRequired || isComputedOnly,
						Sensitive:   isSensitive,
					}
					if isComputedOnly {
						sliceAttr.Optional = false
						sliceAttr.Required = false
						sliceAttr.Computed = true
					}
					if defaultValue != "" {
						if fieldType.Elem().Kind() == reflect.String {
							sliceAttr.Default = ListStringDefault{Values: strings.Split(defaultValue, ",")}
						} else if slices.Contains(intTypes, fieldType.Elem().Kind()) {
							intValues := strings.Split(defaultValue, ",")
							int64Values := make([]int64, 0)
							for _, v := range intValues {
								if v != "" {
									int64Value, err := strconv.ParseInt(v, 10, 64)
									if err == nil {
										int64Values = append(int64Values, int64Value)
									}
								}
							}
							sliceAttr.Default = ListNumericDefault{Values: int64Values}
						} else if fieldType.Elem().Kind() == reflect.Bool {
							boolValues := strings.Split(defaultValue, ",")
							boolSlice := make([]bool, 0)
							for _, v := range boolValues {
								if v != "" {
									boolValue, err := strconv.ParseBool(v)
									if err == nil {
										boolSlice = append(boolSlice, boolValue)
									}
								}
							}
							sliceAttr.Default = ListBoolDefault{Values: boolSlice}
						}
						sliceAttr.Required = false
						sliceAttr.Optional = true
						sliceAttr.Computed = true
					}
					if choices != "" {
						sliceAttr.Validators = append(sliceAttr.Validators, SliceInChoicesValidator{Choices: strings.Split(choices, ",")})
					}
					if isImmutable {
						sliceAttr.PlanModifiers = []planmodifier.List{
							ImmutableList(),
						}
					} else if isForceNew {
						sliceAttr.PlanModifiers = []planmodifier.List{
							listplanmodifier.RequiresReplace(),
						}
					}
					attributes[fieldName] = sliceAttr
				}
			}
			if fieldType.Elem().Kind() == reflect.Map {
				if fieldType.Elem().Key().Kind() == reflect.String {
					mapElementType := types.MapType{ElemType: types.StringType}
					if setAsComputed {
						sliceAttr := schema.ListAttribute{
							ElementType: mapElementType,
							Description: desc,
							Optional:    true,
							Computed:    true,
							Sensitive:   isSensitive,
						}
						attributes[fieldName] = sliceAttr
						continue
					}
					sliceAttr := schema.ListAttribute{
						ElementType: mapElementType,
						Description: desc,
						Optional:    !isRequired,
						Required:    isRequired,
						Computed:    !isRequired,
						Sensitive:   isSensitive,
					}
					if isImmutable {
						sliceAttr.PlanModifiers = []planmodifier.List{
							ImmutableList(),
						}
					} else if isForceNew {
						sliceAttr.PlanModifiers = []planmodifier.List{
							listplanmodifier.RequiresReplace(),
						}
					}
					attributes[fieldName] = sliceAttr
				}
			}
			if fieldType.Elem().Kind() == reflect.Struct {
				// Handle nested structs by recursively generating their schema
				nestedSchemaAttrs := resourceSchemaAttrsFromStruct(reflect.New(fieldType.Elem()).Elem().Interface(), setAsComputed, sensitiveAttrs, extraRequiredAttrs, computedAsSetAttrs, immutableAttrs, forceNewAttrs, computedAttrs)
				if setAsComputed {
					attributes[fieldName] = schema.ListNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: nestedSchemaAttrs,
						},
						Description: desc,
						Optional:    true,
						Computed:    true,
						Sensitive:   isSensitive,
					}
					continue
				}
				attributes[fieldName] = schema.ListNestedAttribute{
					NestedObject: schema.NestedAttributeObject{
						Attributes: nestedSchemaAttrs,
					},
					Description: desc,
					Optional:    !isRequired,
					Required:    isRequired,
					Computed:    !isRequired,
					Sensitive:   isSensitive,
				}
			}
		case reflect.Map:
			// Inner dynamic types are not supported in terraform
			if hasInterfaceInnerType(fieldType) {
				if setAsComputed {
					attributes[fieldName] = schema.DynamicAttribute{
						Description: desc,
						Optional:    true,
						Computed:    true,
						Sensitive:   isSensitive,
					}
					continue
				}
				attributes[fieldName] = schema.DynamicAttribute{
					Description: desc,
					Optional:    !isRequired,
					Required:    isRequired,
					Computed:    !isRequired,
					Sensitive:   isSensitive,
				}
				continue
			}
			if slices.Contains(simpleTypes, fieldType.Elem().Kind()) {
				terraType, err := reflectTypeToTerraformType(fieldType.Elem())
				if err != nil {
					continue
				}
				if setAsComputed {
					strAttr := schema.MapAttribute{
						ElementType: terraType,
						Description: desc,
						Optional:    true,
						Computed:    true,
						Sensitive:   isSensitive,
					}
					attributes[fieldName] = strAttr
					continue
				}
				mapAttr := schema.MapAttribute{
					ElementType: terraType,
					Description: desc,
					Optional:    !isRequired,
					Required:    isRequired,
					Computed:    !isRequired,
					Sensitive:   isSensitive,
				}
				if isImmutable {
					mapAttr.PlanModifiers = []planmodifier.Map{
						ImmutableMap(),
					}
				} else if isForceNew {
					mapAttr.PlanModifiers = []planmodifier.Map{
						mapplanmodifier.RequiresReplace(),
					}
				}
				attributes[fieldName] = mapAttr
			} else if fieldType.Elem().Kind() == reflect.Interface {
				if setAsComputed {
					attributes[fieldName] = schema.DynamicAttribute{
						Description: desc,
						Optional:    true,
						Computed:    true,
						Sensitive:   isSensitive,
					}
					continue
				}
				attributes[fieldName] = schema.DynamicAttribute{
					Description: desc,
					Optional:    !isRequired,
					Required:    isRequired,
					Computed:    !isRequired,
					Sensitive:   isSensitive,
				}
			} else if fieldType.Elem().Kind() == reflect.Struct {
				nestedAttrs := resourceSchemaAttrsFromStruct(reflect.New(fieldType.Elem()).Elem().Interface(), setAsComputed, sensitiveAttrs, extraRequiredAttrs, computedAsSetAttrs, immutableAttrs, forceNewAttrs, computedAttrs)
				if setAsComputed {
					complexMapAttr := schema.MapNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: nestedAttrs,
						},
						Description: desc,
						Optional:    true,
						Computed:    true,
						Sensitive:   isSensitive,
					}
					attributes[fieldName] = complexMapAttr
					continue
				}
				complexMapAttr := schema.MapNestedAttribute{
					NestedObject: schema.NestedAttributeObject{
						Attributes: nestedAttrs,
					},
					Description: desc,
					Optional:    !isRequired,
					Required:    isRequired,
					Computed:    !isRequired,
					Sensitive:   isSensitive,
				}
				attributes[fieldName] = complexMapAttr
			}
		case reflect.Struct:
			// Handle nested structs by recursively generating their schema
			nestedSchemaAttrs := resourceSchemaAttrsFromStruct(reflect.New(fieldType).Elem().Interface(), setAsComputed, sensitiveAttrs, extraRequiredAttrs, computedAsSetAttrs, immutableAttrs, forceNewAttrs, computedAttrs)
			if setAsComputed || isComputedOnly {
				attributes[fieldName] = schema.SingleNestedAttribute{
					Attributes:  nestedSchemaAttrs,
					Description: desc,
					Optional:    !isComputedOnly,
					Computed:    true,
					Sensitive:   isSensitive,
				}
				continue
			}
			attributes[fieldName] = schema.SingleNestedAttribute{
				Attributes:  nestedSchemaAttrs,
				Description: desc,
				Optional:    !isRequired,
				Required:    isRequired,
				Computed:    !isRequired || isComputedOnly,
				Sensitive:   isSensitive,
			}
			if isComputedOnly {
				if attr, ok := attributes[fieldName].(schema.SingleNestedAttribute); ok {
					attr.Optional = false
					attr.Required = false
					attr.Computed = true
					attributes[fieldName] = attr
				}
			}
		case reflect.Interface:
			if setAsComputed {
				attributes[fieldName] = schema.DynamicAttribute{
					Description: desc,
					Optional:    true,
					Computed:    true,
					Sensitive:   isSensitive,
				}
				continue
			}
			attributes[fieldName] = schema.DynamicAttribute{
				Description: desc,
				Optional:    !isRequired,
				Required:    isRequired,
				Computed:    !isRequired,
				Sensitive:   isSensitive,
			}
		default:
			continue
		}
	}
	return attributes
}

// forceComputedAttributesReadOnly recursively marks computed-only attributes as read-only
// (Optional=false, Required=false, Computed=true) in both top-level and nested attributes.
// Supports dot-notation paths like "secret_management.last_modified_time" for nested attributes.
func forceComputedAttributesReadOnly(attributes map[string]schema.Attribute, computedAttrs []string) {
	for _, computedAttrPath := range computedAttrs {
		// Check if this is a path (contains a dot)
		if strings.Contains(computedAttrPath, ".") {
			// Handle path-based attribute (e.g., "secret_management.last_modified_time")
			pathParts := strings.SplitN(computedAttrPath, ".", 2)
			nestedAttrName := pathParts[0]
			remainingPath := pathParts[1]

			if attr, exists := attributes[nestedAttrName]; exists {
				// Navigate to the nested attribute
				switch a := attr.(type) {
				case schema.SingleNestedAttribute:
					if a.Attributes != nil {
						// Recursively process with the remaining path
						forceComputedAttributesReadOnly(a.Attributes, []string{remainingPath})
						attributes[nestedAttrName] = a
					}
				case schema.ListNestedAttribute:
					if a.NestedObject.Attributes != nil {
						// Recursively process with the remaining path
						forceComputedAttributesReadOnly(a.NestedObject.Attributes, []string{remainingPath})
						attributes[nestedAttrName] = a
					}
				case schema.MapNestedAttribute:
					if a.NestedObject.Attributes != nil {
						// Recursively process with the remaining path
						forceComputedAttributesReadOnly(a.NestedObject.Attributes, []string{remainingPath})
						attributes[nestedAttrName] = a
					}
				}
			}
			continue
		}

		// Handle simple attribute name (no path)
		if attr, exists := attributes[computedAttrPath]; exists {
			// Use type assertion to update the attribute
			switch a := attr.(type) {
			case schema.StringAttribute:
				a.Optional = false
				a.Required = false
				a.Computed = true
				attributes[computedAttrPath] = a
			case schema.BoolAttribute:
				a.Optional = false
				a.Required = false
				a.Computed = true
				attributes[computedAttrPath] = a
			case schema.Int64Attribute:
				a.Optional = false
				a.Required = false
				a.Computed = true
				attributes[computedAttrPath] = a
			case schema.ListAttribute:
				a.Optional = false
				a.Required = false
				a.Computed = true
				attributes[computedAttrPath] = a
			case schema.SetAttribute:
				a.Optional = false
				a.Required = false
				a.Computed = true
				attributes[computedAttrPath] = a
			case schema.MapAttribute:
				a.Optional = false
				a.Required = false
				a.Computed = true
				attributes[computedAttrPath] = a
			case schema.DynamicAttribute:
				a.Optional = false
				a.Required = false
				a.Computed = true
				attributes[computedAttrPath] = a
			case schema.SingleNestedAttribute:
				// Recursively process nested attributes
				if a.Attributes != nil {
					forceComputedAttributesReadOnly(a.Attributes, computedAttrs)
				}
				a.Optional = false
				a.Required = false
				a.Computed = true
				attributes[computedAttrPath] = a
			case schema.ListNestedAttribute:
				// Recursively process nested attributes
				if a.NestedObject.Attributes != nil {
					forceComputedAttributesReadOnly(a.NestedObject.Attributes, computedAttrs)
				}
				a.Optional = false
				a.Required = false
				a.Computed = true
				attributes[computedAttrPath] = a
			case schema.MapNestedAttribute:
				// Recursively process nested attributes
				if a.NestedObject.Attributes != nil {
					forceComputedAttributesReadOnly(a.NestedObject.Attributes, computedAttrs)
				}
				a.Optional = false
				a.Required = false
				a.Computed = true
				attributes[computedAttrPath] = a
			}
		}
	}

	// Also recursively process all nested attributes to find computed-only fields
	for key, attr := range attributes {
		switch a := attr.(type) {
		case schema.SingleNestedAttribute:
			if a.Attributes != nil {
				forceComputedAttributesReadOnly(a.Attributes, computedAttrs)
				// The map is modified in place, reassign to ensure the attribute is updated
				attributes[key] = a
			}
		case schema.ListNestedAttribute:
			if a.NestedObject.Attributes != nil {
				forceComputedAttributesReadOnly(a.NestedObject.Attributes, computedAttrs)
				// The map is modified in place, but we need to reassign to update the attribute
				attributes[key] = a
			}
		case schema.MapNestedAttribute:
			if a.NestedObject.Attributes != nil {
				forceComputedAttributesReadOnly(a.NestedObject.Attributes, computedAttrs)
				// The map is modified in place, but we need to reassign to update the attribute
				attributes[key] = a
			}
		}
	}
}

// getNestedStructFieldNames collects all field names that belong to nested structs in the state model.
// This is used to identify flattened fields from create/update schemas that should be excluded.
// Returns a set of field names that are part of nested structs (not squashed).
func getNestedStructFieldNames(stateModel interface{}) map[string]bool {
	fieldNames := make(map[string]bool)
	if stateModel == nil {
		return fieldNames
	}
	modelType := reflect.TypeOf(stateModel)
	if modelType.Kind() == reflect.Pointer {
		modelType = modelType.Elem()
	}
	if modelType.Kind() != reflect.Struct {
		return fieldNames
	}
	// Iterate through the original struct fields to identify nested structs (not squashed)
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		if field.PkgPath != "" { // unexported field
			continue
		}
		mapstructureTag := field.Tag.Get("mapstructure")
		if mapstructureTag == "-" { // skip ignored fields
			continue
		}
		// Skip squashed fields - they're already flattened
		if mapstructureTag == ",squash" {
			continue
		}
		fieldType := field.Type
		if fieldType.Kind() == reflect.Pointer {
			fieldType = fieldType.Elem()
		}
		// Check if this is a nested struct field (not squashed)
		if fieldType.Kind() == reflect.Struct {
			// Get all fields from the nested struct
			nestedFields := resolveFieldsSquashed(fieldType)
			for j := range nestedFields {
				nestedFieldName := resolveFieldName(nestedFields[j])
				// Mark this field as belonging to a nested struct
				fieldNames[nestedFieldName] = true
			}
		}
	}
	return fieldNames
}

// GenerateResourceSchemaFromStruct generates a Terraform schema from a Go struct.
func GenerateResourceSchemaFromStruct(createModel interface{}, updateModel interface{}, stateModel interface{}, sensitiveAttrs []string, extraRequiredAttrs []string, computedAsSetAttrs []string, immutableAttrs []string, forceNewAttrs []string, computedAttrs []string) schema.Schema {
	schemaAttrs := resourceSchemaAttrsFromStruct(createModel, false, sensitiveAttrs, extraRequiredAttrs, computedAsSetAttrs, immutableAttrs, forceNewAttrs, computedAttrs)

	// Get field names that belong to nested structs in the state model
	// These should not appear as flattened fields in the final schema
	nestedStructFieldNames := getNestedStructFieldNames(stateModel)

	if updateModel != nil {
		updateModelAttrs := resourceSchemaAttrsFromStruct(updateModel, true, sensitiveAttrs, extraRequiredAttrs, computedAsSetAttrs, immutableAttrs, forceNewAttrs, computedAttrs)
		for key, updateAttr := range updateModelAttrs {
			// Skip flattened fields that belong to nested structs in the state model
			if nestedStructFieldNames[key] {
				continue
			}
			if _, exists := schemaAttrs[key]; !exists {
				schemaAttrs[key] = updateAttr
			}
		}
	}

	// Remove any flattened fields from create schema that belong to nested structs
	for key := range schemaAttrs {
		if nestedStructFieldNames[key] {
			delete(schemaAttrs, key)
		}
	}

	if stateModel != nil {
		outputModelAttrs := resourceSchemaAttrsFromStruct(stateModel, true, sensitiveAttrs, extraRequiredAttrs, computedAsSetAttrs, immutableAttrs, forceNewAttrs, computedAttrs)
		for key, outputAttr := range outputModelAttrs {
			if _, exists := schemaAttrs[key]; !exists {
				schemaAttrs[key] = outputAttr
			}
		}
	}

	// Force computed-only attributes to be read-only (Optional=false, Required=false, Computed=true)
	// This processes both top-level and nested attributes recursively
	forceComputedAttributesReadOnly(schemaAttrs, computedAttrs)

	return schema.Schema{
		Attributes: schemaAttrs,
	}
}

// ResourceSchemaToSchemaAttrTypes converts a Terraform schema to a map of attribute types.
func ResourceSchemaToSchemaAttrTypes(schemaInput schema.Schema) map[string]attr.Type {
	attributes := make(map[string]attr.Type)
	for key, schemaAttr := range schemaInput.Attributes {
		attributes[key] = schemaAttr.GetType()
	}
	return attributes
}
