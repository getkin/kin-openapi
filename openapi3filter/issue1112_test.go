package openapi3filter

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func issue1112ObjectSchema(additional openapi3.AdditionalProperties) *openapi3.SchemaRef {
	return &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type:                 &openapi3.Types{"object"},
			Properties:           make(map[string]*openapi3.SchemaRef),
			AdditionalProperties: additional,
		},
	}
}

func issue1112Decode(t *testing.T, paramName, rawQuery string, schema *openapi3.SchemaRef) (any, bool, error) {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, "http://example.test/?"+rawQuery, nil)
	require.NoError(t, err)

	param := &openapi3.Parameter{
		Name:    paramName,
		In:      openapi3.ParameterInQuery,
		Style:   "deepObject",
		Explode: openapi3.Ptr(true),
		Schema:  schema,
	}

	return decodeStyledParameter(param, &RequestValidationInput{Request: req})
}

func TestIssue1112FreeFormDeepObject(t *testing.T) {
	schema := issue1112ObjectSchema(openapi3.AdditionalProperties{Has: openapi3.Ptr(true)})

	value, found, err := issue1112Decode(
		t,
		"properties",
		"properties[vaccinated]=true&properties[color]=black&properties[coat_length]=large",
		schema,
	)

	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, map[string]any{
		"vaccinated":  "true",
		"color":       "black",
		"coat_length": "large",
	}, value)
	require.NoError(t, schema.Value.VisitJSON(value))
}

func TestIssue1112SingleFreeFormProperty(t *testing.T) {
	schema := issue1112ObjectSchema(openapi3.AdditionalProperties{Has: openapi3.Ptr(true)})

	value, found, err := issue1112Decode(t, "properties", "properties[color]=black", schema)

	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, map[string]any{"color": "black"}, value)
}
