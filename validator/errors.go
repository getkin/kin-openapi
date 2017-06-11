package openapi3filter

import (
	"errors"
	"fmt"
	"github.com/jban332/kinapi/openapi3"
)

var ErrAuthenticationServiceMissing = errors.New("Request validator doesn't have an authentication service defined")

type RequestError struct {
	Request        *Request
	Parameter   *openapi3.Parameter
	RequestBody *openapi3.RequestBody
	Reason      string
	Err         error
}

func (err *RequestError) Error() string {
	reason := err.Reason
	if e := err.Err; e != nil {
		if len(reason) == 0 {
			reason = e.Error()
		} else {
			reason += ": " + e.Error()
		}
	}
	if v := err.Parameter; v != nil {
		return fmt.Sprintf("Parameter '%s' in %s has an error: %s", v.Name, v.In, reason)
	} else if v := err.RequestBody; v != nil {
		return fmt.Sprintf("Request body has an error: %s", reason)
	} else {
		return reason
	}
}

type ResponseError struct {
	Request   *Request
	Reason string
	Err    error
}

func (err *ResponseError) Error() string {
	reason := err.Reason
	if e := err.Err; e != nil {
		if len(reason) == 0 {
			reason = e.Error()
		} else {
			reason += ": " + e.Error()
		}
	}
	return reason
}

type SecurityRequirementsError struct {
	SecurityRequirements openapi3.SecurityRequirements
	Errors               []error
}

func (err *SecurityRequirementsError) Error() string {
	return "Security requirements failed"
}
