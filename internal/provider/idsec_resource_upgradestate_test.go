// Copyright CyberArk. 2026
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"fmt"
	"maps"
	"math/big"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/cyberark/terraform-provider-idsec/internal/actions"
	"github.com/cyberark/terraform-provider-idsec/internal/schemas"
)

// upgradeStateGenModel is the SDK-style model the schema generator reflects over. It deliberately
// does not declare the hash sibling: WriteOnlyHashedAttributes makes the generator synthesize
// "secret_write_only_hash" itself, exactly as it does for idsec_pcloud_account.
// SafeNumber is here so a test can assert what happens to a numeric attribute: without one, a
// number in the prior state is an attribute the current schema does not declare and the decode
// drops it, which would make TestIdsecResource_UpgradeState_TransformPreservesLargeIntegers pass
// whatever the decoder did with it.
type upgradeStateGenModel struct {
	AccountID  string `json:"account_id,omitempty" mapstructure:"account_id" desc:"the account id"`
	Name       string `json:"name,omitempty" mapstructure:"name" desc:"a display name"`
	Secret     string `json:"secret,omitempty" mapstructure:"secret" desc:"the credential"`
	SafeNumber int64  `json:"safe_number,omitempty" mapstructure:"safe_number" desc:"a numeric id"`
}

// upgradeStateTestActionDef builds an action definition whose "secret" is a hash-mode write-only
// attribute at the given ActionVersion, mirroring the pcloud-account shape this upgrader exists for.
//
// priorVersions are the schema versions the resource declares an upgrade from, each a passthrough
// (nil Transform). Passing none mirrors a resource that bumped ActionVersion without declaring any
// upgrade, which must register no upgraders at all.
func upgradeStateTestActionDef(actionVersion int64, priorVersions ...int64) *actions.IdsecServiceTerraformResourceActionDefinition {
	var stateUpgrades map[int64]actions.StateUpgrade
	if len(priorVersions) > 0 {
		stateUpgrades = make(map[int64]actions.StateUpgrade, len(priorVersions))
		for _, priorVersion := range priorVersions {
			stateUpgrades[priorVersion] = actions.StateUpgrade{}
		}
	}

	return &actions.IdsecServiceTerraformResourceActionDefinition{
		IdsecServiceBaseTerraformActionDefinition: actions.IdsecServiceBaseTerraformActionDefinition{
			IdsecServiceBaseActionDefinition: actions.IdsecServiceBaseActionDefinition{
				ActionName:    "test-account",
				ActionVersion: actionVersion,
				Schemas: map[string]interface{}{
					"create-action": upgradeStateGenModel{},
					"update-action": upgradeStateGenModel{},
				},
			},
			StateSchema:        &upgradeStateGenModel{},
			ComputedAttributes: []string{"account_id"},
		},
		SupportedOperations: []actions.IdsecServiceActionOperation{actions.CreateOperation, actions.UpdateOperation},
		ActionsMappings: map[actions.IdsecServiceActionOperation]string{
			actions.CreateOperation: "create-action",
			actions.UpdateOperation: "update-action",
		},
		ImportID:                  "account_id",
		WriteOnlyHashedAttributes: []string{"secret"},
		StateUpgrades:             stateUpgrades,
	}
}

// upgradeStateTestProvider is the smallest provider that can serve the UpgradeResourceState RPC for
// a single IdsecResource. The test drives the real protocol server rather than calling the upgrader
// directly because the step that matters most -- nulling the write-only attribute -- happens in the
// framework after our StateUpgrader returns, and so is invisible to a direct call.
type upgradeStateTestProvider struct {
	res resource.Resource
}

func (p *upgradeStateTestProvider) Metadata(_ context.Context, _ fwprovider.MetadataRequest, resp *fwprovider.MetadataResponse) {
	resp.TypeName = "idsec"
}

func (p *upgradeStateTestProvider) Schema(_ context.Context, _ fwprovider.SchemaRequest, resp *fwprovider.SchemaResponse) {
	resp.Schema = providerschema.Schema{}
}

func (p *upgradeStateTestProvider) Configure(_ context.Context, _ fwprovider.ConfigureRequest, _ *fwprovider.ConfigureResponse) {
}

func (p *upgradeStateTestProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{func() resource.Resource { return p.res }}
}

func (p *upgradeStateTestProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

// TestIdsecResource_UpgradeState_RegistersVersionsOnlyWhenOptedIn asserts the two halves of the
// opt-in model: StateUpgrades alone decides *whether* a resource registers upgraders at all, and
// once it does, ActionVersion decides *which* versions get one -- every version in
// [0, ActionVersion), with undeclared ones synthesized as pure passthroughs.
//
// Gating on StateUpgrades is deliberate. Registering upgraders off ActionVersion alone would hand
// every resource in the provider a passthrough for every earlier version, which changes upgrade
// behaviour provider-wide and, worse, makes a future bump that renames or restructures an attribute
// silently drop the old value: the passthrough decodes against the current schema, where the old
// name no longer exists. A resource that declares nothing keeps Terraform's hard error ("Terraform
// was expecting an implementation for version N upgrade"), and that error is what forces the
// developer to decide whether the bump needs real upgrade logic.
//
// Synthesizing the undeclared versions below ActionVersion is what that decision then buys: having
// opted in, a resource only declares the versions where something actually changed, and intermediate
// bumps that touched nothing a practitioner's state can see are covered without empty boilerplate.
func TestIdsecResource_UpgradeState_RegistersVersionsOnlyWhenOptedIn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		actionDef    *actions.IdsecServiceTerraformResourceActionDefinition
		wantVersions []int64
	}{
		{
			name: "success_registers_the_declared_prior_version",
			// The pcloud-account shape: at version 2, declaring an upgrade from 1 only. Version 0 is
			// registered alongside it because a resource whose ActionVersion was never set reports
			// schema version 0, so state written by that release arrives as version 0.
			actionDef:    upgradeStateTestActionDef(2, 1),
			wantVersions: []int64{0, 1},
		},
		{
			name:         "success_registers_every_declared_prior_version",
			actionDef:    upgradeStateTestActionDef(3, 1, 2),
			wantVersions: []int64{0, 1, 2},
		},
		{
			name: "success_sparse_declaration_synthesizes_the_undeclared_versions",
			// A resource at version 3 that only restructured state in the bump to 3 declares {2} and
			// nothing else; 0 and 1 are covered as passthroughs without a redundant empty entry each.
			actionDef:    upgradeStateTestActionDef(3, 2),
			wantVersions: []int64{0, 1, 2},
		},
		{
			name: "success_explicitly_declaring_version_zero_changes_nothing",
			// Declaring 0 is allowed but redundant: it is registered either way.
			actionDef:    upgradeStateTestActionDef(2, 0, 1),
			wantVersions: []int64{0, 1},
		},
		{
			name: "success_version_bump_without_declared_upgrades_registers_nothing",
			// The behaviour the opt-in exists for: a bump alone must not synthesize upgraders, so
			// Terraform still fails loudly until the developer declares what the upgrade does.
			actionDef:    upgradeStateTestActionDef(2),
			wantVersions: nil,
		},
		{
			name:         "success_unset_version_registers_nothing",
			actionDef:    upgradeStateTestActionDef(0),
			wantVersions: nil,
		},
		{
			name: "success_unset_version_registers_nothing_even_when_declared",
			// [0, ActionVersion) is empty at version 0, so there is no version older than the current
			// schema to upgrade from and the declared entry has nothing to key.
			actionDef:    upgradeStateTestActionDef(0, 0),
			wantVersions: nil,
		},
		{
			name:         "success_nil_action_definition_registers_nothing",
			actionDef:    nil,
			wantVersions: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &IdsecResource{actionDefinition: tt.actionDef}
			upgraders := s.UpgradeState(context.Background())

			got := make([]int64, 0, len(upgraders))
			for version, upgrader := range upgraders {
				got = append(got, version)
				if upgrader.StateUpgrader == nil {
					t.Errorf("upgrader for version %d has a nil StateUpgrader", version)
				}
				if upgrader.PriorSchema != nil {
					t.Errorf("upgrader for version %d set PriorSchema; the passthrough decodes against the current schema", version)
				}
			}
			slices.Sort(got)

			if len(got) == 0 {
				got = nil
			}
			if !reflect.DeepEqual(got, tt.wantVersions) {
				t.Errorf("UpgradeState registered versions %v, want %v", got, tt.wantVersions)
			}
		})
	}
}

// TestIdsecResource_UpgradeState_NullsWriteOnlyAttributeFromPriorState is the regression test for
// the failure this upgrader exists to fix: prior state written before "secret" became write-only
// still holds the credential in plain text, and returning it makes Terraform fail the run with
// "Invalid resource state upgrade ... the provider returned a value for the write-only attribute".
// Every other attribute must survive the upgrade untouched.
func TestIdsecResource_UpgradeState_NullsWriteOnlyAttributeFromPriorState(t *testing.T) {
	t.Parallel()

	const secretLiteral = "distinctive-super-secret-literal-42"
	ctx := context.Background()
	actionDef := upgradeStateTestActionDef(2, 1)
	idsecRes := &IdsecResource{actionDefinition: actionDef}

	schemaResp := &resource.SchemaResponse{}
	idsecRes.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("building the resource schema: %v", schemaResp.Diagnostics.Errors())
	}
	if schemaResp.Schema.Version != 2 {
		t.Fatalf("schema Version = %d, want 2 (ActionVersion must reach the framework schema)", schemaResp.Schema.Version)
	}
	secretAttr, ok := schemaResp.Schema.Attributes["secret"]
	if !ok {
		t.Fatal(`generated schema has no "secret" attribute`)
	}
	if !secretAttr.IsWriteOnly() {
		t.Fatal(`"secret" is not write-only in the generated schema; the rest of this test proves nothing`)
	}

	server, err := providerserver.NewProtocol6WithError(&upgradeStateTestProvider{res: idsecRes})()
	if err != nil {
		t.Fatalf("building the provider server: %v", err)
	}

	// Prior state as the previous provider release would have written it: version 1, "secret" held
	// in plain text, and no hash sibling (that attribute did not exist yet). "removed_attr" stands
	// in for an attribute the current schema no longer declares.
	priorStateJSON := `{"account_id":"acct-1","name":"policy","secret":"` + secretLiteral + `","removed_attr":"stale"}`

	upgradeResp, err := server.UpgradeResourceState(ctx, &tfprotov6.UpgradeResourceStateRequest{
		TypeName: "idsec_test_account",
		Version:  1,
		RawState: &tfprotov6.RawState{JSON: []byte(priorStateJSON)},
	})
	if err != nil {
		t.Fatalf("UpgradeResourceState: %v", err)
	}
	for _, d := range upgradeResp.Diagnostics {
		if d.Severity == tfprotov6.DiagnosticSeverityError {
			t.Fatalf("UpgradeResourceState diagnostic: %s: %s", d.Summary, d.Detail)
		}
	}
	if upgradeResp.UpgradedState == nil {
		t.Fatal("UpgradeResourceState returned no upgraded state")
	}

	upgraded, err := upgradeResp.UpgradedState.Unmarshal(schemaResp.Schema.Type().TerraformType(ctx))
	if err != nil {
		t.Fatalf("unmarshalling the upgraded state: %v", err)
	}
	attrs := map[string]tftypes.Value{}
	if err := upgraded.As(&attrs); err != nil {
		t.Fatalf("reading the upgraded state as an object: %v", err)
	}

	if got, ok := attrs["secret"]; !ok {
		t.Error(`upgraded state has no "secret" attribute`)
	} else if !got.IsNull() {
		t.Errorf("upgraded state returned a value for the write-only \"secret\"; Terraform rejects the run when this happens (got %s)", got.String())
	}

	// A dedicated check on the literal, not just on "secret": the credential must not survive
	// anywhere in the returned state, including an attribute this test did not think to name.
	if strings.Contains(upgraded.String(), secretLiteral) {
		t.Error("the plain text credential appears somewhere in the upgraded state")
	}

	wantPreserved := map[string]string{"account_id": "acct-1", "name": "policy"}
	for name, want := range wantPreserved {
		var got string
		if err := attrs[name].As(&got); err != nil {
			t.Errorf("reading upgraded %q: %v", name, err)
			continue
		}
		if got != want {
			t.Errorf("upgraded %q = %q, want %q; the upgrade must preserve every non-write-only value", name, got, want)
		}
	}

	// The synthesized hash sibling is absent from prior state, so it must not fail the upgrade --
	// and it must not be left null either: a null sibling plans unknown, which triggerFired reads
	// as a rotation and re-sends the credential to the vault on the first apply after the upgrade.
	// The upgrader adopts a tag derived from the plain text the old state still held, so an
	// unchanged secret in configuration plans to the same tag and nothing is sent.
	hashName := schemas.WriteOnlyHashAttributeName("secret")
	got, ok := attrs[hashName]
	if !ok {
		t.Fatalf("upgraded state has no %q attribute", hashName)
	}
	if got.IsNull() {
		t.Fatalf("upgraded %q is null; the stored credential would be re-sent on the first apply after the upgrade", hashName)
	}
	var adoptedTag string
	if err := got.As(&adoptedTag); err != nil {
		t.Fatalf("reading upgraded %q: %v", hashName, err)
	}
	// The tag's derivation from that specific secret is asserted in the schemas package
	// (TestAdoptWriteOnlyHashTagsFromPriorState_DerivesTagFromStoredSecret); here it only has to
	// have survived the framework's write-only nullification pass in a shape this version can read.
	if !strings.HasPrefix(adoptedTag, "v1:") {
		t.Errorf("upgraded %q is not a recognizable hash tag", hashName)
	}
}

// TestIdsecResource_UpgradeStateForVersion_UsesRespStateSchema asserts that upgradeStateForVersion
// reads the schema from resp.State.Schema (which the framework pre-populates with the cached,
// authoritative schema) rather than rebuilding it via s.Schema().
//
// The setup deliberately breaks s.Schema() by setting actionDefinition.Schemas = nil while
// pre-populating resp.State.Schema with the correct schema, matching what
// fwserver.UpgradeResourceState does before calling any provider-defined StateUpgrader.
// If the upgrader calls s.Schema() (current behaviour) it receives an error diagnostic and returns
// early — the test fails. If it reads resp.State.Schema directly (the fix) it never calls
// s.Schema() and the upgrade succeeds.
func TestIdsecResource_UpgradeStateForVersion_UsesRespStateSchema(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Build the valid schema from a correctly-configured resource instance.
	validRes := &IdsecResource{actionDefinition: upgradeStateTestActionDef(2, 1)}
	schemaResp := &resource.SchemaResponse{}
	validRes.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("building reference schema: %v", schemaResp.Diagnostics.Errors())
	}
	validSchema := schemaResp.Schema

	// Build a resource whose s.Schema() will fail: Schemas is nil, which Schema() rejects with
	// "Schemas mappings are not provided." If upgradeStateForVersion calls s.Schema() the error
	// propagates into resp.Diagnostics, which the assertion below catches.
	brokenDef := upgradeStateTestActionDef(2, 1)
	brokenDef.Schemas = nil
	brokenRes := &IdsecResource{actionDefinition: brokenDef}

	upgraders := brokenRes.UpgradeState(ctx)
	upgrader, ok := upgraders[1]
	if !ok {
		t.Fatal("no upgrader registered for version 1")
	}

	priorStateJSON := `{"account_id":"acct-1","name":"policy","secret":"p@ssw0rd"}`

	// Pre-populate resp.State.Schema exactly as fwserver.UpgradeResourceState does before
	// dispatching to the provider-defined StateUpgrader (framework v1.18.0,
	// internal/fwserver/server_upgraderesourcestate.go).
	resp := &resource.UpgradeStateResponse{
		State: tfsdk.State{Schema: validSchema},
	}
	upgrader.StateUpgrader(ctx, resource.UpgradeStateRequest{
		RawState: &tfprotov6.RawState{JSON: []byte(priorStateJSON)},
	}, resp)

	// The fix: upgradeStateForVersion reads resp.State.Schema and never calls s.Schema(), so
	// brokenRes's nil Schemas never surfaces. The current (unfixed) code calls s.Schema() on
	// brokenRes, which adds "Schemas mappings are not provided" to resp.Diagnostics — detected here.
	if resp.Diagnostics.HasError() {
		t.Errorf("upgradeStateForVersion rebuilt the schema via s.Schema() instead of using resp.State.Schema: %v", resp.Diagnostics.Errors())
	}
	if resp.State.Raw.IsNull() || !resp.State.Raw.IsKnown() {
		t.Error("upgrade produced no state: Raw is null or unknown")
	}
}

// TestIdsecResource_UpgradeState_MissingRawState asserts the upgrader fails loudly rather than
// returning an empty state, which the framework would otherwise accept as a deliberate wipe.
func TestIdsecResource_UpgradeState_MissingRawState(t *testing.T) {
	t.Parallel()

	s := &IdsecResource{actionDefinition: upgradeStateTestActionDef(2, 1)}
	upgraders := s.UpgradeState(context.Background())
	upgrader, ok := upgraders[1]
	if !ok {
		t.Fatal("no upgrader registered for version 1")
	}

	resp := &resource.UpgradeStateResponse{}
	upgrader.StateUpgrader(context.Background(), resource.UpgradeStateRequest{RawState: nil}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected an error diagnostic for a nil RawState")
	}
	if !resp.State.Raw.IsNull() && resp.State.Raw.Type() != nil {
		t.Error("expected no state to be returned alongside the error")
	}
}

// upgradeStateSchemaFor builds the framework schema the given definition generates, standing in for
// the cached schema fwserver.UpgradeResourceState pre-populates on resp.State before dispatching to
// a provider-defined StateUpgrader.
func upgradeStateSchemaFor(t *testing.T, actionDef *actions.IdsecServiceTerraformResourceActionDefinition) rschema.Schema {
	t.Helper()

	schemaResp := &resource.SchemaResponse{}
	(&IdsecResource{actionDefinition: actionDef}).Schema(context.Background(), resource.SchemaRequest{}, schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("building the resource schema: %v", schemaResp.Diagnostics.Errors())
	}
	return schemaResp.Schema
}

// TestIdsecResource_UpgradeState_AppliesTransformToRawState asserts that a declared Transform runs
// on the raw prior-state JSON, before it is decoded against the current schema.
//
// The ordering is the whole point and it is why Transform takes a map rather than a tftypes.Value:
// decoding first drops any attribute the current schema no longer declares, so an attribute that was
// renamed in the bump would already be gone by the time a Transform keyed on its old name ran. The
// prior state here carries "legacy_name", which the current schema does not declare, and the
// Transform moves it to "name". A Transform running after the decode would see null and the assertion
// on "name" would fail.
func TestIdsecResource_UpgradeState_AppliesTransformToRawState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var seenPrior map[string]any
	actionDef := upgradeStateTestActionDef(2, 1)
	actionDef.StateUpgrades[1] = actions.StateUpgrade{
		Transform: func(_ context.Context, prior map[string]any) (map[string]any, diag.Diagnostics) {
			seenPrior = maps.Clone(prior)

			upgraded := maps.Clone(prior)
			upgraded["name"] = upgraded["legacy_name"]
			delete(upgraded, "legacy_name")
			return upgraded, nil
		},
	}

	idsecRes := &IdsecResource{actionDefinition: actionDef}
	upgrader, ok := idsecRes.UpgradeState(ctx)[1]
	if !ok {
		t.Fatal("no upgrader registered for version 1")
	}

	schema := upgradeStateSchemaFor(t, upgradeStateTestActionDef(2, 1))
	resp := &resource.UpgradeStateResponse{State: tfsdk.State{Schema: schema}}
	upgrader.StateUpgrader(ctx, resource.UpgradeStateRequest{
		RawState: &tfprotov6.RawState{
			JSON: []byte(`{"account_id":"acct-1","legacy_name":"policy","secret":"p@ssw0rd"}`),
		},
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("upgrade returned error diagnostics: %v", resp.Diagnostics.Errors())
	}

	if _, held := seenPrior["legacy_name"]; !held {
		t.Errorf(`Transform received %v, which does not hold "legacy_name"; it must see the raw prior state, not state already decoded against the current schema`, seenPrior)
	}

	attrs := map[string]tftypes.Value{}
	if err := resp.State.Raw.As(&attrs); err != nil {
		t.Fatalf("reading the upgraded state as an object: %v", err)
	}
	var name string
	if err := attrs["name"].As(&name); err != nil {
		t.Fatalf(`reading upgraded "name": %v`, err)
	}
	if name != "policy" {
		t.Errorf(`upgraded "name" = %q, want "policy"; the Transform's rewrite did not reach the decoded state`, name)
	}

	// The hash adoption still has to run after the Transform: a Transform that restructures state is
	// no reason to re-send the credential on the first apply after the upgrade.
	hashName := schemas.WriteOnlyHashAttributeName("secret")
	hashValue, ok := attrs[hashName]
	if !ok {
		t.Fatalf("upgraded state has no %q attribute", hashName)
	}
	if hashValue.IsNull() {
		t.Errorf("upgraded %q is null; hash adoption must still run on a transformed state", hashName)
	}
}

// TestIdsecResource_UpgradeState_TransformPreservesLargeIntegers asserts that routing the prior
// state through a Transform does not corrupt integers past 2^53.
//
// The Transform here is a pure passthrough and never reads "safe_number", which is the point: the
// whole prior state round-trips through map[string]any to reach the Transform, so a decode that
// turned numbers into float64 would rewrite state the upgrade was never asked to touch.
// 9007199254740993 is the smallest int64 a float64 cannot represent -- it would come back as
// ...992, a wrong value Terraform then persists with no error anywhere.
func TestIdsecResource_UpgradeState_TransformPreservesLargeIntegers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	const largeInteger int64 = 9007199254740993

	var seenKind string
	actionDef := upgradeStateTestActionDef(2, 1)
	actionDef.StateUpgrades[1] = actions.StateUpgrade{
		Transform: func(_ context.Context, prior map[string]any) (map[string]any, diag.Diagnostics) {
			seenKind = fmt.Sprintf("%T", prior["safe_number"])
			return maps.Clone(prior), nil
		},
	}

	idsecRes := &IdsecResource{actionDefinition: actionDef}
	upgrader, ok := idsecRes.UpgradeState(ctx)[1]
	if !ok {
		t.Fatal("no upgrader registered for version 1")
	}

	resp := &resource.UpgradeStateResponse{
		State: tfsdk.State{Schema: upgradeStateSchemaFor(t, upgradeStateTestActionDef(2, 1))},
	}
	upgrader.StateUpgrader(ctx, resource.UpgradeStateRequest{
		RawState: &tfprotov6.RawState{
			JSON: []byte(fmt.Sprintf(`{"account_id":"acct-1","name":"policy","secret":"p@ssw0rd","safe_number":%d}`, largeInteger)),
		},
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("upgrade returned error diagnostics: %v", resp.Diagnostics.Errors())
	}

	// Pins the contract documented on actions.StateUpgrade.Transform: a Transform that needs the
	// numeric value asserts on json.Number, so a decode that stopped using UseNumber would break
	// implementors rather than silently truncating here.
	if want := "json.Number"; seenKind != want {
		t.Errorf("Transform saw \"safe_number\" as %s, want %s; the prior state must decode with UseNumber", seenKind, want)
	}

	attrs := map[string]tftypes.Value{}
	if err := resp.State.Raw.As(&attrs); err != nil {
		t.Fatalf("reading the upgraded state as an object: %v", err)
	}
	var safeNumber big.Float
	if err := attrs["safe_number"].As(&safeNumber); err != nil {
		t.Fatalf(`reading upgraded "safe_number": %v`, err)
	}
	if got, _ := safeNumber.Int64(); got != largeInteger {
		t.Errorf(`upgraded "safe_number" = %d, want %d; the prior-state round trip truncated it`, got, largeInteger)
	}
}

// TestIdsecResource_UpgradeState_TransformErrorAbortsTheUpgrade asserts a Transform that cannot
// rewrite the prior state fails the run rather than letting a partially upgraded state through.
// Returning the untransformed state would be worse than failing: Terraform would accept it and the
// practitioner would silently lose whatever the Transform was meant to carry across.
func TestIdsecResource_UpgradeState_TransformErrorAbortsTheUpgrade(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	actionDef := upgradeStateTestActionDef(2, 1)
	actionDef.StateUpgrades[1] = actions.StateUpgrade{
		Transform: func(_ context.Context, _ map[string]any) (map[string]any, diag.Diagnostics) {
			var diags diag.Diagnostics
			diags.AddError("Unable to Upgrade Resource State", "the prior state is not in a shape this upgrade understands")
			return nil, diags
		},
	}

	idsecRes := &IdsecResource{actionDefinition: actionDef}
	upgrader, ok := idsecRes.UpgradeState(ctx)[1]
	if !ok {
		t.Fatal("no upgrader registered for version 1")
	}

	resp := &resource.UpgradeStateResponse{
		State: tfsdk.State{Schema: upgradeStateSchemaFor(t, upgradeStateTestActionDef(2, 1))},
	}
	upgrader.StateUpgrader(ctx, resource.UpgradeStateRequest{
		RawState: &tfprotov6.RawState{JSON: []byte(`{"account_id":"acct-1","name":"policy"}`)},
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected the Transform's error diagnostic to fail the upgrade")
	}
	if !resp.State.Raw.IsNull() {
		t.Error("expected no state to be returned alongside the error")
	}
}
