package openapi3

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue746(t *testing.T) {
	schema := &Schema{}
	err := schema.UnmarshalJSON([]byte(`{"additionalProperties": false}`))
	require.NoError(t, err)

	var value any
	err = json.Unmarshal([]byte(`{"foo": "bar"}`), &value)
	require.NoError(t, err)

	err = schema.VisitJSON(value)
	require.Error(t, err)

	schemaErr := &SchemaError{}
	require.ErrorAs(t, err, &schemaErr)
	require.Equal(t, "properties", schemaErr.SchemaField)
	require.Equal(t, `property "foo" is unsupported`, schemaErr.Reason)
}
