package openapi3filter_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
)

// A deepObject parameter whose name is not valid UTF-8 must not panic, and
// validate.
func TestDeepObjectParameterNameNotUTF8(t *testing.T) {
	for _, name := range []string{"p", "\xff", "p.q", "p(", "p|"} {
		t.Run(name, func(t *testing.T) {
			explode := true
			param := &openapi3.Parameter{
				Name: name, In: openapi3.ParameterInQuery, Style: "deepObject", Explode: &explode,
				Schema: &openapi3.SchemaRef{Value: &openapi3.Schema{
					Type:       &openapi3.Types{"object"},
					Properties: openapi3.Schemas{"a": openapi3.NewSchemaRef("", openapi3.NewIntegerSchema())},
					Required:   []string{"a"},
				}},
			}
			req, err := http.NewRequest("GET", "http://example.com/x", nil)
			require.NoError(t, err)
			req.URL.RawQuery = "a=1&" + escapeQueryKey(name+"[a]") + "=1"
			input := &openapi3filter.RequestValidationInput{Request: req, Route: &routers.Route{}}
			require.NotPanics(t, func() {
				require.NoError(t, openapi3filter.ValidateParameter(t.Context(), input, param))
			})
		})
	}
}

func escapeQueryKey(key string) string {
	var b []byte
	for i := 0; i < len(key); i++ {
		c := key[i]
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '.' || c == '[' || c == ']' {
			b = append(b, c)
		} else {
			b = append(b, '%', "0123456789ABCDEF"[c>>4], "0123456789ABCDEF"[c&15])
		}
	}
	return string(b)
}
