package jsoninfo

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

var (
	typeInfos      = map[reflect.Type]*TypeInfo{}
	typeInfosMutex sync.RWMutex
)

// TypeInfo contains information about JSON serialization of a type
type TypeInfo struct {
	MultipleFields   bool // Whether multiple Go fields share the same JSON name
	Type             reflect.Type
	Extensions       []int
	Fields           []FieldInfo
	typeExtensions   []typeExtension
	Schema           interface{}
	SchemaMutex      sync.RWMutex
	fieldNamesString string // For debug messages
}

type typeExtension struct {
	factory   func() Extension
	jsonNames []string
}

func (typeInfo *TypeInfo) AddExtensionFactory(f func() Extension) {
	example := f()
	fields := GetTypeInfoForValue(example).Fields
	jsonNames := make([]string, len(fields))
	for i, field := range fields {
		jsonNames[i] = field.JSONName
	}
	ext := typeExtension{
		factory:   f,
		jsonNames: jsonNames,
	}
	typeInfo.typeExtensions = append(typeInfo.typeExtensions, ext)
}

// FieldInfo contains information about JSON serialization of a field.
type FieldInfo struct {
	MultipleFields     bool // Whether multiple Go fields share this JSON name
	Index              []int
	JSONName           string
	JSONOmitEmpty      bool
	JSONString         bool
	JSONNoRef          bool
	TypeIsMarshaller   bool
	TypeIsUnmarshaller bool
	Type               reflect.Type
}

type sortableFieldInfos []FieldInfo

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

func GetTypeInfoForValue(value interface{}) *TypeInfo {
	return GetTypeInfo(reflect.TypeOf(value))
}

// GetTypeInfo returns TypeInfo for the given type.
func GetTypeInfo(t reflect.Type) *TypeInfo {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	typeInfosMutex.RLock()
	typeInfo, exists := typeInfos[t]
	typeInfosMutex.RUnlock()
	if exists {
		return typeInfo
	}
	if t.Kind() != reflect.Struct {
		typeInfo = &TypeInfo{
			Type:             t,
			fieldNamesString: fmt.Sprintf("(%s is not a struct)", t.String()),
		}
	} else {
		// Allocate
		typeInfo = &TypeInfo{
			Type:   t,
			Fields: make([]FieldInfo, 0, 16),
		}

		// Add fields
		typeInfo.addFields(nil, t)

		// Sort fields
		sort.Sort(sortableFieldInfos(typeInfo.Fields))

		fields := typeInfo.Fields
		fieldNames := make([]string, len(fields))
		for i, field := range fields {
			fieldNames[i] = field.JSONName
		}
		fieldNamesString := "-"
		if len(fieldNames) > 0 {
			fieldNamesString = "'" + strings.Join(fieldNames, "', '") + "'"
		}
		typeInfo.fieldNamesString = fieldNamesString
	}

	// Publish
	typeInfosMutex.Lock()
	typeInfos[t] = typeInfo
	typeInfosMutex.Unlock()
	return typeInfo
}

func (typeInfo *TypeInfo) addFields(parentIndex []int, t reflect.Type) {
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
			if f.Tag.Get("json") == "-" {
				continue
			}
			ft := f.Type
			if ft == extensionPropsType {
				typeInfo.Extensions = index
			} else {
				typeInfo.addFields(index, ft)
			}
			continue iteration
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
		field := FieldInfo{
			Index:    index,
			Type:     f.Type,
			JSONName: f.Name,
		}

		// Read "json" tag
		jsonTag := f.Tag.Get("json")

		// Read our custom "multijson" tag that
		// allows multiple fields with the same name.
		if v := f.Tag.Get("multijson"); len(v) > 0 {
			field.MultipleFields = true
			jsonTag = v
		}

		// Handle "-"
		if jsonTag == "-" {
			continue
		}

		// Parse the tag
		if len(jsonTag) > 0 {
			for i, part := range strings.Split(jsonTag, ",") {
				if i == 0 {
					if len(part) > 0 {
						field.JSONName = part
					}
				} else {
					switch part {
					case "omitempty":
						field.JSONOmitEmpty = true
					case "string":
						field.JSONString = true
					case "noref":
						field.JSONNoRef = true
					}
				}
			}
		}

		if _, ok := field.Type.MethodByName("MarshalJSON"); ok {
			field.TypeIsMarshaller = true
		}
		if _, ok := field.Type.MethodByName("UnmarshalJSON"); ok {
			field.TypeIsUnmarshaller = true
		}

		// Field is done
		typeInfo.Fields = append(typeInfo.Fields, field)
	}
}

var extensionPropsType = reflect.TypeOf(ExtensionProps{})
