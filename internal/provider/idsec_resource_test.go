// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/cyberark/idsec-sdk-golang/pkg/services"
	"github.com/cyberark/terraform-provider-idsec/internal/actions"
	"github.com/cyberark/terraform-provider-idsec/internal/schemas"
)

// CreateTestIdsecResource creates a new IdsecResource instance for testing.
//
// Parameters:
//   - serviceConfig: Configuration for the IDSEC service
//   - actionDefinition: Definition of the Terraform resource action
//
// Returns a new IdsecResource instance configured with the provided parameters.
func CreateTestIdsecResource(
	serviceConfig *services.IdsecServiceConfig,
	actionDefinition *actions.IdsecServiceTerraformResourceActionDefinition,
) resource.Resource {
	return NewIdsecResource(serviceConfig, actionDefinition)
}

// CreateTestServiceConfig creates a test service configuration.
//
// Parameters:
//   - serviceName: Name of the service
//
// Returns a new IdsecServiceConfig instance for testing.
func CreateTestServiceConfig(serviceName string) *services.IdsecServiceConfig {
	return &services.IdsecServiceConfig{
		ServiceName: serviceName,
	}
}

// CreateTestActionDefinition creates a test action definition.
//
// Parameters:
//   - actionName: Name of the action
//   - actionDescription: Description of the action
//
// Returns a new IdsecServiceTerraformResourceActionDefinition instance for testing.
func CreateTestActionDefinition(actionName, actionDescription string) *actions.IdsecServiceTerraformResourceActionDefinition {
	return &actions.IdsecServiceTerraformResourceActionDefinition{
		IdsecServiceBaseTerraformActionDefinition: actions.IdsecServiceBaseTerraformActionDefinition{
			IdsecServiceBaseActionDefinition: actions.IdsecServiceBaseActionDefinition{
				ActionName:        actionName,
				ActionDescription: actionDescription,
			},
		},
	}
}

// CreateTestActionDefinitionWithImportID creates a test action definition with ImportID set.
//
// Parameters:
//   - actionName: Name of the action
//   - actionDescription: Description of the action
//   - importIDAttribute: Import ID attribute name to set
//
// Returns a new IdsecServiceTerraformResourceActionDefinition instance for testing.
func CreateTestActionDefinitionWithImportID(actionName, actionDescription, importIDAttribute string) *actions.IdsecServiceTerraformResourceActionDefinition {
	actionDef := CreateTestActionDefinition(actionName, actionDescription)
	// Use reflection to set ImportID if the field exists
	val := reflect.ValueOf(actionDef).Elem()
	field := val.FieldByName("ImportID")
	if field.IsValid() && field.CanSet() {
		field.SetString(importIDAttribute)
	}
	return actionDef
}

// CreateTestActionDefinitionWithOperations creates a test action definition with specified supported operations.
//
// Parameters:
//   - actionName: Name of the action
//   - actionDescription: Description of the action
//   - supportedOperations: List of supported operations
//
// Returns a new IdsecServiceTerraformResourceActionDefinition instance for testing.
func CreateTestActionDefinitionWithOperations(actionName, actionDescription string, supportedOperations []actions.IdsecServiceActionOperation) *actions.IdsecServiceTerraformResourceActionDefinition {
	actionDef := CreateTestActionDefinition(actionName, actionDescription)
	actionDef.SupportedOperations = supportedOperations
	return actionDef
}

// CreateTestActionDefinitionWithImportIDAndOperations creates a test action definition with ImportID and supported operations set.
//
// Parameters:
//   - actionName: Name of the action
//   - actionDescription: Description of the action
//   - importIDAttribute: Import ID attribute name to set
//   - supportedOperations: List of supported operations
//
// Returns a new IdsecServiceTerraformResourceActionDefinition instance for testing.
func CreateTestActionDefinitionWithImportIDAndOperations(actionName, actionDescription, importIDAttribute string, supportedOperations []actions.IdsecServiceActionOperation) *actions.IdsecServiceTerraformResourceActionDefinition {
	actionDef := CreateTestActionDefinitionWithOperations(actionName, actionDescription, supportedOperations)
	// Use reflection to set ImportID if the field exists
	val := reflect.ValueOf(actionDef).Elem()
	field := val.FieldByName("ImportID")
	if field.IsValid() && field.CanSet() {
		field.SetString(importIDAttribute)
	}
	return actionDef
}

// TestIdsecResource_Metadata tests the Metadata function of IdsecResource.
//
// This test validates that the Metadata function correctly generates the TypeName
// for the Terraform resource based on the provider type name and action name.
// It tests various scenarios including normal action names, names with hyphens,
// and edge cases like empty names.
func TestIdsecResource_Metadata(t *testing.T) {
	tests := []struct {
		name             string
		providerTypeName string
		actionName       string
		expectedTypeName string
	}{
		{
			name:             "success_normal_action_name_with_hyphens",
			providerTypeName: "idsec",
			actionName:       "test-action",
			expectedTypeName: "idsec_test_action",
		},
		{
			name:             "success_action_name_without_hyphens",
			providerTypeName: "idsec",
			actionName:       "testaction",
			expectedTypeName: "idsec_testaction",
		},
		{
			name:             "success_action_name_with_multiple_hyphens",
			providerTypeName: "idsec",
			actionName:       "test-action-with-hyphens",
			expectedTypeName: "idsec_test_action_with_hyphens",
		},
		{
			name:             "success_empty_action_name",
			providerTypeName: "idsec",
			actionName:       "",
			expectedTypeName: "idsec_",
		},
		{
			name:             "success_action_name_with_underscores",
			providerTypeName: "idsec",
			actionName:       "test_action",
			expectedTypeName: "idsec_test_action",
		},
		{
			name:             "success_different_provider_name",
			providerTypeName: "custom_provider",
			actionName:       "my-resource",
			expectedTypeName: "custom_provider_my_resource",
		},
		{
			name:             "success_action_name_with_mixed_separators",
			providerTypeName: "idsec",
			actionName:       "test-action_mixed",
			expectedTypeName: "idsec_test_action_mixed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			ctx := context.Background()
			serviceConfig := CreateTestServiceConfig("test-service")
			actionDefinition := CreateTestActionDefinition(tt.actionName, "Test action description")
			idsecResource := CreateTestIdsecResource(serviceConfig, actionDefinition)

			// Create request and response
			req := resource.MetadataRequest{
				ProviderTypeName: tt.providerTypeName,
			}
			resp := &resource.MetadataResponse{}

			// Execute
			idsecResource.Metadata(ctx, req, resp)

			// Validate
			if resp.TypeName != tt.expectedTypeName {
				t.Errorf("Expected TypeName '%s', got '%s'", tt.expectedTypeName, resp.TypeName)
			}
		})
	}
}

// TestIdsecResource_getImportID tests the getImportID function of IdsecResource.
//
// This test validates that the getImportID function correctly retrieves the ImportID
// from the action definition using reflection, and returns an empty string if not configured.
func TestIdsecResource_getImportID(t *testing.T) {
	tests := []struct {
		name              string
		importIDAttribute string
		expectedAttribute string
		description       string
	}{
		{
			name:              "success_with_import_id_attribute_set",
			importIDAttribute: "safe_id",
			expectedAttribute: "safe_id",
			description:       "Returns configured ImportID when set",
		},
		{
			name:              "success_with_different_attribute",
			importIDAttribute: "platform_id",
			expectedAttribute: "platform_id",
			description:       "Returns configured ImportID for different attribute",
		},
		{
			name:              "success_with_standard_id",
			importIDAttribute: "id",
			expectedAttribute: "id",
			description:       "Returns configured ImportID when set to standard id",
		},
		{
			name:              "success_with_empty_string",
			importIDAttribute: "",
			expectedAttribute: "",
			description:       "Returns empty string when ImportID is empty",
		},
		{
			name:              "success_without_import_id_attribute",
			importIDAttribute: "", // Field not set in struct
			expectedAttribute: "",
			description:       "Returns empty string when ImportID field does not exist (backward compatibility)",
		},
		{
			name:              "success_with_multi_attribute_import_id",
			importIDAttribute: "safe_id:member_name",
			expectedAttribute: "safe_id:member_name",
			description:       "Returns full ImportID string including colon for multi-attribute import",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			serviceConfig := CreateTestServiceConfig("test-service")
			var actionDefinition *actions.IdsecServiceTerraformResourceActionDefinition

			// Create action definition with or without ImportID based on test case
			if tt.name == "success_without_import_id_attribute" {
				// Test case where field is not set at all
				actionDefinition = CreateTestActionDefinition("test-action", "Test action description")
			} else {
				// Test cases where we want to set ImportID
				actionDefinition = CreateTestActionDefinitionWithImportID("test-action", "Test action description", tt.importIDAttribute)
				// Check if field exists - if not, skip test (SDK version doesn't support it)
				val := reflect.ValueOf(actionDefinition).Elem()
				field := val.FieldByName("ImportID")
				if !field.IsValid() && tt.importIDAttribute != "" {
					t.Skipf("ImportID field not available in this SDK version")
				}
			}

			idsecResource := CreateTestIdsecResource(serviceConfig, actionDefinition)
			// Cast to *IdsecResource to access private method
			resource, ok := idsecResource.(*IdsecResource)
			if !ok {
				t.Fatalf("Failed to cast resource to *IdsecResource")
			}

			// Execute
			result := resource.getImportID()

			// Validate
			if result != tt.expectedAttribute {
				t.Errorf("Expected ImportID '%s', got '%s'", tt.expectedAttribute, result)
			}
		})
	}
}

// TestIdsecResource_ImportState tests the ImportState function of IdsecResource.
//
// This test validates that the ImportState function correctly handles various scenarios:
// - Empty import ID validation.
// - Read operation support validation.
// - ImportID configuration validation.
// - Successful import state setting.
func TestIdsecResource_ImportState(t *testing.T) {
	tests := []struct {
		name                string
		importID            string
		supportedOperations []actions.IdsecServiceActionOperation
		importIDAttribute   string
		expectedError       string
		expectedAttribute   string
		expectedValue       string
		expectedAttribute2  string // For multi-attribute import success
		expectedValue2      string
		skipIfNoField       bool
		description         string
	}{
		{
			name:                "error_empty_import_id",
			importID:            "",
			supportedOperations: []actions.IdsecServiceActionOperation{actions.ReadOperation},
			importIDAttribute:   "safe_id",
			expectedError:       "Invalid Import ID",
			expectedAttribute:   "",
			expectedValue:       "",
			skipIfNoField:       false,
			description:         "Returns error when import ID is empty",
		},
		{
			name:                "error_no_read_operation",
			importID:            "test-id-123",
			supportedOperations: []actions.IdsecServiceActionOperation{actions.CreateOperation, actions.UpdateOperation},
			importIDAttribute:   "safe_id",
			expectedError:       "Import Not Supported",
			expectedAttribute:   "",
			expectedValue:       "",
			skipIfNoField:       false,
			description:         "Returns error when resource does not support Read operation",
		},
		{
			name:                "error_no_import_id_attribute",
			importID:            "test-id-123",
			supportedOperations: []actions.IdsecServiceActionOperation{actions.ReadOperation},
			importIDAttribute:   "",
			expectedError:       "Import Not Supported",
			expectedAttribute:   "",
			expectedValue:       "",
			skipIfNoField:       true,
			description:         "Returns error when ImportID is not configured",
		},
		{
			name:                "success_with_safe_id",
			importID:            "test-safe-id-123",
			supportedOperations: []actions.IdsecServiceActionOperation{actions.ReadOperation},
			importIDAttribute:   "safe_id",
			expectedError:       "",
			expectedAttribute:   "safe_id",
			expectedValue:       "test-safe-id-123",
			skipIfNoField:       true,
			description:         "Successfully sets safe_id attribute during import",
		},
		{
			name:                "success_with_platform_id",
			importID:            "test-platform-id-456",
			supportedOperations: []actions.IdsecServiceActionOperation{actions.ReadOperation},
			importIDAttribute:   "platform_id",
			expectedError:       "",
			expectedAttribute:   "platform_id",
			expectedValue:       "test-platform-id-456",
			skipIfNoField:       true,
			description:         "Successfully sets platform_id attribute during import",
		},
		{
			name:                "success_with_standard_id",
			importID:            "test-standard-id-789",
			supportedOperations: []actions.IdsecServiceActionOperation{actions.ReadOperation},
			importIDAttribute:   "id",
			expectedError:       "",
			expectedAttribute:   "id",
			expectedValue:       "test-standard-id-789",
			skipIfNoField:       true,
			description:         "Successfully sets id attribute during import",
		},
		{
			name:                "success_with_nested_attribute",
			importID:            "policy-id-123",
			supportedOperations: []actions.IdsecServiceActionOperation{actions.ReadOperation},
			importIDAttribute:   "metadata.policy_id",
			expectedError:       "",
			expectedAttribute:   "metadata.policy_id",
			expectedValue:       "policy-id-123",
			skipIfNoField:       true,
			description:         "Successfully sets nested metadata.policy_id attribute during import",
		},
		// Multi-attribute import success
		{
			name:                "success_multi_attribute_import",
			importID:            "safe-123:user@example.com",
			supportedOperations: []actions.IdsecServiceActionOperation{actions.ReadOperation},
			importIDAttribute:   "safe_id:member_name",
			expectedError:       "",
			expectedAttribute:   "safe_id",
			expectedValue:       "safe-123",
			expectedAttribute2:  "member_name",
			expectedValue2:      "user@example.com",
			skipIfNoField:       true,
			description:         "Successfully sets multiple attributes during import (safe_id:member_name)",
		},
		// Multi-attribute import errors
		{
			name:                "error_multi_attribute_import_id_without_colon",
			importID:            "single-value-only",
			supportedOperations: []actions.IdsecServiceActionOperation{actions.ReadOperation},
			importIDAttribute:   "safe_id:member_name",
			expectedError:       "Invalid Import ID",
			expectedAttribute:   "",
			expectedValue:       "",
			skipIfNoField:       true,
			description:         "Returns error when ImportID has multiple attributes but import ID has no colon",
		},
		{
			name:                "error_multi_attribute_part_count_mismatch",
			importID:            "only-one-value",
			supportedOperations: []actions.IdsecServiceActionOperation{actions.ReadOperation},
			importIDAttribute:   "safe_id:member_name",
			expectedError:       "Invalid Import ID",
			expectedAttribute:   "",
			expectedValue:       "",
			skipIfNoField:       true,
			description:         "Returns error when import ID has fewer parts than required attributes",
		},
		{
			name:                "error_multi_attribute_too_many_parts",
			importID:            "a:b:c",
			supportedOperations: []actions.IdsecServiceActionOperation{actions.ReadOperation},
			importIDAttribute:   "safe_id:member_name",
			expectedError:       "Invalid Import ID",
			expectedAttribute:   "",
			expectedValue:       "",
			skipIfNoField:       true,
			description:         "Returns error when import ID has more parts than required attributes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			ctx := context.Background()
			serviceConfig := CreateTestServiceConfig("test-service")
			var actionDefinition *actions.IdsecServiceTerraformResourceActionDefinition

			// Create action definition based on test case
			if tt.importIDAttribute != "" {
				actionDefinition = CreateTestActionDefinitionWithImportIDAndOperations(
					"test-action",
					"Test action description",
					tt.importIDAttribute,
					tt.supportedOperations,
				)
			} else {
				actionDefinition = CreateTestActionDefinitionWithOperations(
					"test-action",
					"Test action description",
					tt.supportedOperations,
				)
			}

			// Check if ImportID field exists (for skip logic)
			if tt.skipIfNoField {
				val := reflect.ValueOf(actionDefinition).Elem()
				field := val.FieldByName("ImportID")
				if !field.IsValid() {
					t.Skipf("ImportID field not available in this SDK version")
				}
			}

			idsecResource := CreateTestIdsecResource(serviceConfig, actionDefinition)
			idsecRes, ok := idsecResource.(*IdsecResource)
			if !ok {
				t.Fatalf("Failed to cast resource to *IdsecResource")
			}

			// Create request and response
			req := resource.ImportStateRequest{
				ID: tt.importID,
			}
			resp := &resource.ImportStateResponse{}

			// Minimal schema and raw state for SetAttribute during import.
			testSchema, rawValue := buildImportTestState(tt.expectedAttribute, tt.expectedAttribute2)
			resp.State = tfsdk.State{
				Raw:    rawValue,
				Schema: testSchema,
			}

			// Execute
			idsecRes.ImportState(ctx, req, resp)

			// Validate errors
			if tt.expectedError != "" {
				if !resp.Diagnostics.HasError() {
					t.Errorf("Expected error '%s', but no error was returned", tt.expectedError)
					return
				}
				errorFound := false
				for _, diag := range resp.Diagnostics.Errors() {
					if diag.Summary() == tt.expectedError {
						errorFound = true
						break
					}
				}
				if !errorFound {
					t.Errorf("Expected error '%s', but got: %v", tt.expectedError, resp.Diagnostics.Errors())
				}
			} else {
				// Validate success - check that state was set correctly
				if resp.Diagnostics.HasError() {
					t.Errorf("Expected no errors, but got: %v", resp.Diagnostics.Errors())
					return
				}

				// Verify the state attribute(s) were set
				if tt.expectedAttribute != "" {
					assertImportStateString(t, ctx, resp.State, tt.expectedAttribute, tt.expectedValue)
				}
				if tt.expectedAttribute2 != "" {
					assertImportStateString(t, ctx, resp.State, tt.expectedAttribute2, tt.expectedValue2)
				}
			}
		})
	}
}

func assertImportStateString(t *testing.T, ctx context.Context, state tfsdk.State, attributePath, expectedValue string) {
	t.Helper()

	attrPath, err := schemas.ParseImportAttributePath(attributePath)
	if err != nil {
		t.Fatalf("failed to parse attribute path %q: %v", attributePath, err)
	}

	var attrValue types.String
	diags := state.GetAttribute(ctx, attrPath, &attrValue)
	if diags.HasError() {
		t.Fatalf("failed to get attribute %q from state: %v", attributePath, diags.Errors())
	}
	if attrValue.ValueString() != expectedValue {
		t.Fatalf("expected attribute %q to be %q, got %q", attributePath, expectedValue, attrValue.ValueString())
	}
}

type testPrivateData struct {
	data map[string][]byte
}

func (t *testPrivateData) GetKey(_ context.Context, key string) ([]byte, diag.Diagnostics) {
	if t.data == nil {
		t.data = map[string][]byte{}
	}
	return t.data[key], nil
}

func (t *testPrivateData) SetKey(_ context.Context, key string, value []byte) diag.Diagnostics {
	if t.data == nil {
		t.data = map[string][]byte{}
	}
	t.data[key] = value
	return nil
}

func TestIdsecResource_seedUserSetHistoryFromState(t *testing.T) {
	t.Parallel()

	type createUpdateSchema struct {
		ID   string `json:"id,omitempty" mapstructure:"id"`
		Name string `json:"name,omitempty" mapstructure:"name"`
	}
	type stateSchema struct {
		ID     string `json:"id,omitempty" mapstructure:"id"`
		Name   string `json:"name,omitempty" mapstructure:"name"`
		Status string `json:"status,omitempty" mapstructure:"status"`
	}

	actionDef := &actions.IdsecServiceTerraformResourceActionDefinition{
		IdsecServiceBaseTerraformActionDefinition: actions.IdsecServiceBaseTerraformActionDefinition{
			IdsecServiceBaseActionDefinition: actions.IdsecServiceBaseActionDefinition{
				ActionName: "test-action",
				Schemas: map[string]interface{}{
					"create-action": createUpdateSchema{},
					"update-action": createUpdateSchema{},
				},
			},
			StateSchema:               &stateSchema{},
			HistoryComputedAttributes: []string{"status"},
		},
		SupportedOperations: []actions.IdsecServiceActionOperation{actions.CreateOperation, actions.UpdateOperation},
		ActionsMappings: map[actions.IdsecServiceActionOperation]string{
			actions.CreateOperation: "create-action",
			actions.UpdateOperation: "update-action",
		},
		ImportID: "id",
	}
	idsecRes := &IdsecResource{
		actionDefinition: actionDef,
	}

	state := tfsdk.State{
		Schema: schema.Schema{
			Attributes: map[string]schema.Attribute{
				"id":     schema.StringAttribute{},
				"name":   schema.StringAttribute{},
				"status": schema.StringAttribute{},
			},
		},
		Raw: tftypes.NewValue(
			tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"id":     tftypes.String,
					"name":   tftypes.String,
					"status": tftypes.String,
				},
			},
			map[string]tftypes.Value{
				"id":     tftypes.NewValue(tftypes.String, "rid-1"),
				"name":   tftypes.NewValue(tftypes.String, "policy"),
				"status": tftypes.NewValue(tftypes.String, "Active"),
			},
		),
	}

	private := &testPrivateData{data: map[string][]byte{}}
	var diagnostics diag.Diagnostics
	idsecRes.seedUserSetHistoryFromState(context.Background(), &state, private, private, &diagnostics)
	if diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics errors: %v", diagnostics.Errors())
	}
	got := schemas.ReadUserSetPaths(context.Background(), private)
	if !reflect.DeepEqual(got, map[string]bool{"name": true}) {
		t.Fatalf("seeded history = %v, want only name", got)
	}

	existingBlob, err := schemas.MarshalUserSetHistory([]string{"existing_path"}, "1.0.0")
	if err != nil {
		t.Fatalf("MarshalUserSetHistory error = %v", err)
	}
	private.data[schemas.UserSetAttrsPrivateKey] = existingBlob
	idsecRes.seedUserSetHistoryFromState(context.Background(), &state, private, private, &diagnostics)
	got = schemas.ReadUserSetPaths(context.Background(), private)
	if !reflect.DeepEqual(got, map[string]bool{"existing_path": true}) {
		t.Fatalf("existing history should be preserved, got %v", got)
	}
}
