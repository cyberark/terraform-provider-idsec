// Copyright (c) CyberArk.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	terraformprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/cyberark/idsec-sdk-golang/pkg/auth"
	sdkconfig "github.com/cyberark/idsec-sdk-golang/pkg/config"
	"github.com/cyberark/idsec-sdk-golang/pkg/models"
	authmodels "github.com/cyberark/idsec-sdk-golang/pkg/models/auth"
	"github.com/cyberark/idsec-sdk-golang/pkg/services"
	provideractions "github.com/cyberark/terraform-provider-idsec/internal/actions"
	"github.com/cyberark/terraform-provider-idsec/internal/schemas"
	_ "github.com/cyberark/terraform-provider-idsec/internal/tfactions"
)

// Environment variables for Idsec provider configuration.
const (
	// IdsecCacheAuthenticationEnvVar Environment variable decides whether to cache authentication for idsec or not.
	IdsecCacheAuthenticationEnvVar = "IDSEC_CACHE_AUTHENTICATION"
	// IdsecCacheAuthenticationDefault Default value for cache authentication.
	IdsecCacheAuthenticationDefault = true

	// IdsecAuthMethodEnvVar Environment variable for authentication method, e.g., identity, identity_service_user.
	IdsecAuthMethodEnvVar = "IDSEC_AUTH_METHOD"

	// IdsecSubdomainEnvVar Environment variable for tenant subdomain.
	IdsecSubdomainEnvVar = "IDSEC_SUBDOMAIN"

	// IdsecUsernameEnvVar Environment variable for username, used for identity authentication.
	IdsecUsernameEnvVar = "IDSEC_USERNAME"

	// IdsecSecretEnvVar Environment variable for secret, used for identity authentication.
	IdsecSecretEnvVar = "IDSEC_SECRET"

	// IdsecServiceUserEnvVar Environment variable for service user, used for identity service user authentication.
	IdsecServiceUserEnvVar = "IDSEC_SERVICE_USER"

	// IdsecServiceTokenEnvVar Environment variable for service token, used for identity service user authentication.
	IdsecServiceTokenEnvVar = "IDSEC_SERVICE_TOKEN" // #nosec G101

	// IdsecServiceAuthorizedAppEnvVar Environment variable for authorized application, used for identity service user authentication.
	IdsecServiceAuthorizedAppEnvVar = "IDSEC_SERVICE_AUTHORIZED_APP"
	// IdsecServiceAuthorizedAppDefault Default value for authorized application.
	IdsecServiceAuthorizedAppDefault = "__idaptive_cybr_user_oidc"

	// IdsecPVWAURLEnvVar Environment variable for PVWA URL, used for PVWA authentication.
	IdsecPVWAURLEnvVar = "IDSEC_PVWA_URL"

	// IdsecPVWALoginMethodEnvVar Environment variable for PVWA login method, used for PVWA authentication.
	IdsecPVWALoginMethodEnvVar = "IDSEC_PVWA_LOGIN_METHOD"

	// IdsecPVWALoginMethodDefault Default value for PVWA login method.
	IdsecPVWALoginMethodDefault = "cyberark"
)

const (
	authRetryCount = 3
)

var (
	authRetryableErrrors = []string{
		"invalid keyring",
	}
)

// Ensure IdsecProvider satisfies various provider interfaces.
var _ terraformprovider.Provider = &IdsecProvider{}

// providerVersion holds the version of the Terraform provider.
// This is set during provider configuration and used by resources and data sources for telemetry.
var providerVersion string

// IdsecProviderSchema defines the schema for the Idsec provider configuration.
type IdsecProviderSchema struct {
	AuthMethod           types.String `tfsdk:"auth_method"`
	UserName             types.String `tfsdk:"username"`
	Secret               types.String `tfsdk:"secret"`
	ServiceUser          types.String `tfsdk:"service_user"`
	ServiceToken         types.String `tfsdk:"service_token"`
	ServiceAuthorizedApp types.String `tfsdk:"service_authorized_app"`
	Subdomain            types.String `tfsdk:"subdomain"`
	CacheAuthentication  types.Bool   `tfsdk:"cache_authentication"`
	PVWAURL              types.String `tfsdk:"pvwa_url"`
	PVWALoginMethod      types.String `tfsdk:"pvwa_login_method"`
}

// IdsecProviderConfig holds the configuration for the Idsec provider.
type IdsecProviderConfig struct {
	Version   string `json:"version" mapstructure:"version"`
	GitCommit string `json:"git_commit" mapstructure:"git_commit"`
	BuildDate string `json:"build_date" mapstructure:"build_date"`
}

// IdsecProvider is the main struct for the Idsec provider.
type IdsecProvider struct {
	terraformprovider.Provider
	ispAuth  *auth.IdsecISPAuth
	pvwaAuth *auth.IdsecPVWAAuth
	config   IdsecProviderConfig
}

// NewIdsecProvider creates a new instance of the Idsec provider.
func NewIdsecProvider(config IdsecProviderConfig) func() terraformprovider.Provider {
	return func() terraformprovider.Provider {
		return &IdsecProvider{
			config: config,
		}
	}
}

func (p *IdsecProvider) resolveTerraformStringVar(variable types.String, envVar string) types.String {
	if variable.IsNull() {
		if val, ok := os.LookupEnv(envVar); ok {
			return types.StringValue(val)
		}
	}
	return variable
}

func (p *IdsecProvider) resolveTerraformBoolVar(variable types.Bool, envVar string, defaultVal bool) types.Bool {
	if variable.IsNull() {
		if val, ok := os.LookupEnv(envVar); ok {
			boolVal, err := strconv.ParseBool(val)
			if err != nil {
				return variable
			}
			return types.BoolValue(boolVal)
		}
		return types.BoolValue(defaultVal)
	}
	return variable
}

// authCredentials holds the parsed authentication credentials.
type authCredentials struct {
	userName           string
	secret             string
	authMethod         authmodels.IdsecAuthMethod
	authMethodSettings authmodels.IdsecAuthMethodSettings
}

// IdsecAuthenticator is an interface for authentication providers.
type IdsecAuthenticator interface {
	Authenticate(profile *models.IdsecProfile, authProfile *authmodels.IdsecAuthProfile, secret *authmodels.IdsecSecret, forceRetry bool, forceReauth bool) (*authmodels.IdsecToken, error)
}

// parseIdentityAuth parses and validates identity authentication configuration.
func (p *IdsecProvider) parseIdentityAuth(ctx context.Context, config *IdsecProviderSchema) (*authCredentials, string) {
	tflog.Info(ctx, "Parsing identity authentication method")
	config.UserName = p.resolveTerraformStringVar(config.UserName, IdsecUsernameEnvVar)
	config.Secret = p.resolveTerraformStringVar(config.Secret, IdsecSecretEnvVar)
	if config.UserName.IsNull() || config.Secret.IsNull() {
		return nil, "Username and Secret are required for identity authentication."
	}
	creds := &authCredentials{
		userName:   config.UserName.ValueString(),
		secret:     config.Secret.ValueString(),
		authMethod: authmodels.IdsecAuthMethod("identity"),
		authMethodSettings: &authmodels.IdentityIdsecAuthMethodSettings{
			IdentityTenantSubdomain: config.Subdomain.ValueString(),
		},
	}
	tflog.Info(ctx, fmt.Sprintf("Using identity authentication method with username: %s", creds.userName))
	return creds, ""
}

// parseIdentityServiceUserAuth parses and validates identity service user authentication configuration.
func (p *IdsecProvider) parseIdentityServiceUserAuth(ctx context.Context, config *IdsecProviderSchema) (*authCredentials, string) {
	tflog.Info(ctx, "Parsing identity service user authentication method")
	config.ServiceUser = p.resolveTerraformStringVar(config.ServiceUser, IdsecServiceUserEnvVar)
	config.ServiceToken = p.resolveTerraformStringVar(config.ServiceToken, IdsecServiceTokenEnvVar)
	config.ServiceAuthorizedApp = p.resolveTerraformStringVar(config.ServiceAuthorizedApp, IdsecServiceAuthorizedAppEnvVar)
	if config.ServiceUser.IsNull() || config.ServiceToken.IsNull() {
		return nil, "Service User and Service Token are required for identity service user authentication."
	}
	if config.ServiceAuthorizedApp.IsNull() {
		config.ServiceAuthorizedApp = types.StringValue(IdsecServiceAuthorizedAppDefault)
	}
	creds := &authCredentials{
		userName:   config.ServiceUser.ValueString(),
		secret:     config.ServiceToken.ValueString(),
		authMethod: authmodels.IdsecAuthMethod("identity_service_user"),
		authMethodSettings: &authmodels.IdentityServiceUserIdsecAuthMethodSettings{
			IdentityTenantSubdomain:          config.Subdomain.ValueString(),
			IdentityAuthorizationApplication: config.ServiceAuthorizedApp.ValueString(),
		},
	}
	tflog.Info(ctx, fmt.Sprintf("Using identity service user authentication method with service user: %s", creds.userName))
	return creds, ""
}

// parsePVWAAuth parses and validates PVWA authentication configuration.
func (p *IdsecProvider) parsePVWAAuth(ctx context.Context, config *IdsecProviderSchema) (*authCredentials, string) {
	tflog.Info(ctx, "Parsing PVWA authentication method")
	config.UserName = p.resolveTerraformStringVar(config.UserName, IdsecUsernameEnvVar)
	config.Secret = p.resolveTerraformStringVar(config.Secret, IdsecSecretEnvVar)
	config.PVWAURL = p.resolveTerraformStringVar(config.PVWAURL, IdsecPVWAURLEnvVar)
	config.PVWALoginMethod = p.resolveTerraformStringVar(config.PVWALoginMethod, IdsecPVWALoginMethodEnvVar)
	if config.UserName.IsNull() || config.Secret.IsNull() {
		return nil, "Username and Secret are required for PVWA authentication."
	}
	if config.PVWAURL.IsNull() {
		return nil, "PVWA URL is required for PVWA authentication."
	}
	if config.PVWALoginMethod.IsNull() {
		config.PVWALoginMethod = types.StringValue(IdsecPVWALoginMethodDefault)
	}
	creds := &authCredentials{
		userName:   config.UserName.ValueString(),
		secret:     config.Secret.ValueString(),
		authMethod: authmodels.PVWA,
		authMethodSettings: &authmodels.PVWAIdsecAuthMethodSettings{
			PVWAURL:         config.PVWAURL.ValueString(),
			PVWALoginMethod: config.PVWALoginMethod.ValueString(),
		},
	}
	tflog.Info(ctx, fmt.Sprintf("Using PVWA authentication method with username: %s, PVWA URL: %s", creds.userName, config.PVWAURL.ValueString()))
	return creds, ""
}

// authenticateWithRetry performs authentication with retry logic for transient errors.
func (p *IdsecProvider) authenticateWithRetry(ctx context.Context, authenticator IdsecAuthenticator, creds *authCredentials, authType string) error {
	tflog.Info(ctx, fmt.Sprintf("Performing %s authentication", authType))
	var lastErr error
	for attempt := 1; attempt <= authRetryCount; attempt++ {
		forceRetry := attempt > 1
		if forceRetry {
			tflog.Info(ctx, fmt.Sprintf("Retrying %s authentication, attempt %d", authType, attempt))
		}
		_, err := authenticator.Authenticate(
			nil, // profile
			&authmodels.IdsecAuthProfile{
				Username:           creds.userName,
				AuthMethod:         creds.authMethod,
				AuthMethodSettings: creds.authMethodSettings,
			},
			&authmodels.IdsecSecret{
				Secret: creds.secret,
			},
			forceRetry,
			false,
		)
		if err == nil {
			tflog.Info(ctx, fmt.Sprintf("Successfully authenticated with %s", authType))
			return nil
		}
		lastErr = err
		// Check if error is retryable
		shouldRetry := false
		for _, retryableError := range authRetryableErrrors {
			if strings.Contains(err.Error(), retryableError) {
				tflog.Warn(ctx, fmt.Sprintf("Retrying %s authentication due to retryable error: %s [%v]", authType, retryableError, err))
				shouldRetry = true
				break
			}
		}
		if !shouldRetry {
			return fmt.Errorf("failed to authenticate with %s: %w", authType, err)
		}
	}
	return fmt.Errorf("failed to authenticate with %s, retries exhausted: %w", authType, lastErr)
}

// Metadata returns the provider's metadata.
func (p *IdsecProvider) Metadata(ctx context.Context, req terraformprovider.MetadataRequest, resp *terraformprovider.MetadataResponse) {
	resp.TypeName = "idsec"
	resp.Version = p.config.Version
}

// Schema returns the provider's schema.
func (p *IdsecProvider) Schema(ctx context.Context, req terraformprovider.SchemaRequest, resp *terraformprovider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Idsec provider for managing CyberArk resources.",
		MarkdownDescription: "Idsec provider for managing CyberArk resources.",
		Attributes: map[string]schema.Attribute{
			"auth_method": schema.StringAttribute{
				Optional:            true,
				Description:         "Authentication method. Defaults to 'identity'. When set to 'identity', both 'username' and 'secret' are required. When set to 'identity_service_user', both 'service_user' and 'service_token' are required. When set to 'pvwa', both 'pvwa_url' and 'username'/'secret' are required. Resolved from environment variable IDSEC_AUTH_METHOD.",
				MarkdownDescription: "Authentication method. Defaults to `identity`. When set to `identity`, both `username` and `secret` are **required**. When set to `identity_service_user`, both `service_user` and `service_token` are **required**. When set to `pvwa`, `pvwa_url`, `username`, and `secret` are **required**. Resolved from environment variable `IDSEC_AUTH_METHOD`.",
				Validators: []validator.String{
					schemas.StringInChoicesValidator{Choices: []string{"identity", "identity_service_user", "pvwa"}},
				},
			},
			"subdomain": schema.StringAttribute{
				Optional:            true,
				Description:         "Tenant subdomain for authentication. Optional, typically used for external IDP authentication. Resolved from environment variable IDSEC_SUBDOMAIN.",
				MarkdownDescription: "Tenant subdomain for authentication. Optional, typically used for external IDP authentication. Resolved from environment variable `IDSEC_SUBDOMAIN`.",
			},
			"username": schema.StringAttribute{
				Optional:            true,
				Description:         "Username for identity authentication. Required when 'auth_method' is 'identity' (default). Resolved from environment variable IDSEC_USERNAME.",
				MarkdownDescription: "Username for identity authentication. **Required** when `auth_method` is `identity` (default). Resolved from environment variable `IDSEC_USERNAME`.",
			},
			"secret": schema.StringAttribute{
				Optional:            true,
				Description:         "Secret for identity authentication. Required when 'auth_method' is 'identity' (default). Resolved from environment variable IDSEC_SECRET.",
				MarkdownDescription: "Secret for identity authentication. **Required** when `auth_method` is `identity` (default). Resolved from environment variable `IDSEC_SECRET`.",
				Sensitive:           true,
			},
			"service_user": schema.StringAttribute{
				Optional:            true,
				Description:         "Service user for identity service user authentication. Required when 'auth_method' is 'identity_service_user'. Resolved from environment variable IDSEC_SERVICE_USER.",
				MarkdownDescription: "Service user for identity service user authentication. **Required** when `auth_method` is `identity_service_user`. Resolved from environment variable `IDSEC_SERVICE_USER`.",
			},
			"service_token": schema.StringAttribute{
				Optional:            true,
				Description:         "Service token for identity service user authentication. Required when 'auth_method' is 'identity_service_user'. Resolved from environment variable IDSEC_SERVICE_TOKEN.",
				MarkdownDescription: "Service token for identity service user authentication. **Required** when `auth_method` is `identity_service_user`. Resolved from environment variable `IDSEC_SERVICE_TOKEN`.",
				Sensitive:           true,
			},
			"service_authorized_app": schema.StringAttribute{
				Optional:            true,
				Description:         "Authorized application for identity service user authentication. Used when 'auth_method' is 'identity_service_user'. Defaults to '__idaptive_cybr_user_oidc'. Resolved from environment variable IDSEC_SERVICE_AUTHORIZED_APP.",
				MarkdownDescription: "Authorized application for identity service user authentication. Used when `auth_method` is `identity_service_user`. Defaults to `__idaptive_cybr_user_oidc`. Resolved from environment variable `IDSEC_SERVICE_AUTHORIZED_APP`.",
			},
			"cache_authentication": schema.BoolAttribute{
				Optional:            true,
				Description:         "Cache authentication for the provider. Defaults to true. Resolved from environment variable IDSEC_CACHE_AUTHENTICATION.",
				MarkdownDescription: "Cache authentication for the provider. Defaults to `true`. Resolved from environment variable `IDSEC_CACHE_AUTHENTICATION`.",
			},
			"pvwa_url": schema.StringAttribute{
				Optional:            true,
				Description:         "PVWA base URL for PVWA authentication. Required when 'auth_method' is 'pvwa'. Resolved from environment variable IDSEC_PVWA_URL.",
				MarkdownDescription: "PVWA base URL for PVWA authentication. **Required** when `auth_method` is `pvwa`. Resolved from environment variable `IDSEC_PVWA_URL`.",
			},
			"pvwa_login_method": schema.StringAttribute{
				Optional:            true,
				Description:         "PVWA login method for PVWA authentication. Valid values: 'cyberark', 'ldap', 'windows'. Defaults to 'cyberark'. Used when 'auth_method' is 'pvwa'. Resolved from environment variable IDSEC_PVWA_LOGIN_METHOD.",
				MarkdownDescription: "PVWA login method for PVWA authentication. Valid values: `cyberark`, `ldap`, `windows`. Defaults to `cyberark`. Used when `auth_method` is `pvwa`. Resolved from environment variable `IDSEC_PVWA_LOGIN_METHOD`.",
				Validators: []validator.String{
					schemas.StringInChoicesValidator{Choices: []string{"cyberark", "ldap", "windows"}},
				},
			},
		},
	}
}

// Configure configures the provider with the given context and request.
func (p *IdsecProvider) Configure(ctx context.Context, req terraformprovider.ConfigureRequest, resp *terraformprovider.ConfigureResponse) {
	// Set the tool type for telemetry reporting
	// This ensures runtime report as Terraform Provider
	sdkconfig.SetIdsecToolInUse(sdkconfig.IdsecToolTerraformProvider)

	// Generate a unique correlation ID for this Terraform execution
	sdkconfig.GenerateCorrelationID()

	var config IdsecProviderSchema
	tflog.Info(ctx, "Configuring Idsec provider")
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Failed to read provider configuration")
		return
	}

	// Resolve common configuration from environment variables
	config.CacheAuthentication = p.resolveTerraformBoolVar(config.CacheAuthentication, IdsecCacheAuthenticationEnvVar, IdsecCacheAuthenticationDefault)
	config.AuthMethod = p.resolveTerraformStringVar(config.AuthMethod, IdsecAuthMethodEnvVar)
	config.Subdomain = p.resolveTerraformStringVar(config.Subdomain, IdsecSubdomainEnvVar)

	if config.AuthMethod.IsNull() {
		resp.Diagnostics.AddError("Invalid Configuration", "Auth method is required.")
		return
	}

	// Parse authentication credentials based on auth method
	var creds *authCredentials
	var parseErr string
	switch config.AuthMethod.ValueString() {
	case "identity":
		creds, parseErr = p.parseIdentityAuth(ctx, &config)
	case "identity_service_user":
		creds, parseErr = p.parseIdentityServiceUserAuth(ctx, &config)
	case "pvwa":
		creds, parseErr = p.parsePVWAAuth(ctx, &config)
	default:
		resp.Diagnostics.AddError("Invalid Configuration", "Unsupported auth method.")
		return
	}

	if parseErr != "" {
		resp.Diagnostics.AddError("Invalid Configuration", parseErr)
		return
	}

	// Perform authentication based on the auth method
	if config.AuthMethod.ValueString() == "pvwa" {
		p.configurePVWAAuth(ctx, &config, creds, resp)
	} else {
		p.configureISPAuth(ctx, &config, creds, resp)
	}
}

// configurePVWAAuth configures PVWA authentication for the provider.
func (p *IdsecProvider) configurePVWAAuth(ctx context.Context, config *IdsecProviderSchema, creds *authCredentials, resp *terraformprovider.ConfigureResponse) {
	pvwaAuth, ok := auth.NewIdsecPVWAAuth(config.CacheAuthentication.ValueBool()).(*auth.IdsecPVWAAuth)
	if !ok {
		resp.Diagnostics.AddError("Authentication Error", "Failed to create PVWA authentication.")
		return
	}
	p.pvwaAuth = pvwaAuth

	if err := p.authenticateWithRetry(ctx, pvwaAuth, creds, "PVWA"); err != nil {
		resp.Diagnostics.AddError("Authentication Error", err.Error())
		return
	}

	providerVersion = p.config.Version
	resp.ResourceData = p.pvwaAuth
	resp.DataSourceData = p.pvwaAuth
}

// configureISPAuth configures ISP (Identity) authentication for the provider.
func (p *IdsecProvider) configureISPAuth(ctx context.Context, config *IdsecProviderSchema, creds *authCredentials, resp *terraformprovider.ConfigureResponse) {
	ispAuth, ok := auth.NewIdsecISPAuth(config.CacheAuthentication.ValueBool()).(*auth.IdsecISPAuth)
	if !ok {
		resp.Diagnostics.AddError("Authentication Error", "Failed to create ISP authentication.")
		return
	}
	p.ispAuth = ispAuth

	if err := p.authenticateWithRetry(ctx, ispAuth, creds, "ISP"); err != nil {
		resp.Diagnostics.AddError("Authentication Error", err.Error())
		return
	}

	// Guard against edge cases where authentication succeeds but the Token field
	// on the auth object is not populated (e.g. keyring deserialization issues).
	// FromISPAuth in the SDK dereferences Token without a nil check, so we must
	// ensure it is set before any service tries to use it.
	if ispAuth.Token == nil {
		tflog.Debug(ctx, "ISP auth token not populated after authentication, forcing fresh authentication")
		_, err := ispAuth.Authenticate(
			nil,
			&authmodels.IdsecAuthProfile{
				Username:           creds.userName,
				AuthMethod:         creds.authMethod,
				AuthMethodSettings: creds.authMethodSettings,
			},
			&authmodels.IdsecSecret{
				Secret: creds.secret,
			},
			true,
			true,
		)
		if err != nil {
			resp.Diagnostics.AddError("Authentication Error", fmt.Sprintf("ISP token was nil after initial auth, forced re-auth also failed: %s", err.Error()))
			return
		}
		if ispAuth.Token == nil {
			resp.Diagnostics.AddError("Authentication Error", "ISP auth token is nil even after forced re-authentication")
			return
		}
	}

	providerVersion = p.config.Version
	resp.ResourceData = p.ispAuth
	resp.DataSourceData = p.ispAuth
}

func (p *IdsecProvider) collectTfResources() []schemas.Tuple[*services.IdsecServiceConfig, *provideractions.IdsecServiceTerraformResourceActionDefinition] {
	collected := make([]schemas.Tuple[*services.IdsecServiceConfig, *provideractions.IdsecServiceTerraformResourceActionDefinition], 0)
	for _, config := range provideractions.AllTerraformConfigs() {
		serviceConfig, err := services.GetServiceConfig(config.ServiceName)
		if err != nil {
			continue
		}
		for _, res := range config.Resources {
			found := false
			for _, existing := range collected {
				if existing.Second.ActionName == res.ActionName {
					found = true
					break
				}
			}
			if !found {
				collected = append(collected, schemas.Tuple[*services.IdsecServiceConfig, *provideractions.IdsecServiceTerraformResourceActionDefinition]{
					First:  &serviceConfig,
					Second: res,
				})
			}
		}
	}
	return collected
}

func (p *IdsecProvider) collectTfDataSources() []schemas.Tuple[*services.IdsecServiceConfig, *provideractions.IdsecServiceTerraformDataSourceActionDefinition] {
	collected := make([]schemas.Tuple[*services.IdsecServiceConfig, *provideractions.IdsecServiceTerraformDataSourceActionDefinition], 0)
	for _, config := range provideractions.AllTerraformConfigs() {
		serviceConfig, err := services.GetServiceConfig(config.ServiceName)
		if err != nil {
			continue
		}
		for _, ds := range config.DataSources {
			found := false
			for _, existing := range collected {
				if existing.Second.ActionName == ds.ActionName {
					found = true
					break
				}
			}
			if !found {
				collected = append(collected, schemas.Tuple[*services.IdsecServiceConfig, *provideractions.IdsecServiceTerraformDataSourceActionDefinition]{
					First:  &serviceConfig,
					Second: ds,
				})
			}
		}
	}
	return collected
}

// Resources returns the resources supported by the provider.
func (p *IdsecProvider) Resources(ctx context.Context) []func() resource.Resource {
	collectedResources := p.collectTfResources()
	tflog.Info(ctx, fmt.Sprintf("Collected %d resources from service configurations", len(collectedResources)))
	resourcesFunctions := make([]func() resource.Resource, 0, len(collectedResources))
	for _, resourceDef := range collectedResources {
		tflog.Info(ctx, fmt.Sprintf("Adding resource: %s", resourceDef.Second.ActionName))
		resourcesFunctions = append(resourcesFunctions, func() resource.Resource {
			return NewIdsecResource(resourceDef.First, resourceDef.Second)
		})
	}
	return resourcesFunctions
}

// DataSources returns the data sources supported by the provider.
func (p *IdsecProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	collectedDataSources := p.collectTfDataSources()
	tflog.Info(ctx, fmt.Sprintf("Collected %d data sources from service configurations", len(collectedDataSources)))
	dataSourceFunctions := make([]func() datasource.DataSource, 0, len(collectedDataSources))
	for _, dataSourceDef := range collectedDataSources {
		tflog.Info(ctx, fmt.Sprintf("Adding data source: %s", dataSourceDef.Second.ActionName))
		dataSourceFunctions = append(dataSourceFunctions, func() datasource.DataSource {
			return NewIdsecDataSource(dataSourceDef.First, dataSourceDef.Second)
		})
	}
	return dataSourceFunctions
}
