package openapi3filter

import (
	"net/http"
	"net/url"
)

type RequestValidationInput struct {
	Request     *http.Request
	PathParams  map[string]string
	QueryParams url.Values
	Route       *Route
	Options     *Options
}

func (input *RequestValidationInput) GetQueryParams() url.Values {
	q := input.QueryParams
	if q == nil {
		q = input.Request.URL.Query()
		input.QueryParams = q
	}
	return q
}
