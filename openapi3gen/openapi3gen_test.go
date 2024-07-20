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

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
)

func ExampleGenerator_SchemaRefs() {
	type SomeOtherType string
	type Embedded struct {
		Z string `json:"z"`
	}
	type Embedded2 struct {
		A string `json:"a"`
	}
	type EmbeddedNonStruct string
	type EmbeddedNonStructPtr string
	type Embedded3 struct {
		EmbeddedNonStruct
		*EmbeddedNonStructPtr
	}
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

		Embedded `json:"embedded"`

		Embedded2

		Embedded3 `json:"embedded3"`

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
	// g.SchemaRefs: 17
	// schemaRef: {
	//   "properties": {
	//     "a": {
	//       "type": "string"
	//     },
	//     "bool": {
	//       "type": "boolean"
	//     },
	//     "bytes": {
	//       "format": "byte",
	//       "type": "string"
	//     },
	//     "embedded": {
	//       "properties": {
	//         "z": {
	//           "type": "string"
	//         }
	//       },
	//       "type": "object"
	//     },
	//     "embedded3": {},
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
	//         "nullable": true,
	//         "type": "string"
	//       },
	//       "type": "object"
	//     },
	//     "ptr": {
	//       "nullable": true,
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
	//       "nullable": true,
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
	//         "nullable": true,
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
		Type: &openapi3.Types{"object"},
		Properties: map[string]*openapi3.SchemaRef{
			"A":           {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
			"another":     {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
			"even_a_yaml": {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
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
	require.Equal(t, &openapi3.Types{"array"}, schemaRef.Value.Properties["SliceCycle"].Value.Type)
	require.Equal(t, "#/components/schemas/ObjectDiff", schemaRef.Value.Properties["SliceCycle"].Value.Items.Ref)

	require.NotNil(t, schemaRef.Value.Properties["MapCycle"])
	require.Equal(t, &openapi3.Types{"object"}, schemaRef.Value.Properties["MapCycle"].Value.Type)
	require.Equal(t, "#/components/schemas/ObjectDiff", schemaRef.Value.Properties["MapCycle"].Value.AdditionalProperties.Schema.Ref)
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

func TestSchemaCustomizerExcludeSchema(t *testing.T) {
	type Bla struct {
		Str string
	}

	customizer := openapi3gen.SchemaCustomizer(func(name string, ft reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
		return nil
	})
	schema, err := openapi3gen.NewSchemaRefForValue(&Bla{}, nil, openapi3gen.UseAllExportedFields(), customizer)
	require.NoError(t, err)
	require.Equal(t, &openapi3.SchemaRef{Value: &openapi3.Schema{
		Type: &openapi3.Types{"object"},
		Properties: map[string]*openapi3.SchemaRef{
			"Str": {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
		}}}, schema)

	customizer = openapi3gen.SchemaCustomizer(func(name string, ft reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
		return &openapi3gen.ExcludeSchemaSentinel{}
	})
	schema, err = openapi3gen.NewSchemaRefForValue(&Bla{}, nil, openapi3gen.UseAllExportedFields(), customizer)
	require.NoError(t, err)
	require.Nil(t, schema)
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

type ID [16]byte

// T implements SetSchemar, allowing it to set an OpenAPI schema.
type T struct {
	ID ID `json:"id"`
}

func (_ *ID) SetSchema(schema *openapi3.Schema) {
	schema.Type = &openapi3.Types{"string"} // Assuming this matches your custom implementation
	schema.Format = "uuid"
}

func ExampleSetSchemar() {
	schemas := make(openapi3.Schemas)
	instance := &T{
		ID: ID{},
	}

	// Generate the schema for the instance
	schemaRef, err := openapi3gen.NewSchemaRefForValue(instance, schemas)
	if err != nil {
		panic(err)
	}
	data, err := json.MarshalIndent(schemaRef, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("schemaRef: %s\n", data)
	// Output:
	// schemaRef: {
	//   "properties": {
	//     "id": {
	//       "format": "uuid",
	//       "type": "string"
	//     }
	//   },
	//   "type": "object"
	// }
}
