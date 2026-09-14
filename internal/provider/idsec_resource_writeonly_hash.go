// Copyright CyberArk. 2026
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"maps"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// setWriteOnlyHashTagsInState returns obj with each sibling in plannedHashTags overwritten to its
// resolved value. It is the write-once counterpart of nullWriteOnlyAttributesInState and must run
// after MergePlanToStateObject: that merge demotes an unknown planned sibling to null, and this
// call restores it to the concrete apply-time value.
//
// Returns obj unchanged (with a logged warning) on any error so that a derivation failure
// degrades to a null hash rather than crashing the apply.
func (s *IdsecResource) setWriteOnlyHashTagsInState(ctx context.Context, obj types.Object, plannedHashTags map[string]types.String) types.Object {
	if len(plannedHashTags) == 0 {
		return obj
	}
	attrTypes := obj.AttributeTypes(ctx)
	// obj.Attributes() may return the live internal map in some framework versions; clone before
	// mutating, mirroring nullWriteOnlyAttributesInState's identical defensive copy.
	attrs := maps.Clone(obj.Attributes())
	for siblingName, tagValue := range plannedHashTags {
		if _, exists := attrTypes[siblingName]; !exists {
			tflog.Warn(ctx, "setWriteOnlyHashTagsInState: sibling not in object type, skipping",
				map[string]interface{}{"sibling": siblingName})
			continue
		}
		attrs[siblingName] = tagValue
	}
	newObj, diags := types.ObjectValue(attrTypes, attrs)
	if diags.HasError() {
		// Deliberately NOT logging the diags bundle wholesale: tflog does not redact, and attrs
		// above holds the resolved, credential-derived tags. Every diag.Diagnostic ObjectValue can
		// currently emit only carries attribute names and type strings, but that is a property of
		// today's framework implementation, not a guarantee -- a future framework version, or a
		// custom attr.Type, could echo a value into a diagnostic's detail string. Logging only the
		// sibling names and each diagnostic's Summary() keeps this line unable to widen into a
		// leak channel no matter what the bundle itself ends up containing.
		summaries := make([]string, 0, len(diags.Errors()))
		for _, d := range diags.Errors() {
			summaries = append(summaries, d.Summary())
		}
		tflog.Warn(ctx, "setWriteOnlyHashTagsInState: failed to rebuild object, hash siblings will be null",
			map[string]interface{}{
				"siblings":             slices.Sorted(maps.Keys(plannedHashTags)),
				"diagnostic_summaries": summaries,
			})
		return obj
	}
	return newObj
}
