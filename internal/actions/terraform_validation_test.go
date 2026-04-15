// Copyright (c) CyberArk.
// SPDX-License-Identifier: Apache-2.0

package actions_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/cyberark/idsec-sdk-golang/pkg/common"
	"github.com/cyberark/terraform-provider-idsec/internal/actions"

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

				readActionName, hasRead := resourceDef.ActionsMappings[actions.ReadOperation]
				if !hasRead {
					t.Errorf("ImportID '%s' is configured but Read operation is not supported for resource '%s' in service '%s'",
						resourceDef.ImportID, resourceDef.ActionName, config.ServiceName)
					return
				}

				readSchema, hasReadSchema := resourceDef.Schemas[readActionName]
				if !hasReadSchema || readSchema == nil {
					t.Errorf("ImportID '%s' is configured but Read schema '%s' is not defined for resource '%s' in service '%s'",
						resourceDef.ImportID, readActionName, resourceDef.ActionName, config.ServiceName)
					return
				}

				importFields := strings.Split(resourceDef.ImportID, ":")

				schemaType := reflect.TypeOf(readSchema)
				if schemaType.Kind() == reflect.Ptr {
					schemaType = schemaType.Elem()
				}

				for _, fieldName := range importFields {
					fieldName = strings.TrimSpace(fieldName)
					if fieldName == "" {
						continue
					}

					field := common.FindFieldByName(schemaType, fieldName)
					if field == nil {
						t.Errorf("ImportID field '%s' does not exist in Read schema %s for resource '%s' in service '%s'",
							fieldName, schemaType.Name(), resourceDef.ActionName, config.ServiceName)
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

		field := common.FindFieldByName(schemaType, fieldName)
		if field == nil {
			t.Errorf("%s field '%s' does not exist in schema %s for %s '%s' in service '%s'",
				attributeType, fieldName, schemaType.Name(), actionType, actionName, serviceName)
		}
	}
}
