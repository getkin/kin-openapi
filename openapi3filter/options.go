package openapi3filter

// DefaultOptions do not set an AuthenticationFunc.
// A spec with security schemes defined will not pass validation
// unless an AuthenticationFunc is defined.
var DefaultOptions = &Options{}

// Options used by ValidateRequest and ValidateResponse
type Options struct {
	ExcludeRequestBody    bool
	ExcludeResponseBody   bool
	IncludeResponseStatus bool

	MultiError             bool
	FailOnUnknownParameter bool

	// See NoopAuthenticationFunc
	AuthenticationFunc AuthenticationFunc
}
