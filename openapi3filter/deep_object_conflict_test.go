package openapi3filter_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
)

// A deepObject property that is both a value and the parent of other
// properties is an invalid request, whichever way the query is walked.
func TestDeepObjectPropertyBothValueAndObject(t *testing.T) {
	explode := true
	param := &openapi3.Parameter{
		Name: "p", In: openapi3.ParameterInQuery, Style: "deepObject", Explode: &explode,
		Schema: openapi3.NewSchemaRef("", openapi3.NewObjectSchema().WithAnyAdditionalProperties()),
	}
	// Map iteration order is random: repeat so both orders are tried.
	for range 50 {
		req, err := http.NewRequest("GET", "http://example.com/x?p[a]=1&p[a][b]=2", nil)
		require.NoError(t, err)
		input := &openapi3filter.RequestValidationInput{Request: req, Route: &routers.Route{}}
		require.NotPanics(t, func() {
			require.Error(t, openapi3filter.ValidateParameter(t.Context(), input, param))
		})
	}
}
