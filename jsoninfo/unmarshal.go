package jsoninfo

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// UnmarshalStructFields function:
//   * Unmarshals struct fields, ignoring UnmarshalJSON(...) and fields without 'json' tag.
//   * Correctly handles StrictStruct
func UnmarshalStructFields(data []byte, v StrictStruct) error {
	reflection := reflect.ValueOf(v)
	if reflection.Kind() != reflect.Ptr {
		panic(fmt.Errorf("Value %T is not a pointer", v))
	}
	if reflection.IsNil() {
		panic(fmt.Errorf("Value %T is nil", v))
	}
	reflection = reflection.Elem()
	for (reflection.Kind() == reflect.Interface || reflection.Kind() == reflect.Ptr) && !reflection.IsNil() {
		reflection = reflection.Elem()
	}
	reflectionType := reflection.Type()
	if reflectionType.Kind() != reflect.Struct {
		panic(fmt.Errorf("Value %T is not a struct", v))
	}
	typeInfo := GetTypeInfo(reflectionType)

	// Unmarshal everything
	var remainingFields map[string]json.RawMessage
	err := json.Unmarshal(data, &remainingFields)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal extension properties: %v\nInput: %s", err, string(data))
	}

	// Supported fields
	fields := typeInfo.Fields
	for fieldIndex, field := range fields {
		fieldData, exists := remainingFields[field.JSONName]
		if !exists {
			continue
		}
		if field.TypeIsUnmarshaller {
			fieldType := field.Type
			isPtr := false
			if fieldType.Kind() == reflect.Ptr {
				fieldType = fieldType.Elem()
				isPtr = true
			}
			fieldValue := reflect.New(fieldType)
			err := fieldValue.Interface().(json.Unmarshaler).UnmarshalJSON(fieldData)
			if err != nil {
				if field.MultipleFields {
					i := fieldIndex + 1
					if i < len(fields) && fields[i].JSONName == field.JSONName {
						continue
					}
				}
				return fmt.Errorf("Error while unmarshalling property '%s' (%s): %v",
					field.JSONName, fieldValue.Type().String(), err)
			}
			if !isPtr {
				fieldValue = fieldValue.Elem()
			}
			reflection.FieldByIndex(field.Index).Set(fieldValue)

			// Remove the field from remaining fields
			delete(remainingFields, field.JSONName)
		} else {
			fieldPtr := reflection.FieldByIndex(field.Index)
			if fieldPtr.Kind() != reflect.Ptr || fieldPtr.IsNil() {
				fieldPtr = fieldPtr.Addr()
			}
			err := json.Unmarshal(fieldData, fieldPtr.Interface())
			if err != nil {
				if field.MultipleFields {
					i := fieldIndex + 1
					if i < len(fields) && fields[i].JSONName == field.JSONName {
						continue
					}
				}
				return fmt.Errorf("Error while unmarshalling property '%s' (%s): %v",
					field.JSONName, fieldPtr.Type().String(), err)
			}

			// Remove the field from remaining fields
			delete(remainingFields, field.JSONName)
		}
	}

	// Extensions
	return v.UnmarshalJSONUnsupportedFields(data, remainingFields)
}
