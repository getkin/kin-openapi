package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestNullableNullStillObeysEnumAndNot(t *testing.T) {
	t.Parallel()

	schema := openapi3.NewStringSchema().WithNullable().WithEnum("ON")
	err := schema.VisitJSON(nil)
	require.ErrorContains(t, err, "enum")

	schema.Enum = append(schema.Enum, nil)
	require.NoError(t, schema.VisitJSON(nil))
	require.NoError(t, schema.VisitJSON("ON"))

	schema = openapi3.NewStringSchema().WithNullable()
	schema.Not = &openapi3.SchemaRef{Value: openapi3.NewStringSchema().WithNullable()}
	err = schema.VisitJSON(nil)
	require.ErrorContains(t, err, "not")

	require.NoError(t, openapi3.NewStringSchema().WithNullable().VisitJSON(nil))
}
