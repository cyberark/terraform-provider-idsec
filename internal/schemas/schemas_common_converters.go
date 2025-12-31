// Copyright (c) CyberArk.
// SPDX-License-Identifier: Apache-2.0

package schemas

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/iancoleman/strcase"
	"github.com/mitchellh/mapstructure"
)

func resolveFieldsSquashed(schema reflect.Type) []reflect.StructField {
	var fields []reflect.StructField
	if schema.Kind() == reflect.Pointer {
		schema = schema.Elem()
	}
	for i := 0; i < schema.NumField(); i++ {
		field := schema.Field(i)
		if field.Tag.Get("mapstructure") == ",squash" {
			nestedFields := resolveFieldsSquashed(field.Type)
			fields = append(fields, nestedFields...)
			continue
		}
		if field.Tag.Get("mapstructure") == "-" { // skip fields marked to be ignored
			continue
		}
		if field.PkgPath != "" { // unexported field
			continue
		}
		fields = append(fields, field)
	}
	return fields
}

func resolveFieldsValueSquashed(value reflect.Value) []reflect.Value {
	var fields []reflect.Value
	if value.Kind() == reflect.Pointer {
		value = value.Elem()
	}
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldType := value.Type().Field(i)
		if fieldType.Tag.Get("mapstructure") == ",squash" {
			nestedFields := resolveFieldsValueSquashed(field)
			fields = append(fields, nestedFields...)
			continue
		}
		if fieldType.Tag.Get("mapstructure") == "-" { // skip fields marked to be ignored
			continue
		}
		if fieldType.PkgPath != "" { // unexported field
			continue
		}
		fields = append(fields, field)
	}
	return fields
}

func resolveFieldName(field reflect.StructField) string {
	fieldName := strings.Split(field.Tag.Get("mapstructure"), ",")[0]
	if fieldName != "" {
		return strcase.ToSnake(fieldName)
	}
	fieldName = strings.Split(field.Tag.Get("flag"), ",")[0]
	if fieldName != "" {
		return strcase.ToSnake(fieldName)
	}
	fieldName = strings.Split(field.Tag.Get("json"), ",")[0]
	if fieldName != "" {
		return strcase.ToSnake(fieldName)
	}
	return strcase.ToSnake(field.Name)
}

func isType[T any](t attr.Type) bool {
	_, ok := t.(T)
	return ok
}

func asType[T any](t attr.Type) (T, error) {
	if typed, ok := t.(T); ok {
		return typed, nil
	}
	var zero T
	return zero, fmt.Errorf("failed to cast to %T", zero)
}

func getNullValue(t attr.Type) (attr.Value, error) {
	switch {
	case t.Equal(types.StringType):
		return types.StringNull(), nil
	case t.Equal(types.BoolType):
		return types.BoolNull(), nil
	case t.Equal(types.Int64Type):
		return types.Int64Null(), nil
	case t.Equal(types.NumberType):
		return types.NumberNull(), nil
	case t.Equal(types.Float64Type):
		return types.Float64Null(), nil
	case isType[types.ListType](t):
		typed, err := asType[types.ListType](t)
		if err != nil {
			return nil, err
		}
		return types.ListNull(typed.ElemType), nil
	case isType[types.SetType](t):
		typed, err := asType[types.SetType](t)
		if err != nil {
			return nil, err
		}
		return types.SetNull(typed.ElemType), nil
	case isType[types.TupleType](t):
		typed, err := asType[types.TupleType](t)
		if err != nil {
			return nil, err
		}
		return types.TupleNull(typed.ElemTypes), nil
	case isType[types.MapType](t):
		typed, err := asType[types.MapType](t)
		if err != nil {
			return nil, err
		}
		return types.MapNull(typed.ElemType), nil
	case isType[types.ObjectType](t):
		typed, err := asType[types.ObjectType](t)
		if err != nil {
			return nil, err
		}
		return types.ObjectNull(typed.AttrTypes), nil
	case isType[basetypes.DynamicTypable](t):
		return types.DynamicNull(), nil
	default:
		return nil, fmt.Errorf("unsupported attribute type: %T", t)
	}
}

func objectToMap(obj types.Object, prototype interface{}) (map[string]interface{}, error) {
	if obj.IsNull() || obj.IsUnknown() {
		return nil, fmt.Errorf("object is null or unknown")
	}
	result := make(map[string]interface{})
	for key, val := range obj.Attributes() {
		if val.IsNull() {
			result[key] = nil
			continue
		}
		if val.IsUnknown() {
			continue
		}
		goVal, err := attrToInterface(key, val, prototype)
		if err != nil {
			return nil, fmt.Errorf("error converting attribute %q: %w", key, err)
		}
		if goVal == nil {
			continue
		}
		actualField := findFieldByName(prototype, key)
		if actualField != nil && actualField.Type.Kind() == reflect.Pointer {
			goValReflect := reflect.ValueOf(goVal)
			if goValReflect.Kind() != reflect.Pointer {
				ptrVal := reflect.New(goValReflect.Type())
				ptrVal.Elem().Set(goValReflect)
				result[key] = ptrVal.Interface()
			} else {
				result[key] = goVal
			}
		} else {
			goValReflect := reflect.ValueOf(goVal)
			if goValReflect.Kind() == reflect.Pointer {
				if !goValReflect.IsNil() {
					result[key] = goValReflect.Elem().Interface()
				}
			} else {
				if goValReflect.Kind() == reflect.Bool || !goValReflect.IsZero() {
					result[key] = goVal
				}
			}
		}
	}
	return result, nil
}

func findFieldByName(schema interface{}, name string) *reflect.StructField {
	if schema == nil || name == "" {
		return nil
	}
	schemaType := reflect.TypeOf(schema)
	if schemaType == nil {
		return nil
	}
	if schemaType.Kind() == reflect.Pointer {
		schemaType = schemaType.Elem()
	}
	if schemaType.Kind() != reflect.Struct {
		return nil
	}
	for i := 0; i < schemaType.NumField(); i++ {
		field := schemaType.Field(i)

		if field.PkgPath != "" && !field.Anonymous {
			continue
		}
		flagName := field.Tag.Get("mapstructure")
		if flagName == "" {
			flagName = field.Tag.Get("json")
		}
		if flagName == "" {
			flagName = field.Name
		}
		if strings.Split(flagName, ",")[0] == name {
			return &field
		}
		if field.Tag.Get("mapstructure") == ",squash" {
			subSchema := reflect.New(field.Type).Interface()
			return findFieldByName(subSchema, name)
		}
	}
	return nil
}

func attrToInterface(key string, val attr.Value, prototype interface{}) (interface{}, error) {
	if val.IsNull() || val.IsUnknown() {
		return nil, nil
	}
	actualField := findFieldByName(prototype, key)
	switch v := val.(type) {
	case types.String:
		return v.ValueString(), nil
	case types.Int64:
		return v.ValueInt64(), nil
	case types.Bool:
		return v.ValueBool(), nil
	case types.Float64:
		return v.ValueFloat64(), nil
	case types.Object:
		if actualField != nil {
			nestedPrototype := reflect.New(actualField.Type).Interface()
			return objectToMap(v, nestedPrototype)
		}
		return objectToMap(v, prototype)
	case types.Dynamic:
		if s, ok := v.UnderlyingValue().(types.String); ok {
			var result interface{}
			err := json.Unmarshal([]byte(s.ValueString()), &result)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal dynamic JSON value: %w", err)
			}
			return result, nil
		}
		underlying := v.UnderlyingValue()
		if actualField != nil {
			nestedPrototype := reflect.New(actualField.Type).Interface()
			return attrToInterface(key, underlying, nestedPrototype)
		}
		return attrToInterface(key, underlying, prototype)
	case types.Map:
		attrMap := v.Elements()
		m := make(map[string]interface{}, len(attrMap))
		var elemPrototype interface{}
		if actualField != nil && actualField.Type.Kind() == reflect.Map {
			elemPrototype = reflect.New(actualField.Type.Elem()).Interface()
		}
		for k, elem := range attrMap {
			converted, err := attrToInterface(k, elem, elemPrototype)
			if err != nil {
				return nil, err
			}
			m[k] = converted
		}
		return m, nil
	case types.List, types.Set, types.Tuple:
		var elems []attr.Value
		if l, ok := v.(types.List); ok {
			elems = l.Elements()
		} else if s, ok := v.(types.Set); ok {
			elems = s.Elements()
		} else if t, ok := v.(types.Tuple); ok {
			elems = t.Elements()
		}
		list := make([]interface{}, len(elems))
		var elemPrototype interface{}
		if actualField != nil {
			fieldType := actualField.Type
			for fieldType.Kind() == reflect.Pointer {
				fieldType = fieldType.Elem()
			}
			if fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Array {
				elemType := fieldType.Elem()
				if elemType.Kind() == reflect.Pointer {
					elemPrototype = reflect.New(elemType.Elem()).Interface()
				} else {
					elemPrototype = reflect.New(elemType).Interface()
				}
			}
		}
		for i, elem := range elems {
			converted, err := attrToInterface("", elem, elemPrototype)
			if err != nil {
				return nil, err
			}
			list[i] = converted
		}
		return list, nil
	default:
		return nil, fmt.Errorf("unsupported attribute type: %T", val)
	}
}

func reflectTypeToTerraformType(t reflect.Type) (attr.Type, error) {
	if t == nil {
		return types.StringType, nil
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.String:
		return types.StringType, nil
	case reflect.Bool:
		return types.BoolType, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return types.Int64Type, nil
	case reflect.Slice, reflect.Array:
		elemType, err := reflectTypeToTerraformType(t.Elem())
		if err != nil {
			return nil, err
		}
		return types.ListType{ElemType: elemType}, nil
	case reflect.Map:
		if t.Key().Kind() != reflect.String {
			return nil, fmt.Errorf("map key type must be string")
		}
		elemType, err := reflectTypeToTerraformType(t.Elem())
		if err != nil {
			return nil, err
		}
		return types.MapType{ElemType: elemType}, nil
	case reflect.Struct:
		attrTypes := map[string]attr.Type{}
		actualFields := resolveFieldsSquashed(t)
		for i := range actualFields {
			field := actualFields[i]
			if field.PkgPath != "" {
				continue
			}
			fieldType, err := reflectTypeToTerraformType(field.Type)
			if err != nil {
				return nil, err
			}
			attrTypes[resolveFieldName(field)] = fieldType
		}
		return types.ObjectType{AttrTypes: attrTypes}, nil
	case reflect.Interface:
		// For interfaces, we can return a dynamic type, but it might be better to use a more specific type if known.
		return types.DynamicType, nil
	default:
		return nil, fmt.Errorf("unsupported type: %s", t.Kind())
	}
}

func interfaceTypeToAttr(ctx context.Context, val interface{}, t attr.Type) (attr.Value, error) {
	valReflect := reflect.ValueOf(val)
	if valReflect.Kind() == reflect.Pointer {
		if valReflect.IsNil() {
			return getNullValue(t)
		}
		valReflect = valReflect.Elem()
	}
	switch {
	case t.Equal(types.StringType):
		return types.StringValue(fmt.Sprintf("%v", valReflect.String())), nil
	case t.Equal(types.Int64Type):
		switch valReflect.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return types.Int64Value(valReflect.Int()), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			uintVal := valReflect.Uint()
			if uintVal > uint64(math.MaxInt64) {
				return nil, fmt.Errorf("uint value %d overflows int64", uintVal)
			}
			return types.Int64Value(int64(uintVal)), nil
		default:
			return nil, fmt.Errorf("unsupported kind %v for Int64Type", valReflect.Kind())
		}
	case t.Equal(types.BoolType):
		return types.BoolValue(valReflect.Bool()), nil
	case isType[types.ObjectType](t):
		typed, err := asType[types.ObjectType](t)
		if err != nil {
			return nil, err
		}
		attrs := make(map[string]attr.Type)
		values := make(map[string]attr.Value)
		for key, attrType := range typed.AttrTypes {
			attrs[strcase.ToSnake(key)] = attrType
		}
		actualFields := resolveFieldsSquashed(valReflect.Type())
		actualFieldValues := resolveFieldsValueSquashed(valReflect)
		for i := range actualFieldValues {
			field := actualFieldValues[i]
			if !field.IsValid() || !field.CanInterface() {
				continue
			}
			tagName := resolveFieldName(actualFields[i])
			if attrType, ok := attrs[tagName]; ok {
				attrVal, err := interfaceTypeToAttr(ctx, field.Interface(), attrType)
				if err != nil {
					return nil, fmt.Errorf("field '%s': %w", tagName, err)
				}
				values[tagName] = attrVal
			} else {
				tflog.Warn(ctx, fmt.Sprintf("Field '%s' not found in schema attributes", tagName))
			}
		}
		objVal, diag := types.ObjectValue(attrs, values)
		if diag.HasError() {
			return nil, fmt.Errorf("failed to convert object: %v", diag)
		}
		return objVal, nil
	case isType[types.ListType](t):
		typed, err := asType[types.ListType](t)
		if err != nil {
			return nil, err
		}
		var elems []attr.Value
		for i := 0; i < valReflect.Len(); i++ {
			elemAttr, err := interfaceTypeToAttr(ctx, valReflect.Index(i).Interface(), typed.ElemType)
			if err != nil {
				return nil, err
			}
			elems = append(elems, elemAttr)
		}
		listVal, diag := types.ListValue(typed.ElemType, elems)
		if diag.HasError() {
			return nil, fmt.Errorf("failed to convert list: %v", diag)
		}
		return listVal, nil
	case isType[types.SetType](t):
		typed, err := asType[types.SetType](t)
		if err != nil {
			return nil, err
		}
		var elems []attr.Value
		for i := 0; i < valReflect.Len(); i++ {
			elemAttr, err := interfaceTypeToAttr(ctx, valReflect.Index(i).Interface(), typed.ElemType)
			if err != nil {
				return nil, err
			}
			elems = append(elems, elemAttr)
		}
		setVal, diag := types.SetValue(typed.ElemType, elems)
		if diag.HasError() {
			return nil, fmt.Errorf("failed to convert set: %v", diag)
		}
		return setVal, nil
	case isType[types.TupleType](t):
		typed, err := asType[types.TupleType](t)
		if err != nil {
			return nil, err
		}
		var elems []attr.Value
		for i := 0; i < valReflect.Len(); i++ {
			elemAttr, err := interfaceTypeToAttr(ctx, valReflect.Index(i).Interface(), typed.ElemTypes[i])
			if err != nil {
				return nil, err
			}
			elems = append(elems, elemAttr)
		}
		tupleVal, diag := types.TupleValue(typed.ElemTypes, elems)
		if diag.HasError() {
			return nil, fmt.Errorf("failed to convert tuple: %v", diag)
		}
		return tupleVal, nil
	case isType[types.MapType](t):
		typed, err := asType[types.MapType](t)
		if err != nil {
			return nil, err
		}
		result := make(map[string]attr.Value)
		for _, key := range valReflect.MapKeys() {
			elemAttr, err := interfaceTypeToAttr(ctx, valReflect.MapIndex(key).Interface(), typed.ElemType)
			if err != nil {
				return nil, err
			}
			result[key.String()] = elemAttr
		}
		mapVal, diag := types.MapValue(typed.ElemType, result)
		if diag.HasError() {
			return nil, fmt.Errorf("failed to convert map: %v", diag)
		}
		return mapVal, nil
	case isType[basetypes.DynamicTypable](t):
		jsonBytes, err := json.Marshal(val)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal dynamic value to JSON: %w", err)
		}
		dyn := basetypes.NewDynamicValue(types.StringValue(string(jsonBytes)))
		return dyn, nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", t)
	}
}

func setTargetValueFromPlanAndState(planVal reflect.Value, stateVal reflect.Value, target reflect.Value) {
	// Dereference pointers
	for planVal.Kind() == reflect.Pointer {
		if planVal.IsNil() {
			return
		}
		planVal = planVal.Elem()
	}

	if !stateVal.IsValid() {
		if target.Type().Kind() == reflect.Pointer {
			if planVal.Kind() != reflect.Pointer && planVal.CanAddr() {
				planVal = planVal.Addr()
			}
		}
		target.Set(planVal)
		return
	}

	// Handle basic types
	if slices.Contains(simpleTypes, planVal.Kind()) {
		if planVal.Kind() == reflect.Bool {
			if target.Type().Kind() == reflect.Pointer {
				if planVal.CanAddr() {
					target.Set(planVal.Addr())
				} else {
					ptrVal := reflect.New(planVal.Type())
					ptrVal.Elem().Set(planVal)
					target.Set(ptrVal)
				}
			} else {
				target.Set(planVal)
			}
		} else if !planVal.IsZero() {
			if stateVal.IsZero() || !reflect.DeepEqual(planVal.Interface(), stateVal.Interface()) {
				if target.Type().Kind() == reflect.Pointer {
					if planVal.Kind() != reflect.Pointer {
						if planVal.CanAddr() {
							target.Set(planVal.Addr())
						} else {
							ptrVal := reflect.New(planVal.Type())
							ptrVal.Elem().Set(planVal)
							target.Set(ptrVal)
						}
					} else {
						target.Set(planVal)
					}
				} else {
					target.Set(planVal)
				}
			}
		}
		return
	}

	// Handle complex types
	switch planVal.Kind() {
	case reflect.Slice, reflect.Map, reflect.Interface, reflect.Chan:
		if !planVal.IsNil() {
			if target.Type().Kind() == reflect.Pointer && planVal.Kind() != reflect.Pointer {
				if planVal.CanAddr() {
					target.Set(planVal.Addr())
				} else {
					ptrVal := reflect.New(planVal.Type())
					ptrVal.Elem().Set(planVal)
					target.Set(ptrVal)
				}
			} else {
				target.Set(planVal)
			}
		}
		return
	case reflect.Struct:
		if stateVal.IsZero() {
			if target.Type().Kind() == reflect.Pointer && planVal.Kind() != reflect.Pointer {
				if planVal.CanAddr() {
					target.Set(planVal.Addr())
				} else {
					ptrVal := reflect.New(planVal.Type())
					ptrVal.Elem().Set(planVal)
					target.Set(ptrVal)
				}
			} else {
				target.Set(planVal)
			}
			return
		}
		stateValNonPtr := stateVal
		if stateValNonPtr.Kind() == reflect.Pointer {
			stateValNonPtr = stateValNonPtr.Elem()
		}
		targetNonPtr := target
		if targetNonPtr.Type().Kind() == reflect.Pointer {
			targetNonPtr = targetNonPtr.Elem()
		}
		for i := 0; i < planVal.NumField(); i++ {
			field := planVal.Type().Field(i)
			if field.PkgPath != "" {
				continue
			}
			planField := planVal.Field(i)
			stateField := stateValNonPtr.FieldByName(field.Name)
			targetField := targetNonPtr.FieldByName(field.Name)
			if targetField.IsValid() && targetField.CanSet() {
				setTargetValueFromPlanAndState(planField, stateField, targetField)
			}
		}
	default:
		return
	}
}

// FindMethodByName finds a method by name in the given value.
func FindMethodByName(value reflect.Value, methodName string) (*reflect.Value, error) {
	actionMethod := value.MethodByName(methodName)
	if !actionMethod.IsValid() {
		for i := 0; i < value.NumMethod(); i++ {
			method := value.Type().Method(i)
			if strings.EqualFold(method.Name, methodName) {
				actionMethod = value.MethodByName(method.Name)
				break
			}
		}
		if !actionMethod.IsValid() {
			return nil, fmt.Errorf("method %s not found", methodName)
		}
	}
	return &actionMethod, nil
}

// StructFromPlanObject converts a Terraform plan object to a Go struct.
func StructFromPlanObject(ctx context.Context, plan *tfsdk.Plan, prototype interface{}) (interface{}, error) {
	var planObj types.Object
	diags := plan.Get(ctx, &planObj)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to get full plan object: %v", diags)
	}
	dataMap, err := objectToMap(planObj, prototype)
	if err != nil {
		return nil, fmt.Errorf("failed to convert plan object to map: %v", err)
	}
	protoType := reflect.TypeOf(prototype)
	if protoType.Kind() == reflect.Pointer {
		protoType = protoType.Elem()
	}
	newStruct := reflect.New(protoType).Interface()
	err = mapstructure.Decode(dataMap, newStruct)
	if err != nil {
		return nil, fmt.Errorf("failed to decode map to struct: %w", err)
	}
	return newStruct, nil
}

// StructFromStateObject converts a Terraform state object to a Go struct.
func StructFromStateObject(ctx context.Context, state *tfsdk.State, prototype interface{}) (interface{}, error) {
	var stateObj types.Object
	diags := state.Get(ctx, &stateObj)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to get full state object: %v", diags)
	}
	dataMap, err := objectToMap(stateObj, prototype)
	if err != nil {
		return nil, fmt.Errorf("failed to convert plan object to map: %v", diags)
	}
	protoType := reflect.TypeOf(prototype)
	if protoType.Kind() == reflect.Pointer {
		protoType = protoType.Elem()
	}
	newStruct := reflect.New(protoType).Interface()
	err = mapstructure.Decode(dataMap, newStruct)
	if err != nil {
		return nil, fmt.Errorf("failed to decode map to struct: %w", err)
	}
	return reflect.ValueOf(newStruct).Elem().Interface(), nil
}

// StructFromConfigObject converts a Terraform config object to a Go struct.
func StructFromConfigObject(ctx context.Context, config *tfsdk.Config, prototype interface{}) (interface{}, error) {
	var stateObj types.Object
	diags := config.Get(ctx, &stateObj)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to get full state object: %v", diags)
	}
	dataMap, err := objectToMap(stateObj, prototype)
	if err != nil {
		return nil, fmt.Errorf("failed to convert plan object to map: %v", diags)
	}
	protoType := reflect.TypeOf(prototype)
	newStruct := reflect.New(protoType).Interface()
	err = mapstructure.Decode(dataMap, newStruct)
	if err != nil {
		return nil, fmt.Errorf("failed to decode map to struct: %w", err)
	}
	return reflect.ValueOf(newStruct).Elem().Interface(), nil
}

// StructFromPlanAndStateObject converts a Terraform plan and state object to a Go struct.
func StructFromPlanAndStateObject(ctx context.Context, plan *tfsdk.Plan, state *tfsdk.State, planPrototype interface{}, statePrototype interface{}) (interface{}, error) {
	var stateObj types.Object
	var planObj types.Object
	diags := state.Get(ctx, &stateObj)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to get full state object: %v", diags)
	}
	diags = plan.Get(ctx, &planObj)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to get full plan object: %v", diags)
	}
	stateDataMap, err := objectToMap(stateObj, statePrototype)
	if err != nil {
		return nil, fmt.Errorf("failed to convert state object to map: %v", diags)
	}
	planDataMap, err := objectToMap(planObj, planPrototype)
	if err != nil {
		return nil, fmt.Errorf("failed to convert plan object to map: %v", diags)
	}
	planReflectedPrototype := reflect.TypeOf(planPrototype)
	if planReflectedPrototype.Kind() == reflect.Pointer {
		planReflectedPrototype = planReflectedPrototype.Elem()
	}
	planNewStruct := reflect.New(planReflectedPrototype).Interface()
	err = mapstructure.Decode(planDataMap, planNewStruct)
	if err != nil {
		return nil, fmt.Errorf("failed to decode map to struct: %w", err)
	}

	stateReflectedPrototype := reflect.TypeOf(statePrototype)
	if stateReflectedPrototype.Kind() == reflect.Pointer {
		stateReflectedPrototype = stateReflectedPrototype.Elem()
	}
	stateNewStruct := reflect.New(stateReflectedPrototype).Interface()
	err = mapstructure.Decode(stateDataMap, stateNewStruct)
	if err != nil {
		return nil, fmt.Errorf("failed to decode map to struct: %w", err)
	}
	planFinalizedStruct := reflect.New(planReflectedPrototype).Elem()
	stateValue := reflect.ValueOf(stateNewStruct).Elem()
	actualFields := resolveFieldsSquashed(stateValue.Type())
	actualValueFields := resolveFieldsValueSquashed(stateValue)
	for i := range actualFields {
		field := actualFields[i]
		if newField := planFinalizedStruct.FieldByName(field.Name); newField.IsValid() && newField.CanSet() {
			if newField.Type().Kind() == reflect.Pointer && actualValueFields[i].Kind() != reflect.Pointer {
				actualValueFields[i] = actualValueFields[i].Addr()
			}
			newField.Set(actualValueFields[i])
		}
	}
	planValue := reflect.ValueOf(planNewStruct).Elem()
	actualPlanFields := resolveFieldsSquashed(planValue.Type())
	actualPlanValueFields := resolveFieldsValueSquashed(planValue)
	for i := 0; i < len(actualPlanFields); i++ {
		field := actualPlanFields[i]
		if newField := planFinalizedStruct.FieldByName(field.Name); newField.IsValid() && newField.CanSet() {
			setTargetValueFromPlanAndState(actualPlanValueFields[i], stateValue.FieldByName(field.Name), newField)
		}
	}
	return planFinalizedStruct.Addr().Interface(), nil
}

// StructToStateObject converts a Go struct to a Terraform state object.
func StructToStateObject(ctx context.Context, input interface{}, state *tfsdk.State, plan *tfsdk.Plan, schemaAttrs map[string]attr.Type) (types.Object, error) {
	var stateObj types.Object
	var planObj types.Object
	if state != nil {
		diags := state.Get(ctx, &stateObj)
		if diags.HasError() {
			return types.Object{}, fmt.Errorf("object value getting error: %v", diags)
		}
	}
	if plan != nil {
		diags := plan.Get(ctx, &planObj)
		if diags.HasError() {
			return types.Object{}, fmt.Errorf("object value getting error: %v", diags)
		}
	}
	val := reflect.ValueOf(input)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	typ := val.Type()
	valueMap := make(map[string]attr.Value)
	actualFields := resolveFieldsSquashed(typ)
	actualValueFields := resolveFieldsValueSquashed(val)
	for i := range actualFields {
		field := actualFields[i]
		fieldVal := actualValueFields[i]
		tagName := resolveFieldName(field)
		attrType, ok := schemaAttrs[tagName]
		if !ok {
			tflog.Warn(ctx, fmt.Sprintf("Field '%s' not found in schema attributes", tagName))
			continue
		}
		if !fieldVal.IsValid() || !fieldVal.CanInterface() {
			valueMap[tagName], _ = getNullValue(schemaAttrs[tagName])
			continue
		}
		attrVal, err := interfaceTypeToAttr(ctx, fieldVal.Interface(), attrType)
		if err != nil {
			return types.Object{}, fmt.Errorf("field '%s': %w", tagName, err)
		}
		valueMap[tagName] = attrVal
	}

	for attrName, attrType := range schemaAttrs {
		if _, exists := valueMap[attrName]; !exists {
			if plan != nil {
				if attrValue, ok := planObj.Attributes()[attrName]; ok {
					valueMap[attrName] = attrValue
					continue
				}
			}
			if state != nil {
				if attrValue, ok := stateObj.Attributes()[attrName]; ok {
					valueMap[attrName] = attrValue
					continue
				}
			}
			nullVal, err := getNullValue(attrType)
			if err != nil {
				return types.Object{}, fmt.Errorf("failed to create null value for attribute %q: %w", attrName, err)
			}
			valueMap[attrName] = nullVal
		}
	}
	objVal, diag := types.ObjectValue(schemaAttrs, valueMap)
	if diag.HasError() {
		return types.Object{}, fmt.Errorf("object value creation error: %v", diag)
	}
	return objVal, nil
}

// mergePlanAndStateMap recursively merges plan attributes into existing state attributes.
//
// This function performs a deep merge of Terraform plan values into existing state values,
// handling different attribute types (objects, maps, lists, sets) appropriately. For nested
// structures, it recursively merges their contents rather than replacing them entirely.
//
// Parameters:
//   - ctx: Context for logging and type operations
//   - existingAttrs: Map of existing state attributes to be updated in-place
//   - attrsToMerge: Map of plan attributes to merge into the existing attributes
func mergePlanAndStateMap(ctx context.Context, existingAttrs map[string]attr.Value, attrsToMerge map[string]attr.Value) {
	for key, planVal := range attrsToMerge {
		if planVal.IsNull() || planVal.IsUnknown() {
			continue
		}

		if isType[types.ObjectType](planVal.Type(ctx)) {
			mergeObjectAttribute(ctx, existingAttrs, key, planVal)
			continue
		}

		if isType[types.MapType](planVal.Type(ctx)) {
			mergeMapAttribute(ctx, existingAttrs, key, planVal)
			continue
		}

		if isType[types.ListType](planVal.Type(ctx)) {
			mergeListAttribute(ctx, existingAttrs, key, planVal)
			continue
		}

		if isType[types.SetType](planVal.Type(ctx)) {
			mergeSetAttribute(ctx, existingAttrs, key, planVal)
			continue
		}

		// Scalars / other non-object types: plan overrides existing
		existingAttrs[key] = planVal
	}
}

// mergeObjectAttribute merges a nested object attribute from plan into existing state.
//
// This function performs a deep merge of object attributes by recursively merging
// their nested attributes rather than replacing the entire object.
//
// Parameters:
//   - ctx: Context for type operations
//   - existingAttrs: Map of existing state attributes to be updated in-place
//   - key: Attribute key being merged
//   - planVal: Plan value to merge (must be types.Object type)
func mergeObjectAttribute(ctx context.Context, existingAttrs map[string]attr.Value, key string, planVal attr.Value) {
	planObj, ok := planVal.(types.Object)
	if !ok {
		existingAttrs[key] = planVal
		return
	}

	existingVal, exists := existingAttrs[key]
	if !exists || !isType[types.ObjectType](existingVal.Type(ctx)) {
		existingAttrs[key] = planVal
		return
	}

	existingObj, ok := existingVal.(types.Object)
	if !ok {
		existingAttrs[key] = planVal
		return
	}

	mergedInner := make(map[string]attr.Value, len(existingObj.Attributes()))
	for k, v := range existingObj.Attributes() {
		mergedInner[k] = v
	}
	mergePlanAndStateMap(ctx, mergedInner, planObj.Attributes())
	newObj, _ := types.ObjectValue(existingObj.AttributeTypes(ctx), mergedInner)
	existingAttrs[key] = newObj
}

// mergeMapAttribute merges a map attribute from plan into existing state.
//
// This function performs a deep merge of map attributes. If the map contains object values,
// it recursively merges nested objects. Otherwise, plan values override state values.
//
// Parameters:
//   - ctx: Context for type operations
//   - existingAttrs: Map of existing state attributes to be updated in-place
//   - key: Attribute key being merged
//   - planVal: Plan value to merge (must be types.Map type)
func mergeMapAttribute(ctx context.Context, existingAttrs map[string]attr.Value, key string, planVal attr.Value) {
	planMap, ok := planVal.(types.Map)
	if !ok {
		existingAttrs[key] = planVal
		return
	}

	existingVal, exists := existingAttrs[key]
	if !exists || !isType[types.MapType](existingVal.Type(ctx)) {
		existingAttrs[key] = planVal
		return
	}

	existingMap, ok := existingVal.(types.Map)
	if !ok {
		existingAttrs[key] = planVal
		return
	}

	mapType, err := asType[types.MapType](planMap.Type(ctx))
	if err != nil {
		existingAttrs[key] = planVal
		return
	}

	if !isType[types.ObjectType](mapType.ElemType) {
		existingAttrs[key] = planVal
		return
	}

	mergedMapValues := make(map[string]attr.Value, len(existingMap.Elements()))
	for k, v := range existingMap.Elements() {
		mergedMapValues[k] = v
	}

	for k, planMapVal := range planMap.Elements() {
		if planMapVal.IsNull() || planMapVal.IsUnknown() {
			continue
		}

		existingMapVal, exists := mergedMapValues[k]
		if !exists {
			mergedMapValues[k] = planMapVal
			continue
		}

		planObj, planOk := planMapVal.(types.Object)
		existingObj, existingOk := existingMapVal.(types.Object)
		if !planOk || !existingOk {
			mergedMapValues[k] = planMapVal
			continue
		}

		mergedNestedAttrs := make(map[string]attr.Value, len(existingObj.Attributes()))
		for nestedKey, nestedVal := range existingObj.Attributes() {
			mergedNestedAttrs[nestedKey] = nestedVal
		}
		mergePlanAndStateMap(ctx, mergedNestedAttrs, planObj.Attributes())
		mergedObj, _ := types.ObjectValue(existingObj.AttributeTypes(ctx), mergedNestedAttrs)
		mergedMapValues[k] = mergedObj
	}

	newMap, _ := types.MapValue(mapType.ElemType, mergedMapValues)
	existingAttrs[key] = newMap
}

// mergeListAttribute merges a list attribute from plan into existing state.
//
// This function performs a deep merge of list attributes by index. If list elements are
// objects, it recursively merges them. Otherwise, plan values override state values.
//
// Parameters:
//   - ctx: Context for type operations
//   - existingAttrs: Map of existing state attributes to be updated in-place
//   - key: Attribute key being merged
//   - planVal: Plan value to merge (must be types.List type)
func mergeListAttribute(ctx context.Context, existingAttrs map[string]attr.Value, key string, planVal attr.Value) {
	planList, ok := planVal.(types.List)
	if !ok {
		existingAttrs[key] = planVal
		return
	}

	listType, err := asType[types.ListType](planList.Type(ctx))
	if err != nil {
		existingAttrs[key] = planVal
		return
	}

	if !isType[types.ObjectType](listType.ElemType) {
		existingAttrs[key] = planVal
		return
	}

	existingVal, exists := existingAttrs[key]
	if !exists || !isType[types.ListType](existingVal.Type(ctx)) {
		existingAttrs[key] = planVal
		return
	}

	existingList, ok := existingVal.(types.List)
	if !ok || existingList.IsNull() || existingList.IsUnknown() {
		existingAttrs[key] = planVal
		return
	}

	planElems := planList.Elements()
	existingElems := existingList.Elements()
	mergedElems := make([]attr.Value, len(planElems))

	for i, planElem := range planElems {
		if planElem.IsNull() || planElem.IsUnknown() {
			if i < len(existingElems) {
				mergedElems[i] = existingElems[i]
			} else {
				mergedElems[i] = planElem
			}
			continue
		}

		planObj, planOk := planElem.(types.Object)
		if !planOk || i >= len(existingElems) {
			mergedElems[i] = planElem
			continue
		}

		existingObj, existingOk := existingElems[i].(types.Object)
		if !existingOk || existingObj.IsNull() || existingObj.IsUnknown() {
			mergedElems[i] = planElem
			continue
		}

		mergedNestedAttrs := make(map[string]attr.Value, len(existingObj.Attributes()))
		for nestedKey, nestedVal := range existingObj.Attributes() {
			mergedNestedAttrs[nestedKey] = nestedVal
		}
		mergePlanAndStateMap(ctx, mergedNestedAttrs, planObj.Attributes())
		mergedObj, _ := types.ObjectValue(existingObj.AttributeTypes(ctx), mergedNestedAttrs)
		mergedElems[i] = mergedObj
	}

	newList, _ := types.ListValue(listType.ElemType, mergedElems)
	existingAttrs[key] = newList
}

// mergeSetAttribute merges a set attribute from plan into existing state.
//
// This function filters out null and unknown values from set elements. If set elements
// are objects, it preserves the plan values after filtering.
//
// Parameters:
//   - ctx: Context for type operations
//   - existingAttrs: Map of existing state attributes to be updated in-place
//   - key: Attribute key being merged
//   - planVal: Plan value to merge (must be types.Set type)
func mergeSetAttribute(ctx context.Context, existingAttrs map[string]attr.Value, key string, planVal attr.Value) {
	planSet, ok := planVal.(types.Set)
	if !ok {
		existingAttrs[key] = planVal
		return
	}

	setType, err := asType[types.SetType](planSet.Type(ctx))
	if err != nil {
		existingAttrs[key] = planVal
		return
	}

	if !isType[types.ObjectType](setType.ElemType) {
		existingAttrs[key] = planVal
		return
	}

	planElems := planSet.Elements()
	cleanedElems := make([]attr.Value, 0, len(planElems))

	for _, elem := range planElems {
		if elem.IsNull() || elem.IsUnknown() {
			continue
		}
		cleanedElems = append(cleanedElems, elem)
	}

	newSet, _ := types.SetValue(setType.ElemType, cleanedElems)
	existingAttrs[key] = newSet
}

// MergePlanToStateObject merges a Terraform plan object with a state object.
func MergePlanToStateObject(ctx context.Context, plan *tfsdk.Plan, stateResult types.Object, schemaAttrs map[string]attr.Type) (types.Object, error) {
	var planObj types.Object
	diags := plan.Get(ctx, &planObj)
	if diags.HasError() {
		return types.Object{}, fmt.Errorf("failed to get full plan object: %v", diags)
	}
	mergedAttrsValues := make(map[string]attr.Value)
	for key, val := range stateResult.Attributes() {
		if val.IsNull() || val.IsUnknown() {
			continue
		}
		mergedAttrsValues[key] = val
	}
	mergePlanAndStateMap(ctx, mergedAttrsValues, planObj.Attributes())
	for key, attrType := range schemaAttrs {
		if _, exists := mergedAttrsValues[key]; !exists {
			nullVal, err := getNullValue(attrType)
			if err != nil {
				return types.Object{}, fmt.Errorf("failed to create null value for attribute %q: %w", key, err)
			}
			mergedAttrsValues[key] = nullVal
		}
	}
	for key := range mergedAttrsValues {
		if _, exists := schemaAttrs[key]; !exists {
			delete(mergedAttrsValues, key)
		}
	}
	objVal, diag := types.ObjectValue(schemaAttrs, mergedAttrsValues)
	if diag != nil && diag.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Object value creation error: %v", diag))
		return types.Object{}, fmt.Errorf("object value creation error: %v", diag)
	}
	return objVal, nil
}

// SchemaByPath retrieves a schema value by its path in a nested structure.
func SchemaByPath(schema interface{}, path string) (interface{}, error) {
	keys := strings.Split(path, ".")
	current := schema

	for _, key := range keys {
		switch v := current.(type) {
		case map[string]interface{}:
			if next, ok := v[key]; ok {
				current = next
			} else {
				return nil, fmt.Errorf("key %q not found in map", key)
			}
		default:
			val := reflect.ValueOf(current)
			if val.Kind() == reflect.Pointer {
				val = val.Elem()
			}
			if val.Kind() == reflect.Struct {
				actualValueFields := resolveFieldsValueSquashed(val)
				actualFields := resolveFieldsSquashed(val.Type())
				for i := 0; i < len(actualValueFields); i++ {
					field := actualValueFields[i]
					fieldName := resolveFieldName(actualFields[i])
					if fieldName == key {
						current = field.Interface()
						break
					}
				}
				if reflect.ValueOf(current).IsZero() {
					return nil, fmt.Errorf("field %q not found in struct", key)
				}
			} else {
				return nil, fmt.Errorf("unsupported type %T for key %q", current, key)
			}
		}
	}

	return current, nil
}

func DeepCopy(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	val := reflect.ValueOf(v)
	if !val.IsValid() {
		return nil
	}
	return deepCopy(val).Interface()
}

func deepCopy(v reflect.Value) reflect.Value {
	if !v.IsValid() {
		return v
	}

	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		elemCopy := deepCopy(v.Elem())
		ptrCopy := reflect.New(elemCopy.Type())
		ptrCopy.Elem().Set(elemCopy)
		return ptrCopy.Convert(v.Type())

	case reflect.Interface:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		innerCopy := deepCopy(v.Elem())
		return innerCopy.Convert(v.Type())

	case reflect.Slice:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		cpy := reflect.MakeSlice(v.Type(), v.Len(), v.Cap())
		for i := 0; i < v.Len(); i++ {
			cpy.Index(i).Set(deepCopy(v.Index(i)))
		}
		return cpy

	case reflect.Array:
		cpy := reflect.New(v.Type()).Elem()
		for i := 0; i < v.Len(); i++ {
			cpy.Index(i).Set(deepCopy(v.Index(i)))
		}
		return cpy

	case reflect.Map:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		cpy := reflect.MakeMapWithSize(v.Type(), v.Len())
		for _, key := range v.MapKeys() {
			valCopy := deepCopy(v.MapIndex(key))
			keyCopy := deepCopy(key)
			cpy.SetMapIndex(keyCopy, valCopy)
		}
		return cpy

	case reflect.Struct:
		cpy := reflect.New(v.Type()).Elem()
		for i := 0; i < v.NumField(); i++ {
			if !cpy.Field(i).CanSet() {
				continue
			}
			cpy.Field(i).Set(deepCopy(v.Field(i)))
		}
		return cpy

	default:
		return v
	}
}
