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

// forceComputedAttributesReadOnlyDataSource recursively marks computed-only attributes as read-only
// (Optional=false, Required=false, Computed=true) in both top-level and nested attributes.
// Supports dot-notation paths like "secret_management.last_modified_time" for nested attributes.
func forceComputedAttributesReadOnlyDataSource(attributes map[string]schema.Attribute, computedAttrs []string) {
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
						forceComputedAttributesReadOnlyDataSource(a.Attributes, []string{remainingPath})
						attributes[nestedAttrName] = a
					}
				case schema.ListNestedAttribute:
					if a.NestedObject.Attributes != nil {
						// Recursively process with the remaining path
						forceComputedAttributesReadOnlyDataSource(a.NestedObject.Attributes, []string{remainingPath})
						attributes[nestedAttrName] = a
					}
				case schema.MapNestedAttribute:
					if a.NestedObject.Attributes != nil {
						// Recursively process with the remaining path
						forceComputedAttributesReadOnlyDataSource(a.NestedObject.Attributes, []string{remainingPath})
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
				if a.Attributes != nil {
					forceComputedAttributesReadOnlyDataSource(a.Attributes, computedAttrs)
				}
				a.Optional = false
				a.Required = false
				a.Computed = true
				attributes[computedAttrPath] = a
			case schema.ListNestedAttribute:
				if a.NestedObject.Attributes != nil {
					forceComputedAttributesReadOnlyDataSource(a.NestedObject.Attributes, computedAttrs)
				}
				a.Optional = false
				a.Required = false
				a.Computed = true
				attributes[computedAttrPath] = a
			case schema.MapNestedAttribute:
				if a.NestedObject.Attributes != nil {
					forceComputedAttributesReadOnlyDataSource(a.NestedObject.Attributes, computedAttrs)
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
				forceComputedAttributesReadOnlyDataSource(a.Attributes, computedAttrs)
				// The map is modified in place, reassign to ensure the attribute is updated
				attributes[key] = a
			}
		case schema.ListNestedAttribute:
			if a.NestedObject.Attributes != nil {
				forceComputedAttributesReadOnlyDataSource(a.NestedObject.Attributes, computedAttrs)
				// The map is modified in place, but we need to reassign to update the attribute
				attributes[key] = a
			}
		case schema.MapNestedAttribute:
			if a.NestedObject.Attributes != nil {
				forceComputedAttributesReadOnlyDataSource(a.NestedObject.Attributes, computedAttrs)
				// The map is modified in place, but we need to reassign to update the attribute
				attributes[key] = a
			}
		}
	}
}

// collectAllNestedAttributePaths recursively collects all attribute paths within a nested attribute.
func collectAllNestedAttributePaths(attrs map[string]schema.Attribute, prefix string) []string {
	paths := make([]string, 0)

	for key, attr := range attrs {
		fullPath := key
		if prefix != "" {
			fullPath = prefix + "." + key
		}
		paths = append(paths, fullPath)

		// Recursively collect nested attribute paths
		switch a := attr.(type) {
		case schema.SingleNestedAttribute:
			if a.Attributes != nil {
				nestedPaths := collectAllNestedAttributePaths(a.Attributes, fullPath)
				paths = append(paths, nestedPaths...)
			}
		case schema.ListNestedAttribute:
			if a.NestedObject.Attributes != nil {
				nestedPaths := collectAllNestedAttributePaths(a.NestedObject.Attributes, fullPath)
				paths = append(paths, nestedPaths...)
			}
		case schema.MapNestedAttribute:
			if a.NestedObject.Attributes != nil {
				nestedPaths := collectAllNestedAttributePaths(a.NestedObject.Attributes, fullPath)
				paths = append(paths, nestedPaths...)
			}
		}
	}

	return paths
}

// mergeNestedAttributesAndFindReadOnly recursively merges nested attributes from state model into input model
// and returns a list of attribute paths (using dot notation) that exist only in the state model.
func mergeNestedAttributesAndFindReadOnly(inputAttrs map[string]schema.Attribute, stateAttrs map[string]schema.Attribute, prefix string) []string {
	readOnlyAttrs := make([]string, 0)

	for key, stateAttr := range stateAttrs {
		fullPath := key
		if prefix != "" {
			fullPath = prefix + "." + key
		}

		inputAttr, existsInInput := inputAttrs[key]

		if !existsInInput {
			// This attribute only exists in state model, add it and mark as read-only
			inputAttrs[key] = stateAttr
			readOnlyAttrs = append(readOnlyAttrs, fullPath)

			// If it's a nested attribute, also collect all nested paths within it
			switch a := stateAttr.(type) {
			case schema.SingleNestedAttribute:
				if a.Attributes != nil {
					nestedPaths := collectAllNestedAttributePaths(a.Attributes, fullPath)
					readOnlyAttrs = append(readOnlyAttrs, nestedPaths...)
				}
			case schema.ListNestedAttribute:
				if a.NestedObject.Attributes != nil {
					nestedPaths := collectAllNestedAttributePaths(a.NestedObject.Attributes, fullPath)
					readOnlyAttrs = append(readOnlyAttrs, nestedPaths...)
				}
			case schema.MapNestedAttribute:
				if a.NestedObject.Attributes != nil {
					nestedPaths := collectAllNestedAttributePaths(a.NestedObject.Attributes, fullPath)
					readOnlyAttrs = append(readOnlyAttrs, nestedPaths...)
				}
			}
		} else {
			// Both exist, check if they're nested attributes and merge recursively
			switch stateNested := stateAttr.(type) {
			case schema.SingleNestedAttribute:
				if inputNested, ok := inputAttr.(schema.SingleNestedAttribute); ok {
					if stateNested.Attributes != nil {
						if inputNested.Attributes == nil {
							inputNested.Attributes = make(map[string]schema.Attribute)
						}
						// Recursively merge nested attributes
						nestedReadOnly := mergeNestedAttributesAndFindReadOnly(inputNested.Attributes, stateNested.Attributes, fullPath)
						readOnlyAttrs = append(readOnlyAttrs, nestedReadOnly...)
						// Update the input attribute with merged nested attributes
						inputAttrs[key] = inputNested
					}
				}
			case schema.ListNestedAttribute:
				if inputNested, ok := inputAttr.(schema.ListNestedAttribute); ok {
					if stateNested.NestedObject.Attributes != nil {
						if inputNested.NestedObject.Attributes == nil {
							inputNested.NestedObject.Attributes = make(map[string]schema.Attribute)
						}
						// Recursively merge nested attributes
						nestedReadOnly := mergeNestedAttributesAndFindReadOnly(inputNested.NestedObject.Attributes, stateNested.NestedObject.Attributes, fullPath)
						readOnlyAttrs = append(readOnlyAttrs, nestedReadOnly...)
						// Update the input attribute with merged nested attributes
						inputAttrs[key] = inputNested
					}
				}
			case schema.MapNestedAttribute:
				if inputNested, ok := inputAttr.(schema.MapNestedAttribute); ok {
					if stateNested.NestedObject.Attributes != nil {
						if inputNested.NestedObject.Attributes == nil {
							inputNested.NestedObject.Attributes = make(map[string]schema.Attribute)
						}
						// Recursively merge nested attributes
						nestedReadOnly := mergeNestedAttributesAndFindReadOnly(inputNested.NestedObject.Attributes, stateNested.NestedObject.Attributes, fullPath)
						readOnlyAttrs = append(readOnlyAttrs, nestedReadOnly...)
						// Update the input attribute with merged nested attributes
						inputAttrs[key] = inputNested
					}
				}
			}
		}
	}

	return readOnlyAttrs
}

// GenerateDataSourceSchemaFromStruct generates a Terraform schema from a Go struct.
func GenerateDataSourceSchemaFromStruct(inputModel interface{}, stateModel interface{}, sensitiveAttrs []string, extraRequiredAttrs []string, computedAsSetAttrs []string) schema.Schema {
	inputModelAttrs := make(map[string]schema.Attribute)
	if inputModel != nil {
		inputModelAttrs = dataSourceSchemaAttrsFromStruct(inputModel, false, sensitiveAttrs, extraRequiredAttrs, computedAsSetAttrs)
	}
	outputModelAttrs := dataSourceSchemaAttrsFromStruct(stateModel, true, sensitiveAttrs, extraRequiredAttrs, computedAsSetAttrs)

	// Track which attributes are only in the state model (read-only)
	// This function will merge nested attributes and identify read-only ones
	readOnlyAttrs := mergeNestedAttributesAndFindReadOnly(inputModelAttrs, outputModelAttrs, "")

	// Mark all attributes that are only in state model as read-only (Optional=false, Required=false, Computed=true)
	forceComputedAttributesReadOnlyDataSource(inputModelAttrs, readOnlyAttrs)

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
