// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package cce

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
	"github.com/cyberark/terraform-provider-idsec/internal/schemas"
)

// servicesDataCases lists every details data source that must expose services_data, together with
// the resource that shares its state struct and must NOT gain the attribute (backward compatibility).
var servicesDataCases = []struct {
	service    string
	actionName string
}{
	{"cce-azure", "cce-azure-entra"},
	{"cce-azure", "cce-azure-management-group"},
	{"cce-azure", "cce-azure-subscription"},
	{"cce-aws", "cce-aws-account"},
}

// TestDetailsDataSourcesExposeServicesData verifies that the details data sources expose a typed
// services_data list with the per-service name, status, version and available_versions (EV-150750).
func TestDetailsDataSourcesExposeServicesData(t *testing.T) {
	for _, tc := range servicesDataCases {
		t.Run(tc.actionName, func(t *testing.T) {
			ds := cceDataSourceByName(t, tc.service, tc.actionName)
			generated := schemas.GenerateDataSourceSchemaFromStruct(nil, ds.StateSchema, nil, ds.ExtraRequiredAttributes, nil)

			servicesData, ok := generated.Attributes["services_data"]
			if !ok {
				t.Fatalf("services_data attribute missing from the %s data source schema", tc.actionName)
			}
			if !servicesData.IsComputed() {
				t.Errorf("services_data must be computed on the %s data source", tc.actionName)
			}

			nested, ok := servicesData.(schema.ListNestedAttribute)
			if !ok {
				t.Fatalf("services_data must be a typed nested list, got %T", servicesData)
			}

			for _, attr := range []string{"name", "status", "version"} {
				got, ok := nested.NestedObject.Attributes[attr]
				if !ok {
					t.Fatalf("services_data[].%s missing", attr)
				}
				if got.GetType().String() != "basetypes.StringType" {
					t.Errorf("services_data[].%s must be a string, got %s", attr, got.GetType())
				}
			}

			available, ok := nested.NestedObject.Attributes["available_versions"]
			if !ok {
				t.Fatal("services_data[].available_versions missing")
			}
			if _, isList := available.(schema.ListAttribute); !isList {
				t.Errorf("services_data[].available_versions must be a list of strings, got %T", available)
			}

			// Existing attributes must be untouched.
			for _, attr := range []string{"id", "onboarding_type", "status", "cce_version"} {
				if _, ok := generated.Attributes[attr]; !ok {
					t.Errorf("existing attribute %q missing from the %s data source schema", attr, tc.actionName)
				}
			}
		})
	}
}

// upgradeInfoCases lists every details data source that must expose the workspace-level upgrade
// attributes, together with the resource that shares its base state struct.
var upgradeInfoCases = append([]struct {
	service    string
	actionName string
}{{"cce-aws", "cce-aws-organization"}}, servicesDataCases...)

// upgradeInfoTags are the workspace-level upgrade attributes added to the details data sources (EV-150750).
var upgradeInfoTags = []string{"cce_available_versions", "upgrade_available"}

// TestDetailsDataSourcesExposeUpgradeInfo verifies that the details data sources expose the computed
// cce_available_versions list and the upgrade_available boolean.
func TestDetailsDataSourcesExposeUpgradeInfo(t *testing.T) {
	for _, tc := range upgradeInfoCases {
		t.Run(tc.actionName, func(t *testing.T) {
			ds := cceDataSourceByName(t, tc.service, tc.actionName)
			generated := schemas.GenerateDataSourceSchemaFromStruct(nil, ds.StateSchema, nil, ds.ExtraRequiredAttributes, nil)

			versions, ok := generated.Attributes["cce_available_versions"]
			if !ok {
				t.Fatalf("cce_available_versions attribute missing from the %s data source schema", tc.actionName)
			}
			if _, isList := versions.(schema.ListAttribute); !isList {
				t.Errorf("cce_available_versions must be a list of strings, got %T", versions)
			}
			if !versions.IsComputed() {
				t.Errorf("cce_available_versions must be computed on the %s data source", tc.actionName)
			}

			upgrade, ok := generated.Attributes["upgrade_available"]
			if !ok {
				t.Fatalf("upgrade_available attribute missing from the %s data source schema", tc.actionName)
			}
			if _, isBool := upgrade.(schema.BoolAttribute); !isBool {
				t.Errorf("upgrade_available must be a boolean, got %T", upgrade)
			}
			if !upgrade.IsComputed() {
				t.Errorf("upgrade_available must be computed on the %s data source", tc.actionName)
			}
		})
	}
}

// TestDetailsResourcesDoNotExposeServicesData guards backward compatibility: only the data sources
// gain services_data and the upgrade attributes, the resources keep their existing state schema.
func TestDetailsResourcesDoNotExposeServicesData(t *testing.T) {
	forbidden := append([]string{"services_data"}, upgradeInfoTags...)
	for _, tc := range upgradeInfoCases {
		t.Run(tc.actionName, func(t *testing.T) {
			var resource *tfactions.IdsecServiceTerraformResourceActionDefinition
			for _, cfg := range tfactions.AllTerraformConfigs() {
				if cfg.ServiceName != tc.service {
					continue
				}
				for _, r := range cfg.Resources {
					if r.ActionName == tc.actionName {
						resource = r
					}
				}
			}
			if resource == nil {
				t.Fatalf("resource %q not registered under service %q", tc.actionName, tc.service)
			}

			typ := reflect.TypeOf(resource.StateSchema)
			for typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}
			for i := 0; i < typ.NumField(); i++ {
				name := strings.Split(typ.Field(i).Tag.Get("mapstructure"), ",")[0]
				if slices.Contains(forbidden, name) {
					t.Errorf("resource %s state struct must not contain %s", tc.actionName, name)
				}
			}
		})
	}
}
