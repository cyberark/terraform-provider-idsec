// Copyright CyberArk. 2026
// SPDX-License-Identifier: Apache-2.0

package actions_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

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

// generateSchemaForResource replicates the arguments internal/provider passes at its three
// schemas.GenerateResourceSchemaFromStruct call sites, so this package's tests see the same
// generator inputs a real apply would, without importing internal/provider (an import cycle).
//
// ForceNewAttributes is passed as nil because the field does not exist on the action definition
// today; see IdsecResource.getForceNewAttributes.
//
// The error reports a malformed resource registration, which is distinct from the diagnostics:
// those report that the generated schema itself is invalid.
func generateSchemaForResource(resourceDef *actions.IdsecServiceTerraformResourceActionDefinition, writeOnlyOverride map[string]string) (rschema.Schema, diag.Diagnostics, error) {
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
	)
	return generated, diags, nil
}

// TestAllWriteOnlyAttributesAreValid fails if any registered resource declares a
// WriteOnlyAttributes entry that the schema generator rejects. It passes trivially today because
// no resource declares one yet; it is armed for the first that does, so the eligibility and
// trigger rules become a CI gate rather than documentation a reviewer might miss.
func TestAllWriteOnlyAttributesAreValid(t *testing.T) {
	allConfigs := actions.AllTerraformConfigs()

	if len(allConfigs) == 0 {
		t.Skip("No Terraform service configurations registered")
	}

	for _, config := range allConfigs {
		for _, resourceDef := range config.Resources {
			if len(resourceDef.WriteOnlyAttributes) == 0 {
				continue
			}

			t.Run(config.ServiceName+"/"+resourceDef.ActionName, func(t *testing.T) {
				_, diags, err := generateSchemaForResource(resourceDef, resourceDef.WriteOnlyAttributes)
				if err != nil {
					t.Fatalf("could not generate schema: %v", err)
				}
				if diags.HasError() {
					t.Errorf("resource %q in service %q declares an invalid WriteOnlyAttributes entry:\n%s",
						resourceDef.ActionName, config.ServiceName, diags.Errors())
				}
			})
		}
	}
}
