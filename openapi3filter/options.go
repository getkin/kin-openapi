package openapi3filter

import "github.com/getkin/kin-openapi/openapi3"

// DefaultOptions do not set an AuthenticationFunc.
// A spec with security schemes defined will not pass validation
// unless an AuthenticationFunc is defined.
var DefaultOptions = &Options{}

// Options used by ValidateRequest and ValidateResponse
type Options struct {
	// Set ExcludeRequestBody so ValidateRequest skips request body validation
	ExcludeRequestBody bool

	// Set ExcludeResponseBody so ValidateResponse skips response body validation
	ExcludeResponseBody bool

	// Set IncludeResponseStatus so ValidateResponse fails on response
	// status not defined in OpenAPI spec
	IncludeResponseStatus bool

	MultiError bool

	// See NoopAuthenticationFunc
	AuthenticationFunc AuthenticationFunc

	// Sets a function to override the schema error message.
	CustomSchemaErrorFunc func(err *openapi3.SchemaError) string
}
