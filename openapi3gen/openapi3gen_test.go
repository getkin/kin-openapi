package openapi3gen

import (
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
	schemaRef, refsMap, err := NewSchemaRefForValue(&CyclicType0{})
	require.IsType(t, &CycleError{}, err)
	require.Nil(t, schemaRef)
	require.Empty(t, refsMap)
}

func TestExportedNonTagged(t *testing.T) {
	type Bla struct {
		A          string
		Another    string `json:"another"`
		yetAnother string
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
