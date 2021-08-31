package openapi3gen

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

type CyclicType0 struct {
	CyclicField *CyclicType1 `json:"a"`
}
type CyclicType1 struct {
	CyclicField *CyclicType0 `json:"b"`
}

func TestCyclic(t *testing.T) {
	schemaRef, refsMap, err := NewSchemaRefForValue(&CyclicType0{}, ThrowErrorOnCycle())
	require.IsType(t, &CycleError{}, err)
	require.Nil(t, schemaRef)
	require.Empty(t, refsMap)
}

func TestExportedNonTagged(t *testing.T) {
	type Bla struct {
		A          string
		Another    string `json:"another"`
		yetAnother string // unused because unexported
		EvenAYaml  string `yaml:"even_a_yaml"`
	}

	schemaRef, _, err := NewSchemaRefForValue(&Bla{}, UseAllExportedFields())
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

	schemaRef, _, err := NewSchemaRefForValue(&UnsignedIntStruct{}, UseAllExportedFields())
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
	type Bla struct {
		UntaggedStringField string
		AnonStruct          struct {
			InnerFieldWithoutTag int
			InnerFieldWithTag    int `mymintag:"-1" mymaxtag:"50"`
		}
		EnumField string `json:"another" myenumtag:"a,b"`
	}

	schemaRef, _, err := NewSchemaRefForValue(&Bla{}, UseAllExportedFields(), SchemaCustomizer(func(name string, ft reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
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
        }
      },
      "type": "object"
    },
    "UntaggedStringField": {
      "type": "string"
    },
    "another": {
      "enum": [
        "a",
        "b"
      ],
      "type": "string"
    }
  },
  "type": "object"
}`, string(jsonSchema))
}

func TestSchemaCustomizerError(t *testing.T) {
	type Bla struct{}
	_, _, err := NewSchemaRefForValue(&Bla{}, UseAllExportedFields(), SchemaCustomizer(func(name string, ft reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
		return errors.New("test error")
	}))
	require.EqualError(t, err, "test error")
}
