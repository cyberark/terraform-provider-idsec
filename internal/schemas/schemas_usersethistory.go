// Copyright CyberArk. 2026
// SPDX-License-Identifier: Apache-2.0

package schemas

import (
	"bytes"
	"context"
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// UserSetAttrsPrivateKey is the resource private-state key under which the set of attribute paths
// the user explicitly set (non-null in configuration) at the last successful create/update is stored.
const UserSetAttrsPrivateKey = "idsec_user_set_attrs"

// UserSetHistorySchemaVersion is the version of the user-set history envelope format persisted in
// private state. Bump it whenever the on-disk shape changes so future reads can migrate older blobs.
const UserSetHistorySchemaVersion = 1

// userSetHistory is the self-describing envelope persisted under UserSetAttrsPrivateKey. Wrapping the
// bare path list with metadata (schema version, save time, provider version) lets future provider
// versions detect and migrate the format, and aids debugging by recording when and by which provider
// build the history was written. Older blobs stored as a bare JSON array are still read (see
// parseUserSetHistory) for backward compatibility.
type userSetHistory struct {
	// SchemaVersion is the envelope format version (UserSetHistorySchemaVersion at write time).
	SchemaVersion int `json:"schema_version"`
	// SavedAt is the RFC3339 UTC timestamp of when the history was written.
	SavedAt string `json:"saved_at"`
	// ProviderVersion is the provider build that wrote the history, empty when unknown (e.g. dev builds).
	ProviderVersion string `json:"provider_version,omitempty"`
	// Paths is the sorted, de-duplicated list of user-set attribute paths.
	Paths []string `json:"paths"`
}

// nowFunc returns the current time; it is a package variable so tests can inject a deterministic
// timestamp into the persisted history envelope.
var nowFunc = time.Now

// privateStateReader is the minimal read surface of the framework's resource private-state provider
// data. It is satisfied by *privatestate.ProviderData (an internal framework type exposed on plan
// modifier and resource requests), letting this package consume private state without importing the
// internal package.
type privateStateReader interface {
	GetKey(ctx context.Context, key string) ([]byte, diag.Diagnostics)
}

// CollectUserSetPaths returns the sorted, de-duplicated set of dot-notation attribute paths that the
// user explicitly set in configuration (i.e. are non-null and known). Nested objects are descended
// into so that both the object path and its set leaf paths are recorded (e.g. "data" and
// "data.connection_config"). The result is intended to be persisted as user-set history.
func CollectUserSetPaths(ctx context.Context, config *tfsdk.Config) ([]string, error) {
	if config == nil {
		return nil, nil
	}
	var configObj types.Object
	if diags := config.Get(ctx, &configObj); diags.HasError() {
		return nil, diagsToError(diags)
	}
	if configObj.IsNull() || configObj.IsUnknown() {
		return nil, nil
	}
	paths := map[string]bool{}
	collectUserSetPaths(configObj.Attributes(), "", paths)
	out := make([]string, 0, len(paths))
	for p := range paths {
		out = append(out, p)
	}
	sort.Strings(out)
	return out, nil
}

// collectUserSetPaths recursively records every non-null, known attribute path into paths.
func collectUserSetPaths(attrs map[string]attr.Value, prefix string, paths map[string]bool) {
	for key, val := range attrs {
		if val == nil || val.IsNull() || val.IsUnknown() {
			continue
		}
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}
		paths[path] = true
		if nestedObj, ok := val.(types.Object); ok && !nestedObj.IsNull() && !nestedObj.IsUnknown() {
			collectUserSetPaths(nestedObj.Attributes(), path, paths)
		}
	}
}

// MarshalUserSetHistory encodes the user-set attribute paths together with metadata (schema version,
// save timestamp, and the provider version that wrote them) as a JSON envelope for private state.
//
// Parameters:
//   - paths: the user-set attribute paths to persist; nil/empty is encoded as an empty list so the key
//     is still written and later reads can distinguish "recorded, nothing set" from "never recorded"
//     (private state SetKey treats empty/nil bytes as a delete, but the envelope is never empty).
//   - providerVersion: the provider build identifier, omitted from the JSON when empty (e.g. dev builds).
//
// Returns the encoded JSON bytes, or an error if marshaling fails.
func MarshalUserSetHistory(paths []string, providerVersion string) ([]byte, error) {
	if paths == nil {
		paths = []string{}
	}
	return json.Marshal(userSetHistory{
		SchemaVersion:   UserSetHistorySchemaVersion,
		SavedAt:         nowFunc().UTC().Format(time.RFC3339),
		ProviderVersion: providerVersion,
		Paths:           paths,
	})
}

// ReadUserSetPaths reads and parses the user-set history from private state into a membership set.
// A nil reader, a missing key, or unparseable data yields a nil set, which callers treat as "no
// history available" (gating disabled, the safe default that never spuriously nulls attributes).
// Both the current envelope format and the legacy bare JSON array are accepted.
func ReadUserSetPaths(ctx context.Context, reader privateStateReader) map[string]bool {
	if reader == nil {
		return nil
	}
	raw, diags := reader.GetKey(ctx, UserSetAttrsPrivateKey)
	if diags.HasError() || len(raw) == 0 {
		return nil
	}
	paths, ok := parseUserSetHistory(raw)
	if !ok {
		return nil
	}
	set := make(map[string]bool, len(paths))
	for _, p := range paths {
		set[p] = true
	}
	return set
}

// parseUserSetHistory decodes the persisted user-set history, accepting both the current envelope
// object ({"schema_version":..,"paths":[..]}) and the legacy bare JSON array ([".."]) for backward
// compatibility. It returns the extracted paths and true on success, or nil and false when the data
// is empty or cannot be parsed in either form.
func parseUserSetHistory(raw []byte) ([]string, bool) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return nil, false
	}
	switch trimmed[0] {
	case '{':
		var h userSetHistory
		if err := json.Unmarshal(trimmed, &h); err != nil {
			return nil, false
		}
		return h.Paths, true
	case '[':
		var paths []string
		if err := json.Unmarshal(trimmed, &paths); err != nil {
			return nil, false
		}
		return paths, true
	default:
		return nil, false
	}
}

// normalizeAttrPath converts a framework attribute path string into the schema-relative dot path
// used by the user-set history, dropping list/set/map element indices (e.g. "targets[0].name"
// becomes "targets.name"). This lets a single recorded schema path match every element instance.
func normalizeAttrPath(p string) string {
	if !strings.ContainsAny(p, "[\"") {
		return p
	}
	var b strings.Builder
	depth := 0
	for _, r := range p {
		switch r {
		case '[':
			depth++
		case ']':
			if depth > 0 {
				depth--
			}
		default:
			if depth == 0 {
				b.WriteRune(r)
			}
		}
	}
	// Collapse any duplicate dots that can arise from removed bracket segments.
	out := strings.ReplaceAll(b.String(), "..", ".")
	return strings.Trim(out, ".")
}

// pathInUserSetHistory reports whether the given attribute path was recorded as user-set in history.
// A nil history (no prior record) returns false so removal-to-null stays inert until history exists.
func pathInUserSetHistory(history map[string]bool, attrPath string) bool {
	if history == nil {
		return false
	}
	return history[normalizeAttrPath(attrPath)]
}

// diagsToError flattens diagnostics into a single error for use in non-diagnostic call sites.
func diagsToError(diags diag.Diagnostics) error {
	if !diags.HasError() {
		return nil
	}
	msgs := make([]string, 0, len(diags))
	for _, d := range diags {
		msgs = append(msgs, d.Summary()+": "+d.Detail())
	}
	return &userSetHistoryError{msg: strings.Join(msgs, "; ")}
}

type userSetHistoryError struct{ msg string }

func (e *userSetHistoryError) Error() string { return e.msg }
