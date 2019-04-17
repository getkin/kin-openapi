package openapi3filter

import (
	"net/http"
	"net/url"

	"github.com/getkin/kin-openapi/openapi3"
)

// This function takes a parameter definition from the swagger spec, and
// the value which we received for it. It is expected to return the
// value unmarshaled into an interface which can be traversed for
// validation, it should also return the content type, eg (application/json)
// as a string, so that we know which schema to use for validation.
// If a query parameter appears multiple times, values[] will have more
// than one  value, but for all other parameter types it should have just
// one.
type ContentParameterDecoder func(param *openapi3.Parameter, values []string) (interface{}, string, error)

type RequestValidationInput struct {
	Request      *http.Request
	PathParams   map[string]string
	QueryParams  url.Values
	Route        *Route
	Options      *Options
	ParamDecoder ContentParameterDecoder
}

func (input *RequestValidationInput) GetQueryParams() url.Values {
	q := input.QueryParams
	if q == nil {
		q = input.Request.URL.Query()
		input.QueryParams = q
	}
	return q
}
