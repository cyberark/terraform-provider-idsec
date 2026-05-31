// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// buildImportTestState builds a minimal Terraform state schema and raw value for import tests.
// Attribute paths may be top-level (e.g. "safe_id") or nested (e.g. "metadata.policy_id").
func buildImportTestState(attributePaths ...string) (schema.Schema, tftypes.Value) {
	schemaAttrs := map[string]schema.Attribute{}
	attrTypes := map[string]tftypes.Type{}
	rawAttrs := map[string]tftypes.Value{}

	for _, attributePath := range attributePaths {
		if attributePath == "" {
			continue
		}
		mergeImportTestAttribute(schemaAttrs, attrTypes, rawAttrs, strings.Split(attributePath, "."))
	}

	return schema.Schema{Attributes: schemaAttrs}, tftypes.NewValue(
		tftypes.Object{AttributeTypes: attrTypes},
		rawAttrs,
	)
}

func mergeImportTestAttribute(
	schemaAttrs map[string]schema.Attribute,
	attrTypes map[string]tftypes.Type,
	rawAttrs map[string]tftypes.Value,
	segments []string,
) {
	if len(segments) == 0 {
		return
	}

	name := segments[0]
	if len(segments) == 1 {
		schemaAttrs[name] = schema.StringAttribute{}
		attrTypes[name] = tftypes.String
		rawAttrs[name] = tftypes.NewValue(tftypes.String, nil)
		return
	}

	nestedSchemaAttrs := map[string]schema.Attribute{}
	nestedAttrTypes := map[string]tftypes.Type{}
	nestedRawAttrs := map[string]tftypes.Value{}

	if existingNested, ok := schemaAttrs[name].(schema.SingleNestedAttribute); ok {
		nestedSchemaAttrs = existingNested.Attributes
		for attrName, attrType := range nestedAttrTypesFromSchema(existingNested) {
			nestedAttrTypes[attrName] = attrType
		}
		for attrName, attrValue := range nestedRawAttrsFromSchema(existingNested) {
			nestedRawAttrs[attrName] = attrValue
		}
	}

	mergeImportTestAttribute(nestedSchemaAttrs, nestedAttrTypes, nestedRawAttrs, segments[1:])

	schemaAttrs[name] = schema.SingleNestedAttribute{Attributes: nestedSchemaAttrs}
	nestedObjectType := tftypes.Object{AttributeTypes: nestedAttrTypes}
	attrTypes[name] = nestedObjectType
	rawAttrs[name] = tftypes.NewValue(nestedObjectType, nestedRawAttrs)
}

func nestedAttrTypesFromSchema(nested schema.SingleNestedAttribute) map[string]tftypes.Type {
	attrTypes := make(map[string]tftypes.Type, len(nested.Attributes))
	for attrName := range nested.Attributes {
		attrTypes[attrName] = tftypes.String
	}
	return attrTypes
}

func nestedRawAttrsFromSchema(nested schema.SingleNestedAttribute) map[string]tftypes.Value {
	rawAttrs := make(map[string]tftypes.Value, len(nested.Attributes))
	for attrName := range nested.Attributes {
		rawAttrs[attrName] = tftypes.NewValue(tftypes.String, nil)
	}
	return rawAttrs
}
