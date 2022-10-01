package openapi3

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

const (
	// FormatOfStringForUUIDOfRFC4122 is an optional predefined format for UUID v1-v5 as specified by RFC4122
	FormatOfStringForUUIDOfRFC4122 = `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`
)

// FormatCallback performs custom checks on exotic formats
type FormatCallback func(value string) error

// Format represents a format validator registered by either DefineStringFormat or DefineStringFormatCallback
type Format struct {
	regexp   *regexp.Regexp
	callback FormatCallback
}

// SchemaStringFormats allows for validating string formats
var SchemaStringFormats = make(map[string]Format, 4)

// DefineStringFormat defines a new regexp pattern for a given format
func DefineStringFormat(name string, pattern string) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		err := fmt.Errorf("format %q has invalid pattern %q: %w", name, pattern, err)
		panic(err)
	}
	SchemaStringFormats[name] = Format{regexp: re}
}

// DefineStringFormatCallback adds a validation function for a specific schema format entry
func DefineStringFormatCallback(name string, callback FormatCallback) {
	SchemaStringFormats[name] = Format{callback: callback}
}

func validateIP(ip string) error {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return &SchemaError{
			Value:  ip,
			Reason: "Not an IP address",
		}
	}
	return nil
}

func validateIPv4(ip string) error {
	if err := validateIP(ip); err != nil {
		return err
	}

	if !(strings.Count(ip, ":") < 2) {
		return &SchemaError{
			Value:  ip,
			Reason: "Not an IPv4 address (it's IPv6)",
		}
	}
	return nil
}

func validateIPv6(ip string) error {
	if err := validateIP(ip); err != nil {
		return err
	}

	if !(strings.Count(ip, ":") >= 2) {
		return &SchemaError{
			Value:  ip,
			Reason: "Not an IPv6 address (it's IPv4)",
		}
	}
	return nil
}

func init() {
	// This pattern catches only some suspiciously wrong-looking email addresses.
	// Use DefineStringFormat(...) if you need something stricter.
	DefineStringFormat("email", `^[^@]+@[^@<>",\s]+$`)

	// Base64
	// The pattern supports base64 and b./ase64url. Padding ('=') is supported.
	DefineStringFormat("byte", `(^$|^[a-zA-Z0-9+/\-_]*=*$)`)

	// date
	DefineStringFormat("date", `^[0-9]{4}-(0[1-9]|1[012])-(0[1-9]|[12][0-9]|3[01])$`)

	// date-time
	DefineStringFormat("date-time", `^[0-9]{4}-(0[1-9]|1[012])-(0[1-9]|[12][0-9]|3[01])T([01][0-9]|2[0-3])(:[0-5][0-9]){2}(\.[0-9]+)?(Z|(\+|-)[0-9]{2}:[0-9]{2})?$`)

	// uuid
	DefineStringFormat(`uuid`, FormatOfStringForUUIDOfRFC4122)

	// ipv4/v6
	DefineIPv4Format()
	DefineIPv6Format()
}

// DefineIPv4Format opts in ipv4 format validation on top of OAS 3 spec
func DefineIPv4Format() {
	DefineStringFormatCallback("ipv4", validateIPv4)
}

// DefineIPv6Format opts in ipv6 format validation on top of OAS 3 spec
func DefineIPv6Format() {
	DefineStringFormatCallback("ipv6", validateIPv6)
}
