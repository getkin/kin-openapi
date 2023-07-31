package openapi3

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

const (
	// FormatOfStringForUUIDOfRFC4122 is an optional predefined format for UUID v1-v5 as specified by RFC4122
	FormatOfStringForUUIDOfRFC4122 = `^(?:[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}|00000000-0000-0000-0000-000000000000)$`

	// FormatOfStringForEmail pattern catches only some suspiciously wrong-looking email addresses.
	// Use DefineStringFormat(...) if you need something stricter.
	FormatOfStringForEmail = `^[^@]+@[^@<>",\s]+$`
)

// FormatCallback performs custom checks on exotic formats
type FormatCallback func(value string) error

// Format represents a format validator registered by either DefineStringFormat or DefineStringFormatCallback
type Format struct {
	versionedFormats []*versionedFormat
}

type versionedFormat struct {
	regexp   *regexp.Regexp
	callback FormatCallback
}

func (format *Format) add(minMinorVersion uint64, vFormat *versionedFormat) {
	if format != nil {
		if format.versionedFormats == nil {
			format.versionedFormats = make([]*versionedFormat, minMinorVersion+1)
			format.versionedFormats[minMinorVersion] = vFormat
		} else {
			numVersionedFormats := uint64(len(format.versionedFormats))
			if minMinorVersion >= numVersionedFormats {
				// grow array
				lastValue := format.versionedFormats[numVersionedFormats-1]
				additionalEntries := make([]*versionedFormat, minMinorVersion+1-numVersionedFormats)
				if lastValue != nil {
					for i := 0; i < len(additionalEntries); i++ {
						additionalEntries[i] = lastValue
					}
				}
				format.versionedFormats = append(format.versionedFormats, additionalEntries...)
				format.versionedFormats[minMinorVersion] = vFormat
				return
			}
			for i := minMinorVersion; i < numVersionedFormats; i++ {
				format.versionedFormats[i] = vFormat
			}
		}
	}
}

func (format Format) get(minorVersion uint64) *versionedFormat {
	if format.versionedFormats != nil {
		if minorVersion >= uint64(len(format.versionedFormats)) {
			return format.versionedFormats[len(format.versionedFormats)-1]
		}
		return format.versionedFormats[minorVersion]
	}
	return nil
}

func (format Format) DefinedForMinorVersion(minorVersion uint64) bool {
	return format.get(minorVersion) != nil
}

// SchemaStringFormats allows for validating string formats
var SchemaStringFormats = make(map[string]*Format, 4)
var defaultSchemaStringFormats map[string]*Format

// DefineStringFormat defines a new regexp pattern for a given format
// Will enforce regexp usage for minor versions of OpenAPI (3.Y.Z)
func DefineStringFormat(name string, pattern string, options ...SchemaFormatOption) {
	var schemaFormatOptions SchemaFormatOptions
	for _, option := range options {
		option(&schemaFormatOptions)
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		err := fmt.Errorf("format %q has invalid pattern %q: %w", name, pattern, err)
		panic(err)
	}
	updateSchemaStringFormats(name, schemaFormatOptions.fromOpenAPIMinorVersion, &versionedFormat{regexp: re})
}

func getSchemaStringFormats(name string, minorVersion uint64) *versionedFormat {
	if currentStringFormat, found := SchemaStringFormats[name]; found {
		return currentStringFormat.get(minorVersion)
	}
	return nil
}

func updateSchemaStringFormats(name string, minMinorVersion uint64, vFormat *versionedFormat) {
	if currentStringFormat, found := SchemaStringFormats[name]; found {
		currentStringFormat.add(minMinorVersion, vFormat)
		return
	}
	var newFormat Format
	newFormat.add(minMinorVersion, vFormat)
	SchemaStringFormats[name] = &newFormat
}

// DefineStringFormatCallback adds a validation function for a specific schema format entry
// Will enforce regexp usage for minor versions of OpenAPI (3.Y.Z)
func DefineStringFormatCallback(name string, callback FormatCallback, options ...SchemaFormatOption) {
	var schemaFormatOptions SchemaFormatOptions
	for _, option := range options {
		option(&schemaFormatOptions)
	}
	updateSchemaStringFormats(name, schemaFormatOptions.fromOpenAPIMinorVersion, &versionedFormat{callback: callback})
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

// SaveStringFormats allows to save (obtain a deep copy) of your current string formats
// so you can later restore it if needed
func SaveStringFormats(map[string]*Format) map[string]*Format {
	savedStringFormats := map[string]*Format{}
	for name, value := range SchemaStringFormats {
		var savedFormat Format
		savedFormat.versionedFormats = make([]*versionedFormat, len(value.versionedFormats))
		for index, versionedFormatValue := range value.versionedFormats {
			if versionedFormatValue != nil {
				savedVersionedFormat := versionedFormat{
					regexp:   versionedFormatValue.regexp,
					callback: versionedFormatValue.callback,
				}
				savedFormat.versionedFormats[index] = &savedVersionedFormat
			}
		}
		savedStringFormats[name] = &savedFormat
	}
	return savedStringFormats
}

// RestoreStringFormats allows to restore string format back to default values
func RestoreStringFormats(formatToRestore map[string]*Format) {
	restoredStringFormats := map[string]*Format{}
	for name, value := range formatToRestore {
		var restoredFormat Format
		restoredFormat.versionedFormats = make([]*versionedFormat, len(value.versionedFormats))
		for index, versionedFormatValue := range value.versionedFormats {
			if versionedFormatValue != nil {
				restoredVersionedFormat := versionedFormat{
					regexp:   versionedFormatValue.regexp,
					callback: versionedFormatValue.callback,
				}
				restoredFormat.versionedFormats[index] = &restoredVersionedFormat
			}
		}
		restoredStringFormats[name] = &restoredFormat
	}
	SchemaStringFormats = restoredStringFormats
}

// RestoreDefaultStringFormats allows to restore string format back to default values
func RestoreDefaultStringFormats() {
	RestoreStringFormats(defaultSchemaStringFormats)
}
func init() {
	// Base64
	// The pattern supports base64 and b./ase64url. Padding ('=') is supported.
	DefineStringFormat("byte", `(^$|^[a-zA-Z0-9+/\-_]*=*$)`)

	// date
	DefineStringFormat("date", `^[0-9]{4}-(0[0-9]|10|11|12)-([0-2][0-9]|30|31)$`)

	// date-time
	DefineStringFormat("date-time", `^[0-9]{4}-(0[0-9]|10|11|12)-([0-2][0-9]|30|31)T[0-9]{2}:[0-9]{2}:[0-9]{2}(\.[0-9]+)?(Z|(\+|-)[0-9]{2}:[0-9]{2})?$`)

	defaultSchemaStringFormats = SaveStringFormats(SchemaStringFormats)
}

// DefineIPv4Format opts in ipv4 format validation on top of OAS 3 spec
func DefineIPv4Format() {
	DefineStringFormatCallback("ipv4", validateIPv4)
}

// DefineIPv6Format opts in ipv6 format validation on top of OAS 3 spec
func DefineIPv6Format() {
	DefineStringFormatCallback("ipv6", validateIPv6)
}
