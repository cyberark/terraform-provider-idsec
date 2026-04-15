// Copyright (c) CyberArk.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	api "github.com/cyberark/idsec-sdk-golang/pkg"
	"github.com/cyberark/idsec-sdk-golang/pkg/auth"
	"github.com/cyberark/idsec-sdk-golang/pkg/services"
	"github.com/cyberark/terraform-provider-idsec/internal/actions"
	"github.com/cyberark/terraform-provider-idsec/internal/featureadoption"
	"github.com/cyberark/terraform-provider-idsec/internal/schemas"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// IdsecDataSource is a struct that implements the datasource.DataSource interface.
type IdsecDataSource struct {
	datasource.DataSourceWithConfigure
	IdsecServiceHelper
	serviceConfig    *services.IdsecServiceConfig
	actionDefinition *actions.IdsecServiceTerraformDataSourceActionDefinition
	idsecAPI         *api.IdsecAPI
}

// NewIdsecDataSource creates a new instance of IdsecDataSource.
func NewIdsecDataSource(serviceConfig *services.IdsecServiceConfig,
	actionDefinition *actions.IdsecServiceTerraformDataSourceActionDefinition) datasource.DataSource {
	return &IdsecDataSource{
		IdsecServiceHelper: IdsecServiceHelper{
			serviceConfig: serviceConfig,
		},
		serviceConfig:    serviceConfig,
		actionDefinition: actionDefinition,
	}
}

// setTerraformContext sets terraform context on the service for telemetry.
func (s *IdsecDataSource) setTerraformContext(operation string) {
	service := s.getService()
	if service == nil {
		return
	}

	s.addTelemetryContextField(service, "terraform_data_source", "tfd", s.getTerraformTypeName(s.actionDefinition.ActionName))
	s.addTelemetryContextField(service, "terraform_operation", "tfo", operation)
	s.addTelemetryContextField(service, "provider_version", "tfv", providerVersion)
}

// clearTerraformContext clears terraform context from the SDK's telemetry.
func (s *IdsecDataSource) clearTerraformContext() {
	service := s.getService()
	if service == nil {
		return
	}

	s.clearTelemetryContext(service)
}

// Metadata defines the resource type name.
func (s *IdsecDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, strings.ReplaceAll(s.actionDefinition.ActionName, "-", "_"))
}

// Schema dynamically generates the resource schema using `generateSchemaFromStruct`.
func (s *IdsecDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	if s.actionDefinition.StateSchema == nil || s.actionDefinition.DataSourceAction == "" {
		resp.Diagnostics.AddError("Schema Error", "Data source schema are not provided.")
		return
	}
	inputScheme, ok := s.actionDefinition.Schemas[s.actionDefinition.DataSourceAction]
	if !ok {
		resp.Diagnostics.AddError("Schema Error", fmt.Sprintf("Data source schema for action %s is not provided.", s.actionDefinition.DataSourceAction))
		return
	}
	resp.Schema = schemas.GenerateDataSourceSchemaFromStruct(
		inputScheme,
		s.actionDefinition.StateSchema,
		s.actionDefinition.SensitiveAttributes,
		s.actionDefinition.ExtraRequiredAttributes,
		s.actionDefinition.ComputedAsSetAttributes,
	)
	resp.Schema.Description = s.actionDefinition.ActionDescription
}

// Configure initializes the resource with the necessary dependencies.
func (s *IdsecDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	ispAuth, ok := req.ProviderData.(*auth.IdsecISPAuth)
	if !ok {
		// Try PVWA auth
		pvwaAuth, ok := req.ProviderData.(*auth.IdsecPVWAAuth)
		if !ok {
			resp.Diagnostics.AddError("Authentication Error", "Unable to authenticate with the provided credentials.")
			return
		}
		var err error
		s.idsecAPI, err = api.NewIdsecAPI([]auth.IdsecAuth{pvwaAuth}, nil)
		if err != nil {
			resp.Diagnostics.AddError("Service Initialization Error", fmt.Sprintf("Unable to create API: %s", err.Error()))
			return
		}
	} else {
		var err error
		s.idsecAPI, err = api.NewIdsecAPI([]auth.IdsecAuth{ispAuth}, nil)
		if err != nil {
			resp.Diagnostics.AddError("Service Initialization Error", fmt.Sprintf("Unable to create API: %s", err.Error()))
			return
		}
	}

	// Configure the service instance using the helper
	err := s.configureService(s.idsecAPI)
	if err != nil {
		resp.Diagnostics.AddError("Service Configuration Error", fmt.Sprintf("Unable to configure service: %s", err.Error()))
		return
	}
}

func (s *IdsecDataSource) parseConfig(ctx context.Context, diagnostics *diag.Diagnostics, config tfsdk.Config) (interface{}, error) {
	tflog.Info(ctx, "Parsing input actionDefinition")
	inputScheme, ok := s.actionDefinition.Schemas[s.actionDefinition.DataSourceAction]
	if !ok || inputScheme == nil {
		diagnostics.AddError("Schema Error", fmt.Sprintf("Data source schema for action %s is not provided.", s.actionDefinition.DataSourceAction))
		return nil, fmt.Errorf("data source schema for action %s is not provided", s.actionDefinition.DataSourceAction)
	}
	inputConfigSchema, err := schemas.StructFromConfigObject(ctx, &config, inputScheme)
	if err != nil {
		diagnostics.AddError("Config Copy Error", fmt.Sprintf("Failed to copy actionDefinition: %s", err.Error()))
		return nil, err
	}
	return inputConfigSchema, nil
}

// Read is called when the provider must read data source values in.
func (s *IdsecDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	s.setTerraformContext("Read")
	defer s.clearTerraformContext()
	defer featureadoption.ReportOperationDefer(ctx, s.idsecAPI, &resp.Diagnostics, s.buildFASTags(s.actionDefinition.ActionName, "Read"))()

	tflog.Info(ctx, "Triggering datasource read")
	operationSchemaInput, err := s.parseConfig(ctx, &resp.Diagnostics, req.Config)
	if resp.Diagnostics.HasError() || err != nil {
		tflog.Error(ctx, "Failed to get operation schema input")
		return
	}

	titleCase := cases.Title(language.English)
	actionNameTitled := strings.ReplaceAll(titleCase.String(s.actionDefinition.DataSourceAction), "-", "")
	serviceNameTitled := s.getServiceNameTitled()
	tflog.Info(ctx, fmt.Sprintf("Searching for Service Name: %s, Action Name: %s", serviceNameTitled, actionNameTitled))

	// Get the service from the helper
	service := s.getServiceInstance()
	if service == nil {
		resp.Diagnostics.AddError("Service Error", "Service instance not configured")
		return
	}

	// Get the method from the service
	actionMethod, err := schemas.FindMethodByName(reflect.ValueOf(service), actionNameTitled)
	if err != nil {
		resp.Diagnostics.AddError("Action Method Error", fmt.Sprintf("Unable to find action method: %s", err.Error()))
		return
	}
	actionArgs := []reflect.Value{reflect.ValueOf(operationSchemaInput)}
	tflog.Info(ctx, "Calling action method")
	result := actionMethod.Call(actionArgs)
	for _, res := range result {
		if err, ok := res.Interface().(error); ok && err != nil {
			tflog.Error(ctx, fmt.Sprintf("Failed to call action method: %s", err.Error()))
			resp.Diagnostics.AddError("Action Error", fmt.Sprintf("Unable to call action method: %s", err.Error()))
			return
		}
	}
	if len(result) < 1 {
		tflog.Info(ctx, "No result returned from action method")
		return
	}
	resultElem := result[0]
	if _, ok := resultElem.Interface().(error); ok {
		return
	}
	tflog.Info(ctx, "Managed to call action successfully with result")
	if resultElem.Kind() == reflect.Pointer {
		resultElem = resultElem.Elem()
	}
	tflog.Info(ctx, "Converting result to state object")
	inputScheme, ok := s.actionDefinition.Schemas[s.actionDefinition.DataSourceAction]
	if !ok {
		resp.Diagnostics.AddError("Schema Error", fmt.Sprintf("Data source schema for action %s is not provided.", s.actionDefinition.DataSourceAction))
		return
	}
	outputSchemaDef := schemas.GenerateDataSourceSchemaFromStruct(
		inputScheme,
		s.actionDefinition.StateSchema,
		s.actionDefinition.SensitiveAttributes,
		s.actionDefinition.ExtraRequiredAttributes,
		s.actionDefinition.ComputedAsSetAttributes,
	)
	schemaAttrs := schemas.DataSourceSchemaToSchemaAttrTypes(outputSchemaDef)
	stateResult, err := schemas.StructToStateObject(ctx, resultElem.Interface(), nil, nil, schemaAttrs)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Failed to convert struct to state object: %s", err.Error()))
		resp.Diagnostics.AddError("State Conversion Error", fmt.Sprintf("Failed to convert struct to state object: %s", err.Error()))
		return
	}
	diags := resp.State.Set(ctx, stateResult)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Failed to set state: %s", diags))
	}
	resp.Diagnostics.Append(diags...)
}
