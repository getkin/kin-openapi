package openapi3

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAllOfErrorPreserved(t *testing.T) {
	// language=json
	raw := `
{
	"foo": {
		"key1": "foo",
		"key2": 1
	}
}
`

	// language=json
	schema := `
{
	"type": "object",
	"properties": {
		"foo": {
			"allOf": [
				{
					"type": "object",
					"properties": {
						"key1": {
							"type": "number"
						}
					}
				},
				{
					"type": "object",
					"properties": {
						"key2": {
							"type": "string"
						}
					}
				}
			]
		}
	}
}
`

	s := NewSchema()
	err := s.UnmarshalJSON([]byte(schema))
	require.NoError(t, err)
	err = s.Validate(t.Context())
	require.NoError(t, err)

	obj := make(map[string]any)
	err = json.Unmarshal([]byte(raw), &obj)
	require.NoError(t, err)

	err = s.VisitJSON(obj, MultiErrors(), DisableSchemaErrorDetails())
	require.Error(t, err)

	var multiError MultiError
	ok := errors.As(err, &multiError)
	require.True(t, ok)
	var schemaErr *SchemaError
	ok = errors.As(multiError[0], &schemaErr)
	require.True(t, ok)

	require.Equal(t, "allOf", schemaErr.SchemaField)
	require.Equal(t, `doesn't match all schemas from "allOf"`, schemaErr.Reason)
	require.EqualError(t, schemaErr, `Error at "/foo": doesn't match schema due to: Error at "/key1": value must be a number And Error at "/key2": value must be a string`)

	var me multiErrorForAllOf
	ok = errors.As(err, &me)
	require.True(t, ok)
	require.EqualError(t, me[0], `Error at "/foo/key1": value must be a number`)
	require.EqualError(t, me[1], `Error at "/foo/key2": value must be a string`)
}
