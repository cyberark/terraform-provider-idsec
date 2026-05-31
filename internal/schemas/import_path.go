// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package schemas

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
)

// ParseImportAttributePath converts a dot-separated Terraform attribute path into a framework path.
// Top-level attributes use a single segment (e.g. "safe_id"); nested attributes use multiple
// segments (e.g. "metadata.policy_id").
func ParseImportAttributePath(attributePath string) (path.Path, error) {
	attributePath = strings.TrimSpace(attributePath)
	if attributePath == "" {
		return path.Path{}, fmt.Errorf("import attribute path cannot be empty")
	}

	segments := strings.Split(attributePath, ".")
	for _, segment := range segments {
		if strings.TrimSpace(segment) == "" {
			return path.Path{}, fmt.Errorf("import attribute path %q contains an empty segment", attributePath)
		}
	}

	attrPath := path.Root(segments[0])
	for _, segment := range segments[1:] {
		attrPath = attrPath.AtName(segment)
	}
	return attrPath, nil
}

// SplitImportIDAttributes splits an ImportID spec into individual attribute paths.
// Colon separates multiple attributes (e.g. "safe_id:member_name" or "metadata.policy_id:other_id").
func SplitImportIDAttributes(importID string) []string {
	parts := strings.Split(importID, ":")
	attributes := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			attributes = append(attributes, part)
		}
	}
	return attributes
}

// ValidateStateSchemaImportAttribute verifies that attributePath exists on the Terraform state
// schema and resolves to a string or integer field suitable for import ID values.
func ValidateStateSchemaImportAttribute(stateSchema interface{}, attributePath string) error {
	attributePath = strings.TrimSpace(attributePath)
	if attributePath == "" {
		return fmt.Errorf("import attribute path cannot be empty")
	}

	leafType, err := stateSchemaFieldTypeByPath(reflect.TypeOf(stateSchema), attributePath)
	if err != nil {
		return fmt.Errorf("import attribute path %q: %w", attributePath, err)
	}
	for leafType.Kind() == reflect.Ptr {
		leafType = leafType.Elem()
	}
	if !isImportIDCompatibleType(leafType.Kind()) {
		return fmt.Errorf("import attribute path %q must reference a string or integer field, got %s", attributePath, leafType.Kind())
	}
	return nil
}

func isImportIDCompatibleType(kind reflect.Kind) bool {
	switch kind {
	case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	default:
		return false
	}
}

func stateSchemaFieldTypeByPath(schemaType reflect.Type, attributePath string) (reflect.Type, error) {
	if schemaType == nil {
		return nil, fmt.Errorf("state schema type is nil")
	}
	if schemaType.Kind() == reflect.Ptr {
		schemaType = schemaType.Elem()
	}

	for _, key := range strings.Split(attributePath, ".") {
		key = strings.TrimSpace(key)
		if key == "" {
			return nil, fmt.Errorf("contains an empty segment")
		}
		if schemaType.Kind() != reflect.Struct {
			return nil, fmt.Errorf("field %q is not in a struct", key)
		}

		fieldType, found := stateSchemaStructFieldType(schemaType, key)
		if !found {
			return nil, fmt.Errorf("field %q not found in struct", key)
		}
		schemaType = fieldType
	}

	return schemaType, nil
}

func stateSchemaStructFieldType(schemaType reflect.Type, key string) (reflect.Type, bool) {
	actualFields := resolveFieldsSquashed(schemaType)
	for i := range actualFields {
		if resolveFieldName(actualFields[i]) == key {
			return actualFields[i].Type, true
		}
	}
	return nil, false
}
