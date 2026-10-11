// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package cce

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	awsactions "github.com/cyberark/idsec-sdk-golang/pkg/services/cce/aws/actions"
	azureactions "github.com/cyberark/idsec-sdk-golang/pkg/services/cce/azure/actions"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
	"github.com/cyberark/terraform-provider-idsec/internal/schemas"
)

// cceVersionOptInResources are the CCE resources whose update model declares CCEVersion as a
// pointer, so that UpdateOnlyModified leaves an unchanged cce_version out of the request. The API
// reads the mere presence of a cceVersion as a request to upgrade 'cce', so re-sending the deployed
// value fails an ordinary service upgrade with 501 FEATURE_NOT_IMPLEMENTED.
var cceVersionOptInResources = []struct {
	serviceName string
	actionName  string
	schemaMap   map[string]interface{}
}{
	{"cce-aws", "cce-aws-account", awsactions.ActionToSchemaMap},
	{"cce-aws", "cce-aws-organization", awsactions.ActionToSchemaMap},
	{"cce-azure", "cce-azure-entra", azureactions.ActionToSchemaMap},
	{"cce-azure", "cce-azure-management-group", azureactions.ActionToSchemaMap},
	{"cce-azure", "cce-azure-subscription", azureactions.ActionToSchemaMap},
}

func cceResourceByName(t *testing.T, serviceName, actionName string) *tfactions.IdsecServiceTerraformResourceActionDefinition {
	t.Helper()
	for _, cfg := range tfactions.AllTerraformConfigs() {
		if cfg.ServiceName != serviceName {
			continue
		}
		for _, res := range cfg.Resources {
			if res.ActionName == actionName {
				return res
			}
		}
	}
	t.Fatalf("resource %q not registered under service %q", actionName, serviceName)
	return nil
}

func nullAttrValue(t *testing.T, ctx context.Context, attrType attr.Type) attr.Value {
	t.Helper()
	v, err := attrType.ValueFromTerraform(ctx, tftypes.NewValue(attrType.TerraformType(ctx), nil))
	if err != nil {
		t.Fatalf("failed to build null value for %s: %v", attrType, err)
	}
	return v
}

// oneServiceValue builds a services value holding a single dpa entry, matching whatever
// collection and element type the generated schema chose for the attribute.
func oneServiceValue(t *testing.T, ctx context.Context, servicesType attr.Type) attr.Value {
	t.Helper()
	if _, isDynamic := servicesType.(basetypes.DynamicType); isDynamic {
		// The shape Terraform produces for `services = [{ service_name = "dpa", ... }]` in HCL.
		elemTypes := map[string]attr.Type{
			"service_name": types.StringType,
			"version":      types.StringType,
			"resources":    types.ObjectType{AttrTypes: map[string]attr.Type{}},
		}
		elem := types.ObjectValueMust(elemTypes, map[string]attr.Value{
			"service_name": types.StringValue("dpa"),
			"version":      types.StringValue("0.0.3"),
			"resources":    types.ObjectValueMust(map[string]attr.Type{}, map[string]attr.Value{}),
		})
		return types.DynamicValue(types.TupleValueMust([]attr.Type{types.ObjectType{AttrTypes: elemTypes}}, []attr.Value{elem}))
	}
	typed, ok := servicesType.(attr.TypeWithElementType)
	if !ok {
		t.Fatalf("services attribute type %T has no element type", servicesType)
	}
	elemType, ok := typed.ElementType().(types.ObjectType)
	if !ok {
		t.Fatalf("services element type %T is not an object", typed.ElementType())
	}
	elemAttrs := make(map[string]attr.Value, len(elemType.AttrTypes))
	for name, attrType := range elemType.AttrTypes {
		elemAttrs[name] = nullAttrValue(t, ctx, attrType)
	}
	elemAttrs["service_name"] = types.StringValue("dpa")
	elemAttrs["version"] = types.StringValue("0.0.3")
	elem := types.ObjectValueMust(elemType.AttrTypes, elemAttrs)
	switch servicesType.(type) {
	case types.ListType:
		return types.ListValueMust(elemType, []attr.Value{elem})
	case types.SetType:
		return types.SetValueMust(elemType, []attr.Value{elem})
	default:
		t.Fatalf("unsupported services collection type %T", servicesType)
		return nil
	}
}

// rawObject builds a resource-level value with every attribute null except the given overrides.
func rawObject(t *testing.T, ctx context.Context, s schema.Schema, overrides map[string]attr.Value) tftypes.Value {
	t.Helper()
	objType, ok := s.Type().(types.ObjectType)
	if !ok {
		t.Fatalf("schema type %T is not an object", s.Type())
	}
	values := make(map[string]attr.Value, len(objType.AttrTypes))
	for name, attrType := range objType.AttrTypes {
		values[name] = nullAttrValue(t, ctx, attrType)
	}
	for name, v := range overrides {
		values[name] = v
	}
	raw, err := types.ObjectValueMust(objType.AttrTypes, values).ToTerraformValue(ctx)
	if err != nil {
		t.Fatalf("failed to convert object to terraform value: %v", err)
	}
	return raw
}

// cceVersionOf returns the CCEVersion field of a CCE update request, failing if it is not *string.
func cceVersionOf(t *testing.T, request interface{}) *string {
	t.Helper()
	field := reflect.ValueOf(request).Elem().FieldByName("CCEVersion")
	if !field.IsValid() {
		t.Fatalf("%T has no CCEVersion field", request)
	}
	version, ok := field.Interface().(*string)
	if !ok {
		t.Fatalf("%T.CCEVersion must be *string so UpdateOnlyModified can omit it, got %s", request, field.Type())
	}
	return version
}

func servicesLenOf(request interface{}) int {
	return reflect.ValueOf(request).Elem().FieldByName("Services").Len()
}

// TestCCEResources_UpdateOnlyModifiedOmitsUnchangedCCEVersion runs each opted-in CCE resource's real
// generated schema and update model through the update conversion path.
func TestCCEResources_UpdateOnlyModifiedOmitsUnchangedCCEVersion(t *testing.T) {
	ctx := context.Background()

	for _, rc := range cceVersionOptInResources {
		t.Run(rc.actionName, func(t *testing.T) {
			res := cceResourceByName(t, rc.serviceName, rc.actionName)
			if !res.UpdateOnlyModified {
				t.Fatal("UpdateOnlyModified must be enabled, or the deployed cce_version is re-sent on every update")
			}
			createModel := rc.schemaMap[res.ActionsMappings[tfactions.CreateOperation]]
			updateModel := rc.schemaMap[res.ActionsMappings[tfactions.UpdateOperation]]
			if createModel == nil || updateModel == nil {
				t.Fatalf("create/update models not found for %q", rc.actionName)
			}
			s, diags := schemas.GenerateResourceSchemaFromStruct(createModel, updateModel, res.StateSchema, nil,
				res.ExtraRequiredAttributes, nil, nil, nil, res.ComputedAttributes, nil, nil, nil)
			if diags.HasError() {
				t.Fatalf("schema generation failed: %v", diags)
			}
			objType, ok := s.Type().(types.ObjectType)
			if !ok {
				t.Fatalf("schema type = %T, want types.ObjectType", s.Type())
			}
			services := oneServiceValue(t, ctx, objType.AttrTypes["services"])
			diffOpts := schemas.DiffOnUpdateOptions{Enabled: res.UpdateOnlyModified}
			convert := func(planCCEVersion attr.Value) interface{} {
				t.Helper()
				id := types.StringValue("onboarding-1")
				state := &tfsdk.State{Schema: s, Raw: rawObject(t, ctx, s, map[string]attr.Value{
					"id": id, "services": services, "cce_version": types.StringValue("0.0.1"),
				})}
				plan := &tfsdk.Plan{Schema: s, Raw: rawObject(t, ctx, s, map[string]attr.Value{
					"id": id, "services": services, "cce_version": planCCEVersion,
				})}
				request, err := schemas.StructFromPlanAndStateObject(ctx, plan, state, updateModel, res.StateSchema, diffOpts)
				if err != nil {
					t.Fatalf("StructFromPlanAndStateObject() error = %v", err)
				}
				return request
			}

			t.Run("unchanged_cce_version_is_nil_services_still_sent", func(t *testing.T) {
				request := convert(types.StringValue("0.0.1"))
				if v := cceVersionOf(t, request); v != nil {
					t.Errorf("CCEVersion = %q, want nil (unchanged cce_version must not be sent)", *v)
				}
				if n := servicesLenOf(request); n != 1 {
					t.Errorf("len(Services) = %d, want 1 (services is the full desired set and must always be sent)", n)
				}
			})

			t.Run("changed_cce_version_is_sent", func(t *testing.T) {
				request := convert(types.StringValue("0.2.0"))
				if v := cceVersionOf(t, request); v == nil || *v != "0.2.0" {
					t.Errorf("CCEVersion = %v, want 0.2.0", v)
				}
			})

			// A cce_version the user had configured and then removed plans as unknown, so the converter
			// prefills it from state; ClearRemovedAttributes, which runs next on every update, must then
			// clear it so the deployed version is not re-sent.
			t.Run("cce_version_removed_from_config_is_cleared", func(t *testing.T) {
				request := convert(types.StringUnknown())
				if v := cceVersionOf(t, request); v == nil || *v != "0.0.1" {
					t.Fatalf("precondition: CCEVersion = %v, want prefilled 0.0.1 from state", v)
				}
				config := &tfsdk.Config{Schema: s, Raw: rawObject(t, ctx, s, map[string]attr.Value{
					"id": types.StringValue("onboarding-1"), "services": services,
				})}
				state := &tfsdk.State{Schema: s, Raw: rawObject(t, ctx, s, map[string]attr.Value{
					"id": types.StringValue("onboarding-1"), "services": services, "cce_version": types.StringValue("0.0.1"),
				})}
				err := schemas.ClearRemovedAttributes(ctx, request, config, state, res.ComputedAttributes, nil, map[string]bool{"cce_version": true})
				if err != nil {
					t.Fatalf("ClearRemovedAttributes() error = %v", err)
				}
				if v := cceVersionOf(t, request); v != nil {
					t.Errorf("CCEVersion = %q, want nil (a removed cce_version must not re-send the deployed version)", *v)
				}
			})
		})
	}
}
