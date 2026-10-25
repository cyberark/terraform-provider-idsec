// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package cce

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	awsactions "github.com/cyberark/idsec-sdk-golang/pkg/services/cce/aws/actions"
	awsmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/cce/aws/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
	"github.com/cyberark/terraform-provider-idsec/internal/schemas"
)

// The organization-account update needs to tell an in-place change to an already-onboarded service
// from an unchanged one, and the API cannot answer that: GET account resolves an org member to its
// parent organization's onboarding and returns the organization's parameters as the account's, so it
// describes management-account roles while Terraform holds the member account's. PriorState()
// supplies the only other record - what the last apply sent. Getting it wrong is silent in both
// directions: prior values that mirror the plan make every service look unchanged and an edit never
// reaches the API, while prior values that go missing make every service look changed and the
// unchanged ones ride along, which the API rejects with 501 FEATURE_NOT_IMPLEMENTED when the
// tenant's update feature flag is off. The tests below pin both directions.

// orgAccountResource returns the organization-account resource definition, its real generated
// schema and its update model.
func orgAccountResource(t *testing.T) (*tfactions.IdsecServiceTerraformResourceActionDefinition, schema.Schema, interface{}) {
	t.Helper()
	res := cceResourceByName(t, "cce-aws", "cce-aws-organization-account")
	createModel := awsactions.ActionToSchemaMap[res.ActionsMappings[tfactions.CreateOperation]]
	updateModel := awsactions.ActionToSchemaMap[res.ActionsMappings[tfactions.UpdateOperation]]
	if createModel == nil || updateModel == nil {
		t.Fatal("create/update models not found for cce-aws-organization-account")
	}
	s, diags := schemas.GenerateResourceSchemaFromStruct(createModel, updateModel, res.StateSchema, nil,
		res.ExtraRequiredAttributes, nil, nil, nil, res.ComputedAttributes, nil, nil, nil)
	if diags.HasError() {
		t.Fatalf("schema generation failed: %v", diags)
	}
	return res, s, updateModel
}

// serviceSpec is one `services` element as it would be written in HCL.
type serviceSpec struct {
	name      string
	version   string
	resources map[string]string
}

// servicesValue builds the value Terraform produces for a `services` list. Each element carries only
// the resource keys its own service declares, which is why the attribute is dynamic: a list of
// objects with differing attribute sets has no single element type.
func servicesValue(specs ...serviceSpec) attr.Value {
	elemTypes := make([]attr.Type, 0, len(specs))
	elems := make([]attr.Value, 0, len(specs))
	for _, spec := range specs {
		resourceTypes := make(map[string]attr.Type, len(spec.resources))
		resourceValues := make(map[string]attr.Value, len(spec.resources))
		for key, value := range spec.resources {
			resourceTypes[key] = types.StringType
			resourceValues[key] = types.StringValue(value)
		}
		objectTypes := map[string]attr.Type{
			"service_name": types.StringType,
			"version":      types.StringType,
			"resources":    types.ObjectType{AttrTypes: resourceTypes},
		}
		elemTypes = append(elemTypes, types.ObjectType{AttrTypes: objectTypes})
		elems = append(elems, types.ObjectValueMust(objectTypes, map[string]attr.Value{
			"service_name": types.StringValue(spec.name),
			"version":      types.StringValue(spec.version),
			"resources":    types.ObjectValueMust(resourceTypes, resourceValues),
		}))
	}
	return types.DynamicValue(types.TupleValueMust(elemTypes, elems))
}

// serviceParametersValue builds the value Terraform produces for a `service_parameters` map keyed by
// service name.
func serviceParametersValue(params map[string]map[string]string) attr.Value {
	outerTypes := make(map[string]attr.Type, len(params))
	outerValues := make(map[string]attr.Value, len(params))
	for serviceName, entries := range params {
		innerTypes := make(map[string]attr.Type, len(entries))
		innerValues := make(map[string]attr.Value, len(entries))
		for key, value := range entries {
			innerTypes[key] = types.StringType
			innerValues[key] = types.StringValue(value)
		}
		outerTypes[serviceName] = types.ObjectType{AttrTypes: innerTypes}
		outerValues[serviceName] = types.ObjectValueMust(innerTypes, innerValues)
	}
	return types.DynamicValue(types.ObjectValueMust(outerTypes, outerValues))
}

// convertUpdate runs the plan and state through the real update conversion path and returns the
// request the SDK action would receive.
func convertUpdate(t *testing.T, planAttrs, stateAttrs map[string]attr.Value) *awsmodels.TfIdsecCCEAWSUpdateOrganizationAccount {
	t.Helper()
	ctx := context.Background()
	res, s, updateModel := orgAccountResource(t)

	withID := func(attrs map[string]attr.Value) map[string]attr.Value {
		merged := map[string]attr.Value{"id": types.StringValue("onboarding-1")}
		for name, value := range attrs {
			merged[name] = value
		}
		return merged
	}
	plan := &tfsdk.Plan{Schema: s, Raw: rawObject(t, ctx, s, withID(planAttrs))}
	state := &tfsdk.State{Schema: s, Raw: rawObject(t, ctx, s, withID(stateAttrs))}

	request, err := schemas.StructFromPlanAndStateObject(ctx, plan, state, updateModel, res.StateSchema,
		schemas.DiffOnUpdateOptions{Enabled: res.UpdateOnlyModified})
	if err != nil {
		t.Fatalf("update conversion failed: %v", err)
	}
	typed, ok := request.(*awsmodels.TfIdsecCCEAWSUpdateOrganizationAccount)
	if !ok {
		t.Fatalf("request is %T, want *awsmodels.TfIdsecCCEAWSUpdateOrganizationAccount", request)
	}
	return typed
}

// TestOrganizationAccount_PriorStateIsNotSchemaAttribute pins that the prior-state mechanism stays
// internal. The prior request is held in an unexported field, so mapstructure, encoding/json and
// the schema generator all skip it automatically. This test guards against a regression where
// someone re-introduces exported Previous* fields without the mapstructure:"-" tag.
func TestOrganizationAccount_PriorStateIsNotSchemaAttribute(t *testing.T) {
	_, s, _ := orgAccountResource(t)
	objType, ok := s.Type().(types.ObjectType)
	if !ok {
		t.Fatalf("schema type = %T, want types.ObjectType", s.Type())
	}
	for _, name := range []string{"previous_services", "previous_service_parameters", "prior", "prior_state"} {
		if _, exists := objType.AttrTypes[name]; exists {
			t.Errorf(`attribute %q must not be in the resource schema; prior state must stay internal`, name)
		}
	}
}

// TestOrganizationAccount_PriorStateIsNotSerialized pins that the prior request never appears in
// the request body. The prior is held in an unexported field, so encoding/json skips it by
// language rule; this test guards against a regression where someone re-introduces an exported
// Previous* field without json:"-".
func TestOrganizationAccount_PriorStateIsNotSerialized(t *testing.T) {
	request := convertUpdate(t,
		map[string]attr.Value{"services": servicesValue(serviceSpec{name: "dpa", version: "0.0.3"})},
		map[string]attr.Value{"services": servicesValue(serviceSpec{name: "dpa", version: "0.0.3"})},
	)
	if request.PriorState() == nil || len(request.PriorState().Services) == 0 {
		t.Fatal("precondition: PriorState().Services must be set for this check to mean anything")
	}

	body, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("failed to unmarshal request body: %v", err)
	}
	for _, key := range []string{"previousServices", "previousServiceParameters", "PreviousServices", "PreviousServiceParameters", "prior", "Prior"} {
		if _, exists := payload[key]; exists {
			t.Errorf(`request body contains %q; prior state must not appear in the API request`, key)
		}
	}
}

// TestOrganizationAccount_PriorStateServicesComeFromStateNotPlan is the test that can tell the
// plumbing apart from a no-op. If PriorState().Services were filled from the plan — the default for
// every other field on the model — it would equal Services here, every onboarded service would look
// unchanged, and an edit would silently never be sent.
func TestOrganizationAccount_PriorStateServicesComeFromStateNotPlan(t *testing.T) {
	request := convertUpdate(t,
		map[string]attr.Value{"services": servicesValue(serviceSpec{
			name: "dpa", version: "0.0.3", resources: map[string]string{"DpaRoleArn": "arn:new"},
		})},
		map[string]attr.Value{"services": servicesValue(serviceSpec{
			name: "dpa", version: "0.0.3", resources: map[string]string{"DpaRoleArn": "arn:old"},
		})},
	)

	if len(request.Services) != 1 {
		t.Fatalf("len(Services) = %d, want 1", len(request.Services))
	}
	if got := request.Services[0].Resources["DpaRoleArn"]; got != "arn:new" {
		t.Errorf("Services[0].Resources[DpaRoleArn] = %v, want the planned arn:new", got)
	}

	if request.PriorState() == nil || len(request.PriorState().Services) != 1 {
		priorLen := 0
		if request.PriorState() != nil {
			priorLen = len(request.PriorState().Services)
		}
		t.Fatalf("len(PriorState().Services) = %d, want 1", priorLen)
	}
	if got := request.PriorState().Services[0].Resources["DpaRoleArn"]; got != "arn:old" {
		t.Errorf("PriorState().Services[0].Resources[DpaRoleArn] = %v, want the prior arn:old; "+
			"filling it from the plan makes every service look unchanged", got)
	}
}

// TestOrganizationAccount_PriorStateServiceParametersComeFromStateNotPlan is the same check for
// service_parameters, which carries the settings that are not resource ARNs - sca's sso_enable and
// region, for instance.
func TestOrganizationAccount_PriorStateServiceParametersComeFromStateNotPlan(t *testing.T) {
	services := servicesValue(serviceSpec{name: "sca", version: "0.0.1"})
	request := convertUpdate(t,
		map[string]attr.Value{
			"services":           services,
			"service_parameters": serviceParametersValue(map[string]map[string]string{"sca": {"ssoEnable": "true"}}),
		},
		map[string]attr.Value{
			"services":           services,
			"service_parameters": serviceParametersValue(map[string]map[string]string{"sca": {"ssoEnable": "false"}}),
		},
	)

	if got := request.ServiceParameters["sca"]["ssoEnable"]; got != "true" {
		t.Errorf("ServiceParameters[sca][ssoEnable] = %v, want the planned true", got)
	}
	if request.PriorState() == nil {
		t.Fatal("PriorState() is nil; expected prior service parameters to be set")
	}
	if got := request.PriorState().ServiceParameters["sca"]["ssoEnable"]; got != "false" {
		t.Errorf("PriorState().ServiceParameters[sca][ssoEnable] = %v, want the prior false", got)
	}
}

// TestOrganizationAccount_UnchangedServicesCarryIdenticalPriorState is the no-update case: a re-apply
// with nothing edited. Prior and desired must match exactly, since that equality is what lets the SDK
// drop the service and avoid the 501.
func TestOrganizationAccount_UnchangedServicesCarryIdenticalPriorState(t *testing.T) {
	services := servicesValue(
		serviceSpec{name: "dpa", version: "0.0.3", resources: map[string]string{"DpaRoleArn": "arn:dpa"}},
		serviceSpec{name: "sca", version: "0.0.1", resources: map[string]string{"ScaRoleArn": "arn:sca"}},
	)
	params := serviceParametersValue(map[string]map[string]string{"sca": {"ssoEnable": "false", "ssoRegion": "us-east-1"}})
	attrs := map[string]attr.Value{"services": services, "service_parameters": params}

	request := convertUpdate(t, attrs, attrs)

	if request.PriorState() == nil {
		t.Fatal("PriorState() is nil; an unchanged apply must have prior state set")
	}
	if len(request.PriorState().Services) != len(request.Services) {
		t.Fatalf("len(PriorState().Services) = %d, len(Services) = %d; an unchanged apply must carry both in full",
			len(request.PriorState().Services), len(request.Services))
	}
	for i := range request.Services {
		desired, prior := request.Services[i], request.PriorState().Services[i]
		if desired.ServiceName != prior.ServiceName {
			t.Errorf("Services[%d] is %q but PriorState().Services[%d] is %q; the two lists must stay aligned",
				i, desired.ServiceName, i, prior.ServiceName)
		}
		for key, want := range desired.Resources {
			if got := prior.Resources[key]; got != want {
				t.Errorf("%s prior resource %q = %v, want %v; an unchanged service that looks changed is re-sent and 501s",
					desired.ServiceName, key, got, want)
			}
		}
	}
	for serviceName, entries := range request.ServiceParameters {
		for key, want := range entries {
			if got := request.PriorState().ServiceParameters[serviceName][key]; got != want {
				t.Errorf("%s prior parameter %q = %v, want %v", serviceName, key, got, want)
			}
		}
	}
}

// TestOrganizationAccount_AddedServiceIsAbsentFromPriorState covers adding a service to an account
// that already has one. The new service has no prior entry, which is exactly how the SDK recognises
// it as new, and the existing service's prior entry must still be intact so it is not dragged along.
func TestOrganizationAccount_AddedServiceIsAbsentFromPriorState(t *testing.T) {
	request := convertUpdate(t,
		map[string]attr.Value{"services": servicesValue(
			serviceSpec{name: "dpa", version: "0.0.3", resources: map[string]string{"DpaRoleArn": "arn:dpa"}},
			serviceSpec{name: "sca", version: "0.0.1", resources: map[string]string{"ScaRoleArn": "arn:sca"}},
		)},
		map[string]attr.Value{"services": servicesValue(
			serviceSpec{name: "dpa", version: "0.0.3", resources: map[string]string{"DpaRoleArn": "arn:dpa"}},
		)},
	)

	if len(request.Services) != 2 {
		t.Fatalf("len(Services) = %d, want both the existing and the added service", len(request.Services))
	}
	if request.PriorState() == nil || len(request.PriorState().Services) != 1 {
		priorLen := 0
		if request.PriorState() != nil {
			priorLen = len(request.PriorState().Services)
		}
		t.Fatalf("len(PriorState().Services) = %d, want only the service the last apply sent", priorLen)
	}
	if name := request.PriorState().Services[0].ServiceName; name != "dpa" {
		t.Errorf("PriorState().Services[0].ServiceName = %q, want dpa", name)
	}
	if got := request.PriorState().Services[0].Resources["DpaRoleArn"]; got != "arn:dpa" {
		t.Errorf("the existing service's prior resources were lost: got %v, want arn:dpa; "+
			"it would then look changed and be re-sent alongside the new service", got)
	}
}

// TestOrganizationAccount_RemovedServiceRemainsInPriorState covers dropping a service from the
// configuration. The prior list still holds it, which is the only way to notice it was removed.
func TestOrganizationAccount_RemovedServiceRemainsInPriorState(t *testing.T) {
	request := convertUpdate(t,
		map[string]attr.Value{"services": servicesValue(
			serviceSpec{name: "dpa", version: "0.0.3"},
		)},
		map[string]attr.Value{"services": servicesValue(
			serviceSpec{name: "dpa", version: "0.0.3"},
			serviceSpec{name: "sca", version: "0.0.1"},
		)},
	)

	if len(request.Services) != 1 {
		t.Fatalf("len(Services) = %d, want only the remaining service", len(request.Services))
	}
	if request.PriorState() == nil || len(request.PriorState().Services) != 2 {
		priorLen := 0
		if request.PriorState() != nil {
			priorLen = len(request.PriorState().Services)
		}
		t.Fatalf("len(PriorState().Services) = %d, want both services the last apply sent", priorLen)
	}
}

// TestOrganizationAccount_EmptyPriorServicesWhenStateHasNone covers state predating the PriorState
// mechanism, or an imported resource. There is nothing to compare against, and the SDK treats a
// service it cannot prove changed as unchanged rather than re-sending it on a guess.
func TestOrganizationAccount_EmptyPriorServicesWhenStateHasNone(t *testing.T) {
	request := convertUpdate(t,
		map[string]attr.Value{"services": servicesValue(serviceSpec{name: "dpa", version: "0.0.3"})},
		map[string]attr.Value{},
	)

	if len(request.Services) != 1 {
		t.Fatalf("len(Services) = %d, want 1", len(request.Services))
	}
	priorLen := 0
	if request.PriorState() != nil {
		priorLen = len(request.PriorState().Services)
	}
	if priorLen != 0 {
		t.Errorf("len(PriorState().Services) = %d, want none when prior state has no services", priorLen)
	}
}

// TestOrganizationAccount_VersionChangeIsVisibleInPriorState pins that a per-service version bump is
// comparable too. Version is the one service field that is not a resource, and an upgrade that looked
// unchanged would never be sent.
func TestOrganizationAccount_VersionChangeIsVisibleInPriorState(t *testing.T) {
	request := convertUpdate(t,
		map[string]attr.Value{"services": servicesValue(serviceSpec{name: "dpa", version: "0.0.4"})},
		map[string]attr.Value{"services": servicesValue(serviceSpec{name: "dpa", version: "0.0.3"})},
	)

	if got := request.Services[0].Version; got != "0.0.4" {
		t.Errorf("Services[0].Version = %q, want the planned 0.0.4", got)
	}
	if request.PriorState() == nil || len(request.PriorState().Services) == 0 {
		t.Fatal("PriorState().Services is empty; expected the prior service to be set")
	}
	if got := request.PriorState().Services[0].Version; got != "0.0.3" {
		t.Errorf("PriorState().Services[0].Version = %q, want the prior 0.0.3", got)
	}
}
