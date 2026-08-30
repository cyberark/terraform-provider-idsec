// Copyright CyberArk. 2026
// SPDX-License-Identifier: Apache-2.0

package schemas

import (
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// errWriteOnlySummary is the diagnostic summary used by every error this pass produces.
const errWriteOnlySummary = "Invalid Write-Only Attribute"

const writeOnlyTriggerDescriptionSuffix = "It has no meaning to the API and is stored in state solely to detect that a re-send was requested."

// pathStep is one resolved link in a dotted attribute path.
type pathStep struct {
	name string
	attr schema.Attribute
}

// writeOnlyTriggerPlan is the resolved handling for one key's trigger: either an existing
// attribute name to annotate, or a name to synthesize.
type writeOnlyTriggerPlan struct {
	triggerName    string
	needsSynthesis bool
}

// applyWriteOnlyAttributes validates every entry in opts.WriteOnlyAttributes and, only if
// validation finds no errors, mutates attrs in place to mark each declared key write-only, to
// synthesize any missing trigger attribute, and to note the re-send on an existing trigger's
// description.
//
// generateResourceSchema is its only caller, so the pass is unreachable from any other
// schema-generation call site; see that call site for why that matters.
//
// Every map is iterated in sorted key order, both because attrs is mutated as the pass proceeds
// and so that a resource with several invalid entries reports the same error list on every run.
//
// Returns diagnostics. If any diagnostic is an error, attrs is guaranteed unmodified.
func applyWriteOnlyAttributes(attrs map[string]schema.Attribute, opts resourceSchemaOptions) diag.Diagnostics {
	var diags diag.Diagnostics
	if len(opts.WriteOnlyAttributes) == 0 {
		return diags
	}

	keys := slices.Sorted(maps.Keys(opts.WriteOnlyAttributes))

	for _, key := range keys {
		diags.Append(validateWriteOnlyKey(attrs, key, opts)...)
	}

	// Triggers are checked even for keys that already failed above, so a resource with several
	// bad entries reports all of them in one pass.
	plans := make(map[string]writeOnlyTriggerPlan, len(keys))
	for _, key := range keys {
		plan, triggerDiags := validateWriteOnlyTrigger(attrs, key, opts.WriteOnlyAttributes[key], opts)
		diags.Append(triggerDiags...)
		plans[key] = plan
	}

	if diags.HasError() {
		// Returning before touching attrs matters most for the set-typed case, where a partial
		// rewrite would leave a valid, ordinary *persisted* set attribute instead of an error.
		return diags
	}

	// Grouped by trigger name, not by key, so two keys sharing a trigger produce one attribute
	// rather than a duplicate-key error, and one description naming both keys.
	synthUsers := map[string][]string{}
	existingUsers := map[string][]string{}
	for _, key := range keys {
		users := existingUsers
		if plans[key].needsSynthesis {
			users = synthUsers
		}
		users[plans[key].triggerName] = append(users[plans[key].triggerName], key)
	}
	for _, triggerName := range slices.Sorted(maps.Keys(synthUsers)) {
		attrs[triggerName] = schema.StringAttribute{
			Optional:    true,
			Description: buildSynthesizedTriggerDescription(synthUsers[triggerName]),
		}
	}
	// An existing trigger keeps every other field, but without this its documentation would never
	// mention that changing it re-sends a credential.
	for _, triggerName := range slices.Sorted(maps.Keys(existingUsers)) {
		diags.Append(annotateExistingTrigger(attrs, triggerName, buildExistingTriggerDescriptionSuffix(existingUsers[triggerName]))...)
	}

	for _, key := range keys {
		diags.Append(markWriteOnly(attrs, key, plans[key].triggerName)...)
	}

	return diags
}

// resolveDottedPath descends attrs by splitting dotted on ".", following single, list, and map
// nested containers. It refuses to descend into a SetNestedAttribute, which surfaces as an
// ordinary "not resolvable" result identical to a typo.
//
// Returns the chain of steps from the root to the named attribute, or nil and false if any
// segment does not exist or cannot be traversed.
func resolveDottedPath(attrs map[string]schema.Attribute, dotted string) ([]pathStep, bool) {
	parts := strings.Split(dotted, ".")
	steps := make([]pathStep, 0, len(parts))
	current := attrs
	for i, part := range parts {
		a, ok := current[part]
		if !ok {
			return nil, false
		}
		steps = append(steps, pathStep{name: part, attr: a})
		if i == len(parts)-1 {
			return steps, true
		}
		children, ok := writeOnlyNestedChildren(a)
		if !ok {
			return nil, false
		}
		current = children
	}
	return steps, true
}

// writeOnlyNestedChildren returns the child attributes of a nested container, or nil and false
// for a leaf attribute or a set-based container.
func writeOnlyNestedChildren(a schema.Attribute) (map[string]schema.Attribute, bool) {
	switch t := a.(type) {
	case schema.SingleNestedAttribute:
		return t.Attributes, true
	case schema.ListNestedAttribute:
		return t.NestedObject.Attributes, true
	case schema.MapNestedAttribute:
		return t.NestedObject.Attributes, true
	default:
		return nil, false
	}
}

// hasDefault reports whether a carries a non-nil Default. It inspects the generated attribute
// rather than re-reading struct tags, so it cannot drift from what the generator decided.
func hasDefault(a schema.Attribute) bool {
	switch t := a.(type) {
	case schema.StringAttribute:
		return t.Default != nil
	case schema.BoolAttribute:
		return t.Default != nil
	case schema.Int64Attribute:
		return t.Default != nil
	case schema.ListAttribute:
		return t.Default != nil
	case schema.SetAttribute:
		return t.Default != nil
	default:
		return false
	}
}

// isSetTyped reports whether a is set-based. Neither set type has a WriteOnly field -- their
// IsWriteOnly() methods return a hardcoded false -- so neither can be marked write-only at all.
func isSetTyped(a schema.Attribute) bool {
	switch a.(type) {
	case schema.SetAttribute, schema.SetNestedAttribute:
		return true
	default:
		return false
	}
}

// validateWriteOnlyKey checks that key names an attribute that can legally be made write-only: it
// must exist, no ancestor may be Computed, and it must not be computed-only, set-typed, carrying a
// Default, containing an ineligible descendant, or listed in ImmutableAttributes/ForceNewAttributes.
func validateWriteOnlyKey(attrs map[string]schema.Attribute, key string, opts resourceSchemaOptions) diag.Diagnostics {
	var diags diag.Diagnostics

	steps, ok := resolveDottedPath(attrs, key)
	if !ok {
		diags.AddError(errWriteOnlySummary, fmt.Sprintf(
			"write-only attribute %q does not exist in the generated schema. The key is matched as a "+
				"dotted path from the schema root, never by bare leaf field name, and cannot descend "+
				"into a set-nested container.", key))
		return diags
	}

	for _, ancestor := range steps[:len(steps)-1] {
		if ancestor.attr.IsComputed() {
			diags.AddError(errWriteOnlySummary, fmt.Sprintf(
				"write-only attribute %q cannot be nested under %q because %q is Computed, which the "+
					"framework forbids. This generator makes every optional nested container Computed, so "+
					"restructure the model so %q is Required, or choose a different attribute.",
				key, ancestor.name, ancestor.name, ancestor.name))
			return diags
		}
	}

	leafStep := steps[len(steps)-1]
	leaf := leafStep.attr

	if leaf.IsComputed() && !leaf.IsOptional() && !leaf.IsRequired() {
		diags.AddError(errWriteOnlySummary, fmt.Sprintf(
			"write-only attribute %q is computed-only (server-managed). Only an attribute that is "+
				"already Optional or Required can be made write-only.", key))
		return diags
	}

	if isSetTyped(leaf) {
		diags.AddError(errWriteOnlySummary, fmt.Sprintf(
			"write-only attribute %q is set-typed. SetAttribute and SetNestedAttribute have no "+
				"WriteOnly field to set, so the declaration cannot be honored and the value would "+
				"silently go into state. Restructure it as a list or single nested object.", key))
		return diags
	}

	if hasDefault(leaf) {
		diags.AddError(errWriteOnlySummary, fmt.Sprintf(
			"write-only attribute %q carries a Default, which this generator turns into Computed, and "+
				"WriteOnly cannot be combined with Computed. Remove the `default` struct tag from the "+
				"corresponding SDK model field, or choose a different attribute.", key))
		return diags
	}

	if nested, isNested := writeOnlyNestedChildren(leaf); isNested {
		if d := validateNoIneligibleDescendant(nested, key); d.HasError() {
			return d
		}
	}

	// Both modifiers decide by comparing the planned value against state, and a write-only
	// attribute's planned and prior-state values are permanently null, so neither can observe a
	// change and the intended behavior would be silently lost.
	leafName := leafStep.name
	if slices.Contains(opts.ImmutableAttributes, leafName) {
		diags.AddError(errWriteOnlySummary, fmt.Sprintf(
			"write-only attribute %q is also listed in ImmutableAttributes, so its immutability could "+
				"never be enforced. Remove %q from ImmutableAttributes, or choose a different attribute.",
			key, leafName))
		return diags
	}
	if slices.Contains(opts.ForceNewAttributes, leafName) {
		diags.AddError(errWriteOnlySummary, fmt.Sprintf(
			"write-only attribute %q is also listed in ForceNewAttributes, so replacement could never "+
				"be triggered by it. Remove %q from ForceNewAttributes, or choose a different attribute.",
			key, leafName))
		return diags
	}

	// A Required write-only attribute is deliberately allowed: it is coherent for a credential
	// that must be supplied on every apply.

	return diags
}

// validateNoIneligibleDescendant rejects any descendant that is set-typed or carries a Default,
// since the framework requires every child of a write-only nested attribute to also be
// write-only and neither kind of descendant can ever be.
func validateNoIneligibleDescendant(attrs map[string]schema.Attribute, ancestorKey string) diag.Diagnostics {
	var diags diag.Diagnostics
	for _, name := range slices.Sorted(maps.Keys(attrs)) {
		a := attrs[name]
		path := ancestorKey + "." + name

		if isSetTyped(a) {
			diags.AddError(errWriteOnlySummary, fmt.Sprintf(
				"write-only attribute %q cannot be made write-only because its descendant %q is "+
					"set-typed, and a set-typed attribute can never itself be write-only. Restructure "+
					"%q, or choose a different attribute.", ancestorKey, path, path))
			return diags
		}
		if hasDefault(a) {
			diags.AddError(errWriteOnlySummary, fmt.Sprintf(
				"write-only attribute %q cannot be made write-only because its descendant %q carries a "+
					"Default, which forces it to be Computed, and a Computed attribute can never be "+
					"write-only. Remove the `default` struct tag from the corresponding SDK model field.",
				ancestorKey, path))
			return diags
		}
		if isComputedOnlyAttr(a.IsOptional(), a.IsRequired(), a.IsComputed()) {
			diags.AddError(errWriteOnlySummary, fmt.Sprintf(
				"write-only attribute %q cannot be made write-only because its descendant %q is "+
					"computed-only (server-managed), and a Computed attribute can never be write-only. "+
					"Remove %q from ComputedAttributes, or choose a different attribute.",
				ancestorKey, path, path))
			return diags
		}
		if nested, isNested := writeOnlyNestedChildren(a); isNested {
			if d := validateNoIneligibleDescendant(nested, path); d.HasError() {
				return d
			}
		}
	}
	return diags
}

// validateWriteOnlyTrigger validates the trigger declared for one write-only key and returns the
// plan for the synthesis and marking phases to act on.
func validateWriteOnlyTrigger(attrs map[string]schema.Attribute, key, trigger string, opts resourceSchemaOptions) (writeOnlyTriggerPlan, diag.Diagnostics) {
	var diags diag.Diagnostics

	if trigger == "" {
		diags.AddError(errWriteOnlySummary, fmt.Sprintf(
			"write-only attribute %q declares an empty trigger; a write-only value produces no plan "+
				"difference on its own, so it would never be re-sent after create. Name an existing "+
				"attribute, or a new name to have one synthesized.", key))
		return writeOnlyTriggerPlan{}, diags
	}

	if trigger == key {
		diags.AddError(errWriteOnlySummary, fmt.Sprintf(
			"write-only attribute %q declares itself as its own trigger. It is null in both plan and "+
				"prior state on every call, so the trigger could never fire.", key))
		return writeOnlyTriggerPlan{}, diags
	}

	if steps, ok := resolveDottedPath(attrs, trigger); ok {
		triggerLeaf := steps[len(steps)-1]
		leaf := triggerLeaf.attr

		if isComputedOnlyAttr(leaf.IsOptional(), leaf.IsRequired(), leaf.IsComputed()) {
			diags.AddError(errWriteOnlySummary, fmt.Sprintf(
				"write-only attribute %q designates %q as its trigger, but %q is computed-only "+
					"(server-managed): it changes for reasons unrelated to practitioner intent and is "+
					"stripped from outgoing payloads by ClearComputedAttributes. Designate an attribute "+
					"the practitioner actually sets, or a new name to have one synthesized.",
				key, trigger, trigger))
			return writeOnlyTriggerPlan{}, diags
		}

		if _, isAnotherKey := opts.WriteOnlyAttributes[trigger]; isAnotherKey {
			diags.AddError(errWriteOnlySummary, fmt.Sprintf(
				"write-only attribute %q designates %q as its trigger, but %q is itself declared "+
					"write-only and so is null in both plan and prior state on every call. Designate a "+
					"non-write-only attribute.", key, trigger, trigger))
			return writeOnlyTriggerPlan{}, diags
		}

		if isSetTyped(leaf) {
			diags.AddError(errWriteOnlySummary, fmt.Sprintf(
				"write-only attribute %q designates %q as its trigger, but %q is set-typed. Set element "+
					"ordering churns between plan and state, and a spurious fire re-sends the credential "+
					"and creates a new vault version. Choose a non-set trigger.", key, trigger, trigger))
			return writeOnlyTriggerPlan{}, diags
		}

		if slices.Contains(opts.ImmutableAttributes, triggerLeaf.name) {
			diags.AddError(errWriteOnlySummary, fmt.Sprintf(
				"write-only attribute %q designates %q as its trigger, but %q is also listed in "+
					"ImmutableAttributes, so any plan that changes it is blocked and the write-only value "+
					"could never be rotated. Choose a different trigger, or remove %q from "+
					"ImmutableAttributes.", key, trigger, trigger, trigger))
			return writeOnlyTriggerPlan{}, diags
		}

		// A trigger carrying a `default` tag is fine, even though a Default disqualifies an
		// attribute from BEING write-only: a trigger needs to be persisted and stable, and
		// ApplyRemovedToUnknownModifiers already skips defaulted attributes.
		return writeOnlyTriggerPlan{triggerName: trigger, needsSynthesis: false}, diags
	}

	if strings.Contains(trigger, ".") {
		diags.AddError(errWriteOnlySummary, fmt.Sprintf(
			"write-only attribute %q declares trigger %q, which does not resolve to an existing "+
				"attribute and contains a dot. A trigger name is not an instruction to create "+
				"intermediate containers; name an existing attribute by its full dotted path, or a bare "+
				"top-level name to have one synthesized.", key, trigger))
		return writeOnlyTriggerPlan{}, diags
	}

	if match := normalizedTopLevelCollision(attrs, trigger); match != "" {
		diags.AddError(errWriteOnlySummary, fmt.Sprintf(
			"write-only attribute %q declares trigger %q, which does not exist but matches existing "+
				"top-level attribute %q once lowercased with hyphens and underscores removed. This is "+
				"almost certainly a typo; use %q, or choose a clearly-synthetic name.",
			key, trigger, match, match))
		return writeOnlyTriggerPlan{}, diags
	}

	if modelPath, found := findModelFieldNotInSchema(opts, trigger); found {
		diags.AddError(errWriteOnlySummary, fmt.Sprintf(
			"write-only attribute %q declares trigger %q, which resolves to model field %q but is "+
				"absent from the generated schema. Use the field's actual generated attribute name, its "+
				"full dotted path if it is nested, or a clearly-synthetic name.", key, trigger, modelPath))
		return writeOnlyTriggerPlan{}, diags
	}

	// An ordinary misspelling of an existing name is deliberately not caught, since nothing can
	// tell it apart from a deliberately synthetic name: it synthesizes an attribute nobody sets,
	// so the value is sent on create and never again. Name synthesized triggers conspicuously
	// (secret_rotation_trigger, not secret_typ) so a reviewer can see the difference.
	return writeOnlyTriggerPlan{triggerName: trigger, needsSynthesis: true}, diags
}

// normalizeAttrNameForCollision lowercases s and removes hyphens and underscores.
func normalizeAttrNameForCollision(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "_", "")
	return s
}

// normalizedTopLevelCollision returns the first top-level attribute name in attrs whose
// normalized form matches trigger's, or "". Sorted so the reported match is deterministic.
func normalizedTopLevelCollision(attrs map[string]schema.Attribute, trigger string) string {
	normalizedTrigger := normalizeAttrNameForCollision(trigger)
	for _, name := range slices.Sorted(maps.Keys(attrs)) {
		if normalizeAttrNameForCollision(name) == normalizedTrigger {
			return name
		}
	}
	return ""
}

// findModelFieldNotInSchema looks for a struct field whose resolved schema name equals name, at
// any nesting level of the create, update, or state model. Returns a dotted path for display.
func findModelFieldNotInSchema(opts resourceSchemaOptions, name string) (string, bool) {
	for _, model := range []interface{}{opts.CreateModel, opts.UpdateModel, opts.StateModel} {
		if model == nil {
			continue
		}
		modelType := reflect.TypeOf(model)
		if modelType.Kind() == reflect.Pointer {
			modelType = modelType.Elem()
		}
		if modelType.Kind() != reflect.Struct {
			continue
		}
		if path, ok := findFieldNameInStructType(modelType, name); ok {
			return path, true
		}
	}
	return "", false
}

// findFieldNameInStructType recurses through t's squashed fields and their nested struct, slice,
// and map fields looking for one whose resolved name equals name.
func findFieldNameInStructType(t reflect.Type, name string) (string, bool) {
	for _, field := range resolveFieldsSquashed(t) {
		fieldName := resolveFieldName(field)
		if fieldName == name {
			return fieldName, true
		}

		fieldType := field.Type
		if fieldType.Kind() == reflect.Pointer {
			fieldType = fieldType.Elem()
		}

		switch fieldType.Kind() {
		case reflect.Struct:
			if nestedPath, ok := findFieldNameInStructType(fieldType, name); ok {
				return fieldName + "." + nestedPath, true
			}
		case reflect.Slice, reflect.Array, reflect.Map:
			elemType := fieldType.Elem()
			if elemType.Kind() == reflect.Pointer {
				elemType = elemType.Elem()
			}
			if elemType.Kind() == reflect.Struct {
				if nestedPath, ok := findFieldNameInStructType(elemType, name); ok {
					return fieldName + "." + nestedPath, true
				}
			}
		}
	}
	return "", false
}

// buildSynthesizedTriggerDescription describes a synthesized trigger, naming every write-only key
// that shares it. keys must already be sorted.
func buildSynthesizedTriggerDescription(keys []string) string {
	quoted := make([]string, len(keys))
	for i, k := range keys {
		quoted[i] = fmt.Sprintf("%q", k)
	}
	return fmt.Sprintf("Change this value to re-send %s to the API. %s", strings.Join(quoted, ", "), writeOnlyTriggerDescriptionSuffix)
}

// buildExistingTriggerDescriptionSuffix is appended to the description of an attribute that was
// designated a trigger but already existed. Unlike a synthesized trigger it keeps its own meaning
// to the API, so the text only adds the re-send side effect. keys must already be sorted.
func buildExistingTriggerDescriptionSuffix(keys []string) string {
	quoted := make([]string, len(keys))
	for i, k := range keys {
		quoted[i] = fmt.Sprintf("%q", k)
	}
	return fmt.Sprintf("Changing this value also re-sends %s to the API.", strings.Join(quoted, ", "))
}

// annotateExistingTrigger appends suffix to the description of the attribute at dotted, leaving
// every other field untouched. Any diagnostic indicates a bug, since validateWriteOnlyTrigger
// already resolved this path.
func annotateExistingTrigger(attrs map[string]schema.Attribute, dotted, suffix string) diag.Diagnostics {
	var diags diag.Diagnostics
	parts := strings.Split(dotted, ".")

	container := attrs
	for _, part := range parts[:len(parts)-1] {
		switch t := container[part].(type) {
		case schema.SingleNestedAttribute:
			container = t.Attributes
		case schema.ListNestedAttribute:
			container = t.NestedObject.Attributes
		case schema.MapNestedAttribute:
			container = t.NestedObject.Attributes
		default:
			diags.AddError(errWriteOnlySummary, fmt.Sprintf(
				"internal error: could not descend into %q (type %T) while annotating trigger %q.",
				part, container[part], dotted))
			return diags
		}
	}

	leaf := parts[len(parts)-1]
	annotated, ok := appendAttrDescription(container[leaf], suffix)
	if !ok {
		diags.AddError(errWriteOnlySummary, fmt.Sprintf(
			"internal error: could not append a description to trigger %q (type %T).",
			dotted, container[leaf]))
		return diags
	}
	container[leaf] = annotated
	return diags
}

// appendAttrDescription returns a copy of a with suffix appended to its Description. It works by
// reflection rather than a type switch because a trigger may be any attribute type, including the
// set types that cannot themselves be write-only.
func appendAttrDescription(a schema.Attribute, suffix string) (schema.Attribute, bool) {
	v := reflect.ValueOf(a)
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return a, false
	}

	updated := reflect.New(v.Type())
	updated.Elem().Set(v)

	field := updated.Elem().FieldByName("Description")
	if !field.IsValid() || !field.CanSet() || field.Kind() != reflect.String {
		return a, false
	}
	field.SetString(appendDescription(field.String(), suffix))

	annotated, ok := updated.Elem().Interface().(schema.Attribute)
	return annotated, ok
}

// writeOnlyAttributeDescriptionSuffix names the trigger in the write-only attribute's own
// description, so the rendered documentation carries the semantics without per-resource prose.
func writeOnlyAttributeDescriptionSuffix(triggerName string) string {
	return fmt.Sprintf(
		"Write-only: this value is never stored in Terraform state or plan. It is sent to the API on "+
			"create, and on update only when `%s` changes. Requires Terraform 1.11 or later.",
		triggerName,
	)
}

// markWriteOnly descends attrs by dottedKey and rewrites the attribute found there to be
// write-only. It runs only after every validation has passed, so any diagnostic it produces
// indicates a bug in this pass rather than a schema-declaration error.
func markWriteOnly(attrs map[string]schema.Attribute, dottedKey, triggerName string) diag.Diagnostics {
	return markWriteOnlyAtPath(attrs, strings.Split(dottedKey, "."), triggerName)
}

func markWriteOnlyAtPath(attrs map[string]schema.Attribute, parts []string, triggerName string) diag.Diagnostics {
	var diags diag.Diagnostics
	name := parts[0]
	a := attrs[name]

	if len(parts) > 1 {
		switch t := a.(type) {
		case schema.SingleNestedAttribute:
			diags.Append(markWriteOnlyAtPath(t.Attributes, parts[1:], triggerName)...)
			attrs[name] = t
		case schema.ListNestedAttribute:
			diags.Append(markWriteOnlyAtPath(t.NestedObject.Attributes, parts[1:], triggerName)...)
			attrs[name] = t
		case schema.MapNestedAttribute:
			diags.Append(markWriteOnlyAtPath(t.NestedObject.Attributes, parts[1:], triggerName)...)
			attrs[name] = t
		default:
			diags.AddError(errWriteOnlySummary, fmt.Sprintf(
				"internal error: could not descend into %q (type %T) while marking a write-only path; "+
					"validateWriteOnlyKey should have rejected this schema before mutation began.", name, a))
		}
		return diags
	}

	marked, markDiags := markAttributeWriteOnly(a, triggerName)
	diags.Append(markDiags...)
	attrs[name] = marked
	return diags
}

// markAttributeWriteOnly rewrites the attribute named by a write-only key, extending its
// description to name the trigger.
func markAttributeWriteOnly(a schema.Attribute, triggerName string) (schema.Attribute, diag.Diagnostics) {
	var diags diag.Diagnostics
	marked, ok := setWriteOnly(a, writeOnlyAttributeDescriptionSuffix(triggerName), triggerName)
	if !ok {
		diags.AddError(errWriteOnlySummary, fmt.Sprintf(
			"internal error: cannot mark attribute type %T write-only; validateWriteOnlyKey should "+
				"have rejected this schema before mutation began.", a))
		return a, diags
	}
	return marked, diags
}

// markDescendantsWriteOnly marks every attribute in attrs write-only, passing no description
// suffix: only the container named by the write-only key carries the trigger reference. A type
// that cannot be marked is left alone, which validateNoIneligibleDescendant has already ruled out.
func markDescendantsWriteOnly(attrs map[string]schema.Attribute) {
	for name, a := range attrs {
		if marked, ok := setWriteOnly(a, "", ""); ok {
			attrs[name] = marked
		}
	}
}

// setWriteOnly returns a with WriteOnly=true, Computed=false and plan modifiers cleared,
// recursing into the children of a nested container. A non-empty suffix is appended to the
// description and, for a string, attaches WriteOnlyTriggerValidator (the only attribute type that
// can carry it). Reports false for a type with no WriteOnly field, which includes both set types.
func setWriteOnly(a schema.Attribute, suffix, triggerName string) (schema.Attribute, bool) {
	switch t := a.(type) {
	case schema.StringAttribute:
		t.Computed, t.WriteOnly, t.PlanModifiers = false, true, nil
		if suffix != "" {
			t.Description = appendDescription(t.Description, suffix)
			t.Validators = append(slices.Clone(t.Validators), WriteOnlyTriggerValidator{TriggerPath: triggerName})
		}
		return t, true
	case schema.BoolAttribute:
		t.Computed, t.WriteOnly, t.PlanModifiers = false, true, nil
		t.Description = appendDescription(t.Description, suffix)
		return t, true
	case schema.Int64Attribute:
		t.Computed, t.WriteOnly, t.PlanModifiers = false, true, nil
		t.Description = appendDescription(t.Description, suffix)
		return t, true
	case schema.ListAttribute:
		t.Computed, t.WriteOnly, t.PlanModifiers = false, true, nil
		t.Description = appendDescription(t.Description, suffix)
		return t, true
	case schema.MapAttribute:
		t.Computed, t.WriteOnly, t.PlanModifiers = false, true, nil
		t.Description = appendDescription(t.Description, suffix)
		return t, true
	case schema.DynamicAttribute:
		t.Computed, t.WriteOnly, t.PlanModifiers = false, true, nil
		t.Description = appendDescription(t.Description, suffix)
		return t, true
	case schema.ListNestedAttribute:
		t.Computed, t.WriteOnly, t.PlanModifiers = false, true, nil
		t.Description = appendDescription(t.Description, suffix)
		markDescendantsWriteOnly(t.NestedObject.Attributes)
		return t, true
	case schema.MapNestedAttribute:
		t.Computed, t.WriteOnly, t.PlanModifiers = false, true, nil
		t.Description = appendDescription(t.Description, suffix)
		markDescendantsWriteOnly(t.NestedObject.Attributes)
		return t, true
	case schema.SingleNestedAttribute:
		t.Computed, t.WriteOnly, t.PlanModifiers = false, true, nil
		t.Description = appendDescription(t.Description, suffix)
		markDescendantsWriteOnly(t.Attributes)
		return t, true
	default:
		return a, false
	}
}

// appendDescription joins base and addition with a space, without a leading space if base is empty.
func appendDescription(base, addition string) string {
	if base == "" {
		return addition
	}
	return base + " " + addition
}
