package openapi3filter

import (
	"context"
)

var DefaultOptions = &Options{
	FailFast: true,
}

type Options struct {
	ExcludeRequestBody    bool
	ExcludeResponseBody   bool
	IncludeResponseStatus bool
	FailFast              bool
	AuthenticationFunc    func(c context.Context, input *AuthenticationInput) error
}
