package openapi3gen

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

type CyclicType0 struct {
	CyclicField *CyclicType1 `json:"a"`
}
type CyclicType1 struct {
	CyclicField *CyclicType0 `json:"b"`
}

func TestSimpleStruct(t *testing.T) {
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

	g := NewGenerator()
	schemaRef, err := g.newSchemaRefForValue(&SomeStruct{}, nil)
	require.NoError(t, err)
	require.Len(t, g.SchemaRefs, 15)

	schemaJSON, err := json.Marshal(schemaRef)
	require.NoError(t, err)

	require.JSONEq(t, `
	{
	  "properties": {
	    "bool": {
	      "type": "boolean"
	    },
	    "bytes": {
	      "format": "byte",
	      "type": "string"
	    },
	    "float64": {
	      "format": "double",
	      "type": "number"
	    },
	    "int": {
	      "type": "integer"
	    },
	    "int64": {
	      "format": "int64",
	      "type": "integer"
	    },
	    "json": {},
	    "map": {
	      "additionalProperties": {
	        "type": "string"
	      },
	      "type": "object"
	    },
	    "ptr": {
	      "type": "string"
	    },
	    "slice": {
	      "items": {
	        "type": "string"
	      },
	      "type": "array"
	    },
	    "string": {
	      "type": "string"
	    },
	    "struct": {
	      "properties": {
	        "x": {
	          "type": "string"
	        }
	      },
	      "type": "object"
	    },
	    "structWithoutFields": {},
	    "time": {
	      "format": "date-time",
	      "type": "string"
	    }
	  },
	  "type": "object"
	}
	`, string(schemaJSON))

}

func TestCyclic(t *testing.T) {
	schemas := make(openapi3.Schemas)
	schemaRef, err := NewSchemaRefForValue(&CyclicType0{}, schemas, ThrowErrorOnCycle())
	require.Error(t, err)
	require.IsType(t, &CycleError{}, err)
	require.Empty(t, schemas)

	schemaRef, err = NewSchemaRefForValue(&CyclicType0{}, schemas)
	require.NoError(t, err)
	schemaRefForCyclicType0 := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: "object",
			Properties: openapi3.Schemas{
				"a": {
					Value: &openapi3.Schema{
						Type: "object",
						Properties: openapi3.Schemas{
							"b": {Ref: "#/components/schemas/CyclicType0"},
						},
					},
				},
			},
		},
	}
	require.Equal(t, schemaRefForCyclicType0, schemaRef)
	require.Equal(t, openapi3.Schemas{
		"CyclicType0": schemaRefForCyclicType0,
	}, schemas)
}

func TestExportedNonTagged(t *testing.T) {
	type Bla struct {
		A          string
		Another    string `json:"another"`
		yetAnother string // unused because unexported
		EvenAYaml  string `yaml:"even_a_yaml"`
	}

	schemaRef, err := NewSchemaRefForValue(&Bla{}, nil, UseAllExportedFields())
	require.NoError(t, err)
	require.Equal(t, &openapi3.SchemaRef{Value: &openapi3.Schema{
		Type: "object",
		Properties: map[string]*openapi3.SchemaRef{
			"A":           {Value: &openapi3.Schema{Type: "string"}},
			"another":     {Value: &openapi3.Schema{Type: "string"}},
			"even_a_yaml": {Value: &openapi3.Schema{Type: "string"}},
		}}}, schemaRef)
}

func TestExportUint(t *testing.T) {
	type UnsignedIntStruct struct {
		UnsignedInt uint `json:"uint"`
	}

	schemaRef, err := NewSchemaRefForValue(&UnsignedIntStruct{}, nil, UseAllExportedFields())
	require.NoError(t, err)
	require.Equal(t, &openapi3.SchemaRef{Value: &openapi3.Schema{
		Type: "object",
		Properties: map[string]*openapi3.SchemaRef{
			"uint": {Value: &openapi3.Schema{Type: "integer", Min: &zeroInt}},
		}}}, schemaRef)
}

func TestEmbeddedStructs(t *testing.T) {
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

	generator := NewGenerator(UseAllExportedFields())

	schemaRef, err := generator.GenerateSchemaRef(reflect.TypeOf(instance))
	require.NoError(t, err)

	var ok bool
	_, ok = schemaRef.Value.Properties["Name"]
	require.Equal(t, true, ok)

	_, ok = schemaRef.Value.Properties["ID"]
	require.Equal(t, true, ok)
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

	generator := NewGenerator(UseAllExportedFields())

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

	generator := NewGenerator(UseAllExportedFields())

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

func TestSchemaCustomizer(t *testing.T) {
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

	schemaRef, err := NewSchemaRefForValue(&Bla{}, nil, UseAllExportedFields(), SchemaCustomizer(func(name string, ft reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
		t.Logf("Field=%s,Tag=%s", name, tag)
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
	}))
	require.NoError(t, err)
	jsonSchema, err := json.MarshalIndent(schemaRef, "", "  ")
	require.NoError(t, err)
	require.JSONEq(t, `{
  "properties": {
    "AnonStruct": {
      "properties": {
        "InnerFieldWithTag": {
          "maximum": 50,
          "minimum": -1,
          "type": "integer"
        },
        "InnerFieldWithoutTag": {
          "type": "integer"
        },
				"enum1": {
					"enum": [
						"a",
						"b"
					],
					"type": "string"
				}
      },
      "type": "object"
    },
    "UntaggedStringField": {
      "type": "string"
    },
    "enum2": {
      "enum": [
        "c",
        "d"
      ],
      "type": "string"
    },
    "enum3": {
      "enum": [
        "e",
        "f"
      ],
      "type": "string"
    }
  },
  "type": "object"
}`, string(jsonSchema))
}

func TestSchemaCustomizerError(t *testing.T) {
	type Bla struct{}
	_, err := NewSchemaRefForValue(&Bla{}, nil, UseAllExportedFields(), SchemaCustomizer(func(name string, ft reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
		return errors.New("test error")
	}))
	require.EqualError(t, err, "test error")
}

func TestRecursiveSchema(t *testing.T) {

	type RecursiveType struct {
		Field1     string           `json:"field1"`
		Field2     string           `json:"field2"`
		Field3     string           `json:"field3"`
		Components []*RecursiveType `json:"children,omitempty"`
	}

	schemas := make(openapi3.Schemas)
	schemaRef, err := NewSchemaRefForValue(&RecursiveType{}, schemas)
	require.NoError(t, err)

	jsonSchemas, err := json.MarshalIndent(&schemas, "", "  ")
	require.NoError(t, err)

	jsonSchemaRef, err := json.MarshalIndent(&schemaRef, "", "  ")
	require.NoError(t, err)

	require.JSONEq(t, `{
		"RecursiveType": {
			"properties": {
				"children": {
					"items": {
						"$ref": "#/components/schemas/RecursiveType"
					},
					"type": "array"
				},
				"field1": {
					"type": "string"
				},
				"field2": {
					"type": "string"
				},
				"field3": {
					"type": "string"
				}
			},
			"type": "object"
		}
	}`, string(jsonSchemas))

	require.JSONEq(t, `{
		"properties": {
			"children": {
				"items": {
					"$ref": "#/components/schemas/RecursiveType"
				},
				"type": "array"
			},
			"field1": {
				"type": "string"
			},
			"field2": {
				"type": "string"
			},
			"field3": {
				"type": "string"
			}
		},
		"type": "object"
	}`, string(jsonSchemaRef))

}
