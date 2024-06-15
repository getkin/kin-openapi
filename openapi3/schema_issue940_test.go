package openapi3

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOneOfErrorPreserved(t *testing.T) {

	SchemaErrorDetailsDisabled = true
	defer func() { SchemaErrorDetailsDisabled = false }()

	// language=json
	raw := `
{
	"foo": [ "bar" ]
}
`

	// language=json
	schema := `
{
	"type": "object",
	"properties": {
		"foo": {
			"oneOf": [
				{
					"type": "number"
				},
					{
					"type": "string"
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

	require.Equal(t, "oneOf", schemaErr.SchemaField)
	require.Equal(t, `value doesn't match any schema from "oneOf"`, schemaErr.Reason)
	require.Equal(t, `Error at "/foo": doesn't match schema due to: value must be a number Or value must be a string`, schemaErr.Error())

	var me multiErrorForOneOf
	ok = errors.As(err, &me)
	require.True(t, ok)
	require.Equal(t, `Error at "/foo": value must be a number`, me[0].Error())
	require.Equal(t, `Error at "/foo": value must be a string`, me[1].Error())
}
