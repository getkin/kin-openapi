package openapi3

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAllOfErrorPreserved(t *testing.T) {

	SchemaErrorDetailsDisabled = true
	defer func() { SchemaErrorDetailsDisabled = false }()

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
	err = s.Validate(context.Background())
	require.NoError(t, err)

	obj := make(map[string]any)
	err = json.Unmarshal([]byte(raw), &obj)
	require.NoError(t, err)

	err = s.VisitJSON(obj, MultiErrors())
	require.Error(t, err)

	var multiError MultiError
	ok := errors.As(err, &multiError)
	require.True(t, ok)
	var schemaErr *SchemaError
	ok = errors.As(multiError[0], &schemaErr)
	require.True(t, ok)

	require.Equal(t, "allOf", schemaErr.SchemaField)
	require.Equal(t, `doesn't match all schemas from "allOf"`, schemaErr.Reason)
	require.Equal(t, `Error at "/foo": doesn't match schema due to: Error at "/key1": value must be a number And Error at "/key2": value must be a string`, schemaErr.Error())

	var me multiErrorForAllOf
	ok = errors.As(err, &me)
	require.True(t, ok)
	require.Equal(t, `Error at "/foo/key1": value must be a number`, me[0].Error())
	require.Equal(t, `Error at "/foo/key2": value must be a string`, me[1].Error())
}
