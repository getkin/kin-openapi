package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// A value breaking several properties at once is reported with the same
// error, in the same order, however the validator's maps happen to be walked.
func TestJSONSchema2020_CausesOfAnErrorAreInADeterministicOrder(t *testing.T) {
	t.Parallel()

	schema := openapi3.NewObjectSchema()
	schema.Properties = openapi3.Schemas{}
	for _, name := range []string{"a", "b", "c", "d", "e", "f"} {
		schema.Properties[name] = openapi3.NewSchemaRef("", openapi3.NewStringSchema())
	}
	value := map[string]any{"a": nil, "b": nil, "c": nil, "d": nil, "e": nil, "f": nil}

	first := schema.VisitJSON(value, openapi3.EnableJSONSchema2020())
	require.Error(t, first)
	firstE := first.Error()
	// Map iteration order is random: repeat so several orders are tried.
	for range 100 {
		err := schema.VisitJSON(value, openapi3.EnableJSONSchema2020())
		e := err.Error()
		require.Equal(t, firstE, e)
	}
}
