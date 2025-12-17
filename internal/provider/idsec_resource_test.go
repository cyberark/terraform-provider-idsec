// Copyright (c) CyberArk
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/cyberark/idsec-sdk-golang/pkg/models/actions"
	"github.com/cyberark/idsec-sdk-golang/pkg/services"
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
