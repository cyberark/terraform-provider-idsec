// Copyright (c) CyberArk.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	api "github.com/cyberark/idsec-sdk-golang/pkg"
	"github.com/cyberark/idsec-sdk-golang/pkg/auth"
	"github.com/cyberark/idsec-sdk-golang/pkg/services"
	"github.com/cyberark/terraform-provider-idsec/internal/actions"
	"github.com/cyberark/terraform-provider-idsec/internal/featureadoption"
	"github.com/cyberark/terraform-provider-idsec/internal/schemas"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// IdsecResource is a struct that implements the resource.Resource interface.
type IdsecResource struct {
	resource.ResourceWithConfigure
	IdsecServiceHelper
	serviceConfig    *services.IdsecServiceConfig
	actionDefinition *actions.IdsecServiceTerraformResourceActionDefinition
	idsecAPI         *api.IdsecAPI
}

// NewIdsecResource creates a new instance of IdsecResource.
func NewIdsecResource(serviceConfig *services.IdsecServiceConfig,
	actionDefinition *actions.IdsecServiceTerraformResourceActionDefinition) resource.Resource {
	return &IdsecResource{
		IdsecServiceHelper: IdsecServiceHelper{
			serviceConfig: serviceConfig,
		},
		serviceConfig:    serviceConfig,
		actionDefinition: actionDefinition,
	}
}

// setTerraformContext sets terraform context on the service for telemetry.
func (s *IdsecResource) setTerraformContext(operation string) {
	service := s.getService()
	if service == nil {
		return
	}

	s.addTelemetryContextField(service, "terraform_resource", "tfr", s.getTerraformTypeName(s.actionDefinition.ActionName))
	s.addTelemetryContextField(service, "terraform_operation", "tfo", operation)
	s.addTelemetryContextField(service, "provider_version", "tfv", providerVersion)
}

// clearTerraformContext clears terraform context from the SDK's telemetry.
func (s *IdsecResource) clearTerraformContext() {
	service := s.getService()
	if service == nil {
		return
	}

	s.clearTelemetryContext(service)
}

func (s *IdsecResource) schemaForOperation(operation actions.IdsecServiceActionOperation) (interface{}, error) {
	if !slices.Contains(s.actionDefinition.SupportedOperations, operation) {
		return nil, nil
	}
	operationName, ok := s.actionDefinition.ActionsMappings[operation]
	if !ok {
		return nil, fmt.Errorf("no schema mapping found for operation: %s", operation)
	}
	operationSchema, ok := s.actionDefinition.Schemas[operationName]
	if !ok {
		return nil, fmt.Errorf("no schema mapping found for operation: %s - %s", operationName, operation)
	}
	return schemas.DeepCopy(operationSchema), nil
}

// getStringSliceFromActionDefinition uses reflection to safely read a []string field from
// IdsecServiceBaseTerraformActionDefinition. Provides backward compatibility with SDK
// versions that don't have the field yet.
func (s *IdsecResource) getStringSliceFromActionDefinition(fieldName string) []string {
	val := reflect.ValueOf(s.actionDefinition.IdsecServiceBaseTerraformActionDefinition)
	field := val.FieldByName(fieldName)
	if field.IsValid() && field.Kind() == reflect.Slice {
		if attrs, ok := field.Interface().([]string); ok {
			return attrs
		}
	}
	return []string{}
}

func (s *IdsecResource) getImmutableAttributes() []string {
	return s.getStringSliceFromActionDefinition("ImmutableAttributes")
}

func (s *IdsecResource) getForceNewAttributes() []string {
	return s.getStringSliceFromActionDefinition("ForceNewAttributes")
}

func (s *IdsecResource) getComputedAttributes() []string {
	return s.getStringSliceFromActionDefinition("ComputedAttributes")
}

func (s *IdsecResource) getImportID() string {
	// Use reflection to safely check if ImportID field exists
	// This provides backward compatibility with SDK versions that don't have this field yet
	val := reflect.ValueOf(s.actionDefinition).Elem()
	field := val.FieldByName("ImportID")
	if field.IsValid() && field.Kind() == reflect.String {
		if attr, ok := field.Interface().(string); ok && attr != "" {
			return attr
		}
	}
	return "" // Return empty string if not configured (import not supported)
}

func (s *IdsecResource) parsePlanAndState(ctx context.Context, operation actions.IdsecServiceActionOperation, diagnostics *diag.Diagnostics, plan *tfsdk.Plan, state *tfsdk.State) (interface{}, error) {
	var operationSchemaInput interface{}
	if plan != nil && state != nil {
		tflog.Info(ctx, "Plan and state are not nil")
		operationSchema, err := s.schemaForOperation(operation)
		if err != nil {
			diagnostics.AddError("Schema Error", fmt.Sprintf("No schema mapping found for operation: %s", operation))
			return nil, fmt.Errorf("no schema mapping found for operation: %s", operation)
		}
		operationSchemaInput, err = schemas.StructFromPlanAndStateObject(ctx, plan, state, operationSchema, s.actionDefinition.StateSchema)
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Failed to convert plan and state object to schema: %s", err.Error()))
			diagnostics.AddError("Schema Conversion Error", fmt.Sprintf("Failed to convert plan and state object to schema: %s", err.Error()))
			return nil, err
		}
	} else if plan != nil {
		tflog.Info(ctx, "Plan is not nil")
		operationSchema, err := s.schemaForOperation(operation)
		if err != nil {
			diagnostics.AddError("Schema Error", fmt.Sprintf("No schema mapping found for operation: %s", operation))
			return nil, fmt.Errorf("no schema mapping found for operation: %s", operation)
		}
		operationSchemaInput, err = schemas.StructFromPlanObject(ctx, plan, operationSchema)
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Failed to convert plan object to schema: %s", err.Error()))
			diagnostics.AddError("Schema Conversion Error", fmt.Sprintf("Failed to convert plan object to schema: %s", err.Error()))
			return nil, err
		}
	} else if state != nil {
		tflog.Info(ctx, "State is not nil")
		stateSchema := schemas.DeepCopy(s.actionDefinition.StateSchema)
		if s.actionDefinition.RawStateInference {
			stateSchema = make(map[string]interface{})
		}
		stateSchema, err := schemas.StructFromStateObject(ctx, state, stateSchema)
		if err != nil {
			diagnostics.AddError("Schema Copy Error", fmt.Sprintf("Failed to copy schema: %s", err.Error()))
			return nil, err
		}
		operationSchemaInput, err = s.schemaForOperation(operation)
		if err != nil {
			diagnostics.AddError("Schema Error", fmt.Sprintf("No schema mapping found for operation: %s", operation))
			return nil, fmt.Errorf("no schema mapping found for operation: %s", operation)
		}
		if operation == actions.ReadOperation && s.actionDefinition.ReadSchemaPath != "" {
			stateSchema, err = schemas.SchemaByPath(stateSchema, s.actionDefinition.ReadSchemaPath)
			if err != nil {
				diagnostics.AddError("Schema Path Error", fmt.Sprintf("Failed to apply read path to schema: %s", err.Error()))
				return nil, fmt.Errorf("failed to apply read path to schema: %s", err.Error())
			}
		}
		if operation == actions.DeleteOperation && s.actionDefinition.DeleteSchemaPath != "" {
			stateSchema, err = schemas.SchemaByPath(stateSchema, s.actionDefinition.DeleteSchemaPath)
			if err != nil {
				diagnostics.AddError("Schema Path Error", fmt.Sprintf("Failed to apply delete path to schema: %s", err.Error()))
				return nil, fmt.Errorf("failed to apply delete path to schema: %s", err.Error())
			}
		}
		if operationSchemaInput != nil {
			err = mapstructure.Decode(stateSchema, operationSchemaInput)
			if err != nil {
				diagnostics.AddError("Schema Decode Error", fmt.Sprintf("Failed to decode schema: %s", err.Error()))
				return nil, err
			}
		}
	} else {
		diagnostics.AddError("State Error", "No state or plan provided for operation.")
		return nil, fmt.Errorf("no state or plan provided for operation")
	}
	return operationSchemaInput, nil
}

func (s *IdsecResource) finalizeState(ctx context.Context, operation actions.IdsecServiceActionOperation, originalState basetypes.ObjectValue, respState *tfsdk.State, diagnostics *diag.Diagnostics) {
	if respState != nil && !originalState.IsNull() && operation == actions.UpdateOperation {
		tflog.Info(ctx, "Finalizing failure by reverting to previous state")
		diags := respState.Set(ctx, originalState)
		if diags.HasError() {
			diagnostics.AddError("State Set Error", fmt.Sprintf("Failed to set state after operation failure [%v]", diags))
			diagnostics.Append(diags...)
		}
	}
}

func (s *IdsecResource) finalizeFailure(ctx context.Context, summary string, detail string, operation actions.IdsecServiceActionOperation, originalState basetypes.ObjectValue, respState *tfsdk.State, diagnostics *diag.Diagnostics) {
	tflog.Error(ctx, fmt.Sprintf("%s - %s", summary, detail))
	diagnostics.AddError(summary, detail)
	s.finalizeState(ctx, operation, originalState, respState, diagnostics)
}

func (s *IdsecResource) triggerOperation(ctx context.Context, operation actions.IdsecServiceActionOperation, diagnostics *diag.Diagnostics, plan *tfsdk.Plan, state *tfsdk.State, respState *tfsdk.State) {
	tflog.Info(ctx, fmt.Sprintf("Triggering operation: %s", operation))
	var originalState basetypes.ObjectValue
	if state != nil {
		diags := state.Get(ctx, &originalState)
		if diags.HasError() {
			s.finalizeFailure(ctx, "State Retrieval Error", fmt.Sprintf("Failed to get original state: %v", diags), operation, originalState, respState, diagnostics)
			return
		}
	}
	if !slices.Contains(s.actionDefinition.SupportedOperations, operation) {
		tflog.Info(ctx, fmt.Sprintf("Operation %s is not supported, no action will be made", operation))
		s.finalizeState(ctx, operation, originalState, respState, diagnostics)
		return
	}
	operationSchemaInput, err := s.parsePlanAndState(ctx, operation, diagnostics, plan, state)
	if diagnostics.HasError() || err != nil {
		if err != nil {
			s.finalizeFailure(ctx, "Parsing Error", fmt.Sprintf("Failed to parse plan and state: %s", err.Error()), operation, originalState, respState, diagnostics)
		} else {
			tflog.Error(ctx, "Error parsing plan and state, diagnostics already have errors")
			s.finalizeState(ctx, operation, originalState, respState, diagnostics)
		}
		return
	}
	actionName, ok := s.actionDefinition.ActionsMappings[operation]
	if !ok {
		s.finalizeFailure(ctx, "Action Mapping Error", fmt.Sprintf("No action mapping found for operation: %s", operation), operation, originalState, respState, diagnostics)
		return
	}

	titleCase := cases.Title(language.English)
	actionNameTitled := strings.ReplaceAll(titleCase.String(actionName), "-", "")
	serviceNameTitled := s.getServiceNameTitled()
	tflog.Info(ctx, fmt.Sprintf("Searching for Service Name: %s, Action Name: %s", serviceNameTitled, actionNameTitled))

	// Get the service from the helper
	service := s.getServiceInstance()
	if service == nil {
		s.finalizeFailure(ctx, "Service Error", "Service instance not configured", operation, originalState, respState, diagnostics)
		return
	}

	// Get the method from the service
	actionMethod, err := schemas.FindMethodByName(reflect.ValueOf(service), actionNameTitled)
	if err != nil {
		s.finalizeFailure(ctx, "Action Method Error", fmt.Sprintf("Unable to find action method: %s", err.Error()), operation, originalState, respState, diagnostics)
		return
	}

	var actionArgs []reflect.Value
	if operationSchemaInput != nil {
		actionArgs = append(actionArgs, reflect.ValueOf(operationSchemaInput))
	}
	tflog.Info(ctx, "Calling action method")
	result := actionMethod.Call(actionArgs)
	for _, res := range result {
		if err, ok := res.Interface().(error); ok && err != nil {
			s.finalizeFailure(ctx, "Action Error", fmt.Sprintf("Unable to call action method: %s", err.Error()), operation, originalState, respState, diagnostics)
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
	if respState != nil {
		tflog.Info(ctx, "Converting result to state object")
		createSchema, err := s.schemaForOperation(actions.CreateOperation)
		if err != nil {
			s.finalizeFailure(ctx, "Schema Error", fmt.Sprintf("No schema mapping found for operation: %s", actions.CreateOperation), operation, originalState, respState, diagnostics)
			return
		}
		updateSchema, err := s.schemaForOperation(actions.UpdateOperation)
		if err != nil {
			s.finalizeFailure(ctx, "Schema Error", fmt.Sprintf("No schema mapping found for operation: %s", actions.UpdateOperation), operation, originalState, respState, diagnostics)
			return
		}
		outputSchemaDef := schemas.GenerateResourceSchemaFromStruct(
			createSchema,
			updateSchema,
			s.actionDefinition.StateSchema,
			s.actionDefinition.SensitiveAttributes,
			s.actionDefinition.ExtraRequiredAttributes,
			s.actionDefinition.ComputedAsSetAttributes,
			s.getImmutableAttributes(),
			s.getForceNewAttributes(),
			s.getComputedAttributes(),
		)

		schemaAttrs := schemas.ResourceSchemaToSchemaAttrTypes(outputSchemaDef)
		stateResult, err := schemas.StructToStateObject(ctx, resultElem.Interface(), state, plan, schemaAttrs)
		if err != nil {
			s.finalizeFailure(ctx, "State Conversion Error", fmt.Sprintf("Failed to convert struct to state object: %s", err.Error()), operation, originalState, respState, diagnostics)
			return
		}
		if plan != nil {
			stateResult, err = schemas.MergePlanToStateObject(ctx, plan, stateResult, schemaAttrs)
			if err != nil {
				s.finalizeFailure(ctx, "State Merge Error", fmt.Sprintf("Failed to merge plan to state object: %s", err.Error()), operation, originalState, respState, diagnostics)
				return
			}
		}
		tflog.Info(ctx, "Setting state result")
		diags := respState.Set(ctx, stateResult)
		if diags.HasError() {
			tflog.Error(ctx, fmt.Sprintf("Failed to set state: %s", diags))
		}
		diagnostics.Append(diags...)
	}
}

// Metadata defines the resource type name.
func (s *IdsecResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, strings.ReplaceAll(s.actionDefinition.ActionName, "-", "_"))
}

// Schema dynamically generates the resource schema using `generateSchemaFromStruct`.
func (s *IdsecResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	if s.actionDefinition.Schemas == nil {
		resp.Diagnostics.AddError("Schema Error", "Schemas mappings are not provided.")
		return
	}
	createSchema, err := s.schemaForOperation(actions.CreateOperation)
	if err != nil {
		resp.Diagnostics.AddError("Schema Error", fmt.Sprintf("No schema mapping found for operation: %s - %v", actions.CreateOperation, err))
		return
	}
	updateSchema, err := s.schemaForOperation(actions.UpdateOperation)
	if err != nil {
		resp.Diagnostics.AddError("Schema Error", fmt.Sprintf("No schema mapping found for operation: %s - %v", actions.UpdateOperation, err))
		return
	}
	resp.Schema = schemas.GenerateResourceSchemaFromStruct(
		createSchema,
		updateSchema,
		s.actionDefinition.StateSchema,
		s.actionDefinition.SensitiveAttributes,
		s.actionDefinition.ExtraRequiredAttributes,
		s.actionDefinition.ComputedAsSetAttributes,
		s.getImmutableAttributes(),
		s.getForceNewAttributes(),
		s.getComputedAttributes(),
	)
	resp.Schema.Description = s.actionDefinition.ActionDescription
	if s.actionDefinition.ActionVersion != 0 {
		resp.Schema.Version = s.actionDefinition.ActionVersion
	}
}

// Configure initializes the resource with the necessary dependencies.
func (s *IdsecResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create handles the creation of the resource.
func (s *IdsecResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	s.setTerraformContext("Create")
	defer s.clearTerraformContext()
	defer featureadoption.ReportOperationDefer(ctx, s.idsecAPI, &resp.Diagnostics, s.buildFASTags(s.actionDefinition.ActionName, "Create"))()
	s.triggerOperation(ctx, actions.CreateOperation, &resp.Diagnostics, &req.Plan, nil, &resp.State)
}

// Read handles reading the resource state.
func (s *IdsecResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	s.setTerraformContext("Read")
	defer s.clearTerraformContext()
	defer featureadoption.ReportOperationDefer(ctx, s.idsecAPI, &resp.Diagnostics, s.buildFASTags(s.actionDefinition.ActionName, "Read"))()
	s.triggerOperation(ctx, actions.ReadOperation, &resp.Diagnostics, nil, &req.State, &resp.State)
}

// Update handles updating the resource.
func (s *IdsecResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	s.setTerraformContext("Update")
	defer s.clearTerraformContext()
	defer featureadoption.ReportOperationDefer(ctx, s.idsecAPI, &resp.Diagnostics, s.buildFASTags(s.actionDefinition.ActionName, "Update"))()
	s.triggerOperation(ctx, actions.UpdateOperation, &resp.Diagnostics, &req.Plan, &req.State, &resp.State)
}

// Delete handles deleting the resource.
func (s *IdsecResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	s.setTerraformContext("Delete")
	defer s.clearTerraformContext()
	defer featureadoption.ReportOperationDefer(ctx, s.idsecAPI, &resp.Diagnostics, s.buildFASTags(s.actionDefinition.ActionName, "Delete"))()
	s.triggerOperation(ctx, actions.DeleteOperation, &resp.Diagnostics, nil, &req.State, nil)
}

// ImportState handles importing existing resources into Terraform state.
// This method supports both the `terraform import` command and the `import` block.
func (s *IdsecResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	s.setTerraformContext("Import")
	defer s.clearTerraformContext()
	defer featureadoption.ReportOperationDefer(ctx, s.idsecAPI, &resp.Diagnostics, s.buildFASTags(s.actionDefinition.ActionName, "Import"))()

	tflog.Info(ctx, fmt.Sprintf("Importing resource with ID: %s", req.ID))

	// Validate that the import ID is not empty
	if req.ID == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID cannot be empty. Please provide a valid resource identifier.",
		)
		return
	}

	// Validate that the resource supports Read operation (required for import)
	if !slices.Contains(s.actionDefinition.SupportedOperations, actions.ReadOperation) {
		resp.Diagnostics.AddError(
			"Import Not Supported",
			fmt.Sprintf("This resource type (%s) does not support import because it does not support the Read operation.", s.actionDefinition.ActionName),
		)
		return
	}

	// Get the import ID attribute from action definition
	// Import is only supported if ImportID is explicitly configured
	// If ImportID contains ":", it defines multiple attributes (e.g. "safe_id:member_name").
	// In that case req.ID must contain ":"-separated values in the same order (e.g. "safe-123:member-456").
	importIDAttr := s.getImportID()
	if importIDAttr == "" {
		resp.Diagnostics.AddError(
			"Import Not Supported",
			fmt.Sprintf("This resource type (%s) does not support import. Import support must be explicitly configured in the action definition.", s.actionDefinition.ActionName),
		)
		return
	}

	if strings.Contains(importIDAttr, ":") {
		// Multi-attribute import: split attribute names and values by ":"
		attributes := strings.Split(importIDAttr, ":")
		if !strings.Contains(req.ID, ":") {
			resp.Diagnostics.AddError(
				"Invalid Import ID",
				fmt.Sprintf("This resource requires multiple attributes to import. Use colon-separated values in the same order as the import attributes (%s). Example: value1:value2", importIDAttr),
			)
			return
		}
		values := strings.Split(req.ID, ":")
		if len(attributes) != len(values) {
			resp.Diagnostics.AddError(
				"Invalid Import ID",
				fmt.Sprintf("Import ID has %d part(s) but %d attribute(s) are required (%s). Provide colon-separated values for each attribute.", len(values), len(attributes), importIDAttr),
			)
			return
		}
		for i, attr := range attributes {
			attr = strings.TrimSpace(attr)
			if attr == "" {
				resp.Diagnostics.AddError("Invalid Import ID Attribute", "ImportID contains an empty attribute name.")
				return
			}
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(attr), types.StringValue(values[i]))...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
	} else {
		// Single-attribute import
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(importIDAttr), types.StringValue(req.ID))...)
	}
}
