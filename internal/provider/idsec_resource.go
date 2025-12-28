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
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	api "github.com/cyberark/idsec-sdk-golang/pkg"
	"github.com/cyberark/idsec-sdk-golang/pkg/auth"
	"github.com/cyberark/idsec-sdk-golang/pkg/models/actions"
	"github.com/cyberark/idsec-sdk-golang/pkg/services"
	"github.com/cyberark/terraform-provider-idsec/internal/schemas"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// IdsecResource is a struct that implements the resource.Resource interface.
type IdsecResource struct {
	resource.ResourceWithConfigure
	serviceConfig    *services.IdsecServiceConfig
	actionDefinition *actions.IdsecServiceTerraformResourceActionDefinition
	idsecAPI         *api.IdsecAPI
}

// NewIdsecResource creates a new instance of IdsecResource.
func NewIdsecResource(serviceConfig *services.IdsecServiceConfig,
	actionDefinition *actions.IdsecServiceTerraformResourceActionDefinition) resource.Resource {
	return &IdsecResource{
		serviceConfig:    serviceConfig,
		actionDefinition: actionDefinition,
	}
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

func (s *IdsecResource) getImmutableAttributes() []string {
	// Use reflection to safely check if ImmutableAttributes field exists
	// This provides backward compatibility with SDK versions that don't have this field yet
	val := reflect.ValueOf(s.actionDefinition.IdsecServiceBaseTerraformActionDefinition)
	field := val.FieldByName("ImmutableAttributes")
	if field.IsValid() && field.Kind() == reflect.Slice {
		if attrs, ok := field.Interface().([]string); ok {
			return attrs
		}
	}
	return []string{}
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
	serviceParts := strings.Split(s.serviceConfig.ServiceName, "-")
	actionName, ok := s.actionDefinition.ActionsMappings[operation]
	if !ok {
		s.finalizeFailure(ctx, "Action Mapping Error", fmt.Sprintf("No action mapping found for operation: %s", operation), operation, originalState, respState, diagnostics)
		return
	}

	titleCase := cases.Title(language.English)

	actionNameTitled := strings.ReplaceAll(titleCase.String(actionName), "-", "")
	serviceNameTitled := ""
	for _, part := range serviceParts {
		serviceNameTitled += titleCase.String(part)
	}
	serviceNameTitled = strings.ReplaceAll(serviceNameTitled, "-", "")
	tflog.Info(ctx, fmt.Sprintf("Searching for Service Name: %s, Action Name: %s", serviceNameTitled, actionNameTitled))
	serviceMethod, err := schemas.FindMethodByName(reflect.ValueOf(s.idsecAPI), serviceNameTitled)
	if err != nil {
		s.finalizeFailure(ctx, "Service Method Error", fmt.Sprintf("Unable to find service method: %s", err.Error()), operation, originalState, respState, diagnostics)
		return
	}
	tflog.Info(ctx, fmt.Sprintf("Calling service method: %s", serviceNameTitled))
	// Call the service method to get the actual service
	serviceErr := serviceMethod.Call(nil)

	// Check if the service method returned an error in any of the return data
	service := serviceErr[0]
	if len(serviceErr) > 1 {
		if err, ok := serviceErr[1].Interface().(error); ok && err != nil {
			s.finalizeFailure(ctx, "Service Error", fmt.Sprintf("Unable to call service method: %s", err.Error()), operation, originalState, respState, diagnostics)
			return
		}
	}
	// Get the method from the deduced service above
	actionMethod, err := schemas.FindMethodByName(reflect.ValueOf(service.Interface()), actionNameTitled)
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

// Create handles the creation of the resource.
func (s *IdsecResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	s.triggerOperation(ctx, actions.CreateOperation, &resp.Diagnostics, &req.Plan, nil, &resp.State)
}

// Read handles reading the resource state.
func (s *IdsecResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	s.triggerOperation(ctx, actions.ReadOperation, &resp.Diagnostics, nil, &req.State, &resp.State)
}

// Update handles updating the resource.
func (s *IdsecResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	s.triggerOperation(ctx, actions.UpdateOperation, &resp.Diagnostics, &req.Plan, &req.State, &resp.State)
}

// Delete handles deleting the resource.
func (s *IdsecResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	s.triggerOperation(ctx, actions.DeleteOperation, &resp.Diagnostics, nil, &req.State, nil)
}
