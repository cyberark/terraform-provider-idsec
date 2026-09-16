// Copyright CyberArk. 2026
// SPDX-License-Identifier: Apache-2.0

package actions_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/cyberark/idsec-sdk-golang/pkg/common"
	modelsactions "github.com/cyberark/idsec-sdk-golang/pkg/models/actions"
	"github.com/cyberark/terraform-provider-idsec/internal/actions"
	"github.com/cyberark/terraform-provider-idsec/internal/schemas"

	_ "github.com/cyberark/terraform-provider-idsec/internal/tfactions"
)

// TestAllImportIDAttributesExist validates that all ImportID fields reference valid StateSchema fields.
func TestAllImportIDAttributesExist(t *testing.T) {
	allConfigs := actions.AllTerraformConfigs()

	if len(allConfigs) == 0 {
		t.Skip("No Terraform service configurations registered")
	}

	for _, config := range allConfigs {
		for _, resourceDef := range config.Resources {
			t.Run(config.ServiceName+"/"+resourceDef.ActionName, func(t *testing.T) {
				if resourceDef.ImportID == "" {
					return
				}

				if resourceDef.ImportID == actions.SingletonResourceImportDummyID {
					return
				}

				if !operationSupported(resourceDef.SupportedOperations, actions.ReadOperation) {
					t.Errorf("ImportID '%s' is configured but Read operation is not supported for resource '%s' in service '%s'",
						resourceDef.ImportID, resourceDef.ActionName, config.ServiceName)
					return
				}

				if resourceDef.StateSchema == nil {
					t.Errorf("ImportID '%s' is configured but StateSchema is not defined for resource '%s' in service '%s'",
						resourceDef.ImportID, resourceDef.ActionName, config.ServiceName)
					return
				}

				for _, fieldName := range schemas.SplitImportIDAttributes(resourceDef.ImportID) {
					if err := schemas.ValidateStateSchemaImportAttribute(resourceDef.StateSchema, fieldName); err != nil {
						t.Errorf("ImportID field '%s' is invalid for resource '%s' in service '%s': %v",
							fieldName, resourceDef.ActionName, config.ServiceName, err)
					}
				}
			})
		}
	}
}

// TestAllExtraRequiredAttributesExist validates that all ExtraRequiredAttributes reference valid schema fields.
func TestAllExtraRequiredAttributesExist(t *testing.T) {
	allConfigs := actions.AllTerraformConfigs()

	if len(allConfigs) == 0 {
		t.Skip("No Terraform service configurations registered")
	}

	for _, config := range allConfigs {
		// Validate resources
		for _, resourceDef := range config.Resources {
			t.Run(config.ServiceName+"/"+resourceDef.ActionName, func(t *testing.T) {
				if len(resourceDef.ExtraRequiredAttributes) == 0 {
					return
				}

				createActionName, hasCreate := resourceDef.ActionsMappings[actions.CreateOperation]
				if !hasCreate {
					t.Errorf("ExtraRequiredAttributes configured but Create operation not supported for resource '%s' in service '%s'",
						resourceDef.ActionName, config.ServiceName)
					return
				}

				createSchema, hasCreateSchema := resourceDef.Schemas[createActionName]
				if !hasCreateSchema || createSchema == nil {
					t.Errorf("ExtraRequiredAttributes configured but Create schema '%s' not defined for resource '%s' in service '%s'",
						createActionName, resourceDef.ActionName, config.ServiceName)
					return
				}
				createSchema, _ = modelsactions.UnwrapSchema(createSchema)

				validateAttributeList(t, config.ServiceName, resourceDef.ActionName, "resource",
					resourceDef.ExtraRequiredAttributes, createSchema, "ExtraRequiredAttributes")
			})
		}

		// Validate data sources
		for _, dataSourceDef := range config.DataSources {
			t.Run(config.ServiceName+"/"+dataSourceDef.ActionName+"_datasource", func(t *testing.T) {
				if len(dataSourceDef.ExtraRequiredAttributes) == 0 {
					return
				}

				inputSchema, hasInputSchema := dataSourceDef.Schemas[dataSourceDef.DataSourceAction]
				if !hasInputSchema || inputSchema == nil {
					t.Errorf("ExtraRequiredAttributes configured but DataSource schema '%s' not defined for data_source '%s' in service '%s'",
						dataSourceDef.DataSourceAction, dataSourceDef.ActionName, config.ServiceName)
					return
				}
				inputSchema, _ = modelsactions.UnwrapSchema(inputSchema)

				validateAttributeList(t, config.ServiceName, dataSourceDef.ActionName, "data_source",
					dataSourceDef.ExtraRequiredAttributes, inputSchema, "ExtraRequiredAttributes")
			})
		}
	}
}

// TestAllServiceActionMappingsHaveSchemas validates that all ActionsMappings have corresponding schema entries.
func TestAllServiceActionMappingsHaveSchemas(t *testing.T) {
	allConfigs := actions.AllTerraformConfigs()

	if len(allConfigs) == 0 {
		t.Skip("No Terraform service configurations registered")
	}

	for _, config := range allConfigs {
		for _, resourceDef := range config.Resources {
			t.Run(config.ServiceName+"/"+resourceDef.ActionName, func(t *testing.T) {
				for operation, actionString := range resourceDef.ActionsMappings {
					if _, exists := resourceDef.Schemas[actionString]; !exists {
						t.Errorf("Service '%s': Action '%s' (operation: %v) in resource '%s' is missing from schemas",
							config.ServiceName, actionString, operation, resourceDef.ActionName)
					}
				}
			})
		}

		for _, dataSourceDef := range config.DataSources {
			t.Run(config.ServiceName+"/"+dataSourceDef.ActionName+"_datasource", func(t *testing.T) {
				if dataSourceDef.DataSourceAction != "" {
					if _, exists := dataSourceDef.Schemas[dataSourceDef.DataSourceAction]; !exists {
						t.Errorf("Service '%s': DataSourceAction '%s' in data source '%s' is missing from schemas",
							config.ServiceName, dataSourceDef.DataSourceAction, dataSourceDef.ActionName)
					}
				}
			})
		}
	}
}

func validateAttributeList(t *testing.T, serviceName, actionName, actionType string, attributes []string, schema interface{}, attributeType string) {
	t.Helper()

	if len(attributes) == 0 {
		return
	}

	if schema == nil {
		t.Errorf("%s contains %d attribute(s) but schema is nil for %s '%s' in service '%s'",
			attributeType, len(attributes), actionType, actionName, serviceName)
		return
	}

	schemaType := reflect.TypeOf(schema)
	if schemaType.Kind() == reflect.Ptr {
		schemaType = schemaType.Elem()
	}

	for _, fieldName := range attributes {
		fieldName = strings.TrimSpace(fieldName)
		if fieldName == "" {
			continue
		}

		field := findFieldByTerraformAttributeName(schemaType, fieldName)
		if field == nil {
			t.Errorf("%s field '%s' does not exist in schema %s for %s '%s' in service '%s'",
				attributeType, fieldName, schemaType.Name(), actionType, actionName, serviceName)
		}
	}
}

// findFieldByTerraformAttributeName looks up a struct field by its Terraform attribute name.
//
// Terraform attribute names are derived from struct tags (mapstructure, json) rather than
// Go field names. This function first attempts the PascalCase Go-field-name lookup via
// common.FindFieldByName, then falls back to scanning struct tags directly. This handles
// cases where the Go field name does not match the PascalCase of the attribute name
// (e.g., a field named "ID" with a mapstructure tag of "https_relay_id").
func findFieldByTerraformAttributeName(schemaType reflect.Type, attrName string) *reflect.StructField {
	if field := common.FindFieldByName(schemaType, attrName); field != nil {
		return field
	}

	if schemaType.Kind() == reflect.Ptr {
		schemaType = schemaType.Elem()
	}
	if schemaType.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < schemaType.NumField(); i++ {
		field := schemaType.Field(i)
		if mapTag := strings.Split(field.Tag.Get("mapstructure"), ",")[0]; mapTag == attrName {
			return &field
		}
		if jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]; jsonTag == attrName {
			return &field
		}
	}
	return nil
}

func operationSupported(operations []actions.IdsecServiceActionOperation, operation actions.IdsecServiceActionOperation) bool {
	for _, op := range operations {
		if op == operation {
			return true
		}
	}
	return false
}

// generateSchemaForResource replicates the arguments internal/provider passes at its
// schemas.GenerateResourceSchemaFromStruct call sites, so this package's tests see the same
// generator inputs a real apply would, without importing internal/provider (an import cycle).
//
// ForceNewAttributes is passed as nil because the field does not exist on the action definition
// today; see IdsecResource.getForceNewAttributes.
//
// writeOnlyOverride and writeOnlyHashedOverride let a caller substitute the schema-generation
// inputs under test in place of the resource's own declarations (or nil to use neither), which
// TestAllWriteOnlyAttributesAreValid needs to exercise WriteOnlyAttributes and
// WriteOnlyHashedAttributes together so their cross-field mutual-exclusion rule is reachable.
//
// The error reports a malformed resource registration, which is distinct from the diagnostics:
// those report that the generated schema itself is invalid.
func generateSchemaForResource(resourceDef *actions.IdsecServiceTerraformResourceActionDefinition, writeOnlyOverride map[string]string, writeOnlyHashedOverride []string) (rschema.Schema, diag.Diagnostics, error) {
	createActionName, hasCreate := resourceDef.ActionsMappings[actions.CreateOperation]
	if !hasCreate {
		return rschema.Schema{}, nil, fmt.Errorf("resource %q has no Create operation mapping", resourceDef.ActionName)
	}
	createSchema, hasCreateSchema := resourceDef.Schemas[createActionName]
	if !hasCreateSchema || createSchema == nil {
		return rschema.Schema{}, nil, fmt.Errorf("resource %q: Create schema %q is not registered", resourceDef.ActionName, createActionName)
	}
	createSchema, _ = modelsactions.UnwrapSchema(createSchema)

	var updateSchema interface{}
	if updateActionName, hasUpdate := resourceDef.ActionsMappings[actions.UpdateOperation]; hasUpdate {
		if s, ok := resourceDef.Schemas[updateActionName]; ok && s != nil {
			updateSchema, _ = modelsactions.UnwrapSchema(s)
		}
	}

	generated, diags := schemas.GenerateResourceSchemaFromStruct(
		createSchema,
		updateSchema,
		resourceDef.StateSchema,
		resourceDef.SensitiveAttributes,
		resourceDef.ExtraRequiredAttributes,
		resourceDef.ComputedAsSetAttributes,
		resourceDef.ImmutableAttributes,
		nil,
		resourceDef.ComputedAttributes,
		resourceDef.SemanticEqualityAttributes,
		writeOnlyOverride,
		writeOnlyHashedOverride,
	)
	return generated, diags, nil
}

// TestAllWriteOnlyAttributesAreValid fails if any registered resource declares a
// WriteOnlyAttributes or WriteOnlyHashedAttributes entry that the schema generator rejects. It
// passes trivially today because no resource declares either yet; it is armed for the first that
// does, so the eligibility and trigger rules become a CI gate rather than documentation a
// reviewer might miss.
//
// Both fields are generated together in a single call per resource, rather than in two separate
// calls, because mutual exclusion between them is a cross-field rule: an attribute declared in
// both maps/slices must be rejected, and generating each field in isolation would never exercise
// that check.
func TestAllWriteOnlyAttributesAreValid(t *testing.T) {
	allConfigs := actions.AllTerraformConfigs()

	if len(allConfigs) == 0 {
		t.Skip("No Terraform service configurations registered")
	}

	for _, config := range allConfigs {
		for _, resourceDef := range config.Resources {
			if len(resourceDef.WriteOnlyAttributes) == 0 && len(resourceDef.WriteOnlyHashedAttributes) == 0 {
				continue
			}

			t.Run(config.ServiceName+"/"+resourceDef.ActionName, func(t *testing.T) {
				_, diags, err := generateSchemaForResource(resourceDef, resourceDef.WriteOnlyAttributes, resourceDef.WriteOnlyHashedAttributes)
				if err != nil {
					t.Fatalf("could not generate schema: %v", err)
				}
				if diags.HasError() {
					declared := make([]string, 0, 2)
					if len(resourceDef.WriteOnlyAttributes) > 0 {
						declared = append(declared, "WriteOnlyAttributes")
					}
					if len(resourceDef.WriteOnlyHashedAttributes) > 0 {
						declared = append(declared, "WriteOnlyHashedAttributes")
					}
					t.Errorf("resource %q in service %q declares an invalid %s entry:\n%s",
						resourceDef.ActionName, config.ServiceName, strings.Join(declared, "/"), diags.Errors())
				}
			})
		}
	}
}

// flatAttr is one attribute of a generated schema, reduced to the two facts the sensitivity
// consistency check needs: where it lives and whether it is redacted.
type flatAttr struct {
	owner     string
	path      string
	name      string
	sensitive bool
}

// flattenResourceAttrs walks a generated resource schema, including nested objects, and returns
// one flatAttr per leaf and per nested container.
func flattenResourceAttrs(owner string, attrs map[string]rschema.Attribute, prefix string) []flatAttr {
	out := make([]flatAttr, 0, len(attrs))
	for name, attr := range attrs {
		path := name
		if prefix != "" {
			path = prefix + "." + name
		}
		out = append(out, flatAttr{owner: owner, path: path, name: name, sensitive: attr.IsSensitive()})
		switch a := attr.(type) {
		case rschema.SingleNestedAttribute:
			out = append(out, flattenResourceAttrs(owner, a.Attributes, path)...)
		case rschema.ListNestedAttribute:
			out = append(out, flattenResourceAttrs(owner, a.NestedObject.Attributes, path)...)
		case rschema.SetNestedAttribute:
			out = append(out, flattenResourceAttrs(owner, a.NestedObject.Attributes, path)...)
		case rschema.MapNestedAttribute:
			out = append(out, flattenResourceAttrs(owner, a.NestedObject.Attributes, path)...)
		}
	}
	return out
}

// flattenDataSourceAttrs is flattenResourceAttrs for data-source schemas, which use a parallel
// but distinct set of attribute types.
func flattenDataSourceAttrs(owner string, attrs map[string]dsschema.Attribute, prefix string) []flatAttr {
	out := make([]flatAttr, 0, len(attrs))
	for name, attr := range attrs {
		path := name
		if prefix != "" {
			path = prefix + "." + name
		}
		out = append(out, flatAttr{owner: owner, path: path, name: name, sensitive: attr.IsSensitive()})
		switch a := attr.(type) {
		case dsschema.SingleNestedAttribute:
			out = append(out, flattenDataSourceAttrs(owner, a.Attributes, path)...)
		case dsschema.ListNestedAttribute:
			out = append(out, flattenDataSourceAttrs(owner, a.NestedObject.Attributes, path)...)
		case dsschema.SetNestedAttribute:
			out = append(out, flattenDataSourceAttrs(owner, a.NestedObject.Attributes, path)...)
		case dsschema.MapNestedAttribute:
			out = append(out, flattenDataSourceAttrs(owner, a.NestedObject.Attributes, path)...)
		}
	}
	return out
}

// collectAllAttributes generates every registered resource and data-source schema and returns
// their attributes flattened into a single slice.
func collectAllAttributes(t *testing.T) []flatAttr {
	t.Helper()
	var all []flatAttr
	for _, config := range actions.AllTerraformConfigs() {
		for _, resourceDef := range config.Resources {
			generated, _, err := generateSchemaForResource(resourceDef, resourceDef.WriteOnlyAttributes, resourceDef.WriteOnlyHashedAttributes)
			if err != nil {
				continue
			}
			all = append(all, flattenResourceAttrs("resource "+resourceDef.ActionName, generated.Attributes, "")...)
		}
		for _, dataSourceDef := range config.DataSources {
			if dataSourceDef.StateSchema == nil || dataSourceDef.DataSourceAction == "" {
				continue
			}
			inputSchema, ok := dataSourceDef.Schemas[dataSourceDef.DataSourceAction]
			if !ok {
				continue
			}
			inputSchema, _ = modelsactions.UnwrapSchema(inputSchema)
			generated := schemas.GenerateDataSourceSchemaFromStruct(
				inputSchema,
				dataSourceDef.StateSchema,
				dataSourceDef.SensitiveAttributes,
				dataSourceDef.ExtraRequiredAttributes,
				dataSourceDef.ComputedAsSetAttributes,
			)
			all = append(all, flattenDataSourceAttrs("data source "+dataSourceDef.ActionName, generated.Attributes, "")...)
		}
	}
	return all
}

// TestSensitiveAttributesAreConsistent fails if an attribute name is marked Sensitive somewhere
// in the provider but left unmarked somewhere else.
//
// Sensitivity is declared in two places that can disagree: the SDK's `secret:"true"` struct tag
// and the per-action-definition SensitiveAttributes list. A resource and its data source are
// separate definitions that share one StateSchema, so an untagged credential has to be listed
// twice and silently renders in cleartext wherever the list was not repeated. An unmarked
// attribute is printed verbatim in plan output, CI logs, and pull-request comments, and it does
// not propagate redaction to values derived from it.
//
// The check treats "marked sensitive anywhere" as the provider's own assertion that the value is
// a credential, then requires every other occurrence of that attribute name to agree. Matching is
// on the leaf attribute name rather than the full path because that is the granularity
// SensitiveAttributes itself uses.
//
// It needs no tenant and no Terraform binary: schemas are generated in-process from the
// registrations, so this runs in plain `go test`.
func TestSensitiveAttributesAreConsistent(t *testing.T) {
	t.Parallel()

	allAttrs := collectAllAttributes(t)
	if len(allAttrs) == 0 {
		t.Skip("No Terraform service configurations registered")
	}

	// An attribute name is expected to be sensitive everywhere as soon as it is sensitive once.
	sensitiveNames := map[string]string{}
	for _, a := range allAttrs {
		if a.sensitive {
			if _, seen := sensitiveNames[a.name]; !seen {
				sensitiveNames[a.name] = a.owner + "." + a.path
			}
		}
	}

	for _, a := range allAttrs {
		if a.sensitive {
			continue
		}
		declaredAt, expected := sensitiveNames[a.name]
		if !expected {
			continue
		}
		t.Errorf("%s: attribute %q is not marked Sensitive, but %q is marked Sensitive at %s.\n"+
			"Mark it via the SDK's `secret:\"true\"` struct tag, or add %q to SensitiveAttributes on this definition.",
			a.owner, a.path, a.name, declaredAt, a.name)
	}
}
