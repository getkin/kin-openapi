package openapi3gen_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"github.com/stretchr/testify/require"
)

func ExampleGenerator_SchemaRefs() {
	type SomeOtherType string
	type SomeStruct struct {
		Bool    bool                      `json:"bool"`
		Int     int                       `json:"int"`
		Int64   int64                     `json:"int64"`
		Float64 float64                   `json:"float64"`
		String  string                    `json:"string"`
		Bytes   []byte                    `json:"bytes"`
		JSON    json.RawMessage           `json:"json"`
		Time    time.Time                 `json:"time"`
		Slice   []SomeOtherType           `json:"slice"`
		Map     map[string]*SomeOtherType `json:"map"`

		Struct struct {
			X string `json:"x"`
		} `json:"struct"`

		EmptyStruct struct {
			Y string
		} `json:"structWithoutFields"`

		Ptr *SomeOtherType `json:"ptr"`
	}

	g := openapi3gen.NewGenerator()
	schemaRef, err := g.NewSchemaRefForValue(&SomeStruct{}, nil)
	if err != nil {
		panic(err)
	}

	fmt.Printf("g.SchemaRefs: %d\n", len(g.SchemaRefs))
	var data []byte
	if data, err = json.MarshalIndent(&schemaRef, "", "  "); err != nil {
		panic(err)
	}
	fmt.Printf("schemaRef: %s\n", data)
	// Output:
	// g.SchemaRefs: 15
	// schemaRef: {
	//   "properties": {
	//     "bool": {
	//       "type": "boolean"
	//     },
	//     "bytes": {
	//       "format": "byte",
	//       "type": "string"
	//     },
	//     "float64": {
	//       "format": "double",
	//       "type": "number"
	//     },
	//     "int": {
	//       "type": "integer"
	//     },
	//     "int64": {
	//       "format": "int64",
	//       "type": "integer"
	//     },
	//     "json": {},
	//     "map": {
	//       "additionalProperties": {
	//         "type": "string"
	//       },
	//       "type": "object"
	//     },
	//     "ptr": {
	//       "type": "string"
	//     },
	//     "slice": {
	//       "items": {
	//         "type": "string"
	//       },
	//       "type": "array"
	//     },
	//     "string": {
	//       "type": "string"
	//     },
	//     "struct": {
	//       "properties": {
	//         "x": {
	//           "type": "string"
	//         }
	//       },
	//       "type": "object"
	//     },
	//     "structWithoutFields": {},
	//     "time": {
	//       "format": "date-time",
	//       "type": "string"
	//     }
	//   },
	//   "type": "object"
	// }
}

func ExampleThrowErrorOnCycle() {
	type CyclicType0 struct {
		CyclicField *struct {
			CyclicField *CyclicType0 `json:"b"`
		} `json:"a"`
	}

	schemas := make(openapi3.Schemas)
	schemaRef, err := openapi3gen.NewSchemaRefForValue(&CyclicType0{}, schemas, openapi3gen.ThrowErrorOnCycle())
	if schemaRef != nil || err == nil {
		panic(`With option ThrowErrorOnCycle, an error is returned when a schema reference cycle is found`)
	}
	if _, ok := err.(*openapi3gen.CycleError); !ok {
		panic(`With option ThrowErrorOnCycle, an error of type CycleError is returned`)
	}
	if len(schemas) != 0 {
		panic(`No references should have been collected at this point`)
	}

	if schemaRef, err = openapi3gen.NewSchemaRefForValue(&CyclicType0{}, schemas); err != nil {
		panic(err)
	}

	var data []byte
	if data, err = json.MarshalIndent(schemaRef, "", "  "); err != nil {
		panic(err)
	}
	fmt.Printf("schemaRef: %s\n", data)
	if data, err = json.MarshalIndent(schemas, "", "  "); err != nil {
		panic(err)
	}
	fmt.Printf("schemas: %s\n", data)
	// Output:
	// schemaRef: {
	//   "properties": {
	//     "a": {
	//       "properties": {
	//         "b": {
	//           "$ref": "#/components/schemas/CyclicType0"
	//         }
	//       },
	//       "type": "object"
	//     }
	//   },
	//   "type": "object"
	// }
	// schemas: {
	//   "CyclicType0": {
	//     "properties": {
	//       "a": {
	//         "properties": {
	//           "b": {
	//             "$ref": "#/components/schemas/CyclicType0"
	//           }
	//         },
	//         "type": "object"
	//       }
	//     },
	//     "type": "object"
	//   }
	// }
}

func TestExportedNonTagged(t *testing.T) {
	type Bla struct {
		A          string
		Another    string `json:"another"`
		yetAnother string // unused because unexported
		EvenAYaml  string `yaml:"even_a_yaml"`
	}

	schemaRef, err := openapi3gen.NewSchemaRefForValue(&Bla{}, nil, openapi3gen.UseAllExportedFields())
	require.NoError(t, err)
	require.Equal(t, &openapi3.SchemaRef{Value: &openapi3.Schema{
		Type: "object",
		Properties: map[string]*openapi3.SchemaRef{
			"A":           {Value: &openapi3.Schema{Type: "string"}},
			"another":     {Value: &openapi3.Schema{Type: "string"}},
			"even_a_yaml": {Value: &openapi3.Schema{Type: "string"}},
		}}}, schemaRef)
}

func ExampleUseAllExportedFields() {
	type UnsignedIntStruct struct {
		UnsignedInt uint `json:"uint"`
	}

	schemaRef, err := openapi3gen.NewSchemaRefForValue(&UnsignedIntStruct{}, nil, openapi3gen.UseAllExportedFields())
	if err != nil {
		panic(err)
	}

	var data []byte
	if data, err = json.MarshalIndent(schemaRef, "", "  "); err != nil {
		panic(err)
	}
	fmt.Printf("schemaRef: %s\n", data)
	// Output:
	// schemaRef: {
	//   "properties": {
	//     "uint": {
	//       "minimum": 0,
	//       "type": "integer"
	//     }
	//   },
	//   "type": "object"
	// }
}

func ExampleGenerator_GenerateSchemaRef() {
	type EmbeddedStruct struct {
		ID string
	}

	type ContainerStruct struct {
		Name string
		EmbeddedStruct
	}

	instance := &ContainerStruct{
		Name: "Container",
		EmbeddedStruct: EmbeddedStruct{
			ID: "Embedded",
		},
	}

	generator := openapi3gen.NewGenerator(openapi3gen.UseAllExportedFields())

	schemaRef, err := generator.GenerateSchemaRef(reflect.TypeOf(instance))
	if err != nil {
		panic(err)
	}

	var data []byte
	if data, err = json.MarshalIndent(schemaRef.Value.Properties["Name"].Value, "", "  "); err != nil {
		panic(err)
	}
	fmt.Printf(`schemaRef.Value.Properties["Name"].Value: %s`, data)
	fmt.Println()
	if data, err = json.MarshalIndent(schemaRef.Value.Properties["ID"].Value, "", "  "); err != nil {
		panic(err)
	}
	fmt.Printf(`schemaRef.Value.Properties["ID"].Value: %s`, data)
	fmt.Println()
	// Output:
	// schemaRef.Value.Properties["Name"].Value: {
	//   "type": "string"
	// }
	// schemaRef.Value.Properties["ID"].Value: {
	//   "type": "string"
	// }
}

func TestEmbeddedPointerStructs(t *testing.T) {
	type EmbeddedStruct struct {
		ID string
	}

	type ContainerStruct struct {
		Name string
		*EmbeddedStruct
	}

	instance := &ContainerStruct{
		Name: "Container",
		EmbeddedStruct: &EmbeddedStruct{
			ID: "Embedded",
		},
	}

	generator := openapi3gen.NewGenerator(openapi3gen.UseAllExportedFields())

	schemaRef, err := generator.GenerateSchemaRef(reflect.TypeOf(instance))
	require.NoError(t, err)

	var ok bool
	_, ok = schemaRef.Value.Properties["Name"]
	require.Equal(t, true, ok)

	_, ok = schemaRef.Value.Properties["ID"]
	require.Equal(t, true, ok)
}

// See: https://github.com/getkin/kin-openapi/issues/500
func TestEmbeddedPointerStructsWithSchemaCustomizer(t *testing.T) {
	type EmbeddedStruct struct {
		ID string
	}

	type ContainerStruct struct {
		Name string
		*EmbeddedStruct
	}

	instance := &ContainerStruct{
		Name: "Container",
		EmbeddedStruct: &EmbeddedStruct{
			ID: "Embedded",
		},
	}

	customizerFn := func(name string, t reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
		return nil
	}
	customizerOpt := openapi3gen.SchemaCustomizer(customizerFn)

	generator := openapi3gen.NewGenerator(openapi3gen.UseAllExportedFields(), customizerOpt)

	schemaRef, err := generator.GenerateSchemaRef(reflect.TypeOf(instance))
	require.NoError(t, err)

	var ok bool
	_, ok = schemaRef.Value.Properties["Name"]
	require.Equal(t, true, ok)

	_, ok = schemaRef.Value.Properties["ID"]
	require.Equal(t, true, ok)
}

func TestCyclicReferences(t *testing.T) {
	type ObjectDiff struct {
		FieldCycle *ObjectDiff
		SliceCycle []*ObjectDiff
		MapCycle   map[*ObjectDiff]*ObjectDiff
	}

	instance := &ObjectDiff{
		FieldCycle: nil,
		SliceCycle: nil,
		MapCycle:   nil,
	}

	generator := openapi3gen.NewGenerator(openapi3gen.UseAllExportedFields())

	schemaRef, err := generator.GenerateSchemaRef(reflect.TypeOf(instance))
	require.NoError(t, err)

	require.NotNil(t, schemaRef.Value.Properties["FieldCycle"])
	require.Equal(t, "#/components/schemas/ObjectDiff", schemaRef.Value.Properties["FieldCycle"].Ref)

	require.NotNil(t, schemaRef.Value.Properties["SliceCycle"])
	require.Equal(t, "array", schemaRef.Value.Properties["SliceCycle"].Value.Type)
	require.Equal(t, "#/components/schemas/ObjectDiff", schemaRef.Value.Properties["SliceCycle"].Value.Items.Ref)

	require.NotNil(t, schemaRef.Value.Properties["MapCycle"])
	require.Equal(t, "object", schemaRef.Value.Properties["MapCycle"].Value.Type)
	require.Equal(t, "#/components/schemas/ObjectDiff", schemaRef.Value.Properties["MapCycle"].Value.AdditionalProperties.Ref)
}

func ExampleSchemaCustomizer() {
	type NestedInnerBla struct {
		Enum1Field string `json:"enum1" myenumtag:"a,b"`
	}

	type InnerBla struct {
		UntaggedStringField string
		AnonStruct          struct {
			InnerFieldWithoutTag int
			InnerFieldWithTag    int `mymintag:"-1" mymaxtag:"50"`
			NestedInnerBla
		}
		Enum2Field string `json:"enum2" myenumtag:"c,d"`
	}

	type Bla struct {
		InnerBla
		EnumField3 string `json:"enum3" myenumtag:"e,f"`
	}

	customizer := openapi3gen.SchemaCustomizer(func(name string, ft reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
		if tag.Get("mymintag") != "" {
			minVal, err := strconv.ParseFloat(tag.Get("mymintag"), 64)
			if err != nil {
				return err
			}
			schema.Min = &minVal
		}
		if tag.Get("mymaxtag") != "" {
			maxVal, err := strconv.ParseFloat(tag.Get("mymaxtag"), 64)
			if err != nil {
				return err
			}
			schema.Max = &maxVal
		}
		if tag.Get("myenumtag") != "" {
			for _, s := range strings.Split(tag.Get("myenumtag"), ",") {
				schema.Enum = append(schema.Enum, s)
			}
		}
		return nil
	})

	schemaRef, err := openapi3gen.NewSchemaRefForValue(&Bla{}, nil, openapi3gen.UseAllExportedFields(), customizer)
	if err != nil {
		panic(err)
	}

	var data []byte
	if data, err = json.MarshalIndent(schemaRef, "", "  "); err != nil {
		panic(err)
	}
	fmt.Printf("schemaRef: %s\n", data)
	// Output:
	// schemaRef: {
	//   "properties": {
	//     "AnonStruct": {
	//       "properties": {
	//         "InnerFieldWithTag": {
	//           "maximum": 50,
	//           "minimum": -1,
	//           "type": "integer"
	//         },
	//         "InnerFieldWithoutTag": {
	//           "type": "integer"
	//         },
	//         "enum1": {
	//           "enum": [
	//             "a",
	//             "b"
	//           ],
	//           "type": "string"
	//         }
	//       },
	//       "type": "object"
	//     },
	//     "UntaggedStringField": {
	//       "type": "string"
	//     },
	//     "enum2": {
	//       "enum": [
	//         "c",
	//         "d"
	//       ],
	//       "type": "string"
	//     },
	//     "enum3": {
	//       "enum": [
	//         "e",
	//         "f"
	//       ],
	//       "type": "string"
	//     }
	//   },
	//   "type": "object"
	// }
}

func TestSchemaCustomizerError(t *testing.T) {
	customizer := openapi3gen.SchemaCustomizer(func(name string, ft reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
		return errors.New("test error")
	})

	type Bla struct{}
	_, err := openapi3gen.NewSchemaRefForValue(&Bla{}, nil, openapi3gen.UseAllExportedFields(), customizer)
	require.EqualError(t, err, "test error")
}

func ExampleNewSchemaRefForValue_recursive() {
	type RecursiveType struct {
		Field1     string           `json:"field1"`
		Field2     string           `json:"field2"`
		Field3     string           `json:"field3"`
		Components []*RecursiveType `json:"children,omitempty"`
	}

	schemas := make(openapi3.Schemas)
	schemaRef, err := openapi3gen.NewSchemaRefForValue(&RecursiveType{}, schemas)
	if err != nil {
		panic(err)
	}

	var data []byte
	if data, err = json.MarshalIndent(&schemas, "", "  "); err != nil {
		panic(err)
	}
	fmt.Printf("schemas: %s\n", data)
	if data, err = json.MarshalIndent(&schemaRef, "", "  "); err != nil {
		panic(err)
	}
	fmt.Printf("schemaRef: %s\n", data)
	// Output:
	// schemas: {
	//   "RecursiveType": {
	//     "properties": {
	//       "children": {
	//         "items": {
	//           "$ref": "#/components/schemas/RecursiveType"
	//         },
	//         "type": "array"
	//       },
	//       "field1": {
	//         "type": "string"
	//       },
	//       "field2": {
	//         "type": "string"
	//       },
	//       "field3": {
	//         "type": "string"
	//       }
	//     },
	//     "type": "object"
	//   }
	// }
	// schemaRef: {
	//   "properties": {
	//     "children": {
	//       "items": {
	//         "$ref": "#/components/schemas/RecursiveType"
	//       },
	//       "type": "array"
	//     },
	//     "field1": {
	//       "type": "string"
	//     },
	//     "field2": {
	//       "type": "string"
	//     },
	//     "field3": {
	//       "type": "string"
	//     }
	//   },
	//   "type": "object"
	// }
}
