package openapi3gen

import (
	"reflect"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

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

func TestCircularReferences(t *testing.T) {
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