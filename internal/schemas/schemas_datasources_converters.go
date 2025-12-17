// Copyright (c) CyberArk.
// SPDX-License-Identifier: Apache-2.0

package schemas

import (
	"reflect"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func dataSourceSchemaAttrsFromStruct(inputModel interface{}, setAsComputed bool, sensitiveAttrs []string, extraRequiredAttrs []string, computedAsSetAttrs []string) map[string]schema.Attribute {
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
		fieldName := resolveFieldName(field)
		isRequired := strings.Contains(required, "true") || strings.Contains(validate, "required") || slices.Contains(extraRequiredAttrs, fieldName)
		isSensitive := slices.Contains(sensitiveAttrs, fieldName)
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
			if choices != "" {
				strAttr.Validators = append(strAttr.Validators, StringInChoicesValidator{Choices: strings.Split(choices, ",")})
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
					sliceAttr := schema.SetAttribute{
						ElementType: terraType,
						Description: desc,
						Optional:    !isRequired,
						Required:    isRequired,
						Computed:    !isRequired,
						Sensitive:   isSensitive,
					}
					if choices != "" {
						sliceAttr.Validators = append(sliceAttr.Validators, SliceInSetValidator{Choices: strings.Split(choices, ",")})
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
					if choices != "" {
						sliceAttr.Validators = append(sliceAttr.Validators, SliceInChoicesValidator{Choices: strings.Split(choices, ",")})
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
					attributes[fieldName] = sliceAttr
				}
			}
			if fieldType.Elem().Kind() == reflect.Struct {
				// Handle nested structs by recursively generating their schema
				nestedSchemaAttrs := dataSourceSchemaAttrsFromStruct(reflect.New(fieldType.Elem()).Elem().Interface(), setAsComputed, sensitiveAttrs, extraRequiredAttrs, computedAsSetAttrs)
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
				nestedAttrs := dataSourceSchemaAttrsFromStruct(reflect.New(fieldType.Elem()).Elem().Interface(), setAsComputed, sensitiveAttrs, extraRequiredAttrs, computedAsSetAttrs)
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
			nestedSchemaAttrs := dataSourceSchemaAttrsFromStruct(reflect.New(fieldType).Elem().Interface(), setAsComputed, sensitiveAttrs, extraRequiredAttrs, computedAsSetAttrs)
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

// GenerateDataSourceSchemaFromStruct generates a Terraform schema from a Go struct.
func GenerateDataSourceSchemaFromStruct(inputModel interface{}, stateModel interface{}, sensitiveAttrs []string, extraRequiredAttrs []string, computedAsSetAttrs []string) schema.Schema {
	inputModelAttrs := make(map[string]schema.Attribute)
	if inputModel != nil {
		inputModelAttrs = dataSourceSchemaAttrsFromStruct(inputModel, false, sensitiveAttrs, extraRequiredAttrs, computedAsSetAttrs)
	}
	outputModelAttrs := dataSourceSchemaAttrsFromStruct(stateModel, true, sensitiveAttrs, extraRequiredAttrs, computedAsSetAttrs)
	for key, updateAttr := range outputModelAttrs {
		if _, exists := inputModelAttrs[key]; !exists {
			inputModelAttrs[key] = updateAttr
		}
	}
	return schema.Schema{
		Attributes: inputModelAttrs,
	}
}

// DataSourceSchemaToSchemaAttrTypes converts a Terraform schema to a map of attribute types.
func DataSourceSchemaToSchemaAttrTypes(schemaInput schema.Schema) map[string]attr.Type {
	attributes := make(map[string]attr.Type)
	for key, schemaAttr := range schemaInput.Attributes {
		attributes[key] = schemaAttr.GetType()
	}
	return attributes
}
