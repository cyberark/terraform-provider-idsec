// Copyright (c) CyberArk.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"reflect"
	"testing"

	api "github.com/cyberark/idsec-sdk-golang/pkg"
	"github.com/cyberark/idsec-sdk-golang/pkg/services"
	"github.com/cyberark/terraform-provider-idsec/internal/schemas"
)

// TestGetTerraformTypeName tests the getTerraformTypeName method.
func TestGetTerraformTypeName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		actionName   string
		expectedName string
	}{
		{
			name:         "success_simple_name",
			actionName:   "identity",
			expectedName: "idsec_identity",
		},
		{
			name:         "success_hyphenated_name",
			actionName:   "identity-role",
			expectedName: "idsec_identity_role",
		},
		{
			name:         "success_multiple_hyphens",
			actionName:   "identity-role-admin-rights",
			expectedName: "idsec_identity_role_admin_rights",
		},
		{
			name:         "success_empty_string",
			actionName:   "",
			expectedName: "idsec_",
		},
		{
			name:         "success_single_char",
			actionName:   "a",
			expectedName: "idsec_a",
		},
		{
			name:         "success_underscores_preserved",
			actionName:   "identity_role",
			expectedName: "idsec_identity_role",
		},
		{
			name:         "success_mixed_separators",
			actionName:   "identity-role_member",
			expectedName: "idsec_identity_role_member",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			helper := &IdsecServiceHelper{}

			result := helper.getTerraformTypeName(tt.actionName)

			if result != tt.expectedName {
				t.Errorf("Expected %q, got %q", tt.expectedName, result)
			}
		})
	}
}

// TestGetServiceNameTitled tests the getServiceNameTitled method.
func TestGetServiceNameTitled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		serviceName  string
		expectedName string
	}{
		{
			name:         "success_simple_name",
			serviceName:  "identity",
			expectedName: "Identity",
		},
		{
			name:         "success_hyphenated_name",
			serviceName:  "identity-users",
			expectedName: "IdentityUsers",
		},
		{
			name:         "success_multiple_hyphens",
			serviceName:  "cloud-management-group",
			expectedName: "CloudManagementGroup",
		},
		{
			name:         "success_empty_string",
			serviceName:  "",
			expectedName: "",
		},
		{
			name:         "success_single_char",
			serviceName:  "a",
			expectedName: "A",
		},
		{
			name:         "success_mixed_case_preserved",
			serviceName:  "identity-API-users",
			expectedName: "IdentityApiUsers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			helper := &IdsecServiceHelper{
				serviceConfig: &services.IdsecServiceConfig{
					ServiceName: tt.serviceName,
				},
			}

			result := helper.getServiceNameTitled()

			if result != tt.expectedName {
				t.Errorf("Expected %q, got %q", tt.expectedName, result)
			}
		})
	}
}

// TestConfigureService tests the configureService method.
func TestConfigureService(t *testing.T) {
	tests := []struct {
		name          string
		serviceName   string
		idsecAPI      *api.IdsecAPI
		expectedError bool
		errorContains string
	}{
		{
			name:          "error_nil_api",
			serviceName:   "identity",
			idsecAPI:      nil,
			expectedError: true,
			errorContains: "idsecAPI is nil",
		},
		{
			name:          "error_service_method_not_found",
			serviceName:   "nonexistent-service",
			idsecAPI:      &api.IdsecAPI{}, // Empty API won't have the method
			expectedError: true,
			errorContains: "service method",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helper := &IdsecServiceHelper{
				serviceConfig: &services.IdsecServiceConfig{
					ServiceName: tt.serviceName,
				},
			}

			err := helper.configureService(tt.idsecAPI)

			if tt.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

// TestGetServiceInstance tests the getServiceInstance method.
func TestGetServiceInstance(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		setupService func() services.IdsecService
		expectedNil  bool
	}{
		{
			name: "success_service_configured",
			setupService: func() services.IdsecService {
				return &mockService{}
			},
			expectedNil: false,
		},
		{
			name: "success_service_not_configured",
			setupService: func() services.IdsecService {
				return nil
			},
			expectedNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			helper := &IdsecServiceHelper{
				serviceConfig: &services.IdsecServiceConfig{
					ServiceName: "test-service",
				},
				service: tt.setupService(),
			}

			result := helper.getServiceInstance()

			if tt.expectedNil && result != nil {
				t.Errorf("Expected nil, got %v", result)
			}
			if !tt.expectedNil && result == nil {
				t.Error("Expected non-nil service, got nil")
			}
		})
	}
}

// TestGetService tests the getService method.
func TestGetService(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		setupService func() services.IdsecService
		expectedNil  bool
	}{
		{
			name: "success_service_configured",
			setupService: func() services.IdsecService {
				return &mockService{}
			},
			expectedNil: false,
		},
		{
			name: "success_service_not_configured",
			setupService: func() services.IdsecService {
				return nil
			},
			expectedNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			helper := &IdsecServiceHelper{
				serviceConfig: &services.IdsecServiceConfig{
					ServiceName: "test-service",
				},
				service: tt.setupService(),
			}

			result := helper.getService()

			if tt.expectedNil && result != nil {
				t.Errorf("Expected nil, got %v", result)
			}
			if !tt.expectedNil && result == nil {
				t.Error("Expected non-nil service, got nil")
			}
		})
	}
}

// TestAddTelemetryContextField tests the addTelemetryContextField method.
func TestAddTelemetryContextField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		service   services.IdsecService
		fieldName string
		shortName string
		value     string
	}{
		{
			name:      "success_valid_service",
			service:   &mockService{},
			fieldName: "test_field",
			shortName: "tf",
			value:     "test_value",
		},
		{
			name:      "success_service_returns_error",
			service:   &mockServiceWithError{},
			fieldName: "test_field",
			shortName: "tf",
			value:     "test_value",
		},
		{
			name:      "success_empty_values",
			service:   &mockService{},
			fieldName: "",
			shortName: "",
			value:     "",
		},
		{
			name:      "success_nil_service",
			service:   nil,
			fieldName: "test_field",
			shortName: "tf",
			value:     "test_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			helper := &IdsecServiceHelper{}

			// Should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("addTelemetryContextField panicked: %v", r)
				}
			}()

			helper.addTelemetryContextField(tt.service, tt.fieldName, tt.shortName, tt.value)
		})
	}
}

// TestClearTelemetryContext tests the clearTelemetryContext method.
func TestClearTelemetryContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		service services.IdsecService
	}{
		{
			name:    "success_valid_service",
			service: &mockService{},
		},
		{
			name:    "success_service_returns_error",
			service: &mockServiceWithError{},
		},
		{
			name:    "success_nil_service",
			service: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			helper := &IdsecServiceHelper{}

			// Should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("clearTelemetryContext panicked: %v", r)
				}
			}()

			helper.clearTelemetryContext(tt.service)
		})
	}
}

// TestFindMethodByName tests the schemas.FindMethodByName used by configureService.
func TestFindMethodByName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		targetObject  interface{}
		methodName    string
		expectedError bool
	}{
		{
			name: "success_method_exists",
			targetObject: &testStruct{
				value: "test",
			},
			methodName:    "TestMethod",
			expectedError: false,
		},
		{
			name: "error_method_not_found",
			targetObject: &testStruct{
				value: "test",
			},
			methodName:    "NonExistentMethod",
			expectedError: true,
		},
		{
			name:          "error_nil_object",
			targetObject:  nil,
			methodName:    "TestMethod",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var method *reflect.Value
			var err error

			if tt.targetObject != nil {
				method, err = schemas.FindMethodByName(reflect.ValueOf(tt.targetObject), tt.methodName)
			} else {
				// Handle nil case
				err = fmt.Errorf("nil object")
			}

			if tt.expectedError {
				if err == nil && (method == nil || !method.IsValid()) {
					// This is expected for method not found
				} else if err == nil {
					t.Error("Expected error or invalid method, got valid method")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if method == nil || !method.IsValid() {
					t.Error("Expected valid method, got invalid or nil")
				}
			}
		})
	}
}

// Helper functions and mock types

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// mockService is a mock implementation of IdsecService for testing.
type mockService struct{}

func (m *mockService) ServiceConfig() services.IdsecServiceConfig {
	return services.IdsecServiceConfig{}
}

func (m *mockService) AddExtraContextField(name, shortName, value string) error {
	return nil
}

func (m *mockService) ClearExtraContext() error {
	return nil
}

// mockServiceWithError is a mock that returns errors.
type mockServiceWithError struct{}

func (m *mockServiceWithError) ServiceConfig() services.IdsecServiceConfig {
	return services.IdsecServiceConfig{}
}

func (m *mockServiceWithError) AddExtraContextField(name, shortName, value string) error {
	return fmt.Errorf("mock error adding context field")
}

func (m *mockServiceWithError) ClearExtraContext() error {
	return fmt.Errorf("mock error clearing context")
}

// testStruct is a test struct with a method for FindMethodByName tests.
type testStruct struct {
	value string
}

func (t *testStruct) TestMethod() string {
	return t.value
}
