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

func resourceSchemaAttrsFromStruct(inputModel interface{}, setAsComputed bool, sensitiveAttrs []string, extraRequiredAttrs []string, computedAsSetAttrs []string) map[string]schema.Attribute {
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
		forceNew := field.Tag.Get("forcenew")
		fieldName := resolveFieldName(field)
		isRequired := strings.Contains(required, "true") || strings.Contains(validate, "required") || slices.Contains(extraRequiredAttrs, fieldName)
		isSensitive := slices.Contains(sensitiveAttrs, fieldName)
		isForceNew := strings.Contains(forceNew, "true")
		if fieldType.Kind() == reflect.Pointer {
			fieldType = fieldType.Elem()
		}
		switch fieldType.Kind() {
		case reflect.String:
			if setAsComputed {
				strAttr := schema.StringAttribute{
					Description: desc,
					Optional:    true,
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
				Computed:    !isRequired,
				Sensitive:   isSensitive,
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
			if isForceNew {
				strAttr.PlanModifiers = []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				}
			}
			attributes[fieldName] = strAttr
		case reflect.Bool:
			if setAsComputed {
				boolAttr := schema.BoolAttribute{
					Description: desc,
					Optional:    true,
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
				Computed:    !isRequired,
				Sensitive:   isSensitive,
			}
			if defaultValue != "" {
				boolValue, _ := strconv.ParseBool(defaultValue)
				boolAttr.Default = BoolDefault{Value: boolValue}
				boolAttr.Required = false
				boolAttr.Optional = true
				boolAttr.Computed = true
			}
			if isForceNew {
				boolAttr.PlanModifiers = []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				}
			}
			attributes[fieldName] = boolAttr
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if setAsComputed {
				intAttr := schema.Int64Attribute{
					Description: desc,
					Optional:    true,
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
				Computed:    !isRequired,
				Sensitive:   isSensitive,
			}
			if defaultValue != "" {
				intValue, _ := strconv.ParseInt(defaultValue, 10, 64)
				int64Attr.Default = Int64Default{Value: intValue}
				int64Attr.Required = false
				int64Attr.Optional = true
				int64Attr.Computed = true
			}
			if isForceNew {
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
					if setAsComputed {
						sliceAttr := schema.SetAttribute{
							ElementType: terraType,
							Description: desc,
							Optional:    true,
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
						Computed:    !isRequired,
						Sensitive:   isSensitive,
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
					if isForceNew {
						sliceAttr.PlanModifiers = []planmodifier.Set{
							setplanmodifier.RequiresReplace(),
						}
					}
					attributes[fieldName] = sliceAttr
				} else {
					if setAsComputed {
						sliceAttr := schema.ListAttribute{
							ElementType: terraType,
							Description: desc,
							Optional:    true,
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
						Computed:    !isRequired,
						Sensitive:   isSensitive,
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
					if isForceNew {
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
					if isForceNew {
						sliceAttr.PlanModifiers = []planmodifier.List{
							listplanmodifier.RequiresReplace(),
						}
					}
					attributes[fieldName] = sliceAttr
				}
			}
			if fieldType.Elem().Kind() == reflect.Struct {
				// Handle nested structs by recursively generating their schema
				nestedSchemaAttrs := resourceSchemaAttrsFromStruct(reflect.New(fieldType.Elem()).Elem().Interface(), setAsComputed, sensitiveAttrs, extraRequiredAttrs, computedAsSetAttrs)
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
				if isForceNew {
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
				nestedAttrs := resourceSchemaAttrsFromStruct(reflect.New(fieldType.Elem()).Elem().Interface(), setAsComputed, sensitiveAttrs, extraRequiredAttrs, computedAsSetAttrs)
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
			nestedSchemaAttrs := resourceSchemaAttrsFromStruct(reflect.New(fieldType).Elem().Interface(), setAsComputed, sensitiveAttrs, extraRequiredAttrs, computedAsSetAttrs)
			if setAsComputed {
				attributes[fieldName] = schema.SingleNestedAttribute{
					Attributes:  nestedSchemaAttrs,
					Description: desc,
					Optional:    true,
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
				Computed:    !isRequired,
				Sensitive:   isSensitive,
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

// GenerateResourceSchemaFromStruct generates a Terraform schema from a Go struct.
func GenerateResourceSchemaFromStruct(createModel interface{}, updateModel interface{}, stateModel interface{}, sensitiveAttrs []string, extraRequiredAttrs []string, computedAsSetAttrs []string) schema.Schema {
	schemaAttrs := resourceSchemaAttrsFromStruct(createModel, false, sensitiveAttrs, extraRequiredAttrs, computedAsSetAttrs)
	if updateModel != nil {
		updateModelAttrs := resourceSchemaAttrsFromStruct(updateModel, true, sensitiveAttrs, extraRequiredAttrs, computedAsSetAttrs)
		for key, updateAttr := range updateModelAttrs {
			if _, exists := schemaAttrs[key]; !exists {
				schemaAttrs[key] = updateAttr
			}
		}
	}
	if stateModel != nil {
		outputModelAttrs := resourceSchemaAttrsFromStruct(stateModel, true, sensitiveAttrs, extraRequiredAttrs, computedAsSetAttrs)
		for key, outputAttr := range outputModelAttrs {
			if _, exists := schemaAttrs[key]; !exists {
				schemaAttrs[key] = outputAttr
			}
		}
	}
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
