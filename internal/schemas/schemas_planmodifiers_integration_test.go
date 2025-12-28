// Copyright (c) CyberArk.
// SPDX-License-Identifier: Apache-2.0

// Package schemas provides integration tests for plan modifiers.
//
// This file contains integration tests that validate immutable plan modifier behavior
// using a complete provider and resource implementation. Unlike the unit tests in
// schemas_planmodifiers_test.go which test modifiers in isolation, these tests:
//
//   - Create a real Terraform provider and resource (no network calls)
//   - Verify that resources can be created with immutable fields
//   - Verify that mutable fields can be updated
//   - Verify that attempts to modify immutable fields are blocked with appropriate errors
//   - Test multiple immutable field types (string, int64, bool)
//   - Validate error message content
//
// The tests use resource.UnitTest() and can run without the TF_ACC environment variable.
package schemas

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestImmutableModifier_Integration tests plan modifier behavior with a real provider and resource.
//
// This integration test verifies that immutable modifiers correctly block field changes
// during resource updates. It creates a dummy provider and resource that make no network
// calls, and validates that attempting to modify an immutable field results in an error.
//
// The test covers:
//   - Resource creation succeeds with immutable field set
//   - Resource update with no changes succeeds
//   - Resource update attempting to change immutable field fails with appropriate error
func TestImmutableModifier_Integration(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"test": providerserver.NewProtocol6WithError(&testProvider{}),
		},
		Steps: []resource.TestStep{
			// Step 1: Create resource with immutable_id
			{
				Config: testImmutableResourceConfig("initial-id", "mutable-value"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("test_immutable_resource.test", "immutable_id", "initial-id"),
					resource.TestCheckResourceAttr("test_immutable_resource.test", "mutable_field", "mutable-value"),
				),
			},
			// Step 2: Update only mutable field - should succeed
			{
				Config: testImmutableResourceConfig("initial-id", "updated-value"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("test_immutable_resource.test", "immutable_id", "initial-id"),
					resource.TestCheckResourceAttr("test_immutable_resource.test", "mutable_field", "updated-value"),
				),
			},
			// Step 3: Attempt to update immutable field - should fail
			{
				Config:      testImmutableResourceConfig("changed-id", "updated-value"),
				ExpectError: regexp.MustCompile("Immutable Attribute Cannot Be Changed"),
			},
		},
	})
}

// testImmutableResourceConfig generates Terraform configuration for testing.
//
// Parameters:
//   - immutableID: Value for the immutable_id field
//   - mutableField: Value for the mutable_field field
//
// Returns Terraform configuration string.
func testImmutableResourceConfig(immutableID, mutableField string) string {
	return fmt.Sprintf(`
provider "test" {}

resource "test_immutable_resource" "test" {
  immutable_id  = %[1]q
  mutable_field = %[2]q
}
`, immutableID, mutableField)
}

// testProvider implements a dummy provider for integration testing.
//
// This provider makes no network calls and is used solely for testing plan
// modifier behavior in a realistic provider context.
type testProvider struct{}

// Metadata sets the provider type name.
//
// Parameters:
//   - ctx: Context for the operation
//   - req: Metadata request (unused)
//   - resp: Metadata response where the type name is set
func (p *testProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "test"
}

// Schema returns the provider schema.
//
// This provider has no configuration attributes.
//
// Parameters:
//   - ctx: Context for the operation
//   - req: Schema request (unused)
//   - resp: Schema response where the schema is set
func (p *testProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = providerschema.Schema{
		Description: "Test provider for plan modifier integration testing",
	}
}

// Configure configures the provider.
//
// This provider requires no configuration and makes no network calls.
//
// Parameters:
//   - ctx: Context for the operation
//   - req: Configuration request (unused)
//   - resp: Configuration response (unused)
func (p *testProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// No configuration needed for test provider
}

// Resources returns the list of resources supported by this provider.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns a function that returns the list of resource constructors.
func (p *testProvider) Resources(ctx context.Context) []func() fwresource.Resource {
	return []func() fwresource.Resource{
		func() fwresource.Resource {
			return &testImmutableResource{}
		},
	}
}

// DataSources returns the list of data sources supported by this provider.
//
// This provider has no data sources.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns an empty slice.
func (p *testProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

// testImmutableResource implements a test resource with an immutable field.
//
// This resource makes no network calls and stores all state in memory for testing.
type testImmutableResource struct{}

// testImmutableResourceModel defines the resource's data model.
type testImmutableResourceModel struct {
	ID           types.String `tfsdk:"id"`
	ImmutableID  types.String `tfsdk:"immutable_id"`
	MutableField types.String `tfsdk:"mutable_field"`
}

// Metadata sets the resource type name.
//
// Parameters:
//   - ctx: Context for the operation
//   - req: Metadata request containing provider type name
//   - resp: Metadata response where the type name is set
func (r *testImmutableResource) Metadata(ctx context.Context, req fwresource.MetadataRequest, resp *fwresource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_immutable_resource"
}

// Schema returns the resource schema with an immutable field.
//
// The schema includes:
//   - id: Computed identifier
//   - immutable_id: Required string with ImmutableString modifier
//   - mutable_field: Required string that can be updated
//
// Parameters:
//   - ctx: Context for the operation
//   - req: Schema request (unused)
//   - resp: Schema response where the schema is set
func (r *testImmutableResource) Schema(ctx context.Context, req fwresource.SchemaRequest, resp *fwresource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Test resource with immutable field for plan modifier integration testing",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Resource identifier",
				Computed:    true,
			},
			"immutable_id": schema.StringAttribute{
				Description: "Immutable identifier that cannot be changed after creation",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					ImmutableString(),
				},
			},
			"mutable_field": schema.StringAttribute{
				Description: "Mutable field that can be updated",
				Required:    true,
			},
		},
	}
}

// Create handles resource creation.
//
// This method creates a new resource instance in memory without making any
// network calls. It sets the ID and copies all fields from the plan to state.
//
// Parameters:
//   - ctx: Context for the operation
//   - req: Create request containing the planned resource state
//   - resp: Create response where the new state is set
func (r *testImmutableResource) Create(ctx context.Context, req fwresource.CreateRequest, resp *fwresource.CreateResponse) {
	var plan testImmutableResourceModel

	// Read plan data
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set ID to immutable_id value for simplicity
	plan.ID = plan.ImmutableID

	// Save state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Read handles resource reading.
//
// This method refreshes the resource state. Since the resource is entirely
// in-memory with no external state, it simply returns the current state.
//
// Parameters:
//   - ctx: Context for the operation
//   - req: Read request containing the current state
//   - resp: Read response where the refreshed state is set
func (r *testImmutableResource) Read(ctx context.Context, req fwresource.ReadRequest, resp *fwresource.ReadResponse) {
	var state testImmutableResourceModel

	// Read current state
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// No external state to refresh, just return current state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update handles resource updates.
//
// This method updates the resource state with new values from the plan.
// The immutable_id field is protected by the plan modifier and should not
// change if the modifier is working correctly.
//
// Parameters:
//   - ctx: Context for the operation
//   - req: Update request containing current state and planned changes
//   - resp: Update response where the updated state is set
func (r *testImmutableResource) Update(ctx context.Context, req fwresource.UpdateRequest, resp *fwresource.UpdateResponse) {
	var plan testImmutableResourceModel
	var state testImmutableResourceModel

	// Read current state
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read plan
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update mutable field only (immutable_id should be unchanged)
	state.MutableField = plan.MutableField

	// Save updated state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Delete handles resource deletion.
//
// This method removes the resource from state. No network calls are made.
//
// Parameters:
//   - ctx: Context for the operation
//   - req: Delete request containing the current state
//   - resp: Delete response (state is automatically cleared)
func (r *testImmutableResource) Delete(ctx context.Context, req fwresource.DeleteRequest, resp *fwresource.DeleteResponse) {
	// No-op for in-memory resource
	// State is automatically cleared by the framework
}

// TestImmutableModifier_Integration_MultipleFields tests multiple immutable fields.
//
// This test verifies that multiple fields can be marked as immutable and that
// the plan modifier correctly blocks changes to any of them.
func TestImmutableModifier_Integration_MultipleFields(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"test": providerserver.NewProtocol6WithError(&testProviderMultiField{}),
		},
		Steps: []resource.TestStep{
			// Step 1: Create resource
			{
				Config: testMultiFieldResourceConfig("id-1", 42, true, "mutable"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("test_multifield_resource.test", "immutable_string", "id-1"),
					resource.TestCheckResourceAttr("test_multifield_resource.test", "immutable_int", "42"),
					resource.TestCheckResourceAttr("test_multifield_resource.test", "immutable_bool", "true"),
					resource.TestCheckResourceAttr("test_multifield_resource.test", "mutable_field", "mutable"),
				),
			},
			// Step 2: Update mutable field only - should succeed
			{
				Config: testMultiFieldResourceConfig("id-1", 42, true, "updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("test_multifield_resource.test", "mutable_field", "updated"),
				),
			},
			// Step 3: Attempt to change immutable_string - should fail
			{
				Config:      testMultiFieldResourceConfig("id-2", 42, true, "updated"),
				ExpectError: regexp.MustCompile("Immutable Attribute Cannot Be Changed"),
			},
			// Step 4: Reset and attempt to change immutable_int - should fail
			{
				Config: testMultiFieldResourceConfig("id-1", 42, true, "updated"),
			},
			{
				Config:      testMultiFieldResourceConfig("id-1", 99, true, "updated"),
				ExpectError: regexp.MustCompile("Immutable Attribute Cannot Be Changed"),
			},
			// Step 5: Reset and attempt to change immutable_bool - should fail
			{
				Config: testMultiFieldResourceConfig("id-1", 42, true, "updated"),
			},
			{
				Config:      testMultiFieldResourceConfig("id-1", 42, false, "updated"),
				ExpectError: regexp.MustCompile("Immutable Attribute Cannot Be Changed"),
			},
		},
	})
}

// testMultiFieldResourceConfig generates configuration for multi-field test.
//
// Parameters:
//   - immutableString: Value for immutable_string field
//   - immutableInt: Value for immutable_int field
//   - immutableBool: Value for immutable_bool field
//   - mutableField: Value for mutable_field field
//
// Returns Terraform configuration string.
func testMultiFieldResourceConfig(immutableString string, immutableInt int, immutableBool bool, mutableField string) string {
	return fmt.Sprintf(`
provider "test" {}

resource "test_multifield_resource" "test" {
  immutable_string = %[1]q
  immutable_int    = %[2]d
  immutable_bool   = %[3]t
  mutable_field    = %[4]q
}
`, immutableString, immutableInt, immutableBool, mutableField)
}

// testProviderMultiField implements a provider for multi-field testing.
type testProviderMultiField struct{}

// Metadata sets the provider type name.
//
// Parameters:
//   - ctx: Context for the operation
//   - req: Metadata request (unused)
//   - resp: Metadata response where the type name is set
func (p *testProviderMultiField) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "test"
}

// Schema returns the provider schema.
//
// Parameters:
//   - ctx: Context for the operation
//   - req: Schema request (unused)
//   - resp: Schema response where the schema is set
func (p *testProviderMultiField) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = providerschema.Schema{
		Description: "Test provider for multi-field plan modifier testing",
	}
}

// Configure configures the provider.
//
// Parameters:
//   - ctx: Context for the operation
//   - req: Configuration request (unused)
//   - resp: Configuration response (unused)
func (p *testProviderMultiField) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// No configuration needed
}

// Resources returns the list of resources.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns a function that returns the list of resource constructors.
func (p *testProviderMultiField) Resources(ctx context.Context) []func() fwresource.Resource {
	return []func() fwresource.Resource{
		func() fwresource.Resource {
			return &testMultiFieldResource{}
		},
	}
}

// DataSources returns the list of data sources.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns an empty slice.
func (p *testProviderMultiField) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

// testMultiFieldResource implements a resource with multiple immutable fields.
type testMultiFieldResource struct{}

// testMultiFieldResourceModel defines the resource's data model.
type testMultiFieldResourceModel struct {
	ID              types.String `tfsdk:"id"`
	ImmutableString types.String `tfsdk:"immutable_string"`
	ImmutableInt    types.Int64  `tfsdk:"immutable_int"`
	ImmutableBool   types.Bool   `tfsdk:"immutable_bool"`
	MutableField    types.String `tfsdk:"mutable_field"`
}

// Metadata sets the resource type name.
//
// Parameters:
//   - ctx: Context for the operation
//   - req: Metadata request containing provider type name
//   - resp: Metadata response where the type name is set
func (r *testMultiFieldResource) Metadata(ctx context.Context, req fwresource.MetadataRequest, resp *fwresource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_multifield_resource"
}

// Schema returns the resource schema with multiple immutable fields.
//
// Parameters:
//   - ctx: Context for the operation
//   - req: Schema request (unused)
//   - resp: Schema response where the schema is set
func (r *testMultiFieldResource) Schema(ctx context.Context, req fwresource.SchemaRequest, resp *fwresource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Test resource with multiple immutable fields",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Resource identifier",
				Computed:    true,
			},
			"immutable_string": schema.StringAttribute{
				Description: "Immutable string field",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					ImmutableString(),
				},
			},
			"immutable_int": schema.Int64Attribute{
				Description: "Immutable integer field",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					ImmutableInt64(),
				},
			},
			"immutable_bool": schema.BoolAttribute{
				Description: "Immutable boolean field",
				Required:    true,
				PlanModifiers: []planmodifier.Bool{
					ImmutableBool(),
				},
			},
			"mutable_field": schema.StringAttribute{
				Description: "Mutable field that can be updated",
				Required:    true,
			},
		},
	}
}

// Create handles resource creation.
//
// Parameters:
//   - ctx: Context for the operation
//   - req: Create request containing the planned resource state
//   - resp: Create response where the new state is set
func (r *testMultiFieldResource) Create(ctx context.Context, req fwresource.CreateRequest, resp *fwresource.CreateResponse) {
	var plan testMultiFieldResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set ID
	plan.ID = plan.ImmutableString

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Read handles resource reading.
//
// Parameters:
//   - ctx: Context for the operation
//   - req: Read request containing the current state
//   - resp: Read response where the refreshed state is set
func (r *testMultiFieldResource) Read(ctx context.Context, req fwresource.ReadRequest, resp *fwresource.ReadResponse) {
	var state testMultiFieldResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update handles resource updates.
//
// Parameters:
//   - ctx: Context for the operation
//   - req: Update request containing current state and planned changes
//   - resp: Update response where the updated state is set
func (r *testMultiFieldResource) Update(ctx context.Context, req fwresource.UpdateRequest, resp *fwresource.UpdateResponse) {
	var plan testMultiFieldResourceModel
	var state testMultiFieldResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update mutable field only
	state.MutableField = plan.MutableField

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Delete handles resource deletion.
//
// Parameters:
//   - ctx: Context for the operation
//   - req: Delete request containing the current state
//   - resp: Delete response (state is automatically cleared)
func (r *testMultiFieldResource) Delete(ctx context.Context, req fwresource.DeleteRequest, resp *fwresource.DeleteResponse) {
	// No-op for in-memory resource
}

// TestImmutableModifier_Integration_ErrorMessage validates error message content.
//
// This test verifies that when an immutable field change is blocked, the error
// message contains helpful information about the attribute path and values.
func TestImmutableModifier_Integration_ErrorMessage(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"test": providerserver.NewProtocol6WithError(&testProvider{}),
		},
		Steps: []resource.TestStep{
			// Step 1: Create resource
			{
				Config: testImmutableResourceConfig("original-id", "value"),
			},
			// Step 2: Attempt to change immutable_id and verify error message
			{
				Config:      testImmutableResourceConfig("new-id", "value"),
				ExpectError: regexp.MustCompile("(?s)Immutable Attribute Cannot Be Changed.*immutable_id"),
			},
		},
	})
}

// TestImmutableModifier_Integration_RecreateOnChange validates resource recreation.
//
// This test demonstrates that when an immutable field needs to change, the
// user must recreate the resource by using terraform destroy and recreate.
func TestImmutableModifier_Integration_RecreateOnChange(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"test": providerserver.NewProtocol6WithError(&testProvider{}),
		},
		Steps: []resource.TestStep{
			// Step 1: Create initial resource
			{
				Config: testImmutableResourceConfig("original-id", "value"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("test_immutable_resource.test", "immutable_id", "original-id"),
				),
			},
			// Step 2: Verify that attempting to change immutable_id fails
			{
				Config:      testImmutableResourceConfig("new-id", "value"),
				ExpectError: regexp.MustCompile("Immutable Attribute Cannot Be Changed"),
			},
		},
	})
}
