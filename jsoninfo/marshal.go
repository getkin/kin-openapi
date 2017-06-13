package jsoninfo

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// MarshalStructFields function:
//   * Marshals struct fields, ignoring MarshalJSON() and fields without 'json' tag.
//   * Correctly handles StrictStruct semantics.
func MarshalStructFields(value StrictStruct) ([]byte, error) {
	reflection := reflect.ValueOf(value)

	// Follow "encoding/json" semantics
	if reflection.Kind() != reflect.Ptr {
		// Panic because this is a clear programming error
		panic(fmt.Errorf("Value %s is not a pointer", reflection.Type().String()))
	}
	if reflection.IsNil() {
		// Panic because this is a clear programming error
		panic(fmt.Errorf("Value %s is nil", reflection.Type().String()))
	}

	// Take the element
	reflection = reflection.Elem()

	// Obtain typeInfo
	typeInfo := GetTypeInfo(reflection.Type())

	// Declare result
	result := make(map[string]json.RawMessage)

	// Marshal extensions
	err := value.MarshalJSONUnsupportedFields(result)
	if err != nil {
		return nil, err
	}

	// Supported fields
iteration:
	for _, field := range typeInfo.Fields {
		fieldValue := reflection.FieldByIndex(field.Index)
		if v, ok := fieldValue.Interface().(json.Marshaler); ok {
			if fieldValue.Kind() == reflect.Ptr && fieldValue.IsNil() {
				if field.JSONOmitEmpty {
					continue iteration
				}
				result[field.JSONName] = []byte("null")
				continue
			}
			fieldData, err := v.MarshalJSON()
			if err != nil {
				return nil, err
			}
			result[field.JSONName] = fieldData
			continue
		}
		switch fieldValue.Kind() {
		case reflect.Ptr, reflect.Interface:
			if fieldValue.IsNil() {
				if field.JSONOmitEmpty {
					continue iteration
				}
				result[field.JSONName] = []byte("null")
				continue
			}
		case reflect.Struct:
		case reflect.Map:
			if field.JSONOmitEmpty && (fieldValue.IsNil() || fieldValue.Len() == 0) {
				continue iteration
			}
		case reflect.Slice:
			if field.JSONOmitEmpty && fieldValue.Len() == 0 {
				continue iteration
			}
		case reflect.Bool:
			x := fieldValue.Bool()
			if field.JSONOmitEmpty && x == false {
				continue iteration
			}
			s := "false"
			if x == true {
				s = "true"
			}
			result[field.JSONName] = []byte(s)
			continue iteration
		case reflect.Int64, reflect.Int, reflect.Int32:
			if field.JSONOmitEmpty && fieldValue.Int() == 0 {
				continue iteration
			}
		case reflect.Float64:
			if field.JSONOmitEmpty && fieldValue.Float() == 0.0 {
				continue iteration
			}
		case reflect.String:
			if field.JSONOmitEmpty && len(fieldValue.String()) == 0 {
				continue iteration
			}
		default:
			panic(fmt.Errorf("Field '%s' has unsupported type %s", field.JSONName, field.Type.String()))
		}

		// No special treament is needed
		// Use plain old "encoding/json".Marshal
		fieldData, err := json.Marshal(fieldValue.Addr().Interface())
		if err != nil {
			return nil, err
		}
		result[field.JSONName] = fieldData
	}
	return json.Marshal(result)
}
