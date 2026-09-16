// Copyright CyberArk. 2026
// SPDX-License-Identifier: Apache-2.0

package schemas

import (
	"reflect"
	"slices"

	modelsactions "github.com/cyberark/idsec-sdk-golang/pkg/models/actions"
)

// isFieldSensitive reports whether a model field must be rendered as Terraform-sensitive.
//
// Two independent sources are consulted, and either one is sufficient:
//
//   - The SDK's `secret:"true"` struct tag on the field. Sensitivity is a property of the model
//     field, so a resource and the data source that share a StateSchema cannot disagree about it.
//   - The action definition's SensitiveAttributes list, matched on the attribute name. This
//     remains the escape hatch for fields whose SDK model is not tagged.
//
// The tag is checked first because it is the source that cannot drift. SensitiveAttributes is
// retained rather than replaced because the provider vendors the SDK at a pinned version and
// cannot tag fields in it; TestSensitiveAttributesAreConsistent guards the resulting overlap.
//
// Parameters:
// - field: the reflected model field, used to read the `secret` struct tag
// - fieldName: the resolved Terraform attribute name for the field
// - sensitiveAttrs: the action definition's SensitiveAttributes, may be nil
//
// Returns true when the attribute must carry Sensitive in the generated schema.
func isFieldSensitive(field reflect.StructField, fieldName string, sensitiveAttrs []string) bool {
	return modelsactions.FieldIsSecret(field) || slices.Contains(sensitiveAttrs, fieldName)
}
