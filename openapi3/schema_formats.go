package openapi3

import (
	"regexp"

	"github.com/xeipuuv/gojsonschema"
	// https://github.com/xeipuuv/gojsonschema/pull/297/files discriminator support
)

func init() {
	// gojsonschema.FormatCheckers = gojsonschema.FormatCheckerChain{} FIXME https://github.com/xeipuuv/gojsonschema/pull/326
	gojsonschema.FormatCheckers.Add("byte", byteFormatChecker{})
	gojsonschema.FormatCheckers.Add("date", gojsonschema.DateFormatChecker{})
	gojsonschema.FormatCheckers.Add("date-time", gojsonschema.DateTimeFormatChecker{})
}

type byteFormatChecker struct{}

var _ gojsonschema.FormatChecker = (*byteFormatChecker)(nil)
var reByteFormatChecker = regexp.MustCompile(`(^$|^[a-zA-Z0-9+/\-_]*=*$)`)

// IsFormat supports base64 and base64url. Padding ('=') is supported.
func (byteFormatChecker) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return true
	}

	return reByteFormatChecker.MatchString(asString)
}

// DefineEmailFormat opts-in to checking email format (outside of OpenAPIv3 spec)
func DefineEmailFormat() {
	gojsonschema.FormatCheckers.Add("email", gojsonschema.EmailFormatChecker{})
}

// DefineUUIDFormat opts-in to checking uuid format v1-v5 as specified by RFC4122 (outside of OpenAPIv3 spec)
func DefineUUIDFormat() {
	gojsonschema.FormatCheckers.Add("uuid", gojsonschema.UUIDFormatChecker{})
}

// DefineIPv4Format opts in ipv4 format validation on top of OAS 3 spec
func DefineIPv4Format() {
	gojsonschema.FormatCheckers.Add("ipv4", gojsonschema.IPV4FormatChecker{})
}

// DefineIPv6Format opts in ipv6 format validation on top of OAS 3 spec
func DefineIPv6Format() {
	gojsonschema.FormatCheckers.Add("ipv6", gojsonschema.IPV6FormatChecker{})
}
