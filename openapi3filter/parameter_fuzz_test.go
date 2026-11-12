package openapi3filter_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
)

var (
	paramLocations = []string{openapi3.ParameterInPath, openapi3.ParameterInQuery, openapi3.ParameterInHeader, openapi3.ParameterInCookie}
	paramStyles    = []string{"", "simple", "form", "label", "matrix", "spaceDelimited", "pipeDelimited", "deepObject"}
	paramTypes     = []string{"string", "integer", "number", "boolean", "array", "object"}
)

// FuzzValidateParameter validates an arbitrary raw value against a parameter
// of every location, style, explode setting and schema type: decoding never
// panics, and is deterministic.
func FuzzValidateParameter(f *testing.F) {
	f.Add(uint8(1), uint8(2), uint8(4), true, "a=1&a=2", "a")
	f.Add(uint8(0), uint8(3), uint8(4), false, ".1.2.3", "id")
	f.Add(uint8(0), uint8(4), uint8(5), true, ";a=1;b=2", "id")
	f.Add(uint8(1), uint8(7), uint8(5), true, "p[a]=1&p[b]=2", "p")
	f.Add(uint8(2), uint8(1), uint8(4), false, "1,2,3", "X-Ids")
	f.Add(uint8(3), uint8(2), uint8(0), false, "v", "c")

	f.Fuzz(func(t *testing.T, loc, style, typ uint8, explode bool, raw, name string) {
		if len(raw) > 4<<10 || len(name) > 256 || name == "" {
			t.Skip()
		}
		in := paramLocations[int(loc)%len(paramLocations)]
		param := &openapi3.Parameter{
			Name:     name,
			In:       in,
			Style:    paramStyles[int(style)%len(paramStyles)],
			Explode:  &explode,
			Required: true,
			Schema: &openapi3.SchemaRef{Value: &openapi3.Schema{
				Type:  &openapi3.Types{paramTypes[int(typ)%len(paramTypes)]},
				Items: openapi3.NewSchemaRef("", openapi3.NewIntegerSchema()),
				Properties: openapi3.Schemas{
					"a": openapi3.NewSchemaRef("", openapi3.NewIntegerSchema()),
					"b": openapi3.NewSchemaRef("", openapi3.NewStringSchema()),
				},
			}},
		}
		if in == openapi3.ParameterInPath {
			// A path parameter is always required, and its style is limited.
			param.Required = true
		}

		req, err := http.NewRequest("GET", "http://example.com/x", nil)
		if err != nil {
			t.Skip()
		}
		pathParams := map[string]string{}
		switch in {
		case openapi3.ParameterInPath:
			pathParams[name] = raw
		case openapi3.ParameterInQuery:
			req.URL.RawQuery = raw
		case openapi3.ParameterInHeader:
			req.Header[http.CanonicalHeaderKey(name)] = []string{raw}
		case openapi3.ParameterInCookie:
			req.AddCookie(&http.Cookie{Name: name, Value: url.QueryEscape(raw)})
		}

		input := &openapi3filter.RequestValidationInput{
			Request:    req,
			PathParams: pathParams,
			Route:      &routers.Route{},
		}
		first := openapi3filter.ValidateParameter(t.Context(), input, param)
		second := openapi3filter.ValidateParameter(t.Context(), input, param)
		require.Equal(t, first == nil, second == nil, "validation is not deterministic: %v then %v", first, second)
	})
}
