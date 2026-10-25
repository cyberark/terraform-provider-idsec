// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

// This file is deliberately package schemas_test (not schemas): it needs to import
// internal/actions to walk the real provider registry, and internal/actions imports
// internal/schemas, so an in-package (package schemas) test file importing internal/actions
// would be a compile-time import cycle. The external test package has no such restriction.
package schemas_test

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/cyberark/terraform-provider-idsec/internal/actions"
	"github.com/cyberark/terraform-provider-idsec/internal/schemas"
	// Blank-imported for its init() side effects only: every internal/tfactions/<svc> package
	// registers its resources/data sources with internal/actions on import, and
	// actions.AllTerraformConfigs() below is empty unless something has triggered that.
	_ "github.com/cyberark/terraform-provider-idsec/internal/tfactions"
)

// snakeCaseTagPattern identifies tags that are already valid, unambiguous snake_case (lowercase
// letters, digits, and underscores only) - the exact set schemas.ResolveFieldName must return
// verbatim rather than re-normalizing.
var snakeCaseTagPattern = regexp.MustCompile(`^[a-z0-9_]+$`)

// rawTagName returns the first of mapstructure/flag/json tag values present on field, mirroring
// resolveFieldName's own lookup order and skipping any tag options (e.g. ",omitempty").
func rawTagName(field reflect.StructField) (string, bool) {
	for _, tagKey := range []string{"mapstructure", "flag", "json"} {
		raw := strings.Split(field.Tag.Get(tagKey), ",")[0]
		if raw != "" && raw != "-" {
			return raw, true
		}
	}
	return "", false
}

// TestAllRegisteredSchemaTagsSurviveSnakeCaseNormalization walks every struct type reachable
// from a registered resource/data source's Schemas map or StateSchema (recursively, through
// nested structs, pointers, slices, and maps) and fails if any field whose tag is already valid
// snake_case does not come back from schemas.ResolveFieldName unchanged - the exact bug class
// that mangled "k8s_namespace" into "k_8_s_namespace" and silently dropped "k8s_mfa_caching"
// input. This exercises the real production normalization (not a reimplementation of it), so it
// keeps failing if resolveFieldName/normalizeFieldName's short-circuit is ever removed or
// narrowed, and catches any newly-added already-snake_case/digit-adjacent tag immediately.
func TestAllRegisteredSchemaTagsSurviveSnakeCaseNormalization(t *testing.T) {
	t.Parallel()

	visited := map[reflect.Type]bool{}
	var mismatches []string

	var walk func(reflect.Type)
	walk = func(typ reflect.Type) {
		for typ.Kind() == reflect.Pointer {
			typ = typ.Elem()
		}
		switch typ.Kind() {
		case reflect.Slice, reflect.Array, reflect.Map:
			walk(typ.Elem())
			return
		}
		if typ.Kind() != reflect.Struct {
			return
		}
		if visited[typ] {
			return
		}
		visited[typ] = true

		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			if field.PkgPath != "" && !field.Anonymous { // unexported, non-embedded
				continue
			}
			if field.Tag.Get("mapstructure") == "-" {
				// Matches resolveFieldsSquashed: these fields never reach resolveFieldName in
				// production (they're filtered out before schema/state conversion), so calling
				// ResolveFieldName on them directly here would exercise a code path real
				// callers never hit.
				continue
			}
			if tag, ok := rawTagName(field); ok && snakeCaseTagPattern.MatchString(tag) {
				if resolved := schemas.ResolveFieldName(field); resolved != tag {
					mismatches = append(mismatches, fmt.Sprintf(
						"%s.%s: tag %q is already snake_case but ResolveFieldName returns %q",
						typ.String(), field.Name, tag, resolved,
					))
				}
			}
			walk(field.Type)
		}
	}

	for _, config := range actions.AllTerraformConfigs() {
		for _, res := range config.Resources {
			if res == nil {
				continue
			}
			// Only the specific operation input schemas this resource actually wires up via
			// ActionsMappings are reachable through Terraform - res.Schemas is the whole
			// service's action-to-schema map (actions.ActionToSchemaMap), which also contains
			// other actions never exposed as a Terraform attribute on this resource.
			for _, actionName := range res.ActionsMappings {
				if prototype, ok := res.Schemas[actionName]; ok && prototype != nil {
					walk(reflect.TypeOf(prototype))
				}
			}
			if res.StateSchema != nil {
				walk(reflect.TypeOf(res.StateSchema))
			}
		}
		for _, ds := range config.DataSources {
			if ds == nil {
				continue
			}
			if prototype, ok := ds.Schemas[ds.DataSourceAction]; ok && prototype != nil {
				walk(reflect.TypeOf(prototype))
			}
			if ds.StateSchema != nil {
				walk(reflect.TypeOf(ds.StateSchema))
			}
		}
	}

	if len(mismatches) > 0 {
		t.Fatalf(
			"found %d schema tag(s) that strcase.ToSnake would mangle even though already snake_case "+
				"(these need to go through resolveFieldName's/normalizeFieldName's short-circuit, which "+
				"this test proves is still working):\n%s",
			len(mismatches), strings.Join(mismatches, "\n"),
		)
	}
}
