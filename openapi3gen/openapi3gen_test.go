package openapi3gen

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type CyclicType0 struct {
	CyclicField *CyclicType1 `json:"a"`
}
type CyclicType1 struct {
	CyclicField *CyclicType0 `json:"b"`
}

func TestCyclic(t *testing.T) {
	schema, refsMap, err := NewSchemaRefForValue(&CyclicType0{})
	require.IsType(t, &CycleError{}, err)
	require.Nil(t, schema)
	require.Empty(t, refsMap)
}
