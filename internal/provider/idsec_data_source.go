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
	"github.com/cyberark/idsec-sdk-golang/pkg/models/actions"
	"github.com/cyberark/idsec-sdk-golang/pkg/services"
	"github.com/cyberark/terraform-provider-idsec/internal/schemas"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// IdsecDataSource is a struct that implements the datasource.DataSource interface.
type IdsecDataSource struct {
	datasource.DataSourceWithConfigure
	serviceConfig    *services.IdsecServiceConfig
	actionDefinition *actions.IdsecServiceTerraformDataSourceActionDefinition
	idsecAPI         *api.IdsecAPI
}

// NewIdsecDataSource creates a new instance of IdsecResource.
func NewIdsecDataSource(serviceConfig *services.IdsecServiceConfig,
	actionDefinition *actions.IdsecServiceTerraformDataSourceActionDefinition) datasource.DataSource {
	return &IdsecDataSource{
		serviceConfig:    serviceConfig,
		actionDefinition: actionDefinition,
	}
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
	if !ok || ispAuth == nil {
		resp.Diagnostics.AddError("Authentication Error", "Unable to authenticate with the provided credentials.")
		return
	}
	var err error
	s.idsecAPI, err = api.NewIdsecAPI([]auth.IdsecAuth{ispAuth}, nil)
	if err != nil {
		resp.Diagnostics.AddError("Service Initialization Error", fmt.Sprintf("Unable to create SIA API: %s", err.Error()))
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
	tflog.Info(ctx, "Triggering datasource read")
	operationSchemaInput, err := s.parseConfig(ctx, &resp.Diagnostics, req.Config)
	if resp.Diagnostics.HasError() || err != nil {
		tflog.Error(ctx, "Failed to get operation schema input")
		return
	}
	serviceParts := strings.Split(s.serviceConfig.ServiceName, "-")

	titleCase := cases.Title(language.English)

	actionNameTitled := strings.ReplaceAll(titleCase.String(s.actionDefinition.DataSourceAction), "-", "")
	serviceNameTitled := ""
	for _, part := range serviceParts {
		serviceNameTitled += titleCase.String(part)
	}
	serviceNameTitled = strings.ReplaceAll(serviceNameTitled, "-", "")
	tflog.Info(ctx, fmt.Sprintf("Searching for Service Name: %s, Action Name: %s", serviceNameTitled, actionNameTitled))
	serviceMethod, err := schemas.FindMethodByName(reflect.ValueOf(s.idsecAPI), serviceNameTitled)
	if err != nil {
		resp.Diagnostics.AddError("Service Method Error", fmt.Sprintf("Unable to find service method: %s", err.Error()))
		return
	}
	tflog.Info(ctx, fmt.Sprintf("Calling service method: %s", serviceNameTitled))
	// Call the service method to get the actual service
	serviceErr := serviceMethod.Call(nil)

	// Check if the service method returned an error in any of the return data
	service := serviceErr[0]
	if len(serviceErr) > 1 {
		if err, ok := serviceErr[1].Interface().(error); ok && err != nil {
			tflog.Error(ctx, fmt.Sprintf("Failed to call service method: %s", err.Error()))
			resp.Diagnostics.AddError("Service Error", fmt.Sprintf("Unable to call service method: %s", err.Error()))
			return
		}
	}
	// Get the method from the deduced service above
	actionMethod, err := schemas.FindMethodByName(reflect.ValueOf(service.Interface()), actionNameTitled)
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
