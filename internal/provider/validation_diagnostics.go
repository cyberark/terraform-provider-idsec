// Copyright CyberArk. 2026
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/cyberark/idsec-sdk-golang/pkg/validation"
)

// appendValidationDiagnostics turns a validation error into one
// per-attribute Terraform diagnostic per failed field so the offending
// HCL attribute is highlighted in editors and CLI output.
func appendValidationDiagnostics(diags *diag.Diagnostics, err error) {
	if err == nil {
		return
	}
	var verr *validation.Error
	if !errors.As(err, &verr) || len(verr.Fields()) == 0 {
		diags.AddError("Invalid Configuration", err.Error())
		return
	}
	for _, fe := range verr.Fields() {
		fp := validation.FieldPath(fe)
		diags.AddAttributeError(
			tfPathFromFieldPath(fp),
			fmt.Sprintf("Invalid value for %q", fp),
			renderFieldMessage(fe),
		)
	}
}

// tfPathFromFieldPath maps a dotted, tag-resolved field path
// (e.g. "spec.account_id") to a Terraform path.Path so the diagnostic
// targets the right attribute.
func tfPathFromFieldPath(dotted string) path.Path {
	if dotted == "" {
		return path.Empty()
	}
	parts := strings.Split(dotted, ".")
	if strings.ContainsAny(parts[0], "[]") {
		return path.Empty()
	}
	out := path.Root(parts[0])
	for _, p := range parts[1:] {
		if strings.ContainsAny(p, "[]") {
			break
		}
		out = out.AtName(p)
	}
	return out
}

// renderFieldMessage turns one validator.FieldError into a human sentence
// suitable for a Terraform diagnostic detail.
func renderFieldMessage(e validator.FieldError) string {
	got := formatGotValue(e)
	param := e.Param()
	unit := lengthUnitFor(e)

	switch e.Tag() {
	case "required":
		return "value is required"
	case "max":
		return fmt.Sprintf("value must be at most %s%s; got %s", param, unit, got)
	case "min":
		return fmt.Sprintf("value must be at least %s%s; got %s", param, unit, got)
	case "len":
		return fmt.Sprintf("value must be exactly %s%s; got %s", param, unit, got)
	case "oneof":
		return fmt.Sprintf("value must be one of: %s; got %s", strings.Join(strings.Fields(param), ", "), got)
	default:
		rule := e.Tag()
		if param != "" {
			rule += "=" + param
		}
		return fmt.Sprintf("value must satisfy %q; got %s", rule, got)
	}
}

// lengthUnitFor returns " characters" for string-typed fields and an
// empty string for numeric ones.
func lengthUnitFor(e validator.FieldError) string {
	if e.Kind() == reflect.String {
		return " characters"
	}
	return ""
}

// formatGotValue renders the value the user supplied so it can be
// embedded in an error message.
func formatGotValue(e validator.FieldError) string {
	v := e.Value()
	if v == nil {
		return "(none)"
	}
	if s, ok := v.(string); ok {
		if s == "" {
			return "(empty)"
		}
		return fmt.Sprintf("%q", s)
	}
	return fmt.Sprintf("%v", v)
}
