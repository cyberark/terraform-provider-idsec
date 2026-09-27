// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package cce

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
	"github.com/cyberark/terraform-provider-idsec/internal/schemas"
)

// cceDataSourceByName returns the registered data source definition with the given action name.
func cceDataSourceByName(t *testing.T, serviceName, actionName string) *tfactions.IdsecServiceTerraformDataSourceActionDefinition {
	t.Helper()
	for _, cfg := range tfactions.AllTerraformConfigs() {
		if cfg.ServiceName != serviceName {
			continue
		}
		for _, ds := range cfg.DataSources {
			if ds.ActionName == actionName {
				return ds
			}
		}
	}
	t.Fatalf("data source %q not registered under service %q", actionName, serviceName)
	return nil
}

// TestOrganizationDataSourceExposesServiceVersions guards the input side of EV-143045.
//
// A member account has no version of its own; it inherits the parent organization's. The
// add-account module therefore reads each service's version off this data source and feeds it into
// the account resource's `services`, which is what gives Terraform something to diff when the
// organization is upgraded. If `version` silently disappears from `services_data` the account
// resource goes back to having no config-visible signal, the plan is empty, and the upgrade is a
// silent no-op again - the exact failure this ticket fixed.
func TestOrganizationDataSourceExposesServiceVersions(t *testing.T) {
	ds := cceDataSourceByName(t, "cce-aws", "cce-aws-organization")

	generated := schemas.GenerateDataSourceSchemaFromStruct(
		nil,
		ds.StateSchema,
		nil,
		ds.ExtraRequiredAttributes,
		nil,
	)

	servicesData, ok := generated.Attributes["services_data"]
	if !ok {
		t.Fatal("services_data attribute missing from the organization data source schema")
	}

	nested, ok := servicesData.(schema.ListNestedAttribute)
	if !ok {
		t.Fatalf("services_data must stay a typed nested list so per-service fields are addressable in HCL, got %T", servicesData)
	}

	version, ok := nested.NestedObject.Attributes["version"]
	if !ok {
		t.Fatal("services_data[].version missing: the add-account module cannot derive member-account service versions without it")
	}
	if version.GetType().String() != "basetypes.StringType" {
		t.Errorf("services_data[].version must be a string, got %s", version.GetType())
	}
}
