package jsoninfo

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// MarshalStructFields function:
//   * Marshals struct fields, ignoring MarshalJSON() and fields without 'json' tag.
//   * Ignores fields without 'json' tag (to avoid accidental vulnerabilities).
//   * Correctly handles:
//     * jsoninfo.RefProps
//     * jsoninfo.ExtensionProps
//
// Note that the above rules apply only to the argument!
// "encoding/json".Marshal will be used for all children.
//
// Typical usage looks like:
//
//   func (this *SomeType) MarshalJSON() ([]byte, error) {
//     return jsoninfo.MarshalStructFields(this)
//   }
//
func MarshalStructFields(v interface{}) ([]byte, error) {
	value := reflect.ValueOf(v)
	if value.Kind() != reflect.Ptr {
		panic(fmt.Errorf("Value %s is not a pointer", value.Type().String()))
	}
	if value.IsNil() {
		panic(fmt.Errorf("Value %s is nil", value.Type().String()))
	}
	value = value.Elem()

	// Handle non-blank ref
	if refHolder, ok := v.(RefHolder); ok {
		ref := refHolder.GetRef()
		if len(ref) > 0 {
			return json.Marshal(&RefProps{
				Ref: ref,
			})
		}
	}

	// Obtain typeInfo
	typeInfo := GetTypeInfo(value.Type())
	return marshalStructFields(value, typeInfo)
}

func marshalWithoutRef(value reflect.Value) ([]byte, error) {
loop:
	for {
		switch value.Kind() {
		case reflect.Ptr, reflect.Interface:
			value = value.Elem()
			continue loop
		case reflect.Struct:
			typeInfo := GetTypeInfo(value.Type())
			if len(typeInfo.Ref) > 0 {
				// The function doesn't do special treatment for ref
				return marshalStructFields(value, typeInfo)
			}
		}
		break
	}
	return json.Marshal(value)
}

func marshalStructFields(value reflect.Value, typeInfo *TypeInfo) ([]byte, error) {

	// Declare result
	result := make(map[string]json.RawMessage)

	// Marshal extensions
	if x := typeInfo.Extensions; len(x) > 0 {
		extensionsPtr := value.FieldByIndex(x).Addr().Interface().(*ExtensionProps)
		if m := extensionsPtr.UnsupportedExtensions; m != nil {
			for k, v := range m {
				result[k] = v
			}
		}
		for _, extension := range extensionsPtr.Extensions {
			data, err := json.Marshal(extension)
			if err != nil {
				return nil, err
			}
			err = json.Unmarshal(data, result)
			if err != nil {
				return nil, err
			}
		}
	}

	// Supported fields
iteration:
	for _, field := range typeInfo.Fields {
		fieldValue := value.FieldByIndex(field.Index)
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
			if field.JSONNoRef {
				fieldData, err := marshalWithoutRef(fieldValue.Addr())
				if err != nil {
					return nil, err
				}
				result[field.JSONName] = fieldData
				continue iteration
			}
		case reflect.Struct:
			if field.JSONNoRef {
				fieldData, err := marshalWithoutRef(fieldValue.Addr())
				if err != nil {
					return nil, err
				}
				result[field.JSONName] = fieldData
				continue iteration
			}
		case reflect.Map:
			if field.JSONOmitEmpty && (fieldValue.IsNil() || fieldValue.Len() == 0) {
				continue iteration
			}
			if field.JSONNoRef {
				// We need to build the map manually
				resultMap := make(map[string]json.RawMessage)
				for _, k := range fieldValue.MapKeys() {
					v := fieldValue.MapIndex(k)

					// Marshal without ref
					d, err := marshalWithoutRef(v)
					if err != nil {
						return nil, err
					}

					// Put to map
					switch k.Kind() {
					case reflect.String:
						resultMap[k.String()] = d
					default:
						return nil, fmt.Errorf("Struct field can't have 'noref' and have map key of type '%s'", k.Type().String())
					}
				}

				// Marshal map
				data, err := json.Marshal(resultMap)
				if err != nil {
					return nil, err
				}
				result[field.JSONName] = data
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

// UnmarshalStructFields function:
//   * Unmarshals struct fields, ignoring UnmarshalJSON(...)
//   * Ignores fields without 'json' tag (to avoid accidental vulnerabilities).
//   * Correctly handles the following embedded fields:
//     * jsoninfo.RefProps
//     * jsoninfo.ExtensionProps
//   * Returns an error if a JSON object has a property that not's part of the struct.
//
// Typical usage looks like:
//
//   func (this *SomeType) UnmarshalJSON(data []byte[]), error) {
//     return jsoninfo.UnmarshalStructFields(data, this)
//   }
//
func UnmarshalStructFields(data []byte, v interface{}) error {
	value := reflect.ValueOf(v)
	if value.Kind() != reflect.Ptr {
		panic(fmt.Errorf("Value %T is not a pointer", v))
	}
	if value.IsNil() {
		panic(fmt.Errorf("Value %T is nil", v))
	}
	value = value.Elem()
	for (value.Kind() == reflect.Interface || value.Kind() == reflect.Ptr) && !value.IsNil() {
		value = value.Elem()
	}
	valueType := value.Type()
	if valueType.Kind() != reflect.Struct {
		panic(fmt.Errorf("Value %T is not a struct", v))
	}
	typeInfo := GetTypeInfo(valueType)

	// Unmarshal everything
	var remainingFields map[string]json.RawMessage
	err := json.Unmarshal(data, &remainingFields)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal extension properties: %v\nInput: %s", err, string(data))
	}

	// Ref
	if refIndex := typeInfo.Ref; len(refIndex) > 0 {
		refData, exists := remainingFields["$ref"]
		if exists {
			if len(remainingFields) != 1 {
				return fmt.Errorf("Encountered JSON that has both '$ref' and other properties")
			}
			var ref string
			err := json.Unmarshal(refData, &ref)
			if err != nil {
				return err
			}
			value.FieldByIndex(refIndex).Addr().Interface().(RefHolder).SetRef(ref)
			return nil
		}
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
			value.FieldByIndex(field.Index).Set(fieldValue)

			// Remove the field from remaining fields
			delete(remainingFields, field.JSONName)
		} else {
			fieldPtr := value.FieldByIndex(field.Index)
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
	if x := typeInfo.Extensions; len(x) > 0 {
		extensionPropsPtr := value.FieldByIndex(x).Addr().Interface().(*ExtensionProps)

		// Handle type extensions
		if list := typeInfo.typeExtensions; len(list) > 0 {
			holder := value.Interface().(ExtensionHolder)
			for _, typeExtension := range typeInfo.typeExtensions {
				extension := typeExtension.factory()
				err := extension.UnmarshalJSONExtension(data, holder)
				if err != nil {
					return err
				}
				extensionPropsPtr.Extensions = append(extensionPropsPtr.Extensions, extension)
				for _, name := range typeExtension.jsonNames {
					delete(remainingFields, name)
				}
			}
		}

		for k := range remainingFields {
			// Ensure that this is an extension
			if strings.HasPrefix(k, "x-") == false {
				return fmt.Errorf("Type %s doesn't support property '%s'. Extension properties should start with 'x-'", value.Type().Name(), k)
			}

			// Ensure that this doesn't belong to any namespace that expected to be supported
			for _, prefix := range ExtensionPrefixesExpectedToBeSupported {
				if strings.HasPrefix(k, prefix) {
					return fmt.Errorf("Type %s does not support JSON property '%s'. Extension properties starting with '%s' must be supported", value.Type().Name(), k, prefix)
				}
			}
		}
		extensionPropsPtr.UnsupportedExtensions = remainingFields

		// OK
		return nil
	}

	// The struct doesn't embed ExtensionProps
	// This is ok if no extensions found
	if len(remainingFields) == 0 {
		return nil
	}

	// Unsupported extensions found
	keys := make([]string, 0, len(remainingFields))
	for k := range remainingFields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return fmt.Errorf("Unsupported properties: '%v'\nSupported properties are: %s",
		strings.Join(keys, "', '"),
		typeInfo.fieldNamesString)
}
