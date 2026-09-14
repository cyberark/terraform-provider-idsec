// Copyright CyberArk. 2026
// SPDX-License-Identifier: Apache-2.0

// This file contains end-to-end integration tests for hash-mode write-only attributes
// (WriteOnlyHashedAttributes). Unlike the unit tests in schemas_writeonly_hash_test.go, which call
// the schema pass and plan modifier in isolation, these tests stand up a real Terraform provider
// and resource whose schema is produced by the actual GenerateResourceSchemaFromStruct generator
// with a WriteOnlyHashedAttributes declaration, then drive it through genuine create/update/
// destroy apply cycles using the real terraform CLI (no network, no credentials).
//
// They require a terraform binary that supports write-only attributes (Terraform >= 1.11) on PATH.
package schemas

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const woHashResourceAddr = "test_wohash_resource.test"

// woHashGenModel is the SDK-style model the schema generator reflects over to build the resource
// schema. It carries the same mapstructure/desc tags a real resource model would, and deliberately
// does NOT declare the hash sibling: the generator synthesizes "password_write_only_hash" itself.
type woHashGenModel struct {
	Name     string `mapstructure:"name" desc:"a display name" required:"true"`
	Password string `mapstructure:"password" desc:"the secret whose rotation is detected by hash"`
}

// woHashStateModel is the terraform-plugin-framework model used to read plan/config and write
// state. Its tfsdk tags mirror the generated schema exactly, including the synthesized sibling.
type woHashStateModel struct {
	Name                  types.String `tfsdk:"name"`
	Password              types.String `tfsdk:"password"`
	PasswordWriteOnlyHash types.String `tfsdk:"password_write_only_hash"`
}

// woHashBackend stands in for the remote API. It records what each apply "sent", so the tests can
// assert on rotation behavior that is invisible in Terraform state (the secret is write-only).
type woHashBackend struct {
	mu          sync.Mutex
	createCount int
	updateCount int
	deleteCount int
	lastSecret  string
	sentSecrets []string
}

func (b *woHashBackend) record(op, secret string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	switch op {
	case "create":
		b.createCount++
		b.lastSecret = secret
		b.sentSecrets = append(b.sentSecrets, secret)
	case "update":
		b.updateCount++
		b.lastSecret = secret
		b.sentSecrets = append(b.sentSecrets, secret)
	case "delete":
		b.deleteCount++
	}
}

// TestWriteOnlyHashedAttribute_Integration_ComputedSecret drives a hash-mode write-only resource
// through a full create -> no-op -> rotate -> destroy lifecycle against the real terraform CLI.
//
// The secret is sourced from an upstream managed resource's computed attribute rather than from a
// literal. That is deliberate and load-bearing: it keeps the secret UNKNOWN at plan time on any
// step that changes it, so the hash sibling plans to "known after apply" for the straightforward
// reason that the secret itself is not yet known. See
// TestWriteOnlyHashedAttribute_Integration_LiteralSecret for the contrasting known-at-plan
// case, which reaches the same "known after apply" outcome for a different reason: even though the
// secret is known, hashPlanModifier still cannot commit a concrete tag without a reusable prior
// salt (there is none on first create), and drawing one at plan time would make the plan
// non-deterministic across the two PlanResourceChange calls Terraform issues during a single
// apply. Both cases resolve the concrete tag on the apply side, via ResolveWriteOnlyHashTags.
func TestWriteOnlyHashedAttribute_Integration_ComputedSecret(t *testing.T) {
	// Not parallel: the shared backend records ordered side effects across steps.
	backend := &woHashBackend{}

	var hashAfterCreate string

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"test": providerserver.NewProtocol6WithError(&woHashProvider{backend: backend}),
		},
		CheckDestroy: func(*terraform.State) error {
			backend.mu.Lock()
			defer backend.mu.Unlock()
			if backend.deleteCount < 1 {
				return fmt.Errorf("expected Delete to be called on destroy, got deleteCount=%d", backend.deleteCount)
			}
			return nil
		},
		Steps: []resource.TestStep{
			// Step 1: CREATE with the upstream secret resolving to "s1".
			{
				Config: woHashComputedSecretConfig("db", "s1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(woHashResourceAddr, "name", "db"),
					resource.TestCheckResourceAttrSet(woHashResourceAddr, "password_write_only_hash"),
					// The raw secret is write-only and must never land in state.
					resource.TestCheckNoResourceAttr(woHashResourceAddr, "password"),
					// The backend saw the secret exactly once, via Create.
					checkBackend(backend, func(b *woHashBackend) error {
						if b.createCount != 1 {
							return fmt.Errorf("createCount = %d, want 1", b.createCount)
						}
						if b.updateCount != 0 {
							return fmt.Errorf("updateCount = %d, want 0", b.updateCount)
						}
						if b.lastSecret != "s1" {
							return fmt.Errorf("secret sent to backend = %q, want %q", b.lastSecret, "s1")
						}
						return nil
					}),
					captureAttr(woHashResourceAddr, "password_write_only_hash", &hashAfterCreate),
				),
			},
			// Step 2: NO-OP apply with the identical secret. The plan modifier reuses the salt
			// embedded in the prior tag, so the tag is byte-identical, the plan is empty, and
			// Terraform performs no Update -- proving an unchanged secret is not spuriously re-sent.
			{
				Config: woHashComputedSecretConfig("db", "s1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkBackend(backend, func(b *woHashBackend) error {
						if b.updateCount != 0 {
							return fmt.Errorf("updateCount = %d, want 0 (an unchanged secret must not re-send)", b.updateCount)
						}
						return nil
					}),
					assertHashAttrEquals(&hashAfterCreate, true),
				),
			},
			// Step 3: ROTATE the secret to "s2". The tag changes, so the computed sibling is the
			// only visible diff, Terraform calls Update, and the new secret is re-sent.
			{
				Config: woHashComputedSecretConfig("db", "s2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(woHashResourceAddr, "password_write_only_hash"),
					resource.TestCheckNoResourceAttr(woHashResourceAddr, "password"),
					checkBackend(backend, func(b *woHashBackend) error {
						if b.updateCount < 1 {
							return fmt.Errorf("updateCount = %d, want >= 1 (a changed secret must re-send)", b.updateCount)
						}
						if b.lastSecret != "s2" {
							return fmt.Errorf("secret sent to backend = %q, want %q", b.lastSecret, "s2")
						}
						return nil
					}),
					assertHashAttrEquals(&hashAfterCreate, false),
				),
			},
		},
	})
}

// TestWriteOnlyHashedAttribute_Integration_LiteralSecret drives the known-at-plan-time case (the
// write-only secret is a literal, not sourced from an upstream computed value) through the same
// full create -> no-op -> rotate -> destroy lifecycle as
// TestWriteOnlyHashedAttribute_Integration_ComputedSecret, proving that a literal secret now
// creates successfully: the hash sibling still plans Unknown (there is no reusable prior salt on
// first create), and the concrete tag is derived on the apply side by ResolveWriteOnlyHashTags,
// rather than being committed as a concrete value by the plan modifier.
//
// The literal-secret case is the specific regression this test guards against. It is the shape
// that used to trip Terraform's plan-consistency check: the hash sibling's value used to come
// ONLY from hashPlanModifier at plan time (nothing on the apply side computed it). On first create
// there is no prior-state salt to reuse, so the modifier drew a fresh random salt and committed a
// concrete tag. Terraform then re-plans the resource while expanding the plan during apply; the
// modifier drew a SECOND fresh salt and committed a DIFFERENT concrete tag, and Terraform rejected
// the known-value-changed-to-known-value as "Provider produced inconsistent final plan". (The
// computed-secret case above never hit this, because the secret itself being unknown at plan time
// already forced the sibling unknown for an unrelated reason.)
//
// The fix: hashPlanModifier now plans the sibling unknown whenever there is no reusable prior
// salt, drawing no randomness at plan time at all, so both PlanResourceChange calls agree (both
// unknown). ResolveWriteOnlyHashTags then derives the concrete tag exactly once, on the apply
// side, where drawing a salt is safe. woHashResource.Create/Update below call it just as the real
// IdsecResource.triggerOperation does.
func TestWriteOnlyHashedAttribute_Integration_LiteralSecret(t *testing.T) {
	// Not parallel: the shared backend records ordered side effects across steps.
	backend := &woHashBackend{}

	var hashAfterCreateLiteral string

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"test": providerserver.NewProtocol6WithError(&woHashProvider{backend: backend}),
		},
		CheckDestroy: func(*terraform.State) error {
			backend.mu.Lock()
			defer backend.mu.Unlock()
			if backend.deleteCount < 1 {
				return fmt.Errorf("expected Delete to be called on destroy, got deleteCount=%d", backend.deleteCount)
			}
			return nil
		},
		Steps: []resource.TestStep{
			// Step 1: CREATE with a literal secret, known at plan time. Before the fix, this step
			// alone failed with "Provider produced inconsistent final plan".
			{
				Config: woHashLiteralSecretConfig("db", "s1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(woHashResourceAddr, "name", "db"),
					resource.TestCheckResourceAttrSet(woHashResourceAddr, "password_write_only_hash"),
					// The raw secret is write-only and must never land in state.
					resource.TestCheckNoResourceAttr(woHashResourceAddr, "password"),
					// The backend saw the secret exactly once, via Create.
					checkBackend(backend, func(b *woHashBackend) error {
						if b.createCount != 1 {
							return fmt.Errorf("createCount = %d, want 1", b.createCount)
						}
						if b.updateCount != 0 {
							return fmt.Errorf("updateCount = %d, want 0", b.updateCount)
						}
						if b.lastSecret != "s1" {
							return fmt.Errorf("secret sent to backend = %q, want %q", b.lastSecret, "s1")
						}
						return nil
					}),
					captureAttr(woHashResourceAddr, "password_write_only_hash", &hashAfterCreateLiteral),
				),
			},
			// Step 2: NO-OP apply with the identical literal secret. The plan modifier reuses the
			// salt embedded in the prior tag (resolved on apply after step 1), so the tag is
			// byte-identical, the plan is empty, and Terraform performs no Update.
			{
				Config: woHashLiteralSecretConfig("db", "s1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkBackend(backend, func(b *woHashBackend) error {
						if b.updateCount != 0 {
							return fmt.Errorf("updateCount = %d, want 0 (an unchanged secret must not re-send)", b.updateCount)
						}
						return nil
					}),
					assertHashAttrEquals(&hashAfterCreateLiteral, true),
				),
			},
			// Step 3: ROTATE the literal secret to "s2". The tag changes, so the computed sibling
			// is the only visible diff, Terraform calls Update, and the new secret is re-sent.
			{
				Config: woHashLiteralSecretConfig("db", "s2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(woHashResourceAddr, "password_write_only_hash"),
					resource.TestCheckNoResourceAttr(woHashResourceAddr, "password"),
					checkBackend(backend, func(b *woHashBackend) error {
						if b.updateCount < 1 {
							return fmt.Errorf("updateCount = %d, want >= 1 (a changed secret must re-send)", b.updateCount)
						}
						if b.lastSecret != "s2" {
							return fmt.Errorf("secret sent to backend = %q, want %q", b.lastSecret, "s2")
						}
						return nil
					}),
					assertHashAttrEquals(&hashAfterCreateLiteral, false),
				),
			},
		},
	})
}

// woHashComputedSecretConfig renders HCL where the write-only secret is the computed output of an
// upstream resource, so it is unknown at plan time whenever it changes.
func woHashComputedSecretConfig(name, secret string) string {
	return fmt.Sprintf(`
provider "test" {}

resource "test_secret_source" "src" {
  input = %[2]q
}

resource "test_wohash_resource" "test" {
  name     = %[1]q
  password = test_secret_source.src.result
}
`, name, secret)
}

// woHashLiteralSecretConfig renders HCL where the write-only secret is a literal, so it is known at
// plan time.
func woHashLiteralSecretConfig(name, password string) string {
	return fmt.Sprintf(`
provider "test" {}

resource "test_wohash_resource" "test" {
  name     = %[1]q
  password = %[2]q
}
`, name, password)
}

// checkBackend adapts a backend assertion into a resource.TestCheckFunc.
func checkBackend(b *woHashBackend, fn func(*woHashBackend) error) resource.TestCheckFunc {
	return func(*terraform.State) error {
		b.mu.Lock()
		defer b.mu.Unlock()
		return fn(b)
	}
}

// captureAttr stores the value of an attribute from state into dst, for later cross-step comparison.
func captureAttr(resourceName, attr string, dst *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, err := stateAttr(s, resourceName, attr)
		if err != nil {
			return err
		}
		*dst = v
		return nil
	}
}

// assertHashAttrEquals compares password_write_only_hash in the test resource's state against a
// previously captured value. When wantEqual is true it requires them equal; when false it requires
// them different.
func assertHashAttrEquals(want *string, wantEqual bool) resource.TestCheckFunc {
	const attr = "password_write_only_hash"
	return func(s *terraform.State) error {
		got, err := stateAttr(s, woHashResourceAddr, attr)
		if err != nil {
			return err
		}
		if wantEqual && got != *want {
			return fmt.Errorf("%s.%s = %q, want it unchanged at %q", woHashResourceAddr, attr, got, *want)
		}
		if !wantEqual && got == *want {
			return fmt.Errorf("%s.%s = %q, want it changed from %q", woHashResourceAddr, attr, got, *want)
		}
		return nil
	}
}

func stateAttr(s *terraform.State, resourceName, attr string) (string, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return "", fmt.Errorf("resource %s not found in state", resourceName)
	}
	return rs.Primary.Attributes[attr], nil
}

// woHashProvider is a minimal provider exposing the hash-mode test resource and an upstream
// secret-source resource that produces a computed (plan-unknown-on-change) value.
type woHashProvider struct {
	backend *woHashBackend
}

func (p *woHashProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "test"
}

func (p *woHashProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = providerschema.Schema{Description: "Test provider for write-only hash integration testing"}
}

func (p *woHashProvider) Configure(_ context.Context, _ provider.ConfigureRequest, _ *provider.ConfigureResponse) {
}

func (p *woHashProvider) Resources(_ context.Context) []func() fwresource.Resource {
	return []func() fwresource.Resource{
		func() fwresource.Resource { return &woHashResource{backend: p.backend} },
		func() fwresource.Resource { return &secretSourceResource{} },
	}
}

func (p *woHashProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

// woHashResource is an in-memory resource whose schema is produced by the real generator with a
// WriteOnlyHashedAttributes declaration, so every hash-mode code path is exercised end to end.
type woHashResource struct {
	backend *woHashBackend
}

func (r *woHashResource) Metadata(_ context.Context, req fwresource.MetadataRequest, resp *fwresource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wohash_resource"
}

func (r *woHashResource) Schema(_ context.Context, _ fwresource.SchemaRequest, resp *fwresource.SchemaResponse) {
	generated, diags := GenerateResourceSchemaFromStruct(
		&woHashGenModel{},    // create model
		nil,                  // update model
		nil,                  // state model
		nil,                  // sensitiveAttrs
		nil,                  // extraRequiredAttrs
		nil,                  // computedAsSetAttrs
		nil,                  // immutableAttrs
		nil,                  // forceNewAttrs
		nil,                  // computedAttrs
		nil,                  // semanticEqualityAttrs
		nil,                  // writeOnlyAttrs (manual trigger mode)
		[]string{"password"}, // writeOnlyHashedAttrs (hash mode)
	)
	resp.Diagnostics.Append(diags...)
	resp.Schema = schema.Schema{
		Description: "Test resource with a hash-mode write-only password",
		Attributes:  generated.Attributes,
	}
}

func (r *woHashResource) Create(ctx context.Context, req fwresource.CreateRequest, resp *fwresource.CreateResponse) {
	var plan woHashStateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The write-only secret is null in plan and state by design; it exists only in config.
	var config woHashStateModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.backend.record("create", config.Password.ValueString())

	// Mirrors IdsecResource.triggerOperation: resolve the concrete tag for a sibling that planned
	// unknown (no reusable prior salt exists on create) before writing state.
	plannedHashTags, hashDiags := ResolveWriteOnlyHashTags(ctx, []string{"password"}, &req.Plan, nil, &req.Config)
	resp.Diagnostics.Append(hashDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if tag, ok := plannedHashTags["password_write_only_hash"]; ok {
		plan.PasswordWriteOnlyHash = tag
	}

	// State mirrors the plan: password stays null (write-only), the computed hash sibling carries
	// the resolved tag.
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *woHashResource) Read(ctx context.Context, req fwresource.ReadRequest, resp *fwresource.ReadResponse) {
	var state woHashStateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *woHashResource) Update(ctx context.Context, req fwresource.UpdateRequest, resp *fwresource.UpdateResponse) {
	var plan woHashStateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config woHashStateModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.backend.record("update", config.Password.ValueString())

	// Mirrors IdsecResource.triggerOperation: resolve the concrete tag for a sibling that planned
	// unknown, reusing req.State's prior tag salt when a rotated secret's tag still needs one.
	plannedHashTags, hashDiags := ResolveWriteOnlyHashTags(ctx, []string{"password"}, &req.Plan, &req.State, &req.Config)
	resp.Diagnostics.Append(hashDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if tag, ok := plannedHashTags["password_write_only_hash"]; ok {
		plan.PasswordWriteOnlyHash = tag
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *woHashResource) Delete(_ context.Context, _ fwresource.DeleteRequest, _ *fwresource.DeleteResponse) {
	r.backend.record("delete", "")
}

// secretSourceModel backs secretSourceResource: a required input echoed into a computed result.
type secretSourceModel struct {
	Input  types.String `tfsdk:"input"`
	Result types.String `tfsdk:"result"`
}

// secretSourceResource produces a computed "result" equal to its "input". Because result is
// Computed with no UseStateForUnknown, Terraform plans it as unknown whenever input changes, which
// is exactly what makes the downstream write-only secret unknown at plan time on create and rotate.
type secretSourceResource struct{}

func (r *secretSourceResource) Metadata(_ context.Context, req fwresource.MetadataRequest, resp *fwresource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret_source"
}

func (r *secretSourceResource) Schema(_ context.Context, _ fwresource.SchemaRequest, resp *fwresource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Produces a computed result mirroring its input, unknown at plan time when input changes",
		Attributes: map[string]schema.Attribute{
			"input":  schema.StringAttribute{Required: true},
			"result": schema.StringAttribute{Computed: true},
		},
	}
}

func (r *secretSourceResource) Create(ctx context.Context, req fwresource.CreateRequest, resp *fwresource.CreateResponse) {
	var plan secretSourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Result = plan.Input
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *secretSourceResource) Read(ctx context.Context, req fwresource.ReadRequest, resp *fwresource.ReadResponse) {
	var state secretSourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *secretSourceResource) Update(ctx context.Context, req fwresource.UpdateRequest, resp *fwresource.UpdateResponse) {
	var plan secretSourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Result = plan.Input
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *secretSourceResource) Delete(_ context.Context, _ fwresource.DeleteRequest, _ *fwresource.DeleteResponse) {
}
