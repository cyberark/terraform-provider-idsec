// Copyright CyberArk. 2026
// SPDX-License-Identifier: Apache-2.0

package schemas

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// WriteOnlyHashAdoptPrivateKey is the resource private-state key, set on import, listing the
// hashed write-only attributes whose hash should be adopted from configuration on the next apply
// instead of their value being sent. Without it, an existing resource with a null hash sends the
// value: only an import shows the backend already holds that credential.
const WriteOnlyHashAdoptPrivateKey = "idsec_wo_hash_adopt"

// ReadPendingHashAdoption returns the set of hashed write-only attribute paths recorded under
// WriteOnlyHashAdoptPrivateKey.
//
// A nil reader, a missing key, or an undecodable value all yield nil, which degrades to sending the
// credential rather than silently dropping it.
//
// Parameters:
// - ctx: Request context
// - reader: The resource's private-state data; nil is treated as empty
//
// Returns the set of pending attribute paths, or nil when none are pending.
func ReadPendingHashAdoption(ctx context.Context, reader privateStateReader) map[string]bool {
	if reader == nil {
		return nil
	}
	raw, diags := reader.GetKey(ctx, WriteOnlyHashAdoptPrivateKey)
	if diags.HasError() || len(raw) == 0 {
		return nil
	}
	var attrs []string
	if err := json.Unmarshal(raw, &attrs); err != nil || len(attrs) == 0 {
		return nil
	}
	pending := make(map[string]bool, len(attrs))
	for _, a := range attrs {
		pending[a] = true
	}
	return pending
}

// AdoptWriteOnlyHashTagsFromPriorState returns priorValue with each hashed write-only attribute's
// hash sibling set to a tag derived from the plain-text secret that attribute still holds in
// priorValue. It exists for one caller: the state upgrader that runs when an already-released
// attribute is marked write-only, where the state being upgraded predates hash mode and so carries
// the credential in plain text with a null sibling.
//
// Adopting the tag here is what stops the first apply after the upgrade from re-sending the
// credential to the API: a null sibling always plans unknown, which IdsecResource.triggerFired
// reads as "changed" and so sends the value -- the rotation that a vault may reject outright
// ("the password typed does not adhere to current password policy rules") and that would clobber
// an out-of-band rotation when it succeeds. With the tag adopted, an unchanged secret in
// configuration plans to a byte-identical tag (hashPlanModifier reuses the salt embedded here) and
// the trigger does not fire.
//
// Unlike the import-time marker (WriteOnlyHashAdoptPrivateKey), which suppresses the send
// unconditionally because nothing reveals what the backend holds, this derives the tag from the
// secret the backend was actually last sent. So a practitioner who edits the secret in the same run
// as the provider upgrade still gets it sent: its tag differs from the adopted one. The marker
// route is unavailable here regardless -- UpgradeStateResponse carries no private state.
//
// Every failure mode leaves the sibling null, which degrades to the pre-existing behaviour of
// sending the credential once rather than silently recording a hash for a value the backend may not
// hold. That covers: an attribute whose secret is absent or null in prior state (e.g. an upgrade
// from a release where it was already write-only under a manual trigger, so state never held it),
// a sibling the current schema does not declare, and a salt that could not be generated.
//
// A sibling that is already non-null in prior state is left untouched: that state was written by a
// release that already had hash mode, and its tag is the authoritative rotation baseline.
//
// Parameters:
// - ctx: Request context
// - hashedAttrs: The resource's WriteOnlyHashedAttributes; empty makes this a no-op
// - priorValue: The prior state decoded against the current schema
//
// Returns the (possibly rewritten) state value and any diagnostics. Only warnings are ever
// returned: a failure to adopt must not turn a provider upgrade into a dead end where the resource
// can no longer be planned, refreshed, or destroyed.
func AdoptWriteOnlyHashTagsFromPriorState(ctx context.Context, hashedAttrs []string, priorValue tftypes.Value) (tftypes.Value, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(hashedAttrs) == 0 || priorValue.IsNull() || priorValue.Type() == nil || !priorValue.Type().Is(tftypes.Object{}) {
		return priorValue, diags
	}

	rawAttrs := map[string]tftypes.Value{}
	if err := priorValue.As(&rawAttrs); err != nil {
		tflog.Warn(ctx, fmt.Sprintf("AdoptWriteOnlyHashTagsFromPriorState: prior state could not be read as an object, leaving hash siblings null: %s", err.Error()))
		return priorValue, diags
	}
	// As hands back the value's own attribute map, so mutating it would rewrite priorValue in place
	// and leave the caller unable to fall back to the value it passed in -- which every failure path
	// below does. Clone before touching anything, mirroring setWriteOnlyHashTagsInState.
	attrs := maps.Clone(rawAttrs)

	changed := false
	// Sorted for a stable log and diagnostic order across runs. A dotted (nested) path never
	// matches a key here and so is skipped, which is consistent with hash mode rejecting nested
	// sources at schema-construction time (validateWriteOnlyHashedKey).
	for _, attrPath := range slices.Sorted(slices.Values(hashedAttrs)) {
		siblingName := WriteOnlyHashAttributeName(attrPath)

		sibling, ok := attrs[siblingName]
		if !ok {
			tflog.Warn(ctx, fmt.Sprintf("AdoptWriteOnlyHashTagsFromPriorState: sibling %q is not declared by the current schema, skipping", siblingName))
			continue
		}
		if !sibling.IsNull() {
			continue
		}

		secretVal, ok := attrs[attrPath]
		if !ok || secretVal.IsNull() || !secretVal.IsFullyKnown() || secretVal.Type() == nil || !secretVal.Type().Is(tftypes.String) {
			tflog.Info(ctx, fmt.Sprintf("AdoptWriteOnlyHashTagsFromPriorState: prior state holds no readable value for %q, so its hash cannot be adopted and the configured value will be sent on the next apply", attrPath))
			continue
		}
		var secret string
		if err := secretVal.As(&secret); err != nil {
			tflog.Warn(ctx, fmt.Sprintf("AdoptWriteOnlyHashTagsFromPriorState: the prior value of %q could not be read as a string, so its hash cannot be adopted: %s", attrPath, err.Error()))
			continue
		}

		salt, err := newSalt(nil)
		if err != nil {
			diags.AddAttributeWarning(
				path.Root(siblingName),
				"Unable To Adopt Stored Credential Hash",
				fmt.Sprintf(
					"Could not generate a random salt to hash the previously stored value of %q: %s. "+
						"Its hash stays unset, so the configured value will be sent to the API once on the "+
						"next apply.", attrPath, err),
			)
			continue
		}

		// Values are never logged: the tag is credential-derived and the source is the credential.
		tflog.Info(ctx, fmt.Sprintf("AdoptWriteOnlyHashTagsFromPriorState: adopting the hash of the previously stored value of %q, so an unchanged secret is not re-sent on the next apply", attrPath))
		attrs[siblingName] = tftypes.NewValue(tftypes.String, deriveTag(secret, salt))
		changed = true
	}

	if !changed {
		return priorValue, diags
	}
	return tftypes.NewValue(priorValue.Type(), attrs), diags
}
