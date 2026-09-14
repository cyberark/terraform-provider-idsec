// Copyright CyberArk. 2026
// SPDX-License-Identifier: Apache-2.0

package schemas

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/crypto/pbkdf2"
)

// writeOnlyHashAttributeSuffix is appended to a write-only hashed attribute's own name to name its
// synthesized sibling.
const writeOnlyHashAttributeSuffix = "_write_only_hash"

// WriteOnlyHashAttributeName returns the name of the synthesized hash sibling for the given
// write-only attribute path. Both the schema pass (applyWriteOnlyHashedAttributes) and the
// resource-side resolver (IdsecResource.effectiveWriteOnlyTriggers) must call this; the two sides
// must never diverge, or the schema would declare one trigger name while the apply-side logic
// looks for another.
func WriteOnlyHashAttributeName(attrPath string) string {
	return attrPath + writeOnlyHashAttributeSuffix
}

const (
	// hashFormatVersion is the leading field of every tag deriveTag produces. Bump it, and reject
	// every other value in parseTag, if pbkdf2Iterations, saltLen, or tagLen ever change: an old
	// tag hashed under different parameters must not be compared against a newly derived one.
	hashFormatVersion = "v1"
	pbkdf2Iterations  = 210_000 // GOLDEN VECTOR CONSTANT -- do not change without a format version bump.
	saltLen           = 16
	tagLen            = 32
)

// newSalt returns saltLen cryptographically random bytes read from r, or from crypto/rand when r
// is nil. A short read from r is returned as an error rather than padded or retried, so a caller
// can never be handed a salt that is silently weaker than requested.
func newSalt(r io.Reader) ([]byte, error) {
	if r == nil {
		r = rand.Reader
	}
	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(r, salt); err != nil {
		return nil, fmt.Errorf("failed to generate a %d-byte salt: %w", saltLen, err)
	}
	return salt, nil
}

// deriveTag returns "v1:<base64url-salt>:<base64url-tag>", where tag is the PBKDF2-HMAC-SHA256
// derivation of secret under salt. The salt travels alongside the tag specifically so a later
// call can recover it via parseTag and derive a byte-identical tag for an unchanged secret,
// producing no plan diff.
func deriveTag(secret string, salt []byte) string {
	tag := pbkdf2.Key([]byte(secret), salt, pbkdf2Iterations, tagLen, sha256.New)
	return strings.Join([]string{
		hashFormatVersion,
		base64.RawURLEncoding.EncodeToString(salt),
		base64.RawURLEncoding.EncodeToString(tag),
	}, ":")
}

// parseTag parses a tag string produced by deriveTag. It is strict and never panics: any
// malformed input -- the wrong number of colon-separated parts, an unrecognized version, invalid
// base64, or a decoded segment of the wrong length -- returns ok=false with every other return
// value zeroed, rather than a partially-populated result.
func parseTag(s string) (version string, salt []byte, tag []byte, ok bool) {
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return "", nil, nil, false
	}
	if parts[0] != hashFormatVersion {
		return "", nil, nil, false
	}
	decodedSalt, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || len(decodedSalt) != saltLen {
		return "", nil, nil, false
	}
	decodedTag, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || len(decodedTag) != tagLen {
		return "", nil, nil, false
	}
	return parts[0], decodedSalt, decodedTag, true
}

// reusableSalt returns the salt embedded in a prior state tag, or nil when there is no prior tag
// or it is not a tag this version can parse. nil signals that fresh randomness is required, which
// only the apply side may draw: generating a salt at plan time makes PlanModifyString
// non-deterministic across the two PlanResourceChange calls Terraform issues during apply, and
// Terraform rejects a planned known value that changed.
func reusableSalt(priorTag types.String) []byte {
	if priorTag.IsNull() || priorTag.IsUnknown() {
		return nil
	}
	_, salt, _, ok := parseTag(priorTag.ValueString())
	if !ok {
		return nil
	}
	return salt
}

// isHashableSource reports whether a is eligible to be the source attribute of a write-only hash.
// Hash mode is string-only, for two independent reasons, either of which would be sufficient on
// its own: deriveTag hashes a Go string, and hashPlanModifier.PlanModifyString can only extract one
// from a tftypes.Value via rawVal.As(&secret), so a non-string source can never produce a tag --
// its trigger would stay perpetually null, and the secret would be sent on create and never rotated
// again, which is exactly the silent failure this feature exists to eliminate. And even if it could
// be computed, a bool has only two possible values and is trivially reversible from its tag, so a
// hashed bool is not a meaningful rotation signal; an int64 credential is not a real use case here
// either. validateWriteOnlyHashedKey rejects anything else at schema-construction time instead of
// letting it fail silently at apply time.
func isHashableSource(a schema.Attribute) bool {
	_, ok := a.(schema.StringAttribute)
	return ok
}

// hashedWriteOnlyAttributeDescriptionSuffix names the synthesized hash sibling in the source
// attribute's own description, mirroring writeOnlyAttributeDescriptionSuffix but for hash mode:
// there is no trigger for the practitioner to set, since rotation is detected automatically.
func hashedWriteOnlyAttributeDescriptionSuffix(hashAttrName string) string {
	return fmt.Sprintf(
		"Write-only: this value is never stored in Terraform state or plan. It is sent to the API on "+
			"create, and on update whenever it changes; rotation is detected automatically via the "+
			"computed `%s` attribute, so no manual trigger needs to be bumped. Requires Terraform 1.11 "+
			"or later.",
		hashAttrName,
	)
}

// validateWriteOnlyHashedKey checks that attrPath names an attribute that can legally have a
// write-only hash sibling synthesized for it: it must exist as a top-level string attribute, it
// must not also be declared in WriteOnlyAttributes, its synthesized sibling name must not collide
// with an existing attribute or a shadowed SDK model field, and it must pass every ordinary
// write-only eligibility check (reusing validateWriteOnlyKey, since hash mode's source attribute
// ends up write-only in exactly the same shape a manual declaration would produce).
func validateWriteOnlyHashedKey(attrs map[string]schema.Attribute, attrPath string, opts resourceSchemaOptions) diag.Diagnostics {
	var diags diag.Diagnostics

	hashName := WriteOnlyHashAttributeName(attrPath)

	if strings.Contains(attrPath, ".") {
		diags.AddError(errWriteOnlySummary, fmt.Sprintf(
			"write-only hashed attribute %q is nested, but its synthesized hash sibling %q is always "+
				"top-level, which would put the secret and its trigger in different containers. "+
				"Restructure %q as a top-level attribute, or declare it in WriteOnlyAttributes with a "+
				"manual top-level trigger instead.", attrPath, hashName, attrPath))
		return diags
	}

	a, ok := attrs[attrPath]
	if !ok {
		diags.AddError(errWriteOnlySummary, fmt.Sprintf(
			"write-only hashed attribute %q does not exist in the generated schema. The key is matched "+
				"by its exact top-level attribute name.", attrPath))
		return diags
	}

	if !isHashableSource(a) {
		diags.AddError(errWriteOnlySummary, fmt.Sprintf(
			"write-only hashed attribute %q is type %T, not a string. Hash mode derives its tag from a "+
				"string secret, so a non-string source can never produce one. Declare it in "+
				"WriteOnlyAttributes with a manual trigger instead, or choose a different attribute to "+
				"hash.", attrPath, a))
		return diags
	}

	if _, isManual := opts.WriteOnlyAttributes[attrPath]; isManual {
		diags.AddError(errWriteOnlySummary, fmt.Sprintf(
			"write-only hashed attribute %q is also a key in WriteOnlyAttributes, so it would be marked "+
				"write-only twice with two different triggers. Remove it from one of the two maps.", attrPath))
		return diags
	}

	if match := normalizedTopLevelCollision(attrs, hashName); match != "" {
		diags.AddError(errWriteOnlySummary, fmt.Sprintf(
			"write-only hashed attribute %q would synthesize %q, which collides with existing attribute "+
				"%q once lowercased with hyphens and underscores removed. Rename %q, or choose a "+
				"different attribute to hash.", attrPath, hashName, match, match))
		return diags
	}

	if modelPath, found := findModelFieldNotInSchema(opts, hashName); found {
		diags.AddError(errWriteOnlySummary, fmt.Sprintf(
			"write-only hashed attribute %q would synthesize %q, but that name resolves to SDK model "+
				"field %q, which is absent from the generated schema. Rename the model field, or choose "+
				"a different attribute to hash.", attrPath, hashName, modelPath))
		return diags
	}

	// Every remaining rule -- computed-only, set-typed (already excluded above, but kept for a
	// consistent error voice), a `default` tag, and ImmutableAttributes/ForceNewAttributes
	// membership -- is exactly the ordinary write-only eligibility check: the source attribute
	// ends up write-only in the same shape a manual declaration would produce.
	diags.Append(validateWriteOnlyKey(attrs, attrPath, opts)...)
	return diags
}

// applyWriteOnlyHashedAttributes validates every entry in opts.WriteOnlyHashedAttributes and, only
// if validation finds no errors, mutates attrs in place to synthesize each entry's
// "<attr>_write_only_hash" sibling and mark the source attribute write-only with that sibling as
// its trigger.
//
// generateResourceSchema calls this immediately after applyWriteOnlyAttributes, so this pass sees
// the final computed-only flags and every attribute the manual pass already touched; running any
// earlier could validate against a schema that is about to change underneath it.
//
// Entries are iterated in sorted order over a deduplicated copy, both because attrs is mutated as
// the pass proceeds and so that a resource with several invalid entries reports the same error
// list on every run.
//
// Returns diagnostics. If any diagnostic is an error, attrs is guaranteed unmodified by this
// function. Note that when called from generateResourceSchema, applyWriteOnlyAttributes runs
// first and may have already committed mutations before this pass detects a conflict (e.g. the
// same path in both WriteOnlyAttributes and WriteOnlyHashedAttributes); callers must treat the
// returned schema as unusable when diags.HasError() is true regardless.
func applyWriteOnlyHashedAttributes(attrs map[string]schema.Attribute, opts resourceSchemaOptions) diag.Diagnostics {
	var diags diag.Diagnostics
	if len(opts.WriteOnlyHashedAttributes) == 0 {
		return diags
	}

	counts := make(map[string]int, len(opts.WriteOnlyHashedAttributes))
	for _, attrPath := range opts.WriteOnlyHashedAttributes {
		counts[attrPath]++
	}

	keys := slices.Sorted(maps.Keys(counts))

	for _, attrPath := range keys {
		if counts[attrPath] > 1 {
			diags.AddError(errWriteOnlySummary, fmt.Sprintf(
				"write-only hashed attribute %q is declared %d times in WriteOnlyHashedAttributes. List "+
					"each attribute path once.", attrPath, counts[attrPath]))
			continue
		}
		diags.Append(validateWriteOnlyHashedKey(attrs, attrPath, opts)...)
	}

	if diags.HasError() {
		return diags
	}

	for _, attrPath := range keys {
		hashName := WriteOnlyHashAttributeName(attrPath)

		attrs[hashName] = schema.StringAttribute{
			Computed:  true,
			Sensitive: true,
			// Deliberately NOT Optional: a practitioner who could set this could forge a matching tag
			// and permanently suppress rotation of the secret.
			Description: fmt.Sprintf(
				"Automatically computed hash of %s. Changes when the secret changes; used to detect "+
					"rotation without storing the raw value in state.", attrPath),
			PlanModifiers: []planmodifier.String{newHashPlanModifier(attrPath)},
		}

		marked, markDiags := markAttributeWriteOnly(
			attrs[attrPath], hashName, hashedWriteOnlyAttributeDescriptionSuffix(hashName), false)
		diags.Append(markDiags...)
		attrs[attrPath] = marked
	}

	return diags
}

// hashPlanModifier computes the planned value of a synthesized write-only hash attribute from the
// secret in configuration, reusing the salt embedded in the prior state value so that an unchanged
// secret plans to a byte-identical tag and produces no diff.
//
// It may only ever produce a value it can derive deterministically: when a prior salt is
// reusable, the tag is concrete; otherwise it plans unknown rather than drawing fresh randomness.
// Terraform issues two PlanResourceChange calls during a single apply, and a plan modifier that
// drew a new random salt on each call would commit two different concrete tags, which Terraform
// rejects as an inconsistent final plan. The concrete tag for the no-reusable-salt case (first
// create, or a stored tag this version cannot parse) is instead derived on the apply side by
// ResolveWriteOnlyHashTags, which runs once per apply and so can safely draw randomness.
type hashPlanModifier struct {
	sourcePath string
}

// newHashPlanModifier returns a plan modifier that hashes the secret at sourcePath.
func newHashPlanModifier(sourcePath string) hashPlanModifier {
	return hashPlanModifier{sourcePath: sourcePath}
}

var _ planmodifier.String = hashPlanModifier{}

// Description returns a description of the plan modifier.
func (m hashPlanModifier) Description(_ context.Context) string {
	return fmt.Sprintf(
		"Computes a PBKDF2-HMAC-SHA256 tag of %q, reusing the salt embedded in the prior state value "+
			"so an unchanged secret plans to no diff.", m.sourcePath)
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m hashPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

// PlanModifyString computes the planned hash tag from the secret at m.sourcePath in configuration.
// A schema-generation bug that makes the source path unwalkable, or resolves it to something other
// than a value, must not break planning: this leaves resp.PlanValue at its framework-assigned
// default and returns rather than raising a diagnostic.
//
// This method may only assign resp.PlanValue a value it can derive deterministically: a concrete
// tag when a prior salt is reusable, or types.StringUnknown() otherwise. It never draws fresh
// randomness itself -- see the hashPlanModifier doc comment and ResolveWriteOnlyHashTags for why.
func (m hashPlanModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.Config.Raw.IsNull() {
		return
	}

	sourceAttrPath := tftypes.NewAttributePath()
	for _, part := range strings.Split(m.sourcePath, ".") {
		sourceAttrPath = sourceAttrPath.WithAttributeName(part)
	}

	result, _, err := tftypes.WalkAttributePath(req.Config.Raw, sourceAttrPath)
	if err != nil {
		return
	}
	rawVal, ok := result.(tftypes.Value)
	if !ok {
		return
	}

	if rawVal.IsNull() {
		// No secret configured: no rotation to detect, so the hash tracks whatever is (or is not)
		// already in state.
		resp.PlanValue = req.StateValue
		return
	}

	if !rawVal.IsFullyKnown() {
		resp.PlanValue = types.StringUnknown()
		resp.Diagnostics.AddAttributeWarning(
			req.Path,
			"Secret Not Known At Plan Time",
			fmt.Sprintf(
				"The value of %q is not known until apply, so rotation cannot be evaluated until then. "+
					"Its hash will be recomputed and the value re-sent to the API on this apply.",
				m.sourcePath,
			),
		)
		return
	}

	var secret string
	if err := rawVal.As(&secret); err != nil {
		// validateWriteOnlyHashedKey (via isHashableSource) guarantees the source attribute is a
		// StringAttribute, so a validly generated schema never reaches this branch. It stays as
		// defense against a generator bug -- a schema that somehow attached this plan modifier to a
		// non-string source -- rather than an expected path, so planning degrades to a no-op instead
		// of panicking.
		return
	}

	salt := reusableSalt(req.StateValue)
	if salt == nil {
		// No reusable prior salt: a concrete tag would require fresh randomness, making
		// PlanModifyString non-deterministic across the two PlanResourceChange calls Terraform
		// issues during apply. Plan unknown ("known after apply"); the tag is derived on the
		// apply side by ResolveWriteOnlyHashTags.
		if !req.StateValue.IsNull() && !req.StateValue.IsUnknown() {
			// Existing resource with a stored tag this version cannot parse (garbage or a future
			// hashFormatVersion). The apply-side resolver re-derives with a fresh salt, causing a
			// one-time idempotent re-send of the credential to self-heal.
			resp.Diagnostics.AddAttributeWarning(
				req.Path,
				"Stored Hash Not Recognized",
				fmt.Sprintf(
					"The stored hash for %q uses an unrecognized format. Its hash will be "+
						"recomputed and the value re-sent to the API on this apply to self-heal.",
					m.sourcePath,
				),
			)
		}
		resp.PlanValue = types.StringUnknown()
		return
	}
	resp.PlanValue = types.StringValue(deriveTag(secret, salt))
}

// ResolveWriteOnlyHashTags derives the concrete hash tag to persist for each hashed write-only
// attribute whose planned sibling value is unknown (i.e. there was no reusable prior salt at plan
// time). It returns a map keyed by SIBLING attribute name (e.g. "password_write_only_hash") to
// the resolved types.String value that must be written into state after MergePlanToStateObject.
//
// Must be called before the API request is issued so a derivation failure aborts the operation
// rather than orphaning a created resource with a null hash in state.
//
// Returns an empty map (and no diagnostics) when plan or config is nil, which covers Delete and
// Read where there is nothing to resolve.
func ResolveWriteOnlyHashTags(
	ctx context.Context,
	hashedAttrs []string,
	plan *tfsdk.Plan,
	priorState *tfsdk.State,
	config *tfsdk.Config,
) (map[string]types.String, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(hashedAttrs) == 0 || plan == nil || plan.Raw.IsNull() || config == nil {
		return nil, diags
	}

	result := make(map[string]types.String, len(hashedAttrs))

	for _, attrPath := range slices.Sorted(slices.Values(hashedAttrs)) {
		siblingName := WriteOnlyHashAttributeName(attrPath)

		// Walk the plan to find the sibling's planned value.
		siblingTFPath := tftypes.NewAttributePath().WithAttributeName(siblingName)
		siblingResult, _, err := tftypes.WalkAttributePath(plan.Raw, siblingTFPath)
		if err != nil {
			// The sibling does not exist in the plan -- schema mismatch; skip gracefully.
			tflog.Warn(ctx, fmt.Sprintf("ResolveWriteOnlyHashTags: sibling %q not found in plan: %v", siblingName, err))
			continue
		}
		siblingVal, ok := siblingResult.(tftypes.Value)
		if !ok || (!siblingVal.IsNull() && siblingVal.IsFullyKnown()) {
			// Planned value is already concrete: the modifier committed it (salt-reuse path).
			// Never contradict a known planned value.
			continue
		}

		// Walk the config to get the secret.
		sourceTFPath := tftypes.NewAttributePath().WithAttributeName(attrPath)
		sourceResult, _, err := tftypes.WalkAttributePath(config.Raw, sourceTFPath)
		if err != nil || sourceResult == nil {
			tflog.Warn(ctx, fmt.Sprintf("ResolveWriteOnlyHashTags: source %q not found in config: %v", attrPath, err))
			result[siblingName] = types.StringNull()
			continue
		}
		sourceVal, ok := sourceResult.(tftypes.Value)
		if !ok || sourceVal.IsNull() {
			result[siblingName] = types.StringNull()
			continue
		}
		if !sourceVal.IsFullyKnown() {
			diags.AddAttributeWarning(
				path.Root(attrPath),
				"Secret Not Known At Apply Time",
				fmt.Sprintf("The value of %q is not known during apply; its hash cannot be derived and will be null until the next apply.", attrPath),
			)
			result[siblingName] = types.StringNull()
			continue
		}

		var secret string
		if err := sourceVal.As(&secret); err != nil {
			diags.AddAttributeError(
				path.Root(attrPath),
				"Failed To Read Secret",
				fmt.Sprintf("Could not read the value of %q as a string: %s", attrPath, err),
			)
			continue
		}

		// Determine salt: reuse prior state salt when available (preserves rotation stability for
		// ephemeral secrets), draw a fresh one for first-create and post-import.
		var salt []byte
		if priorState != nil && !priorState.Raw.IsNull() {
			siblingPriorResult, _, priorErr := tftypes.WalkAttributePath(priorState.Raw, siblingTFPath)
			if priorErr == nil {
				if priorVal, ok2 := siblingPriorResult.(tftypes.Value); ok2 {
					var priorTagStr string
					if err2 := priorVal.As(&priorTagStr); err2 == nil {
						salt = reusableSalt(types.StringValue(priorTagStr))
					}
				}
			}
		}
		if salt == nil {
			var saltErr error
			salt, saltErr = newSalt(nil)
			if saltErr != nil {
				diags.AddAttributeError(
					path.Root(siblingName),
					"Failed To Generate Salt",
					fmt.Sprintf("Could not generate a random salt to hash %q: %s. The operation is aborted to avoid persisting a resource without a valid hash.", attrPath, saltErr),
				)
				continue
			}
		}

		result[siblingName] = types.StringValue(deriveTag(secret, salt))
	}

	return result, diags
}
