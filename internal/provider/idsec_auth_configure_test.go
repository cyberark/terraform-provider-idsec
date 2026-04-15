// Copyright (c) CyberArk
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/cyberark/idsec-sdk-golang/pkg/auth"
	"github.com/cyberark/idsec-sdk-golang/pkg/services"
	"github.com/cyberark/terraform-provider-idsec/internal/actions"
)

// createTestResourceForAuth creates a test resource for authentication testing.
func createTestResourceForAuth() *IdsecResource {
	serviceConfig := &services.IdsecServiceConfig{
		ServiceName: "test-service",
	}
	actionDefinition := &actions.IdsecServiceTerraformResourceActionDefinition{
		IdsecServiceBaseTerraformActionDefinition: actions.IdsecServiceBaseTerraformActionDefinition{
			IdsecServiceBaseActionDefinition: actions.IdsecServiceBaseActionDefinition{
				ActionName:        "test-resource",
				ActionDescription: "Test resource for auth testing",
			},
		},
	}
	return &IdsecResource{
		IdsecServiceHelper: IdsecServiceHelper{
			serviceConfig: serviceConfig,
		},
		serviceConfig:    serviceConfig,
		actionDefinition: actionDefinition,
	}
}

// createTestDataSourceForAuth creates a test data source for authentication testing.
func createTestDataSourceForAuth() *IdsecDataSource {
	serviceConfig := &services.IdsecServiceConfig{
		ServiceName: "test-service",
	}
	actionDefinition := &actions.IdsecServiceTerraformDataSourceActionDefinition{
		IdsecServiceBaseTerraformActionDefinition: actions.IdsecServiceBaseTerraformActionDefinition{
			IdsecServiceBaseActionDefinition: actions.IdsecServiceBaseActionDefinition{
				ActionName:        "test-datasource",
				ActionDescription: "Test data source for auth testing",
			},
		},
	}
	return &IdsecDataSource{
		IdsecServiceHelper: IdsecServiceHelper{
			serviceConfig: serviceConfig,
		},
		serviceConfig:    serviceConfig,
		actionDefinition: actionDefinition,
	}
}

// TestIdsecResource_Configure_ISPAuth tests that ISP authentication continues to work
// with resources after the interface-based authentication fix.
//
// This is a regression test to verify that the existing ISP authentication flow
// (identity, identity_service_user) is not broken by the interface change.
func TestIdsecResource_Configure_ISPAuth(t *testing.T) {
	tests := []struct {
		name          string
		providerData  interface{}
		expectError   bool
		errorContains string
		description   string
	}{
		{
			name:         "success_isp_auth_pointer_accepted",
			providerData: auth.NewIdsecISPAuth(false),
			expectError:  false,
			description:  "ISP authentication should be accepted by resource Configure",
		},
		{
			name:         "success_nil_provider_data",
			providerData: nil,
			expectError:  false, // nil provider data returns early, no error
			description:  "Nil provider data should return early without error",
		},
		{
			name:          "error_invalid_auth_type_string",
			providerData:  "invalid_auth",
			expectError:   true,
			errorContains: "Authentication Error",
			description:   "Invalid auth type (string) should produce authentication error",
		},
		{
			name:          "error_invalid_auth_type_int",
			providerData:  12345,
			expectError:   true,
			errorContains: "Authentication Error",
			description:   "Invalid auth type (int) should produce authentication error",
		},
		{
			name:          "error_invalid_auth_type_struct",
			providerData:  struct{ Name string }{"invalid"},
			expectError:   true,
			errorContains: "Authentication Error",
			description:   "Invalid auth type (struct) should produce authentication error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			idsecResource := createTestResourceForAuth()

			req := resource.ConfigureRequest{
				ProviderData: tt.providerData,
			}
			resp := &resource.ConfigureResponse{}

			idsecResource.Configure(ctx, req, resp)

			if tt.expectError {
				if !resp.Diagnostics.HasError() {
					t.Errorf("Expected error containing '%s', but no error was returned", tt.errorContains)
					return
				}
				errorFound := false
				for _, diag := range resp.Diagnostics.Errors() {
					if diag.Summary() == tt.errorContains {
						errorFound = true
						break
					}
				}
				if !errorFound {
					t.Errorf("Expected error containing '%s', but got: %v", tt.errorContains, resp.Diagnostics.Errors())
				}
			} else {
				if resp.Diagnostics.HasError() {
					// Only check for unexpected errors if providerData is not nil
					// (nil providerData returns early, non-nil but invalid types may error)
					if tt.providerData != nil {
						// Check if error is due to API/service initialization (expected, no real auth or service)
						hasExpectedError := false
						for _, diag := range resp.Diagnostics.Errors() {
							if diag.Summary() == "Service Initialization Error" || diag.Summary() == "Service Configuration Error" {
								hasExpectedError = true
								break
							}
						}
						if !hasExpectedError {
							t.Errorf("Expected no errors (or Service Initialization/Configuration Error), but got: %v", resp.Diagnostics.Errors())
						}
					}
				}
			}
		})
	}
}

// TestIdsecResource_Configure_PVWAAuth tests that PVWA authentication passes the
// provider layer correctly.
//
// This test verifies that PVWA authentication is properly accepted by resources
// using the interface-based type assertion.
func TestIdsecResource_Configure_PVWAAuth(t *testing.T) {
	tests := []struct {
		name          string
		providerData  interface{}
		expectError   bool
		errorContains string
		description   string
	}{
		{
			name:         "success_pvwa_auth_pointer_accepted",
			providerData: auth.NewIdsecPVWAAuth(false),
			expectError:  false,
			description:  "PVWA authentication should be accepted by resource Configure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			idsecResource := createTestResourceForAuth()

			req := resource.ConfigureRequest{
				ProviderData: tt.providerData,
			}
			resp := &resource.ConfigureResponse{}

			idsecResource.Configure(ctx, req, resp)

			if tt.expectError {
				if !resp.Diagnostics.HasError() {
					t.Errorf("Expected error containing '%s', but no error was returned", tt.errorContains)
					return
				}
				errorFound := false
				for _, diag := range resp.Diagnostics.Errors() {
					if diag.Summary() == tt.errorContains {
						errorFound = true
						break
					}
				}
				if !errorFound {
					t.Errorf("Expected error containing '%s', but got: %v", tt.errorContains, resp.Diagnostics.Errors())
				}
			} else {
				if resp.Diagnostics.HasError() {
					// Check if error is due to API/service initialization (expected, no real auth or service)
					hasExpectedError := false
					for _, diag := range resp.Diagnostics.Errors() {
						if diag.Summary() == "Service Initialization Error" || diag.Summary() == "Service Configuration Error" {
							hasExpectedError = true
							break
						}
					}
					if !hasExpectedError {
						t.Errorf("Expected no errors (or Service Initialization/Configuration Error), but got: %v", resp.Diagnostics.Errors())
					}
				}
			}
		})
	}
}

// TestIdsecDataSource_Configure_ISPAuth tests that ISP authentication continues to work
// with data sources after the interface-based authentication fix.
//
// This is a regression test to verify that the existing ISP authentication flow
// is not broken by the interface change for data sources.
func TestIdsecDataSource_Configure_ISPAuth(t *testing.T) {
	tests := []struct {
		name          string
		providerData  interface{}
		expectError   bool
		errorContains string
		description   string
	}{
		{
			name:         "success_isp_auth_pointer_accepted",
			providerData: auth.NewIdsecISPAuth(false),
			expectError:  false,
			description:  "ISP authentication should be accepted by data source Configure",
		},
		{
			name:         "success_nil_provider_data",
			providerData: nil,
			expectError:  false, // nil provider data returns early, no error
			description:  "Nil provider data should return early without error",
		},
		{
			name:          "error_invalid_auth_type_string",
			providerData:  "invalid_auth",
			expectError:   true,
			errorContains: "Authentication Error",
			description:   "Invalid auth type (string) should produce authentication error",
		},
		{
			name:          "error_invalid_auth_type_int",
			providerData:  12345,
			expectError:   true,
			errorContains: "Authentication Error",
			description:   "Invalid auth type (int) should produce authentication error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			idsecDataSource := createTestDataSourceForAuth()

			req := datasource.ConfigureRequest{
				ProviderData: tt.providerData,
			}
			resp := &datasource.ConfigureResponse{}

			idsecDataSource.Configure(ctx, req, resp)

			if tt.expectError {
				if !resp.Diagnostics.HasError() {
					t.Errorf("Expected error containing '%s', but no error was returned", tt.errorContains)
					return
				}
				errorFound := false
				for _, diag := range resp.Diagnostics.Errors() {
					if diag.Summary() == tt.errorContains {
						errorFound = true
						break
					}
				}
				if !errorFound {
					t.Errorf("Expected error containing '%s', but got: %v", tt.errorContains, resp.Diagnostics.Errors())
				}
			} else {
				if resp.Diagnostics.HasError() {
					if tt.providerData != nil {
						// Check if error is due to API/service initialization (expected, no real auth or service)
						hasExpectedError := false
						for _, diag := range resp.Diagnostics.Errors() {
							if diag.Summary() == "Service Initialization Error" || diag.Summary() == "Service Configuration Error" {
								hasExpectedError = true
								break
							}
						}
						if !hasExpectedError {
							t.Errorf("Expected no errors (or Service Initialization/Configuration Error), but got: %v", resp.Diagnostics.Errors())
						}
					}
				}
			}
		})
	}
}

// TestIdsecDataSource_Configure_PVWAAuth tests that PVWA authentication passes the
// provider layer correctly for data sources.
//
// This test verifies that PVWA authentication is properly accepted by data sources
// using the interface-based type assertion.
func TestIdsecDataSource_Configure_PVWAAuth(t *testing.T) {
	tests := []struct {
		name          string
		providerData  interface{}
		expectError   bool
		errorContains string
		description   string
	}{
		{
			name:         "success_pvwa_auth_pointer_accepted",
			providerData: auth.NewIdsecPVWAAuth(false),
			expectError:  false,
			description:  "PVWA authentication should be accepted by data source Configure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			idsecDataSource := createTestDataSourceForAuth()

			req := datasource.ConfigureRequest{
				ProviderData: tt.providerData,
			}
			resp := &datasource.ConfigureResponse{}

			idsecDataSource.Configure(ctx, req, resp)

			if tt.expectError {
				if !resp.Diagnostics.HasError() {
					t.Errorf("Expected error containing '%s', but no error was returned", tt.errorContains)
					return
				}
				errorFound := false
				for _, diag := range resp.Diagnostics.Errors() {
					if diag.Summary() == tt.errorContains {
						errorFound = true
						break
					}
				}
				if !errorFound {
					t.Errorf("Expected error containing '%s', but got: %v", tt.errorContains, resp.Diagnostics.Errors())
				}
			} else {
				if resp.Diagnostics.HasError() {
					// Check if error is due to API/service initialization (expected, no real auth or service)
					hasExpectedError := false
					for _, diag := range resp.Diagnostics.Errors() {
						if diag.Summary() == "Service Initialization Error" || diag.Summary() == "Service Configuration Error" {
							hasExpectedError = true
							break
						}
					}
					if !hasExpectedError {
						t.Errorf("Expected no errors (or Service Initialization/Configuration Error), but got: %v", resp.Diagnostics.Errors())
					}
				}
			}
		})
	}
}

// TestAuthInterfaceTypeAssertion tests that both ISP and PVWA auth types
// correctly implement the auth.IdsecAuth interface and can be used interchangeably.
//
// This test validates the core assumption of the fix: that interface-based type
// assertion allows both auth types to pass through the provider layer.
func TestAuthInterfaceTypeAssertion(t *testing.T) {
	tests := []struct {
		name         string
		authProvider auth.IdsecAuth
		expectedName string
		description  string
	}{
		{
			name:         "success_isp_auth_implements_interface",
			authProvider: auth.NewIdsecISPAuth(false),
			expectedName: "isp",
			description:  "IdsecISPAuth should implement auth.IdsecAuth interface",
		},
		{
			name:         "success_pvwa_auth_implements_interface",
			authProvider: auth.NewIdsecPVWAAuth(false),
			expectedName: "pvwa",
			description:  "IdsecPVWAAuth should implement auth.IdsecAuth interface",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Verify the auth provider is not nil
			if tt.authProvider == nil {
				t.Errorf("Expected auth provider to be non-nil")
				return
			}

			// Verify the authenticator name matches expected
			authName := tt.authProvider.AuthenticatorName()
			if authName != tt.expectedName {
				t.Errorf("Expected authenticator name '%s', got '%s'", tt.expectedName, authName)
			}

			// Verify interface type assertion works
			_, ok := interface{}(tt.authProvider).(auth.IdsecAuth)
			if !ok {
				t.Errorf("Expected auth provider to implement auth.IdsecAuth interface")
			}
		})
	}
}

// TestBothAuthTypesPassProviderLayer verifies that both ISP and PVWA authentication
// types can be passed through the provider layer to resources and data sources.
//
// This is the key integration test that validates the fix works correctly
// for both authentication methods.
func TestBothAuthTypesPassProviderLayer(t *testing.T) {
	authTypes := []struct {
		name         string
		authProvider auth.IdsecAuth
	}{
		{
			name:         "isp_auth",
			authProvider: auth.NewIdsecISPAuth(false),
		},
		{
			name:         "pvwa_auth",
			authProvider: auth.NewIdsecPVWAAuth(false),
		},
	}

	for _, authType := range authTypes {
		t.Run("resource_"+authType.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			idsecResource := createTestResourceForAuth()

			req := resource.ConfigureRequest{
				ProviderData: authType.authProvider,
			}
			resp := &resource.ConfigureResponse{}

			idsecResource.Configure(ctx, req, resp)

			// Check that there's no Authentication Error
			// (Service Initialization Error is expected since we don't have real credentials)
			for _, diag := range resp.Diagnostics.Errors() {
				if diag.Summary() == "Authentication Error" {
					t.Errorf("Authentication Error occurred for %s: %s", authType.name, diag.Detail())
				}
			}
		})

		t.Run("datasource_"+authType.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			idsecDataSource := createTestDataSourceForAuth()

			req := datasource.ConfigureRequest{
				ProviderData: authType.authProvider,
			}
			resp := &datasource.ConfigureResponse{}

			idsecDataSource.Configure(ctx, req, resp)

			// Check that there's no Authentication Error
			// (Service Initialization Error is expected since we don't have real credentials)
			for _, diag := range resp.Diagnostics.Errors() {
				if diag.Summary() == "Authentication Error" {
					t.Errorf("Authentication Error occurred for %s: %s", authType.name, diag.Detail())
				}
			}
		})
	}
}
