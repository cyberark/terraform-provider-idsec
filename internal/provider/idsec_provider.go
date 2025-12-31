// Copyright (c) CyberArk.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"

	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	terraformprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/cyberark/idsec-sdk-golang/pkg/auth"
	"github.com/cyberark/idsec-sdk-golang/pkg/models/actions"
	authmodels "github.com/cyberark/idsec-sdk-golang/pkg/models/auth"
	"github.com/cyberark/idsec-sdk-golang/pkg/services"
	"github.com/cyberark/terraform-provider-idsec/internal/schemas"
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
)

// Ensure IdsecProvider satisfies various provider interfaces.
var _ terraformprovider.Provider = &IdsecProvider{}

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
	ispAuth *auth.IdsecISPAuth
	config  IdsecProviderConfig
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
				Description:         "Authentication method. Defaults to 'identity'. When set to 'identity', both 'username' and 'secret' are required. When set to 'identity_service_user', both 'service_user' and 'service_token' are required. Resolved from environment variable IDSEC_AUTH_METHOD.",
				MarkdownDescription: "Authentication method. Defaults to `identity`. When set to `identity`, both `username` and `secret` are **required**. When set to `identity_service_user`, both `service_user` and `service_token` are **required**. Resolved from environment variable `IDSEC_AUTH_METHOD`.",
				Validators: []validator.String{
					schemas.StringInChoicesValidator{Choices: []string{"identity", "identity_service_user"}},
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
		},
	}
}

// Configure configures the provider with the given context and request.
func (p *IdsecProvider) Configure(ctx context.Context, req terraformprovider.ConfigureRequest, resp *terraformprovider.ConfigureResponse) {
	var config IdsecProviderSchema
	tflog.Info(ctx, "Configuring Idsec provider")
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Failed to read provider configuration")
		return
	}
	var userName string
	var secret string
	var authMethodSettings authmodels.IdsecAuthMethodSettings
	config.CacheAuthentication = p.resolveTerraformBoolVar(config.CacheAuthentication, IdsecCacheAuthenticationEnvVar, IdsecCacheAuthenticationDefault)
	config.AuthMethod = p.resolveTerraformStringVar(config.AuthMethod, IdsecAuthMethodEnvVar)
	config.Subdomain = p.resolveTerraformStringVar(config.Subdomain, IdsecSubdomainEnvVar)
	if config.AuthMethod.IsNull() {
		resp.Diagnostics.AddError("Invalid Configuration", "Auth method is required.")
		return
	}
	switch config.AuthMethod.ValueString() {
	case "identity":
		tflog.Info(ctx, "Parsing identity authentication method")
		config.UserName = p.resolveTerraformStringVar(config.UserName, IdsecUsernameEnvVar)
		config.Secret = p.resolveTerraformStringVar(config.Secret, IdsecSecretEnvVar)
		if config.UserName.IsNull() || config.Secret.IsNull() {
			resp.Diagnostics.AddError("Invalid Configuration", "Username and Secret are required for identity authentication.")
			return
		}
		userName = config.UserName.ValueString()
		secret = config.Secret.ValueString()
		authMethodSettings = &authmodels.IdentityIdsecAuthMethodSettings{
			IdentityTenantSubdomain: config.Subdomain.ValueString(),
		}
		tflog.Info(ctx, fmt.Sprintf("Using identity authentication method with username: %s", userName))
	case "identity_service_user":
		tflog.Info(ctx, "Parsing identity service user authentication method")
		config.ServiceUser = p.resolveTerraformStringVar(config.ServiceUser, IdsecServiceUserEnvVar)
		config.ServiceToken = p.resolveTerraformStringVar(config.ServiceToken, IdsecServiceTokenEnvVar)
		config.ServiceAuthorizedApp = p.resolveTerraformStringVar(config.ServiceAuthorizedApp, IdsecServiceAuthorizedAppEnvVar)
		if config.ServiceUser.IsNull() || config.ServiceToken.IsNull() {
			resp.Diagnostics.AddError("Invalid Configuration", "Service User and Service Token are required for identity service user authentication.")
			return
		}
		if config.ServiceAuthorizedApp.IsNull() {
			config.ServiceAuthorizedApp = types.StringValue(IdsecServiceAuthorizedAppDefault)
		}
		userName = config.ServiceUser.ValueString()
		secret = config.ServiceToken.ValueString()
		authMethodSettings = &authmodels.IdentityServiceUserIdsecAuthMethodSettings{
			IdentityTenantSubdomain:          config.Subdomain.ValueString(),
			IdentityAuthorizationApplication: config.ServiceAuthorizedApp.ValueString(),
		}
		tflog.Info(ctx, fmt.Sprintf("Using identity service user authentication method with service user: %s", userName))
	default:
		resp.Diagnostics.AddError("Invalid Configuration", "Unsupported auth method.")
		return
	}
	tflog.Info(ctx, "Performing isp authentication")
	ispAuth, ok := auth.NewIdsecISPAuth(config.CacheAuthentication.ValueBool()).(*auth.IdsecISPAuth)
	if !ok {
		resp.Diagnostics.AddError("Authentication Error", "Failed to create ISP authentication.")
		return
	}
	p.ispAuth = ispAuth
	_, err := p.ispAuth.Authenticate(
		nil,
		&authmodels.IdsecAuthProfile{
			Username:           userName,
			AuthMethod:         authmodels.IdsecAuthMethod(config.AuthMethod.ValueString()),
			AuthMethodSettings: authMethodSettings,
		},
		&authmodels.IdsecSecret{
			Secret: secret,
		},
		false,
		false,
	)
	if err != nil {
		resp.Diagnostics.AddError("Authentication Error", fmt.Sprintf("Failed to authenticate with the provided credentials. [%v]", err))
		return
	}
	tflog.Info(ctx, "Successfully authenticated with ISP")
	resp.ResourceData = p.ispAuth
	resp.DataSourceData = p.ispAuth
}

func (p *IdsecProvider) collectTfItems(actionType actions.IdsecServiceActionType) []schemas.Tuple[*services.IdsecServiceConfig, actions.IdsecServiceActionDefinition] {
	collectedResources := make([]schemas.Tuple[*services.IdsecServiceConfig, actions.IdsecServiceActionDefinition], 0)
	for _, serviceConfig := range services.AllServiceConfigs() {
		if tfResources, ok := serviceConfig.ActionsConfigurations[actionType]; ok {
			for _, tfResourceBase := range tfResources {
				found := false
				for _, collectedResource := range collectedResources {
					if collectedResource.Second.ActionDefinitionName() == tfResourceBase.ActionDefinitionName() {
						found = true
						break
					}
				}
				if !found {
					collectedResources = append(collectedResources, schemas.Tuple[*services.IdsecServiceConfig, actions.IdsecServiceActionDefinition]{
						First:  &serviceConfig,
						Second: tfResourceBase,
					})
				}
			}
		}
	}
	return collectedResources
}

// Resources returns the resources supported by the provider.
func (p *IdsecProvider) Resources(ctx context.Context) []func() resource.Resource {
	collectedResources := p.collectTfItems(actions.IdsecServiceActionTypeTerraformResource)
	tflog.Info(ctx, fmt.Sprintf("Collected %d resources from service configurations", len(collectedResources)))
	resourcesFunctions := make([]func() resource.Resource, 0, len(collectedResources))
	for _, resourceDef := range collectedResources {
		tflog.Info(ctx, fmt.Sprintf("Adding resource: %s", resourceDef.Second.ActionDefinitionName()))
		resourcesFunctions = append(resourcesFunctions, func() resource.Resource {
			second, ok := resourceDef.Second.(*actions.IdsecServiceTerraformResourceActionDefinition)
			if !ok {
				tflog.Error(ctx, fmt.Sprintf("Failed to cast resource definition for resource: %s", resourceDef.Second.ActionDefinitionName()))
				return nil
			}
			return NewIdsecResource(resourceDef.First, second)
		})
	}
	return resourcesFunctions
}

// DataSources returns the data sources supported by the provider.
func (p *IdsecProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	collectedResources := p.collectTfItems(actions.IdsecServiceActionTypeTerraformDataSource)
	tflog.Info(ctx, fmt.Sprintf("Collected %d data sources from service configurations", len(collectedResources)))
	dataSourceFunctions := make([]func() datasource.DataSource, 0, len(collectedResources))
	for _, dataSourceDef := range collectedResources {
		tflog.Info(ctx, fmt.Sprintf("Adding data source: %s", dataSourceDef.Second.ActionDefinitionName()))
		dataSourceFunctions = append(dataSourceFunctions, func() datasource.DataSource {
			second, ok := dataSourceDef.Second.(*actions.IdsecServiceTerraformDataSourceActionDefinition)
			if !ok {
				tflog.Error(ctx, fmt.Sprintf("Failed to cast data source definition for data source: %s", dataSourceDef.Second.ActionDefinitionName()))
				return nil
			}
			return NewIdsecDataSource(dataSourceDef.First, second)
		})
	}
	return dataSourceFunctions
}
