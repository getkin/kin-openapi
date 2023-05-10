package openapi3gen

import (
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

// theFieldInfo contains information about JSON serialization of a field.
type theFieldInfo struct {
	HasJSONTag         bool
	TypeIsMarshaller   bool
	TypeIsUnmarshaller bool
	JSONOmitEmpty      bool
	JSONString         bool
	Index              []int
	Type               reflect.Type
	JSONName           string
}

func appendFields(fields []theFieldInfo, parentIndex []int, t reflect.Type) []theFieldInfo {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return fields
	}

	// For each field
	numField := t.NumField()
iteration:
	for i := 0; i < numField; i++ {
		f := t.Field(i)
		index := make([]int, 0, len(parentIndex)+1)
		index = append(index, parentIndex...)
		index = append(index, i)

		// See whether this is an embedded field
		if f.Anonymous {
			jsonTag := f.Tag.Get("json")
			if jsonTag == "-" {
				continue
			}
			if jsonTag == "" {
				fields = appendFields(fields, index, f.Type)
				continue iteration
			}
		}

		// Ignore certain types
		switch f.Type.Kind() {
		case reflect.Func, reflect.Chan:
			continue iteration
		}

		// Is it a private (lowercase) field?
		firstRune, _ := utf8.DecodeRuneInString(f.Name)
		if unicode.IsLower(firstRune) {
			continue iteration
		}

		// Declare a field
		field := theFieldInfo{
			Index:    index,
			Type:     f.Type,
			JSONName: f.Name,
		}

		// Read "json" tag
		jsonTag := f.Tag.Get("json")

		// Handle "-"
		if jsonTag == "-" {
			continue
		}

		// Parse the tag
		if jsonTag != "" {
			field.HasJSONTag = true
			for i, part := range strings.Split(jsonTag, ",") {
				if i == 0 {
					if part != "" {
						field.JSONName = part
					}
				} else {
					switch part {
					case "omitempty":
						field.JSONOmitEmpty = true
					case "string":
						field.JSONString = true
					}
				}
			}
		}

		_, field.TypeIsMarshaller = field.Type.MethodByName("MarshalJSON")
		_, field.TypeIsUnmarshaller = field.Type.MethodByName("UnmarshalJSON")

		// Field is done
		fields = append(fields, field)
	}

	return fields
}

type sortableFieldInfos []theFieldInfo

func (list sortableFieldInfos) Len() int {
	return len(list)
}

func (list sortableFieldInfos) Less(i, j int) bool {
	return list[i].JSONName < list[j].JSONName
}

func (list sortableFieldInfos) Swap(i, j int) {
	a, b := list[i], list[j]
	list[i], list[j] = b, a
}
