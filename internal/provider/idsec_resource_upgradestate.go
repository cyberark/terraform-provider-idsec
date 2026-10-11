// Copyright CyberArk. 2026
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/cyberark/terraform-provider-idsec/internal/actions"
	"github.com/cyberark/terraform-provider-idsec/internal/schemas"
)

// UpgradeState returns one state upgrader per schema version this resource explicitly declares via
// StateUpgrades, each the value-preserving passthrough (or structural Transform) built by
// upgradeStateForVersion.
//
// Terraform CLI calls UpgradeResourceState for every resource in state on every run, even when the
// stored version already matches the current schema. The framework serves that same-version case
// itself by round-tripping the stored state through the current schema, and that passthrough does
// NOT null out write-only attributes -- only the provider-defined upgrade path does, via
// NullifyWriteOnlyAttributes in fwserver.UpgradeResourceState. So marking an already-released
// attribute write-only without bumping ActionVersion hands Terraform back the plain text value
// still sitting in the old state, and Terraform rejects it:
//
//	Error: Invalid resource state upgrade
//	... the provider returned a value for the write-only attribute
//	"idsec_pcloud_account.pcloud_account.secret". Write-only attributes cannot be read back
//	from the provider.
//
// Every pre-existing resource of that type then fails at plan time with no way forward short of
// hand-editing state. Making a released attribute write-only is therefore a three-part change:
// declare it (WriteOnlyAttributes / WriteOnlyHashedAttributes), bump the resource's ActionVersion,
// AND declare a StateUpgrades entry for the prior version.
//
// Upgraders are opt-in per resource: a resource that declares nothing in StateUpgrades returns nil
// here, keeping Terraform's hard error ("Terraform was expecting an implementation for version N
// upgrade") for any practitioner holding older state. That error is deliberate -- it is what forces
// a developer to explicitly opt in rather than silently accepting a passthrough. Once a resource
// opts in by declaring at least one entry, every version in [0, ActionVersion) is registered:
// explicitly declared entries carry their Transform (or passthrough when nil); undeclared versions
// are synthesized as pure passthroughs. This means a resource that bumps from 1→3 and restructures
// state in the bump to 2 only needs {1: {Transform: ...}}; version 0 is auto-covered.
func (s *IdsecResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	if s.actionDefinition == nil || len(s.actionDefinition.StateUpgrades) == 0 {
		return nil
	}

	// A resource opts in by declaring at least one StateUpgrades entry. Once opted in, every
	// version in [0, ActionVersion) is registered: explicitly declared entries carry their
	// Transform (or passthrough if nil); undeclared versions are synthesized as pure passthroughs.
	upgraders := make(map[int64]resource.StateUpgrader, s.actionDefinition.ActionVersion)
	for priorVersion := int64(0); priorVersion < s.actionDefinition.ActionVersion; priorVersion++ {
		upgrade := s.actionDefinition.StateUpgrades[priorVersion] // zero value (passthrough) if not declared
		upgraders[priorVersion] = resource.StateUpgrader{
			// PriorSchema is deliberately nil: the provider keeps no historical schemas, and
			// decoding the raw state against the current schema with IgnoreUndefinedAttributes buys
			// exactly what a prior schema would here. Attributes the old state does not carry
			// decode as null; attributes the current schema no longer declares are dropped.
			StateUpgrader: s.upgradeStateForVersion(priorVersion, upgrade),
		}
	}
	return upgraders
}

// upgradeStateForVersion returns the StateUpgrader for an upgrade from priorVersion: it optionally
// applies upgrade.Transform to the raw prior-state JSON (when non-nil), then decodes the result
// against the current schema and returns it otherwise unchanged.
//
// Routing the state through this provider-defined path rather than the framework's same-version
// passthrough is the entire point -- the framework nulls every write-only attribute in what we
// return here, which is what stops a previously persisted credential from being read back.
//
// When upgrade.Transform is non-nil it runs on the raw prior-state JSON BEFORE it is decoded
// against the current schema. Ordering matters: decoding first drops any attribute the current
// schema no longer declares, so an attribute that was renamed in the bump would already be gone
// by the time a Transform keyed on its old name ran. Transform receives and returns a
// map[string]any so it can manipulate fields by name without knowledge of tftypes.
//
// For hash-mode write-only attributes (WriteOnlyHashedAttributes) it also adopts a hash for the
// credential the prior state still holds in plain text, via schemas.AdoptWriteOnlyHashTagsFromPriorState.
// Without that step the sibling is null after the upgrade -- no older state ever held one -- a null
// sibling always plans unknown, and triggerFired reads unknown as "changed" and re-sends the
// credential on the first apply after the upgrade. That send is not harmless: a vault may reject it
// against its password policy, failing the apply outright ("The password typed does not adhere to
// current password policy rules"), and when it succeeds it clobbers any out-of-band rotation.
// Import solves the same problem with the private-state marker markHashAdoptionPending writes, which
// is not an option here: UpgradeStateResponse carries no private state. Deriving the tag from the
// stored plain text is the stronger signal anyway -- a secret the practitioner edits in the same run
// as the upgrade hashes differently and is still sent, where the marker would suppress it.
func (s *IdsecResource) upgradeStateForVersion(priorVersion int64, upgrade actions.StateUpgrade) func(context.Context, resource.UpgradeStateRequest, *resource.UpgradeStateResponse) {
	return func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
		typeName := s.getTerraformTypeName(s.actionDefinition.ActionName)

		if req.RawState == nil {
			resp.Diagnostics.AddError(
				"Unable to Upgrade Resource State",
				fmt.Sprintf("No prior state data was provided for the schema version %d upgrade of %q. ", priorVersion, typeName)+
					"Preventing the unexpected loss of resource state data. "+
					"This is always an issue with the Terraform Provider and should be reported to the provider developer.",
			)
			return
		}

		// Apply the declared Transform (if any) to the raw prior-state JSON before decoding it
		// against the current schema. Running Transform first means it sees every attribute the old
		// state held, including attributes the current schema no longer declares that would have
		// been silently dropped by the IgnoreUndefinedAttributes decode below.
		rawJSON := req.RawState.JSON
		if upgrade.Transform != nil {
			// UseNumber is not optional here: the default decode turns every JSON number into a
			// float64, which silently truncates any integer past 2^53 -- and because the whole prior
			// state round-trips through this map, that would corrupt numeric attributes the Transform
			// never touches. json.Number re-serialises as the original literal, so an untouched value
			// reaches the decode below byte for byte. tftypes decodes with UseNumber for this same
			// reason; without it this pre-decode hop would reintroduce the loss it avoids.
			var prior map[string]any
			decoder := json.NewDecoder(bytes.NewReader(rawJSON))
			decoder.UseNumber()
			if err := decoder.Decode(&prior); err != nil {
				resp.Diagnostics.AddError(
					"Unable to Upgrade Resource State",
					fmt.Sprintf("Failed to parse the prior state JSON for the schema version %d upgrade of %q. ", priorVersion, typeName)+
						"This is always an issue with the Terraform Provider and should be reported to the provider developer:\n\n"+err.Error(),
				)
				return
			}

			transformed, transformDiags := upgrade.Transform(ctx, prior)
			resp.Diagnostics.Append(transformDiags...)
			if resp.Diagnostics.HasError() {
				return
			}

			marshaled, err := json.Marshal(transformed)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to Upgrade Resource State",
					fmt.Sprintf("Failed to serialise the transformed state for the schema version %d upgrade of %q. ", priorVersion, typeName)+
						"This is always an issue with the Terraform Provider and should be reported to the provider developer:\n\n"+err.Error(),
				)
				return
			}
			rawJSON = marshaled
		}

		rawStateValue, err := (&tfprotov6.RawState{JSON: rawJSON}).UnmarshalWithOpts(
			resp.State.Schema.Type().TerraformType(ctx),
			tfprotov6.UnmarshalOpts{
				// Attributes in the stored state that the current schema no longer declares are
				// skipped rather than failing the upgrade, matching the framework's own passthrough.
				ValueFromJSONOpts: tftypes.ValueFromJSONOpts{IgnoreUndefinedAttributes: true},
			},
		)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Read Previously Saved State for UpgradeResourceState",
				fmt.Sprintf("There was an error reading the saved resource state for the schema version %d upgrade of %q using the current resource schema.\n\n", priorVersion, typeName)+
					"If you manually modified the resource state, you will need to manually modify it to match the current resource schema. "+
					"Otherwise, please report this to the provider developer:\n\n"+err.Error(),
			)
			return
		}

		// Must run on the decoded prior state, before it is handed back: the framework nulls every
		// write-only attribute in what this returns, so this is the last point at which the
		// plain-text credential the old state holds is still readable. The hash sibling itself is
		// Computed, not write-only, so the adopted tag survives that nullification.
		rawStateValue, adoptDiags := schemas.AdoptWriteOnlyHashTagsFromPriorState(
			ctx, s.actionDefinition.WriteOnlyHashedAttributes, rawStateValue)
		resp.Diagnostics.Append(adoptDiags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Values are never logged here: the state being upgraded is precisely the state that may
		// still hold a credential in plain text.
		tflog.Info(ctx, "Upgrading resource state to the current schema version", map[string]interface{}{
			"resource":        typeName,
			"prior_version":   priorVersion,
			"current_version": s.actionDefinition.ActionVersion,
		})

		resp.State.Raw = rawStateValue
	}
}
