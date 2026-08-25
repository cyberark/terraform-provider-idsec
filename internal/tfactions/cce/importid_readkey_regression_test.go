// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package cce

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/mitchellh/mapstructure"

	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
	"github.com/cyberark/terraform-provider-idsec/internal/schemas"
)

// cceUpdatableResources returns the registered CCE AWS/Azure resources that both
// support Update and treat "id" as a computed attribute. These are precisely the
// resources exposed to the empty-id-on-update regression, so every read-key
// invariant below is asserted against exactly this set.
func cceUpdatableResources(t *testing.T) []*tfactions.IdsecServiceTerraformResourceActionDefinition {
	t.Helper()
	var out []*tfactions.IdsecServiceTerraformResourceActionDefinition
	for _, cfg := range tfactions.AllTerraformConfigs() {
		if cfg.ServiceName != "cce-aws" && cfg.ServiceName != "cce-azure" {
			continue
		}
		for _, res := range cfg.Resources {
			if !slices.Contains(res.SupportedOperations, tfactions.UpdateOperation) {
				continue
			}
			if !slices.Contains(res.ComputedAttributes, "id") {
				continue
			}
			out = append(out, res)
		}
	}
	return out
}

// mapstructureStringField returns the string value of the field on structVal
// whose top-level mapstructure name matches name.
func mapstructureStringField(structVal reflect.Value, name string) (string, bool) {
	for structVal.Kind() == reflect.Pointer {
		if structVal.IsNil() {
			return "", false
		}
		structVal = structVal.Elem()
	}
	if structVal.Kind() != reflect.Struct {
		return "", false
	}
	structType := structVal.Type()
	for i := 0; i < structType.NumField(); i++ {
		tag := structType.Field(i).Tag.Get("mapstructure")
		if tag == "" {
			continue
		}
		tagName, _, _ := strings.Cut(tag, ",")
		if tagName == name && structVal.Field(i).Kind() == reflect.String {
			return structVal.Field(i).String(), true
		}
	}
	return "", false
}

// TestCCEResources_UpdateWithComputedIDDeclareImportID is a regression guard for the
// empty-ID-on-update bug (provider v0.5.0, PR #197 "OLY-17512: Fix mismatch in
// mergePlanAndStateMap"). That change started stripping computed attributes from
// the update payload, preserving only the ImportID read key. The CCE AWS/Azure
// resources declared "id" as computed but did NOT set ImportID, so on update "id"
// was zeroed and the SDK issued a get-details request with an empty id
// (e.g. `GET /api/aws/programmatic/account/`), which the tenant rejects with a
// generic 403.
//
// Invariant: any CCE resource that supports Update AND lists "id" as a computed
// attribute MUST declare an ImportID that includes "id"; otherwise its read key is
// stripped from the update request.
func TestCCEResources_UpdateWithComputedIDDeclareImportID(t *testing.T) {
	var checked int
	for _, cfg := range tfactions.AllTerraformConfigs() {
		if cfg.ServiceName != "cce-aws" && cfg.ServiceName != "cce-azure" {
			continue
		}
		for _, res := range cfg.Resources {
			t.Run(cfg.ServiceName+"/"+res.ActionDefinitionName(), func(t *testing.T) {
				if !slices.Contains(res.SupportedOperations, tfactions.UpdateOperation) {
					t.Skip("resource does not support update")
				}
				if !slices.Contains(res.ComputedAttributes, "id") {
					t.Skip("resource does not treat 'id' as computed")
				}
				checked++
				if res.ImportID == "" {
					t.Fatalf("'id' is computed and the resource supports update, but ImportID is empty: " +
						"the update payload will strip 'id' (empty-id get-details -> 403). Set ImportID: \"id\".")
				}
				if !slices.Contains(schemas.SplitImportIDAttributes(res.ImportID), "id") {
					t.Fatalf("ImportID %q does not include the computed read key 'id': "+
						"update would strip 'id' from the request.", res.ImportID)
				}
			})
		}
	}

	if checked == 0 {
		t.Fatal("expected at least one CCE resource with a computed 'id' and update support to be checked; " +
			"registration may have changed")
	}
}

// TestCCEResources_ImportIDResolvesOnStateSchema guards against a subtler variant
// of the regression: declaring an ImportID whose read key does not actually exist
// on the resource's state schema. Such a key silently fails to protect anything
// from the computed-attribute stripping, reintroducing the empty-id -> 403 bug.
// Every ImportID attribute must resolve to a real string/int field on the SDK
// state schema for both CCE AWS and Azure resources.
func TestCCEResources_ImportIDResolvesOnStateSchema(t *testing.T) {
	resources := cceUpdatableResources(t)
	if len(resources) == 0 {
		t.Fatal("expected at least one updatable CCE resource with a computed 'id'; registration may have changed")
	}
	for _, res := range resources {
		t.Run(res.ActionDefinitionName(), func(t *testing.T) {
			if res.ImportID == "" {
				t.Fatalf("ImportID is empty; cannot protect the computed read key from update stripping")
			}
			if res.StateSchema == nil {
				t.Fatalf("StateSchema is nil; cannot validate ImportID %q", res.ImportID)
			}
			for _, attr := range schemas.SplitImportIDAttributes(res.ImportID) {
				if err := schemas.ValidateStateSchemaImportAttribute(res.StateSchema, attr); err != nil {
					t.Fatalf("ImportID attribute %q does not resolve on the state schema: %v", attr, err)
				}
			}
		})
	}
}

// TestCCEResources_UpdatePreservesReadKeyOnStateSchema exercises the real fix
// end-to-end on each resource's actual SDK state schema: after populating the
// read key and stripping computed attributes with the declared ImportID as the
// skip list (mirroring what the update path does), the "id" read key must
// survive. As a control, the same strip with no skip list must wipe it - the
// exact regression that yielded the empty-id get-details -> 403.
func TestCCEResources_UpdatePreservesReadKeyOnStateSchema(t *testing.T) {
	const readKey = "read-key-123"

	resources := cceUpdatableResources(t)
	if len(resources) == 0 {
		t.Fatal("expected at least one updatable CCE resource with a computed 'id'; registration may have changed")
	}
	for _, res := range resources {
		t.Run(res.ActionDefinitionName(), func(t *testing.T) {
			if res.StateSchema == nil {
				t.Fatalf("StateSchema is nil; cannot exercise clear-computed behavior")
			}
			schemaType := reflect.TypeOf(res.StateSchema)
			for schemaType.Kind() == reflect.Pointer {
				schemaType = schemaType.Elem()
			}

			// Skipped read key survives.
			preserved := reflect.New(schemaType).Interface()
			if err := mapstructure.Decode(map[string]interface{}{"id": readKey}, preserved); err != nil {
				t.Fatalf("failed to seed state schema with id: %v", err)
			}
			if err := schemas.ClearComputedAttributes(preserved, res.ComputedAttributes, schemas.SplitImportIDAttributes(res.ImportID)); err != nil {
				t.Fatalf("ClearComputedAttributes (with skip) returned error: %v", err)
			}
			if got, ok := mapstructureStringField(reflect.ValueOf(preserved), "id"); !ok || got != readKey {
				t.Fatalf("read key 'id' must survive update stripping when declared in ImportID %q, got %q (found=%t)",
					res.ImportID, got, ok)
			}

			// Control: without the skip list the read key is wiped.
			wiped := reflect.New(schemaType).Interface()
			if err := mapstructure.Decode(map[string]interface{}{"id": readKey}, wiped); err != nil {
				t.Fatalf("failed to seed state schema with id: %v", err)
			}
			if err := schemas.ClearComputedAttributes(wiped, res.ComputedAttributes, nil); err != nil {
				t.Fatalf("ClearComputedAttributes (no skip) returned error: %v", err)
			}
			if got, ok := mapstructureStringField(reflect.ValueOf(wiped), "id"); !ok || got != "" {
				t.Fatalf("control: computed 'id' must be wiped without a skip list, got %q (found=%t)", got, ok)
			}
		})
	}
}
