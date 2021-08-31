package openapi3

import (
	"fmt"
	"regexp"
)

var identifierRegExp *regexp.Regexp

// ValidateIdentifier verifies whether Component object key matches identifier pattern according to OpenAPIv3
func ValidateIdentifier(value string) (err error) {
	const re = `^[a-zA-Z0-9._-]+$`
	if identifierRegExp == nil {
		if identifierRegExp, err = regexp.Compile(re); err != nil {
			return
		}
	}

	if identifierRegExp.MatchString(value) {
		return nil
	}
	return fmt.Errorf("identifier %q is not supported by OpenAPIv3 standard (regexp: %q)", value, re)
}
