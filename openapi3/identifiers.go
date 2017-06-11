package openapi3

import (
	"fmt"
	"regexp"
)

const identifierPattern = `^[a-zA-Z0-9.\-_]+$`

var identifierRegExp = regexp.MustCompile(identifierPattern)

func ValidateIdentifier(value string) error {
	if identifierRegExp.MatchString(value) {
		return nil
	}
	return fmt.Errorf("Identifier '%s' is not supported by OpenAPI version 3 standard (regexp: '%s')", value, identifierPattern)
}
