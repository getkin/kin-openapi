package openapi2_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi2"
)

func TestSchemaRef_UnmarshalJSON_Ref(t *testing.T) {
	data := []byte(`{"$ref":"#/definitions/Pet"}`)

	var ref openapi2.SchemaRef
	require.NoError(t, json.Unmarshal(data, &ref))

	assert.Equal(t, "#/definitions/Pet", ref.Ref)
	assert.Nil(t, ref.Value)
}

func TestSchemaRef_UnmarshalJSON_RefWithExtensions(t *testing.T) {
	data := []byte(`{"$ref":"#/definitions/Pet","x-order":1,"something":"extra"}`)

	var ref openapi2.SchemaRef
	require.NoError(t, json.Unmarshal(data, &ref))

	assert.Equal(t, "#/definitions/Pet", ref.Ref)
	assert.Equal(t, float64(1), ref.Extensions["x-order"])
	assert.Nil(t, ref.Extensions["something"])
}

func TestSchemaRef_UnmarshalJSON_Value(t *testing.T) {
	data := []byte(`{"type":"string"}`)

	var ref openapi2.SchemaRef
	require.NoError(t, json.Unmarshal(data, &ref))

	assert.Empty(t, ref.Ref)
	require.NotNil(t, ref.Value)
	require.NotNil(t, ref.Value.Type)
	assert.Contains(t, *ref.Value.Type, "string")
}
