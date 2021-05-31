package openapi3

import (
	"bytes"
	"errors"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

// SchemaValidationError is a collection of errors
type SchemaValidationError []gojsonschema.ResultError

var _ error = (*SchemaValidationError)(nil)

func (e SchemaValidationError) Error() string {
	var buff strings.Builder
	for i, re := range []gojsonschema.ResultError(e) {
		if i != 0 {
			buff.WriteString("\n")
		}
		buff.WriteString(re.String())
	}
	return buff.String()
}

// Errors unwraps into much detailed errors.
// See https://pkg.go.dev/github.com/xeipuuv/gojsonschema#ResultError
func (e SchemaValidationError) Errors() []gojsonschema.ResultError {
	return e
}

// JSONPointer returns a dot (.) delimited "JSON path" to the context of the first error.
func (e SchemaValidationError) JSONPointer() string {
	return []gojsonschema.ResultError(e)[0].Field()
}

func (e SchemaValidationError) asMultiError() MultiError {
	errs := make([]error, 0, len(e))
	for _, re := range e {
		errs = append(errs, errors.New(re.String()))
	}
	return errs
}

// MultiError is a collection of errors, intended for when
// multiple issues need to be reported upstream
type MultiError []error

func (me MultiError) Error() string {
	buff := &bytes.Buffer{}
	for _, e := range me {
		buff.WriteString(e.Error())
		buff.WriteString(" | ")
	}
	return buff.String()
}

//Is allows you to determine if a generic error is in fact a MultiError using `errors.Is()`
//It will also return true if any of the contained errors match target
func (me MultiError) Is(target error) bool {
	if _, ok := target.(MultiError); ok {
		return true
	}
	for _, e := range me {
		if errors.Is(e, target) {
			return true
		}
	}
	return false
}

//As allows you to use `errors.As()` to set target to the first error within the multi error that matches the target type
func (me MultiError) As(target interface{}) bool {
	for _, e := range me {
		if errors.As(e, target) {
			return true
		}
	}
	return false
}
