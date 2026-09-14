// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	api "github.com/cyberark/idsec-sdk-golang/pkg"
	"github.com/cyberark/idsec-sdk-golang/pkg/auth"
	models "github.com/cyberark/idsec-sdk-golang/pkg/models"
	authmodels "github.com/cyberark/idsec-sdk-golang/pkg/models/auth"
	"github.com/cyberark/idsec-sdk-golang/pkg/services"
	"github.com/cyberark/terraform-provider-idsec/internal/actions"
)

// fakeAuthenticator is a test double for IdsecAuthenticator that returns a
// scripted sequence of errors, then nil on subsequent calls.
type fakeAuthenticator struct {
	errors []error // errors to return in order; after exhaustion returns nil
	calls  int
}

func (f *fakeAuthenticator) Authenticate(_ *models.IdsecProfile, _ *authmodels.IdsecAuthProfile, _ *authmodels.IdsecSecret, _, _ bool) (*authmodels.IdsecToken, error) {
	if f.calls < len(f.errors) {
		err := f.errors[f.calls]
		f.calls++
		return nil, err
	}
	f.calls++
	return &authmodels.IdsecToken{}, nil
}

// createTestResource creates a test resource with an optional configureService stub.
func createTestResource(configureServiceFn func(*api.IdsecAPI) error) *IdsecResource {
	serviceConfig := &services.IdsecServiceConfig{ServiceName: "test-service"}
	actionDefinition := &actions.IdsecServiceTerraformResourceActionDefinition{
		IdsecServiceBaseTerraformActionDefinition: actions.IdsecServiceBaseTerraformActionDefinition{
			IdsecServiceBaseActionDefinition: actions.IdsecServiceBaseActionDefinition{
				ActionName: "test-resource",
			},
		},
	}
	return &IdsecResource{
		IdsecServiceHelper: IdsecServiceHelper{
			serviceConfig:      serviceConfig,
			configureServiceFn: configureServiceFn,
		},
		serviceConfig:    serviceConfig,
		actionDefinition: actionDefinition,
	}
}

// createTestDataSource creates a test data source with an optional configureService stub.
func createTestDataSource(configureServiceFn func(*api.IdsecAPI) error) *IdsecDataSource {
	serviceConfig := &services.IdsecServiceConfig{ServiceName: "test-service"}
	actionDefinition := &actions.IdsecServiceTerraformDataSourceActionDefinition{
		IdsecServiceBaseTerraformActionDefinition: actions.IdsecServiceBaseTerraformActionDefinition{
			IdsecServiceBaseActionDefinition: actions.IdsecServiceBaseActionDefinition{
				ActionName: "test-datasource",
			},
		},
	}
	return &IdsecDataSource{
		IdsecServiceHelper: IdsecServiceHelper{
			serviceConfig:      serviceConfig,
			configureServiceFn: configureServiceFn,
		},
		serviceConfig:    serviceConfig,
		actionDefinition: actionDefinition,
	}
}

// createTestResourceForAuth creates a test resource without a configureService stub.
// Use createTestResource(nil) or createTestResource(fn) for new tests.
func createTestResourceForAuth() *IdsecResource { return createTestResource(nil) }

// createTestDataSourceForAuth creates a test data source without a configureService stub.
func createTestDataSourceForAuth() *IdsecDataSource { return createTestDataSource(nil) }

// successfulConfigureService is a configureService stub that always succeeds.
func successfulConfigureService(_ *api.IdsecAPI) error { return nil }

// TestIdsecResource_Configure tests the type-assertion gate in resource Configure.
func TestIdsecResource_Configure(t *testing.T) {
	tests := []struct {
		name          string
		providerData  interface{}
		expectError   bool
		errorContains string
	}{
		{
			name:         "nil_provider_data_returns_early",
			providerData: nil,
			expectError:  false,
		},
		{
			name:          "isp_auth_rejected",
			providerData:  auth.NewIdsecISPAuth(false),
			expectError:   true,
			errorContains: "Unexpected Provider Data",
		},
		{
			name:          "pvwa_auth_rejected",
			providerData:  auth.NewIdsecPVWAAuth(false),
			expectError:   true,
			errorContains: "Unexpected Provider Data",
		},
		{
			name:          "string_rejected",
			providerData:  "invalid",
			expectError:   true,
			errorContains: "Unexpected Provider Data",
		},
		{
			name:          "int_rejected",
			providerData:  12345,
			expectError:   true,
			errorContains: "Unexpected Provider Data",
		},
		{
			name:         "idsec_api_accepted_and_assigned",
			providerData: &api.IdsecAPI{},
			expectError:  false, // configureService stub succeeds
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			// Use a successful stub so the accepted case doesn't error on service init.
			r := createTestResource(successfulConfigureService)
			req := resource.ConfigureRequest{ProviderData: tt.providerData}
			resp := &resource.ConfigureResponse{}

			r.Configure(ctx, req, resp)

			if tt.expectError {
				if !resp.Diagnostics.HasError() {
					t.Errorf("Expected error '%s' but got none", tt.errorContains)
					return
				}
				for _, diag := range resp.Diagnostics.Errors() {
					if diag.Summary() == tt.errorContains {
						return
					}
				}
				t.Errorf("Expected error '%s', got: %v", tt.errorContains, resp.Diagnostics.Errors())
			} else {
				if resp.Diagnostics.HasError() {
					t.Errorf("Expected no error, got: %v", resp.Diagnostics.Errors())
				}
				// For the success case with a real IdsecAPI, verify assignment happened.
				if tt.providerData != nil {
					if r.idsecAPI == nil {
						t.Errorf("Expected idsecAPI to be assigned after successful Configure")
					}
				}
			}
		})
	}
}

// TestIdsecDataSource_Configure tests the type-assertion gate in data source Configure.
func TestIdsecDataSource_Configure(t *testing.T) {
	tests := []struct {
		name          string
		providerData  interface{}
		expectError   bool
		errorContains string
	}{
		{
			name:         "nil_provider_data_returns_early",
			providerData: nil,
			expectError:  false,
		},
		{
			name:          "isp_auth_rejected",
			providerData:  auth.NewIdsecISPAuth(false),
			expectError:   true,
			errorContains: "Unexpected Provider Data",
		},
		{
			name:          "pvwa_auth_rejected",
			providerData:  auth.NewIdsecPVWAAuth(false),
			expectError:   true,
			errorContains: "Unexpected Provider Data",
		},
		{
			name:          "string_rejected",
			providerData:  "invalid",
			expectError:   true,
			errorContains: "Unexpected Provider Data",
		},
		{
			name:          "int_rejected",
			providerData:  12345,
			expectError:   true,
			errorContains: "Unexpected Provider Data",
		},
		{
			name:         "idsec_api_accepted_and_assigned",
			providerData: &api.IdsecAPI{},
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			ds := createTestDataSource(successfulConfigureService)
			req := datasource.ConfigureRequest{ProviderData: tt.providerData}
			resp := &datasource.ConfigureResponse{}

			ds.Configure(ctx, req, resp)

			if tt.expectError {
				if !resp.Diagnostics.HasError() {
					t.Errorf("Expected error '%s' but got none", tt.errorContains)
					return
				}
				for _, diag := range resp.Diagnostics.Errors() {
					if diag.Summary() == tt.errorContains {
						return
					}
				}
				t.Errorf("Expected error '%s', got: %v", tt.errorContains, resp.Diagnostics.Errors())
			} else {
				if resp.Diagnostics.HasError() {
					t.Errorf("Expected no error, got: %v", resp.Diagnostics.Errors())
				}
				if tt.providerData != nil {
					if ds.idsecAPI == nil {
						t.Errorf("Expected idsecAPI to be assigned after successful Configure")
					}
				}
			}
		})
	}
}

// TestIdsecResource_Configure_RetryBehavior tests the retry inversion in resource Configure:
// - non-retryable errors stop immediately and surface as "Service Configuration Error".
// - "unexpected end of JSON input" retries until exhausted, then surfaces as "Service Configuration Error".
// - a transient error that clears on retry results in success.
func TestIdsecResource_Configure_RetryBehavior(t *testing.T) {
	tests := []struct {
		name            string
		configureErrs   []error // sequence returned by stub; nil entry = success
		expectError     bool
		expectedSummary string
		expectedCalls   int
	}{
		{
			name:            "non_retryable_error_stops_after_one_attempt",
			configureErrs:   []error{fmt.Errorf("connection refused")},
			expectError:     true,
			expectedSummary: "Service Configuration Error",
			expectedCalls:   1,
		},
		{
			name: "retryable_json_error_exhausts_all_attempts",
			configureErrs: []error{
				fmt.Errorf("unexpected end of JSON input"),
				fmt.Errorf("unexpected end of JSON input"),
				fmt.Errorf("unexpected end of JSON input"),
				fmt.Errorf("unexpected end of JSON input"),
				fmt.Errorf("unexpected end of JSON input"),
			},
			expectError:     true,
			expectedSummary: "Service Configuration Error",
			expectedCalls:   5,
		},
		{
			name: "transient_json_error_clears_on_retry",
			configureErrs: []error{
				fmt.Errorf("unexpected end of JSON input"),
				nil, // succeeds on second attempt
			},
			expectError:   false,
			expectedCalls: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			calls := 0
			errs := tt.configureErrs
			stub := func(_ *api.IdsecAPI) error {
				var err error
				if calls < len(errs) {
					err = errs[calls]
				}
				calls++
				return err
			}

			ctx := context.Background()
			r := createTestResource(stub)
			req := resource.ConfigureRequest{ProviderData: &api.IdsecAPI{}}
			resp := &resource.ConfigureResponse{}

			r.Configure(ctx, req, resp)

			if tt.expectError {
				if !resp.Diagnostics.HasError() {
					t.Errorf("Expected '%s' error but got none", tt.expectedSummary)
				} else {
					found := false
					for _, diag := range resp.Diagnostics.Errors() {
						if diag.Summary() == tt.expectedSummary {
							found = true
						}
					}
					if !found {
						t.Errorf("Expected error summary '%s', got: %v", tt.expectedSummary, resp.Diagnostics.Errors())
					}
				}
			} else {
				if resp.Diagnostics.HasError() {
					t.Errorf("Expected no error, got: %v", resp.Diagnostics.Errors())
				}
			}
			if calls != tt.expectedCalls {
				t.Errorf("Expected %d configureService calls, got %d", tt.expectedCalls, calls)
			}
		})
	}
}

// TestAuthInterfaceTypeAssertion tests that both ISP and PVWA auth types implement auth.IdsecAuth.
func TestAuthInterfaceTypeAssertion(t *testing.T) {
	tests := []struct {
		name         string
		authProvider auth.IdsecAuth
		expectedName string
	}{
		{
			name:         "isp_auth_implements_interface",
			authProvider: auth.NewIdsecISPAuth(false),
			expectedName: "isp",
		},
		{
			name:         "pvwa_auth_implements_interface",
			authProvider: auth.NewIdsecPVWAAuth(false),
			expectedName: "pvwa",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.authProvider == nil {
				t.Fatalf("Expected auth provider to be non-nil")
			}
			if tt.authProvider.AuthenticatorName() != tt.expectedName {
				t.Errorf("Expected authenticator name '%s', got '%s'", tt.expectedName, tt.authProvider.AuthenticatorName())
			}
			if _, ok := interface{}(tt.authProvider).(auth.IdsecAuth); !ok {
				t.Errorf("Expected auth provider to implement auth.IdsecAuth interface")
			}
		})
	}
}

// TestIdsecAPI_ProviderData verifies that *api.IdsecAPI is accepted by Configure
// and that the error is "Service Configuration Error" (type assertion succeeded),
// not "Unexpected Provider Data" (type assertion failed).
func TestIdsecAPI_ProviderData(t *testing.T) {
	idsecAPI := &api.IdsecAPI{}

	t.Run("resource_accepts_idsec_api_type", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		r := createTestResourceForAuth() // no stub — real configureService, expected to fail
		req := resource.ConfigureRequest{ProviderData: idsecAPI}
		resp := &resource.ConfigureResponse{}

		r.Configure(ctx, req, resp)

		for _, diag := range resp.Diagnostics.Errors() {
			if diag.Summary() == "Unexpected Provider Data" {
				t.Errorf("Got 'Unexpected Provider Data' for *api.IdsecAPI: %s", diag.Detail())
			}
		}
		hasServiceError := false
		for _, diag := range resp.Diagnostics.Errors() {
			if diag.Summary() == "Service Configuration Error" {
				hasServiceError = true
			}
		}
		if !hasServiceError {
			t.Errorf("Expected 'Service Configuration Error' for zero-value IdsecAPI, got: %v", resp.Diagnostics.Errors())
		}
	})

	t.Run("datasource_accepts_idsec_api_type", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		ds := createTestDataSourceForAuth()
		req := datasource.ConfigureRequest{ProviderData: idsecAPI}
		resp := &datasource.ConfigureResponse{}

		ds.Configure(ctx, req, resp)

		for _, diag := range resp.Diagnostics.Errors() {
			if diag.Summary() == "Unexpected Provider Data" {
				t.Errorf("Got 'Unexpected Provider Data' for *api.IdsecAPI: %s", diag.Detail())
			}
		}
		hasServiceError := false
		for _, diag := range resp.Diagnostics.Errors() {
			if diag.Summary() == "Service Configuration Error" {
				hasServiceError = true
			}
		}
		if !hasServiceError {
			t.Errorf("Expected 'Service Configuration Error' for zero-value IdsecAPI, got: %v", resp.Diagnostics.Errors())
		}
	})
}

// TestIdsecAPI_FlowsThroughProviderLayer verifies that *api.IdsecAPI flows through
// both resource and data source Configure without producing "Unexpected Provider Data",
// and that idsecAPI is assigned on the resource/datasource.
// Replaces TestBothAuthTypesPassProviderLayer.
func TestIdsecAPI_FlowsThroughProviderLayer(t *testing.T) {
	idsecAPI := &api.IdsecAPI{}

	t.Run("resource", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		r := createTestResource(successfulConfigureService)
		req := resource.ConfigureRequest{ProviderData: idsecAPI}
		resp := &resource.ConfigureResponse{}

		r.Configure(ctx, req, resp)

		for _, diag := range resp.Diagnostics.Errors() {
			if diag.Summary() == "Unexpected Provider Data" {
				t.Errorf("*api.IdsecAPI rejected by resource Configure: %s", diag.Detail())
			}
		}
		if resp.Diagnostics.HasError() {
			t.Errorf("Expected no error, got: %v", resp.Diagnostics.Errors())
		}
		if r.idsecAPI != idsecAPI {
			t.Errorf("Expected idsecAPI to be assigned on resource")
		}
	})

	t.Run("datasource", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		ds := createTestDataSource(successfulConfigureService)
		req := datasource.ConfigureRequest{ProviderData: idsecAPI}
		resp := &datasource.ConfigureResponse{}

		ds.Configure(ctx, req, resp)

		for _, diag := range resp.Diagnostics.Errors() {
			if diag.Summary() == "Unexpected Provider Data" {
				t.Errorf("*api.IdsecAPI rejected by datasource Configure: %s", diag.Detail())
			}
		}
		if resp.Diagnostics.HasError() {
			t.Errorf("Expected no error, got: %v", resp.Diagnostics.Errors())
		}
		if ds.idsecAPI != idsecAPI {
			t.Errorf("Expected idsecAPI to be assigned on datasource")
		}
	})
}

// TestAuthenticateWithRetry tests the retry inversion in authenticateWithRetry:
// - retryable errors ("invalid keyring", "unexpected end of JSON input") are retried.
// - non-retryable errors stop immediately.
// - a transient retryable error that clears on retry results in success.
func TestAuthenticateWithRetry(t *testing.T) {
	tests := []struct {
		name          string
		authErrors    []error // sequence; after exhaustion Authenticate returns nil
		expectError   bool
		expectedCalls int
	}{
		{
			name:          "non_retryable_error_stops_immediately",
			authErrors:    []error{fmt.Errorf("bad credentials")},
			expectError:   true,
			expectedCalls: 1,
		},
		{
			name: "invalid_keyring_is_retried",
			authErrors: []error{
				fmt.Errorf("invalid keyring"),
				fmt.Errorf("invalid keyring"),
				nil,
			},
			expectError:   false,
			expectedCalls: 3,
		},
		{
			name: "unexpected_json_is_retried",
			authErrors: []error{
				fmt.Errorf("unexpected end of JSON input"),
				nil,
			},
			expectError:   false,
			expectedCalls: 2,
		},
		{
			name: "retryable_error_exhausts_all_attempts",
			authErrors: []error{
				fmt.Errorf("invalid keyring"),
				fmt.Errorf("invalid keyring"),
				fmt.Errorf("invalid keyring"),
			},
			expectError:   true,
			expectedCalls: 3, // authRetryCount = 3
		},
		{
			name:          "success_on_first_attempt",
			authErrors:    nil,
			expectError:   false,
			expectedCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fake := &fakeAuthenticator{errors: tt.authErrors}
			p := &IdsecProvider{}
			creds := &authCredentials{
				userName:   "test",
				secret:     "secret",
				authMethod: "test",
			}

			err := p.authenticateWithRetry(context.Background(), fake, creds, "test")

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
			if fake.calls != tt.expectedCalls {
				t.Errorf("Expected %d Authenticate calls, got %d", tt.expectedCalls, fake.calls)
			}
		})
	}
}
