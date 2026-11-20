package openapi3filter_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
)

// Distinct deepObject keys ("p[a]" and "p[a]b]") may name the same property:
// the outcome must not depend on the random order of the query map, the
// last key in sorted order winning.
func TestDeepObjectKeysNamingTheSamePropertyAreDeterministic(t *testing.T) {
	explode := true
	param := &openapi3.Parameter{
		Name: "p", In: openapi3.ParameterInQuery, Style: "deepObject", Explode: &explode,
		Schema: openapi3.NewSchemaRef("", openapi3.NewObjectSchema().
			WithProperty("a", openapi3.NewIntegerSchema())),
	}
	// Map iteration order is random: repeat so both orders are tried.
	for range 50 {
		req, err := http.NewRequest("GET", "http://example.com/x?p[a]=%10&p[a]b]=2", nil)
		require.NoError(t, err)
		input := &openapi3filter.RequestValidationInput{Request: req, Route: &routers.Route{}}
		require.NoError(t, openapi3filter.ValidateParameter(t.Context(), input, param))
	}
}
