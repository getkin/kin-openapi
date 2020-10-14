package openapi3

import "bytes"

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
