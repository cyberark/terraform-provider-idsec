// Copyright CyberArk. 2026
// SPDX-License-Identifier: Apache-2.0

package schemas

import (
	"fmt"
	"reflect"
	"strings"

	modelsactions "github.com/cyberark/idsec-sdk-golang/pkg/models/actions"
)

// deprecationInfo carries the two distinct strings we derive from an SDK field's
// `deprecated:"<replacement>,<message>"` struct tag.
type deprecationInfo struct {
	runtimeMsg string
	descSuffix string
}

// newDeprecationInfo formats the SDK Deprecation struct for Terraform.
//
// Format precedence:
//
//	replacement + message: `Use "<replacement>" instead. <message>`
//	replacement only:      `Use "<replacement>" instead.`
//	message only:          `<message>`
//	marker only:           runtime="Deprecated.", desc suffix omitted
func newDeprecationInfo(field reflect.StructField) deprecationInfo {
	dep := modelsactions.FieldDeprecation(field)
	if dep == nil {
		return deprecationInfo{}
	}
	switch {
	case dep.Replacement != "" && dep.Message != "":
		s := fmt.Sprintf("Use %q instead. %s", dep.Replacement, dep.Message)
		return deprecationInfo{runtimeMsg: s, descSuffix: s}
	case dep.Replacement != "":
		s := fmt.Sprintf("Use %q instead.", dep.Replacement)
		return deprecationInfo{runtimeMsg: s, descSuffix: s}
	case dep.Message != "":
		return deprecationInfo{runtimeMsg: dep.Message, descSuffix: dep.Message}
	default:
		return deprecationInfo{runtimeMsg: "Deprecated."}
	}
}

// applyDeprecation writes deprecation metadata onto a Plugin Framework
// attribute and returns the modified copy. It is a no-op when info is zero.
func applyDeprecation[T any](attr T, info deprecationInfo) T {
	if info.runtimeMsg == "" {
		return attr
	}
	v := reflect.ValueOf(&attr).Elem()
	if v.Kind() != reflect.Struct {
		return attr
	}
	if f := v.FieldByName("DeprecationMessage"); f.IsValid() && f.CanSet() && f.Kind() == reflect.String {
		f.SetString(info.runtimeMsg)
	}
	if info.descSuffix != "" {
		appendDeprecationToDescription(v, "Description", info.descSuffix)
	}
	return attr
}

// appendDeprecationToDescription appends a `**Deprecated**: <msg>` suffix to a
// string field if it exists and is settable.
func appendDeprecationToDescription(v reflect.Value, fieldName, msg string) {
	f := v.FieldByName(fieldName)
	if !f.IsValid() || !f.CanSet() || f.Kind() != reflect.String {
		return
	}
	suffix := "**Deprecated**: " + msg
	existing := f.String()
	if strings.Contains(existing, suffix) {
		return
	}
	if existing != "" {
		existing += " "
	}
	f.SetString(existing + suffix)
}
