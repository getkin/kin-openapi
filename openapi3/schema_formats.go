package openapi3

import (
	"fmt"
	"net"
	"regexp"
)

const (
	// FormatOfStringForUUIDOfRFC4122 is an optional predefined format for UUID v1-v5 as specified by RFC4122
	FormatOfStringForUUIDOfRFC4122 = `^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`
)

// FormatType is the type of format validation used
type FormatType string

// FormatTypeRe is regexp based validation
const FormatTypeRe = "re"

// FormatTypeCallback is callback based validation
const FormatTypeCallback = "callback"

//FormatCallback custom check on exotic formats
type FormatCallback func(Val string) error

//Format is the format type context fo validate the format
type Format struct {
	regexp   *regexp.Regexp
	callback FormatCallback
}

//SchemaStringFormats allows for validating strings format
var SchemaStringFormats = make(map[string]Format, 8)

//DefineStringFormat Defines a new regexp pattern for a given format
func DefineStringFormat(name string, pattern string) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		err := fmt.Errorf("Format '%v' has invalid pattern '%v': %v", name, pattern, err)
		panic(err)
	}
	f := Format{
		regexp:   re,
		callback: nil}
	SchemaStringFormats[name] = f
}

// DefineStringCallbackFormat define callback based type callback validation
func DefineStringCallbackFormat(name string, callback FormatCallback) {
	f := Format{
		regexp:   nil,
		callback: callback}
	SchemaStringFormats[name] = f
}

func validateIP(ip string) (*net.IP, error) {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return nil, &SchemaError{
			Value:  ip,
			Reason: "Not an IP address",
		}
	}
	return &parsed, nil
}

func validateIPv4(ip string) error {
	parsed, err := validateIP(ip)
	if err != nil {
		return err
	}

	if parsed.To4() == nil {
		return &SchemaError{
			Value:  ip,
			Reason: "Not an IPv4 address (it's IPv6)",
		}
	}
	return nil
}
func validateIPv6(ip string) error {
	parsed, err := validateIP(ip)
	if err != nil {
		return err
	}

	if parsed.To4() != nil {
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
	DefineStringFormat("date", `^[0-9]{4}-(0[0-9]|10|11|12)-([0-2][0-9]|30|31)$`)

	// date-time
	DefineStringFormat("date-time", `^[0-9]{4}-(0[0-9]|10|11|12)-([0-2][0-9]|30|31)T[0-9]{2}:[0-9]{2}:[0-9]{2}(.[0-9]+)?(Z|(\+|-)[0-9]{2}:[0-9]{2})?$`)

	DefineStringCallbackFormat("ipv4", validateIPv4)
	DefineStringCallbackFormat("ipv6", validateIPv6)

}
