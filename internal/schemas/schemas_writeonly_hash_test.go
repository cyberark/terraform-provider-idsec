// Copyright CyberArk. 2026
// SPDX-License-Identifier: Apache-2.0

package schemas

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ---- Hashing primitive tests ----

// TestDeriveTag_Determinism guards the property hashPlanModifier's whole no-diff design depends
// on: the same secret hashed under the same salt, on repeated calls, must produce a byte-identical
// tag every time.
func TestDeriveTag_Determinism(t *testing.T) {
	t.Parallel()

	salt := bytes.Repeat([]byte{0x11}, saltLen)
	var want string
	for i := range 5 {
		got := deriveTag("s3cr3t", salt)
		if i == 0 {
			want = got
			continue
		}
		if got != want {
			t.Fatalf("deriveTag is not deterministic: run %d got %q, want %q", i, got, want)
		}
	}
}

// TestDeriveTag_SaltSensitivity guards against a derivation that ignores its salt argument.
func TestDeriveTag_SaltSensitivity(t *testing.T) {
	t.Parallel()

	saltA := bytes.Repeat([]byte{0x01}, saltLen)
	saltB := bytes.Repeat([]byte{0x02}, saltLen)

	tagA := deriveTag("s3cr3t", saltA)
	tagB := deriveTag("s3cr3t", saltB)
	if tagA == tagB {
		t.Errorf("expected two different salts to produce different tags, both got %q", tagA)
	}
}

// TestDeriveTag_SecretSensitivity guards against a derivation that ignores its secret argument.
func TestDeriveTag_SecretSensitivity(t *testing.T) {
	t.Parallel()

	salt := bytes.Repeat([]byte{0x03}, saltLen)

	tagA := deriveTag("secret-a", salt)
	tagB := deriveTag("secret-b", salt)
	if tagA == tagB {
		t.Errorf("expected two different secrets to produce different tags, both got %q", tagA)
	}
}

// TestParseTag_RoundTrip guards that a tag deriveTag produces can always be parsed back into its
// exact version, salt, and tag bytes -- the property hashPlanModifier's salt-reuse path depends on.
func TestParseTag_RoundTrip(t *testing.T) {
	t.Parallel()

	salt := bytes.Repeat([]byte{0x2a}, saltLen)
	tagStr := deriveTag("s3cr3t", salt)

	version, parsedSalt, parsedTag, ok := parseTag(tagStr)
	if !ok {
		t.Fatalf("expected parseTag to accept a tag deriveTag just produced, got ok=false for %q", tagStr)
	}
	if version != hashFormatVersion {
		t.Errorf("expected version %q, got %q", hashFormatVersion, version)
	}
	if !bytes.Equal(parsedSalt, salt) {
		t.Errorf("expected parsed salt to equal the input salt, got %x want %x", parsedSalt, salt)
	}
	if len(parsedTag) != tagLen {
		t.Errorf("expected a %d-byte tag, got %d bytes", tagLen, len(parsedTag))
	}
}

// TestParseTag_Rejections covers every malformed input parseTag must reject without panicking,
// returning ok=false and every other return value zeroed.
func TestParseTag_Rejections(t *testing.T) {
	t.Parallel()

	validSalt := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0xab}, saltLen))
	validTag := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0xcd}, tagLen))
	shortSalt := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0xab}, saltLen-1))
	shortTag := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0xcd}, tagLen-1))

	tests := []struct {
		name  string
		input string
	}{
		{name: "empty_string", input: ""},
		{name: "one_part", input: "v1"},
		{name: "two_parts", input: "v1:" + validSalt},
		{name: "four_parts", input: "v1:" + validSalt + ":" + validTag + ":extra"},
		{name: "unknown_version", input: "v2:" + validSalt + ":" + validTag},
		{name: "invalid_base64_salt", input: "v1:not-valid-base64!!:" + validTag},
		{name: "invalid_base64_tag", input: "v1:" + validSalt + ":not-valid-base64!!"},
		{name: "salt_wrong_length", input: "v1:" + shortSalt + ":" + validTag},
		{name: "tag_wrong_length", input: "v1:" + validSalt + ":" + shortTag},
	}

	for _, tt := range tests {
		t.Run("rejects_"+tt.name, func(t *testing.T) {
			t.Parallel()

			version, salt, tag, ok := parseTag(tt.input)
			if ok {
				t.Fatalf("expected ok=false for input %q, got version=%q salt=%x tag=%x", tt.input, version, salt, tag)
			}
			if version != "" || salt != nil || tag != nil {
				t.Errorf("expected every return value zeroed on rejection, got version=%q salt=%x tag=%x", version, salt, tag)
			}
		})
	}
}

// TestDeriveTag_GoldenVector cross-checks the Go implementation against an INDEPENDENT PBKDF2-
// HMAC-SHA256 implementation (Python's hashlib), rather than against Go's own output, so this test
// can actually catch a wrong parameter or algorithm rather than only detect a future change to the
// existing (possibly already-wrong) Go output.
//
// The expected literal below was produced by running, on the command line:
//
//	python3 -c "import hashlib,base64; t=hashlib.pbkdf2_hmac('sha256', b'test-secret', b'\x00'*16, 210000, 32); print('v1:'+base64.urlsafe_b64encode(b'\x00'*16).decode().rstrip('=')+':'+base64.urlsafe_b64encode(t).decode().rstrip('='))"
//
// which printed:
//
//	v1:AAAAAAAAAAAAAAAAAAAAAA:7dJMm-qi96TC_tZTyvi47uFCnE3CngcDeBHHxbHBl9g
//
// The constants are asserted here too: changing pbkdf2Iterations, saltLen, tagLen, or
// hashFormatVersion invalidates every "<attr>_write_only_hash" tag already stored in practitioner
// state (an old tag was derived under the old parameters and can never again compare equal to one
// derived under new ones), which silently and permanently disables rotation detection for every
// existing resource using this feature until the practitioner force-replaces it. That is exactly
// the scenario hashFormatVersion exists to make an explicit, versioned migration instead of a
// silent one, so a parameter bump SHOULD fail this test -- the fix is a new format version, not a
// changed expectation here.
func TestDeriveTag_GoldenVector(t *testing.T) {
	t.Parallel()

	if hashFormatVersion != "v1" {
		t.Errorf("hashFormatVersion changed to %q: bump the format version and this golden vector together, since every stored tag under the old parameters is now invalid", hashFormatVersion)
	}
	if pbkdf2Iterations != 210_000 {
		t.Errorf("pbkdf2Iterations changed to %d: bump the format version and this golden vector together, since every stored tag under the old parameters is now invalid", pbkdf2Iterations)
	}
	if saltLen != 16 {
		t.Errorf("saltLen changed to %d: bump the format version and this golden vector together, since every stored tag under the old parameters is now invalid", saltLen)
	}
	if tagLen != 32 {
		t.Errorf("tagLen changed to %d: bump the format version and this golden vector together, since every stored tag under the old parameters is now invalid", tagLen)
	}

	const want = "v1:AAAAAAAAAAAAAAAAAAAAAA:7dJMm-qi96TC_tZTyvi47uFCnE3CngcDeBHHxbHBl9g"
	got := deriveTag("test-secret", make([]byte, saltLen))
	if got != want {
		t.Errorf("deriveTag disagrees with the independent Python PBKDF2-HMAC-SHA256 implementation: got %q, want %q", got, want)
	}
}

// TestNewSalt covers randomness, non-determinism, and strict short-read handling.
func TestNewSalt(t *testing.T) {
	t.Parallel()

	t.Run("nil_reader_uses_crypto_rand_and_returns_saltLen_bytes", func(t *testing.T) {
		t.Parallel()
		salt, err := newSalt(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(salt) != saltLen {
			t.Errorf("expected %d bytes, got %d", saltLen, len(salt))
		}
	})

	t.Run("two_successive_calls_differ", func(t *testing.T) {
		t.Parallel()
		saltA, err := newSalt(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		saltB, err := newSalt(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if bytes.Equal(saltA, saltB) {
			t.Error("expected two successive salts to differ; got the same bytes both times, suggesting a stuck or constant source")
		}
	})

	t.Run("reader_error_returns_error_and_nil_salt", func(t *testing.T) {
		t.Parallel()
		salt, err := newSalt(iotest.ErrReader(errors.New("boom")))
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
		if salt != nil {
			t.Errorf("expected a nil salt on error, got %x", salt)
		}
	})

	t.Run("short_read_then_EOF_returns_error_not_a_short_or_padded_salt", func(t *testing.T) {
		t.Parallel()
		// A bytes.Reader over a buffer shorter than saltLen reads what it has and then reports
		// EOF, exercising io.ReadFull's short-read path distinctly from an immediate error.
		salt, err := newSalt(bytes.NewReader(make([]byte, saltLen-1)))
		if err == nil {
			t.Fatal("expected an error for a short read, got nil")
		}
		if salt != nil {
			t.Errorf("expected a nil salt on a short read, got %x (len=%d)", salt, len(salt))
		}
	})
}

// TestDeriveTag_NotNaiveDigest catches a regression to a plain, unsalted digest of the secret: a
// tag must not be recoverable as (or contain) a bare SHA-256 of the secret, which would make an
// exfiltrated tag directly reversible via a rainbow table rather than requiring 210,000 PBKDF2
// iterations per guess.
func TestDeriveTag_NotNaiveDigest(t *testing.T) {
	t.Parallel()

	secret := "s3cr3t"
	salt := bytes.Repeat([]byte{0x44}, saltLen)

	naiveDigest := sha256.Sum256([]byte(secret))
	naiveDigestB64 := base64.RawURLEncoding.EncodeToString(naiveDigest[:])

	tagStr := deriveTag(secret, salt)
	if strings.Contains(tagStr, naiveDigestB64) {
		t.Errorf("tag %q contains the base64 of a bare SHA-256(secret); expected a PBKDF2-derived tag unrelated to the naive digest", tagStr)
	}

	_, _, tagBytes, ok := parseTag(tagStr)
	if !ok {
		t.Fatalf("expected a parseable tag, got %q", tagStr)
	}
	if bytes.Equal(tagBytes, naiveDigest[:]) {
		t.Error("decoded tag bytes equal a bare SHA-256(secret); expected a PBKDF2-derived tag, not a naive digest")
	}
}

// BenchmarkDeriveTag measures the cost of one PBKDF2-HMAC-SHA256 derivation at the production
// iteration count. Secret and salt are fixed outside the loop so the benchmark measures only the
// derivation itself.
func BenchmarkDeriveTag(b *testing.B) {
	secret := "s3cr3t"
	salt := bytes.Repeat([]byte{0x55}, saltLen)
	for b.Loop() {
		deriveTag(secret, salt)
	}
}

// ---- Schema pass tests ----

// whPasswordModel is the base fixture for hash-mode schema-pass tests: a single hashable string
// source, with no manually-declared trigger, since hash mode never needs one.
type whPasswordModel struct {
	Password string `mapstructure:"password" desc:"a secret"`
}

type whBoolSourceModel struct {
	Password bool `mapstructure:"password" desc:"a secret"`
}

type whInt64SourceModel struct {
	Password int64 `mapstructure:"password" desc:"a secret"`
}

type whDefaultTagModel struct {
	Password string `mapstructure:"password" default:"x" desc:"a secret"`
}

// whExactCollisionModel already declares an attribute with the exact name a hash sibling for
// "password" would synthesize.
type whExactCollisionModel struct {
	Password              string `mapstructure:"password" desc:"a secret"`
	PasswordWriteOnlyHash string `mapstructure:"password_write_only_hash" desc:"existing"`
}

// whNormalizedCollisionModel declares an attribute that only collides with the synthesized
// sibling name once both are lowercased with hyphens and underscores removed.
type whNormalizedCollisionModel struct {
	Password string `mapstructure:"password" desc:"a secret"`
	Existing string `mapstructure:"passwordwriteonlyhash" desc:"existing"`
}

// whBothModesModel carries one attribute for manual write-only mode and one for hash mode, so a
// single generation can exercise both together without either interfering with the other.
type whBothModesModel struct {
	Secret   string `mapstructure:"secret" desc:"a secret"`
	Password string `mapstructure:"password" desc:"a password"`
}

// whShadowedHashSiblingModel has a model field whose mapstructure key matches exactly what
// WriteOnlyHashAttributeName("password") returns ("password_write_only_hash"), but the field
// uses a channel type so the schema generator drops it. This exercises the
// findModelFieldNotInSchema guard inside validateWriteOnlyHashedKey.
type whShadowedHashSiblingModel struct {
	Password              string      `mapstructure:"password" desc:"a secret"`
	PasswordWriteOnlyHash chan string `mapstructure:"password_write_only_hash" desc:"dropped"`
}

// assertSchemaAttributesUnchanged fails the test unless got has exactly the same attribute names
// as baseline, and every shared attribute is unchanged per assertAttributeUnchanged. It is the
// hash-mode analogue of TestApplyWriteOnlyAttributes_UnmodifiedOnError's per-attribute check, but
// compares the whole attribute set: a rejected hashed declaration must add no synthesized sibling
// and touch no existing attribute.
func assertSchemaAttributesUnchanged(t *testing.T, baseline, got map[string]schema.Attribute) {
	t.Helper()
	if len(baseline) != len(got) {
		t.Fatalf("attribute count changed: baseline=%d got=%d (baseline keys=%v, got keys=%v)",
			len(baseline), len(got), mapKeysForDiagnostics(baseline), mapKeysForDiagnostics(got))
	}
	for name, baseAttr := range baseline {
		gotAttr, ok := got[name]
		if !ok {
			t.Fatalf("attribute %q present in baseline but missing from got", name)
			continue
		}
		assertAttributeUnchanged(t, name, baseAttr, gotAttr)
	}
}

func mapKeysForDiagnostics(m map[string]schema.Attribute) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// TestApplyWriteOnlyHashedAttributes_SiblingShape covers the shape of the synthesized
// "<attr>_write_only_hash" sibling.
func TestApplyWriteOnlyHashedAttributes_SiblingShape(t *testing.T) {
	t.Parallel()

	s, diags := generateResourceSchema(resourceSchemaOptions{
		CreateModel:               &whPasswordModel{},
		WriteOnlyHashedAttributes: []string{"password"},
	})
	requireNoErrors(t, diags)

	sibling := stringAttr(t, s.Attributes, "password_write_only_hash")
	if !sibling.Computed {
		t.Error("expected the sibling to be Computed")
	}
	if !sibling.Sensitive {
		t.Error("expected the sibling to be Sensitive")
	}
	if sibling.Optional {
		t.Error("expected the sibling to NOT be Optional: an Optional sibling would let a practitioner forge a matching tag and permanently suppress rotation")
	}
	if sibling.Required {
		t.Error("expected the sibling to NOT be Required")
	}
	if len(sibling.PlanModifiers) != 1 {
		t.Fatalf("expected exactly one plan modifier, got %d: %+v", len(sibling.PlanModifiers), sibling.PlanModifiers)
	}
	hpm, ok := sibling.PlanModifiers[0].(hashPlanModifier)
	if !ok {
		t.Fatalf("expected the plan modifier to be a hashPlanModifier, got %T", sibling.PlanModifiers[0])
	}
	if hpm.sourcePath != "password" {
		t.Errorf("expected sourcePath %q, got %q", "password", hpm.sourcePath)
	}
	if sibling.Description == "" {
		t.Error("expected a non-empty description")
	}
	if !strings.Contains(sibling.Description, "password") {
		t.Errorf("expected description to mention the source attribute, got %q", sibling.Description)
	}
}

// TestApplyWriteOnlyHashedAttributes_SourceShape covers the shape of the source attribute once
// hash mode has marked it write-only.
func TestApplyWriteOnlyHashedAttributes_SourceShape(t *testing.T) {
	t.Parallel()

	s, diags := generateResourceSchema(resourceSchemaOptions{
		CreateModel:               &whPasswordModel{},
		WriteOnlyHashedAttributes: []string{"password"},
	})
	requireNoErrors(t, diags)

	source := stringAttr(t, s.Attributes, "password")
	if !source.WriteOnly {
		t.Error("expected password to be WriteOnly")
	}
	if source.Computed {
		t.Error("expected password to NOT be Computed")
	}
	if !strings.Contains(source.Description, "password_write_only_hash") {
		t.Errorf("expected description to name the synthesized sibling, got %q", source.Description)
	}

	// The sibling is Computed and never appears in configuration at all, so attaching
	// WriteOnlyTriggerValidator here would warn "Write-Only Value Set Without Trigger" on every
	// single plan where the practitioner sets the password: the validator's whole premise -- a
	// trigger the practitioner can set alongside the write-only value -- does not hold in hash
	// mode, where rotation is detected automatically instead.
	if _, found := findValidatorOfType[WriteOnlyTriggerValidator](source.Validators); found {
		t.Error("expected no WriteOnlyTriggerValidator on the hash-mode source attribute")
	}
}

// TestApplyWriteOnlyHashedAttributes_NoTriggerValidatorOnSibling covers the same absence on the
// synthesized sibling itself, which is not a StringAttribute in write-only position and could never
// carry the validator meaningfully either way.
func TestApplyWriteOnlyHashedAttributes_NoTriggerValidatorOnSibling(t *testing.T) {
	t.Parallel()

	s, diags := generateResourceSchema(resourceSchemaOptions{
		CreateModel:               &whPasswordModel{},
		WriteOnlyHashedAttributes: []string{"password"},
	})
	requireNoErrors(t, diags)

	sibling := stringAttr(t, s.Attributes, "password_write_only_hash")
	if _, found := findValidatorOfType[WriteOnlyTriggerValidator](sibling.Validators); found {
		t.Error("expected no WriteOnlyTriggerValidator on the synthesized sibling")
	}
}

// TestApplyWriteOnlyAttributes_UnaffectedByHashModeRefactor is the regression guard for the
// setWriteOnly/markAttributeWriteOnly refactor that made WriteOnlyTriggerValidator attachment
// conditional on attachTriggerValidator: a resource declaring only WriteOnlyAttributes (manual
// mode, no hashed attributes at all) must still get the validator on its write-only attribute.
func TestApplyWriteOnlyAttributes_UnaffectedByHashModeRefactor(t *testing.T) {
	t.Parallel()

	s, diags := generateResourceSchema(resourceSchemaOptions{
		CreateModel:         &woStringModel{},
		WriteOnlyAttributes: map[string]string{"secret": "rotate"},
	})
	requireNoErrors(t, diags)

	secret := stringAttr(t, s.Attributes, "secret")
	if _, found := findValidatorOfType[WriteOnlyTriggerValidator](secret.Validators); !found {
		t.Error("expected manual write-only mode to still attach WriteOnlyTriggerValidator")
	}
}

// TestApplyWriteOnlyHashedAttributes_BothModesOnDifferentAttributes covers a resource that
// declares both a manual write-only attribute and a hashed one, on different attributes: each mode
// must produce its own shape independent of the other.
func TestApplyWriteOnlyHashedAttributes_BothModesOnDifferentAttributes(t *testing.T) {
	t.Parallel()

	s, diags := generateResourceSchema(resourceSchemaOptions{
		CreateModel:               &whBothModesModel{},
		WriteOnlyAttributes:       map[string]string{"secret": "secret_trigger"},
		WriteOnlyHashedAttributes: []string{"password"},
	})
	requireNoErrors(t, diags)

	secret := stringAttr(t, s.Attributes, "secret")
	if !secret.WriteOnly {
		t.Error("expected secret to be write-only")
	}
	if _, found := findValidatorOfType[WriteOnlyTriggerValidator](secret.Validators); !found {
		t.Error("expected secret (manual mode) to carry WriteOnlyTriggerValidator")
	}
	if _, ok := s.Attributes["secret_trigger"]; !ok {
		t.Error("expected secret_trigger to be synthesized")
	}

	password := stringAttr(t, s.Attributes, "password")
	if !password.WriteOnly {
		t.Error("expected password to be write-only")
	}
	if _, found := findValidatorOfType[WriteOnlyTriggerValidator](password.Validators); found {
		t.Error("expected password (hash mode) to NOT carry WriteOnlyTriggerValidator")
	}
	if _, ok := s.Attributes["password_write_only_hash"]; !ok {
		t.Error("expected password_write_only_hash to be synthesized")
	}
}

// TestApplyWriteOnlyHashedAttributes_Rejections covers every hashed declaration this pass refuses.
func TestApplyWriteOnlyHashedAttributes_Rejections(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		opts  resourceSchemaOptions
		error string
	}{
		{
			name:  "nested_dotted_path",
			opts:  resourceSchemaOptions{CreateModel: &whPasswordModel{}, WriteOnlyHashedAttributes: []string{"metadata.password"}},
			error: "is nested",
		},
		{
			name:  "key_not_in_schema",
			opts:  resourceSchemaOptions{CreateModel: &whPasswordModel{}, WriteOnlyHashedAttributes: []string{"nope"}},
			error: "does not exist in the generated schema",
		},
		{
			name:  "non_string_source_bool",
			opts:  resourceSchemaOptions{CreateModel: &whBoolSourceModel{}, WriteOnlyHashedAttributes: []string{"password"}},
			error: "not a string",
		},
		{
			name:  "non_string_source_int64",
			opts:  resourceSchemaOptions{CreateModel: &whInt64SourceModel{}, WriteOnlyHashedAttributes: []string{"password"}},
			error: "not a string",
		},
		{
			name:  "collision_exact_name",
			opts:  resourceSchemaOptions{CreateModel: &whExactCollisionModel{}, WriteOnlyHashedAttributes: []string{"password"}},
			error: "collides with existing attribute",
		},
		{
			name:  "collision_normalized_name",
			opts:  resourceSchemaOptions{CreateModel: &whNormalizedCollisionModel{}, WriteOnlyHashedAttributes: []string{"password"}},
			error: "collides with existing attribute",
		},
		{
			name:  "duplicate_entry",
			opts:  resourceSchemaOptions{CreateModel: &whPasswordModel{}, WriteOnlyHashedAttributes: []string{"password", "password"}},
			error: "declared 2 times",
		},
		{
			name:  "carries_default",
			opts:  resourceSchemaOptions{CreateModel: &whDefaultTagModel{}, WriteOnlyHashedAttributes: []string{"password"}},
			error: "carries a Default",
		},
		{
			name:  "in_immutable_attributes",
			opts:  resourceSchemaOptions{CreateModel: &whPasswordModel{}, ImmutableAttributes: []string{"password"}, WriteOnlyHashedAttributes: []string{"password"}},
			error: "also listed in ImmutableAttributes",
		},
		{
			// whShadowedHashSiblingModel has a chan field named "password_write_only_hash"; the
			// schema generator drops channel fields, so the synthesized sibling name collides with
			// an SDK model field that is absent from the schema.  This exercises the
			// findModelFieldNotInSchema guard path inside validateWriteOnlyHashedKey.
			name:  "hash_sibling_shadows_dropped_model_field",
			opts:  resourceSchemaOptions{CreateModel: &whShadowedHashSiblingModel{}, WriteOnlyHashedAttributes: []string{"password"}},
			error: "absent from the generated schema",
		},
	}

	for _, tt := range tests {
		t.Run("error_"+tt.name, func(t *testing.T) {
			t.Parallel()
			_, diags := generateResourceSchema(tt.opts)
			requireSingleError(t, diags, tt.error)
		})
	}
}

// TestApplyWriteOnlyHashedAttributes_UnmodifiedOnError checks that a rejected hashed declaration
// leaves the schema exactly as it would have been without any hashed entry, for the cases where
// WriteOnlyAttributes is empty so there is no preceding manual-pass mutation to interact with. See
// TestApplyWriteOnlyHashedAttributes_MutualExclusivity for the one rejection case where that does
// not hold and why.
func TestApplyWriteOnlyHashedAttributes_UnmodifiedOnError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		base resourceSchemaOptions
		attr string
	}{
		{
			name: "non_string_source",
			base: resourceSchemaOptions{CreateModel: &whBoolSourceModel{}},
			attr: "password",
		},
		{
			name: "carries_default",
			base: resourceSchemaOptions{CreateModel: &whDefaultTagModel{}},
			attr: "password",
		},
		{
			name: "immutable_conflict",
			base: resourceSchemaOptions{CreateModel: &whPasswordModel{}, ImmutableAttributes: []string{"password"}},
			attr: "password",
		},
		{
			name: "collision_exact_name",
			base: resourceSchemaOptions{CreateModel: &whExactCollisionModel{}},
			attr: "password",
		},
	}

	for _, tt := range tests {
		t.Run("error_"+tt.name+"_leaves_schema_untouched", func(t *testing.T) {
			t.Parallel()
			baseline, baseDiags := generateResourceSchema(tt.base)
			requireNoErrors(t, baseDiags)

			withHashed := tt.base
			withHashed.WriteOnlyHashedAttributes = []string{tt.attr}
			got, diags := generateResourceSchema(withHashed)
			if !diags.HasError() {
				t.Fatal("expected an error diagnostic")
			}
			assertSchemaAttributesUnchanged(t, baseline.Attributes, got.Attributes)
		})
	}
}

// TestApplyWriteOnlyHashedAttributes_MutualExclusivity covers the WriteOnlyAttributes /
// WriteOnlyHashedAttributes mutual-exclusion rule for the same attribute. It is split into two
// sub-tests deliberately: a direct call into applyWriteOnlyHashedAttributes, where the documented
// "attrs unmodified on error" guarantee holds, and a call through the full generateResourceSchema
// pipeline, where it does NOT -- see the FINDING recorded in the delivery report.
func TestApplyWriteOnlyHashedAttributes_MutualExclusivity(t *testing.T) {
	t.Parallel()

	t.Run("error_direct_call_leaves_attrs_untouched", func(t *testing.T) {
		t.Parallel()
		attrs := map[string]schema.Attribute{
			"password": schema.StringAttribute{Optional: true, Description: "a secret"},
		}
		baseline := map[string]schema.Attribute{
			"password": attrs["password"],
		}

		diags := applyWriteOnlyHashedAttributes(attrs, resourceSchemaOptions{
			WriteOnlyAttributes:       map[string]string{"password": "rotate"},
			WriteOnlyHashedAttributes: []string{"password"},
		})
		requireSingleError(t, diags, "is also a key in WriteOnlyAttributes")
		assertSchemaAttributesUnchanged(t, baseline, attrs)
	})

	// This sub-test documents, rather than merely asserts, the observed behavior: see the FINDING
	// in the delivery report about applyWriteOnlyAttributes (the manual pass, which
	// generateResourceSchema always runs first) unconditionally committing its own mutation --
	// marking "password" write-only with a synthesized "rotate" trigger -- before
	// applyWriteOnlyHashedAttributes ever gets a chance to reject the same key for being declared
	// in both maps. The combined call still reports an error, but the schema it returns is a
	// partially-mutated one, not the unmodified schema the sibling doc comment on
	// applyWriteOnlyHashedAttributes promises when read in isolation.
	t.Run("error_through_full_pipeline_the_manual_pass_has_already_mutated_the_source", func(t *testing.T) {
		t.Parallel()
		got, diags := generateResourceSchema(resourceSchemaOptions{
			CreateModel:               &whBothModesModel{},
			WriteOnlyAttributes:       map[string]string{"password": "rotate"},
			WriteOnlyHashedAttributes: []string{"password"},
		})
		if !diags.HasError() {
			t.Fatal("expected an error diagnostic")
		}
		password := stringAttr(t, got.Attributes, "password")
		if !password.WriteOnly {
			t.Errorf("expected to observe the manual pass's mutation surviving the overall error (password.WriteOnly=true); if this now fails, the finding in the report may have been fixed and should be re-verified, got %+v", password)
		}
	})
}

// TestApplyWriteOnlyHashedAttributes_SchemaDeterminism calls generateResourceSchema twice with
// identical options and asserts the two results are identical, guarding against a salt (or any
// other non-deterministic value) leaking into schema generation itself; only PlanModifyString may
// ever draw a salt.
//
// Approach: field-by-field comparison (via assertAttributeUnchanged, this package's own established
// technique) rather than reflect.DeepEqual on the whole schema.Schema. hashPlanModifier itself has
// no func-valued fields and so IS DeepEqual-safe, but other attribute kinds already generated by
// this package embed func-valued plan modifiers or defaults for which Go's DeepEqual considers any
// two non-nil funcs unequal even when behaviorally identical -- exactly the reason
// assertAttributeUnchanged exists in schemas_writeonly_test.go. Reusing it here keeps this test
// correct even if a future change adds a func-valued field to an attribute in this fixture.
func TestApplyWriteOnlyHashedAttributes_SchemaDeterminism(t *testing.T) {
	t.Parallel()

	opts := resourceSchemaOptions{
		CreateModel:               &whBothModesModel{},
		WriteOnlyAttributes:       map[string]string{"secret": "secret_trigger"},
		WriteOnlyHashedAttributes: []string{"password"},
	}

	first, diags := generateResourceSchema(opts)
	requireNoErrors(t, diags)
	second, diags2 := generateResourceSchema(opts)
	requireNoErrors(t, diags2)

	if len(first.Attributes) != len(second.Attributes) {
		t.Fatalf("attribute count differs between runs: %d vs %d", len(first.Attributes), len(second.Attributes))
	}
	for name, a := range first.Attributes {
		b, ok := second.Attributes[name]
		if !ok {
			t.Fatalf("attribute %q present in first run but not second", name)
		}
		assertAttributeUnchanged(t, name, a, b)
	}

	firstSibling := stringAttr(t, first.Attributes, "password_write_only_hash")
	secondSibling := stringAttr(t, second.Attributes, "password_write_only_hash")
	firstModifier, ok := firstSibling.PlanModifiers[0].(hashPlanModifier)
	if !ok {
		t.Fatalf("expected a hashPlanModifier, got %T", firstSibling.PlanModifiers[0])
	}
	secondModifier, ok := secondSibling.PlanModifiers[0].(hashPlanModifier)
	if !ok {
		t.Fatalf("expected a hashPlanModifier, got %T", secondSibling.PlanModifiers[0])
	}
	if firstModifier.sourcePath != secondModifier.sourcePath {
		t.Errorf("sourcePath differs between runs: %q vs %q", firstModifier.sourcePath, secondModifier.sourcePath)
	}
}

// ---- Plan modifier tests ----

// hashModifierConfigRaw builds a single-attribute "password" object Raw value, mirroring the
// tftypes object-construction idiom used by internal/provider's write-only tests.
func hashModifierConfigRaw(password tftypes.Value) tftypes.Value {
	return tftypes.NewValue(
		tftypes.Object{AttributeTypes: map[string]tftypes.Type{"password": tftypes.String}},
		map[string]tftypes.Value{"password": password},
	)
}

func hashKnownString(s string) tftypes.Value { return tftypes.NewValue(tftypes.String, s) }
func hashNullString() tftypes.Value          { return tftypes.NewValue(tftypes.String, nil) }
func hashUnknownString() tftypes.Value       { return tftypes.NewValue(tftypes.String, tftypes.UnknownValue) }

// hashModifierObjectRaw builds a two-attribute "password"/"password_write_only_hash" object Raw
// value. Unlike hashModifierConfigRaw (single-attribute, used only for the plan modifier's
// config-only view), ResolveWriteOnlyHashTags reads the sibling out of plan (and optionally prior
// state) as well as the source out of config, so its tests need both attributes present on
// whichever tfsdk.Plan/State/Config value they build.
func hashModifierObjectRaw(password, hash tftypes.Value) tftypes.Value {
	return tftypes.NewValue(
		tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"password":                 tftypes.String,
			"password_write_only_hash": tftypes.String,
		}},
		map[string]tftypes.Value{
			"password":                 password,
			"password_write_only_hash": hash,
		},
	)
}

// TestHashPlanModifier_SourceNullInConfig covers the source-null branch: no secret configured, so
// the plan simply carries forward whatever tag (or null) is already in state.
func TestHashPlanModifier_SourceNullInConfig(t *testing.T) {
	t.Parallel()

	priorState := types.StringValue("v1:AAAA:BBBB")
	req := planmodifier.StringRequest{
		Path:       path.Root("password_write_only_hash"),
		StateValue: priorState,
		Config:     tfsdk.Config{Raw: hashModifierConfigRaw(hashNullString())},
	}
	resp := &planmodifier.StringResponse{PlanValue: types.StringUnknown()}

	newHashPlanModifier("password").PlanModifyString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no error diagnostics, got %v", resp.Diagnostics.Errors())
	}
	if len(resp.Diagnostics) != 0 {
		t.Errorf("expected zero diagnostics, got %v", resp.Diagnostics)
	}
	if !resp.PlanValue.Equal(priorState) {
		t.Errorf("expected the plan to equal the prior state exactly, got %v want %v", resp.PlanValue, priorState)
	}
}

// TestHashPlanModifier_UnknownSourceInConfig covers the not-yet-known branch: the secret's value
// will not be known until apply, so the plan for the hash must itself be unknown, with exactly one
// warning and zero errors.
func TestHashPlanModifier_UnknownSourceInConfig(t *testing.T) {
	t.Parallel()

	req := planmodifier.StringRequest{
		Path:       path.Root("password_write_only_hash"),
		StateValue: types.StringValue("v1:AAAA:BBBB"),
		Config:     tfsdk.Config{Raw: hashModifierConfigRaw(hashUnknownString())},
	}
	resp := &planmodifier.StringResponse{PlanValue: types.StringUnknown()}

	newHashPlanModifier("password").PlanModifyString(context.Background(), req, resp)

	if !resp.PlanValue.IsUnknown() {
		t.Errorf("expected the planned value to be unknown, got %v", resp.PlanValue)
	}
	if errs := resp.Diagnostics.Errors(); len(errs) != 0 {
		t.Errorf("expected zero errors, got %v", errs)
	}
	if warnings := resp.Diagnostics.Warnings(); len(warnings) != 1 {
		t.Fatalf("expected exactly one warning, got %d: %v", len(warnings), warnings)
	}
}

// TestHashPlanModifier_ReusesPriorSalt covers the core no-diff property: an unchanged secret must
// plan to a byte-identical tag, because the prior state's salt is parsed and reused rather than a
// fresh one drawn.
func TestHashPlanModifier_ReusesPriorSalt(t *testing.T) {
	t.Parallel()

	salt, err := newSalt(bytes.NewReader(bytes.Repeat([]byte{0x07}, saltLen)))
	if err != nil {
		t.Fatalf("failed to build a fixed salt fixture: %v", err)
	}
	priorTag := deriveTag("s3cr3t", salt)

	req := planmodifier.StringRequest{
		Path:       path.Root("password_write_only_hash"),
		StateValue: types.StringValue(priorTag),
		Config:     tfsdk.Config{Raw: hashModifierConfigRaw(hashKnownString("s3cr3t"))},
	}
	resp := &planmodifier.StringResponse{PlanValue: types.StringUnknown()}

	newHashPlanModifier("password").PlanModifyString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error diagnostics: %v", resp.Diagnostics.Errors())
	}
	if resp.PlanValue.ValueString() != priorTag {
		t.Errorf("expected the planned tag to be byte-identical to the prior state tag for an unchanged secret, got %q want %q", resp.PlanValue.ValueString(), priorTag)
	}
	_, plannedSalt, _, ok := parseTag(resp.PlanValue.ValueString())
	if !ok {
		t.Fatalf("expected a valid tag, got %q", resp.PlanValue.ValueString())
	}
	if !bytes.Equal(plannedSalt, salt) {
		t.Errorf("expected the planned salt to be reused from prior state, got %x want %x", plannedSalt, salt)
	}
}

// TestHashPlanModifier_UnknownWhenNoPriorState covers the first-apply path: with no prior state
// value to parse a salt from, a concrete tag would require fresh randomness, which would make
// PlanModifyString non-deterministic across the two PlanResourceChange calls Terraform issues
// during a single apply (Terraform rejects a planned known value that changes between the two).
// The plan value must therefore be unknown ("known after apply"), with zero diagnostics -- no
// warning either, since this is the routine first-create path, not a self-heal from a bad prior
// tag. The concrete tag is derived later, on the apply side, by ResolveWriteOnlyHashTags.
func TestHashPlanModifier_UnknownWhenNoPriorState(t *testing.T) {
	t.Parallel()

	req := planmodifier.StringRequest{
		Path:       path.Root("password_write_only_hash"),
		StateValue: types.StringNull(),
		Config:     tfsdk.Config{Raw: hashModifierConfigRaw(hashKnownString("s3cr3t"))},
	}
	resp := &planmodifier.StringResponse{PlanValue: types.StringUnknown()}

	newHashPlanModifier("password").PlanModifyString(context.Background(), req, resp)

	if !resp.PlanValue.IsUnknown() {
		t.Errorf("expected the planned value to be unknown on first create, got %v", resp.PlanValue)
	}
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error diagnostics: %v", resp.Diagnostics.Errors())
	}
	if len(resp.Diagnostics) != 0 {
		t.Errorf("expected zero diagnostics on first create (no warning either), got %v", resp.Diagnostics)
	}
}

// TestHashPlanModifier_UnknownWhenPriorStateUnparseable covers the format-migration path: a prior
// state value that parseTag rejects (garbage, or a future format version) must not fail the plan,
// and -- like the no-prior-state case -- cannot plan a concrete tag without drawing fresh
// randomness, so it plans unknown too. Unlike first create, this path warns: an existing resource
// whose stored tag suddenly becomes unparseable is a self-heal, worth surfacing to the
// practitioner, whereas first create is the routine, expected case.
func TestHashPlanModifier_UnknownWhenPriorStateUnparseable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		prior string
	}{
		{name: "garbage", prior: "garbage"},
		{name: "future_format_version", prior: "v2:AAAA:BBBB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := planmodifier.StringRequest{
				Path:       path.Root("password_write_only_hash"),
				StateValue: types.StringValue(tt.prior),
				Config:     tfsdk.Config{Raw: hashModifierConfigRaw(hashKnownString("s3cr3t"))},
			}
			resp := &planmodifier.StringResponse{PlanValue: types.StringUnknown()}

			newHashPlanModifier("password").PlanModifyString(context.Background(), req, resp)

			if !resp.PlanValue.IsUnknown() {
				t.Errorf("expected the planned value to be unknown when the prior tag is unparseable, got %v", resp.PlanValue)
			}
			if errs := resp.Diagnostics.Errors(); len(errs) != 0 {
				t.Fatalf("expected zero error diagnostics on the format-migration path, got %v", errs)
			}
			if warnings := resp.Diagnostics.Warnings(); len(warnings) != 1 {
				t.Fatalf("expected exactly one warning diagnostic, got %d: %v", len(warnings), warnings)
			}
		})
	}
}

// TestHashPlanModifier_SecretChanged covers rotation detection: a different secret in config,
// against the same prior state tag, must plan to a different tag (so the trigger fires), while
// still reusing the prior salt.
func TestHashPlanModifier_SecretChanged(t *testing.T) {
	t.Parallel()

	salt, err := newSalt(bytes.NewReader(bytes.Repeat([]byte{0x09}, saltLen)))
	if err != nil {
		t.Fatalf("failed to build a fixed salt fixture: %v", err)
	}
	priorTag := deriveTag("old-secret", salt)

	req := planmodifier.StringRequest{
		Path:       path.Root("password_write_only_hash"),
		StateValue: types.StringValue(priorTag),
		Config:     tfsdk.Config{Raw: hashModifierConfigRaw(hashKnownString("new-secret"))},
	}
	resp := &planmodifier.StringResponse{PlanValue: types.StringUnknown()}

	newHashPlanModifier("password").PlanModifyString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error diagnostics: %v", resp.Diagnostics.Errors())
	}
	if resp.PlanValue.ValueString() == priorTag {
		t.Error("expected a changed secret to plan to a different tag so the trigger fires")
	}
	_, plannedSalt, _, ok := parseTag(resp.PlanValue.ValueString())
	if !ok {
		t.Fatalf("expected a valid tag, got %q", resp.PlanValue.ValueString())
	}
	if !bytes.Equal(plannedSalt, salt) {
		t.Errorf("expected the salt to still be reused from prior state even though the secret changed, got %x want %x", plannedSalt, salt)
	}
}

// TestHashPlanModifier_PerpetualDiffRegression is the most important test in this file: with an
// unchanged secret, two independent plans of the same prior state must produce byte-identical
// results, both to each other and to the prior state value.
//
// A fresh salt drawn on every call -- rather than the prior tag's salt being parsed and reused --
// would fail this assertion. In production that would show up as a non-empty diff on every single
// `terraform plan` of an otherwise-untouched resource, and every `terraform apply` would re-send
// and rotate the credential in the backend, defeating the entire purpose of hash-based rotation
// detection.
func TestHashPlanModifier_PerpetualDiffRegression(t *testing.T) {
	t.Parallel()

	salt, err := newSalt(bytes.NewReader(bytes.Repeat([]byte{0x0a}, saltLen)))
	if err != nil {
		t.Fatalf("failed to build a fixed salt fixture: %v", err)
	}
	priorTag := deriveTag("stable-secret", salt)

	plan := func() types.String {
		req := planmodifier.StringRequest{
			Path:       path.Root("password_write_only_hash"),
			StateValue: types.StringValue(priorTag),
			Config:     tfsdk.Config{Raw: hashModifierConfigRaw(hashKnownString("stable-secret"))},
		}
		resp := &planmodifier.StringResponse{PlanValue: types.StringUnknown()}
		newHashPlanModifier("password").PlanModifyString(context.Background(), req, resp)
		if resp.Diagnostics.HasError() {
			t.Fatalf("unexpected error diagnostics: %v", resp.Diagnostics.Errors())
		}
		return resp.PlanValue
	}

	first := plan()
	second := plan()

	if !first.Equal(second) {
		t.Errorf("expected two plans of an unchanged secret to be byte-identical, got %q vs %q", first.ValueString(), second.ValueString())
	}
	if !first.Equal(types.StringValue(priorTag)) {
		t.Errorf("expected the plan to be byte-identical to the prior state tag, got %q want %q", first.ValueString(), priorTag)
	}
}

// TestHashPlanModifier_ConfigRawNull covers a config with an entirely null Raw value (the zero
// value of tfsdk.Config), which must be a safe no-op rather than a panic.
func TestHashPlanModifier_ConfigRawNull(t *testing.T) {
	t.Parallel()

	req := planmodifier.StringRequest{
		Path:       path.Root("password_write_only_hash"),
		StateValue: types.StringValue("v1:AAAA:BBBB"),
		Config:     tfsdk.Config{},
	}
	resp := &planmodifier.StringResponse{PlanValue: types.StringUnknown()}

	newHashPlanModifier("password").PlanModifyString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("expected no diagnostics, got %v", resp.Diagnostics.Errors())
	}
}

// ---- ResolveWriteOnlyHashTags tests ----

// TestResolveWriteOnlyHashTags_PlanApplyRoundTrip is the core "plan/apply split works end to end"
// test for the create-time fix: hashPlanModifier plans the sibling unknown on first create (no
// reusable prior salt), and ResolveWriteOnlyHashTags must derive a concrete, parseable tag for it
// from the same secret. Feeding that resolved tag back into PlanModifyString as the prior state
// value, with the same secret, must then plan a byte-identical tag -- the salt embedded in the
// apply-time-resolved tag has to be reusable on the very next plan, exactly like a tag drawn by the
// old plan-time-only code path was.
func TestResolveWriteOnlyHashTags_PlanApplyRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	plan := &tfsdk.Plan{Raw: hashModifierObjectRaw(hashKnownString("s3cr3t"), hashUnknownString())}
	config := &tfsdk.Config{Raw: hashModifierConfigRaw(hashKnownString("s3cr3t"))}

	result, diags := ResolveWriteOnlyHashTags(ctx, []string{"password"}, plan, nil, config)
	if diags.HasError() {
		t.Fatalf("unexpected error diagnostics: %v", diags.Errors())
	}
	resolved, ok := result["password_write_only_hash"]
	if !ok {
		t.Fatalf("expected %q in the result map, got %v", "password_write_only_hash", result)
	}
	if resolved.IsNull() || resolved.IsUnknown() {
		t.Fatalf("expected a concrete resolved tag, got %v", resolved)
	}
	resolvedTag := resolved.ValueString()
	if _, _, _, ok := parseTag(resolvedTag); !ok {
		t.Fatalf("expected a valid v1 tag, got %q", resolvedTag)
	}

	// Feed the resolved tag back as the prior state value: the next plan must reuse its salt and
	// produce the exact same tag, so there is no diff on the apply that follows this one.
	req := planmodifier.StringRequest{
		Path:       path.Root("password_write_only_hash"),
		StateValue: types.StringValue(resolvedTag),
		Config:     tfsdk.Config{Raw: hashModifierConfigRaw(hashKnownString("s3cr3t"))},
	}
	resp := &planmodifier.StringResponse{PlanValue: types.StringUnknown()}
	newHashPlanModifier("password").PlanModifyString(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error diagnostics on the follow-up plan: %v", resp.Diagnostics.Errors())
	}
	if resp.PlanValue.IsNull() || resp.PlanValue.IsUnknown() {
		t.Fatalf("expected a concrete planned value on the follow-up plan, got %v", resp.PlanValue)
	}
	if resp.PlanValue.ValueString() != resolvedTag {
		t.Errorf("expected the follow-up plan to be byte-identical to the apply-resolved tag, got %q want %q", resp.PlanValue.ValueString(), resolvedTag)
	}
}

// TestResolveWriteOnlyHashTags_SkipsKnownPlannedValue covers the salt-reuse path: when
// hashPlanModifier already committed a concrete planned value (a reusable prior salt was found at
// plan time), the resolver must never contradict it. The sibling must be entirely absent from the
// result map, not merely left unresolved as null -- the caller (setWriteOnlyHashTagsInState) only
// overwrites keys present in the map, so an absent key is what leaves the already-correct planned
// value untouched.
func TestResolveWriteOnlyHashTags_SkipsKnownPlannedValue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	plan := &tfsdk.Plan{Raw: hashModifierObjectRaw(hashKnownString("s3cr3t"), hashKnownString("v1:AAAA:BBBB"))}
	config := &tfsdk.Config{Raw: hashModifierConfigRaw(hashKnownString("s3cr3t"))}

	result, diags := ResolveWriteOnlyHashTags(ctx, []string{"password"}, plan, nil, config)
	if diags.HasError() {
		t.Fatalf("unexpected error diagnostics: %v", diags.Errors())
	}
	if _, ok := result["password_write_only_hash"]; ok {
		t.Errorf("expected %q to be absent from the result map when already planned to a known value, got %v", "password_write_only_hash", result)
	}
}

// TestResolveWriteOnlyHashTags_ReusesPriorStateSalt covers the priorState salt-reuse branch inside
// ResolveWriteOnlyHashTags itself, distinct from hashPlanModifier's own plan-time salt reuse
// (TestHashPlanModifier_ReusesPriorSalt): when a prior state value for the sibling exists and
// parses, a second apply-time resolution must reuse its salt rather than drawing a fresh one.
//
// This is the only place a salt is ever chosen for a secret that is unknown at plan time on EVERY
// plan -- for example one sourced from an upstream computed attribute that never itself settles to
// a stable value. hashPlanModifier can never commit a concrete planned tag for such a secret, so
// it always plans the sibling unknown, and ResolveWriteOnlyHashTags runs on every single apply.
// Without this reuse, each of those applies would draw a fresh salt and derive a different tag,
// producing a perpetual, spurious diff against the previous state and re-sending the (unchanged)
// credential to the backend forever -- exactly the failure mode hash mode exists to prevent.
func TestResolveWriteOnlyHashTags_ReusesPriorStateSalt(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := &tfsdk.Config{Raw: hashModifierConfigRaw(hashKnownString("s3cr3t"))}

	// First apply: no prior state at all (first create), so a fresh salt must be drawn.
	firstPlan := &tfsdk.Plan{Raw: hashModifierObjectRaw(hashKnownString("s3cr3t"), hashUnknownString())}
	firstResult, diags := ResolveWriteOnlyHashTags(ctx, []string{"password"}, firstPlan, nil, config)
	if diags.HasError() {
		t.Fatalf("unexpected error diagnostics on first resolution: %v", diags.Errors())
	}
	firstTag, ok := firstResult["password_write_only_hash"]
	if !ok || firstTag.IsNull() || firstTag.IsUnknown() {
		t.Fatalf("expected a concrete resolved tag on first resolution, got %v", firstResult)
	}

	// Second apply: the secret is STILL unknown at plan time (e.g. sourced from an upstream
	// computed attribute that never settles), so the sibling plans unknown again and the resolver
	// runs a second time -- but now with the first apply's tag available as prior state.
	secondPlan := &tfsdk.Plan{Raw: hashModifierObjectRaw(hashKnownString("s3cr3t"), hashUnknownString())}
	priorState := &tfsdk.State{Raw: hashModifierObjectRaw(hashNullString(), hashKnownString(firstTag.ValueString()))}

	secondResult, diags2 := ResolveWriteOnlyHashTags(ctx, []string{"password"}, secondPlan, priorState, config)
	if diags2.HasError() {
		t.Fatalf("unexpected error diagnostics on second resolution: %v", diags2.Errors())
	}
	secondTag, ok := secondResult["password_write_only_hash"]
	if !ok || secondTag.IsNull() || secondTag.IsUnknown() {
		t.Fatalf("expected a concrete resolved tag on second resolution, got %v", secondResult)
	}

	if !firstTag.Equal(secondTag) {
		t.Errorf("expected the second resolution to reuse the first tag's salt and produce a byte-identical tag for an unchanged secret, got %q want %q", secondTag.ValueString(), firstTag.ValueString())
	}
}

// TestApplyWriteOnlyHashedAttributes_SiblingNotUseStateForUnknown guards a hazard specific to this
// fix's design: hashPlanModifier now deliberately plans the sibling unknown whenever there is no
// reusable prior salt (see TestHashPlanModifier_UnknownWhenNoPriorState). If the sibling ever also
// carried stringplanmodifier.UseStateForUnknown, that modifier would pin whatever tag was already
// in prior state instead of leaving the value unknown, which would silently and permanently
// suppress ResolveWriteOnlyHashTags from ever running -- the apply-time tag would never be
// computed, and worse, on first create there is no prior state tag to pin, so the practitioner
// would see a stale or null hash forever. The sibling must carry exactly one plan modifier, and it
// must be hashPlanModifier.
func TestApplyWriteOnlyHashedAttributes_SiblingNotUseStateForUnknown(t *testing.T) {
	t.Parallel()

	s, diags := generateResourceSchema(resourceSchemaOptions{
		CreateModel:               &whPasswordModel{},
		WriteOnlyHashedAttributes: []string{"password"},
	})
	requireNoErrors(t, diags)

	sibling := stringAttr(t, s.Attributes, "password_write_only_hash")
	if !sibling.Computed {
		t.Error("expected the sibling to be Computed")
	}
	if sibling.Optional {
		t.Error("expected the sibling to NOT be Optional")
	}
	if len(sibling.PlanModifiers) != 1 {
		t.Fatalf("expected exactly one plan modifier (hashPlanModifier only, no UseStateForUnknown), got %d: %+v", len(sibling.PlanModifiers), sibling.PlanModifiers)
	}
	if _, ok := sibling.PlanModifiers[0].(hashPlanModifier); !ok {
		t.Errorf("expected the sole plan modifier to be hashPlanModifier, got %T -- if this is stringplanmodifier.UseStateForUnknown, it would pin the old tag and silently suppress rotation forever", sibling.PlanModifiers[0])
	}
}
