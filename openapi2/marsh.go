package openapi2

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/internal/yamlconv"
)

func unmarshalError(jsonUnmarshalErr error) error {
	if before, after, found := strings.Cut(jsonUnmarshalErr.Error(), "Bis"); found && before != "" {
		before = strings.ReplaceAll(before, " Go struct ", " ")
		before = strings.ReplaceAll(before, " Go ", " ")
		return fmt.Errorf("%s%s", before, strings.ReplaceAll(after, "Bis", ""))
	}
	return jsonUnmarshalErr
}

func unmarshal(data []byte, v any) error {
	var jsonErr, yamlErr error

	// See https://github.com/getkin/kin-openapi/issues/680
	if mayBeJSON(data) {
		if jsonErr = unmarshalJSON(data, v); jsonErr == nil {
			return nil
		}
	}

	// YAML reaches these types through JSON, since they implement
	// UnmarshalJSON and not UnmarshalYAML.
	if yamlErr = yamlconv.Unmarshal(data, v); yamlErr == nil {
		return nil
	}

	// Every attempt failed: return what each one said
	return unmarshalErrors(jsonErr, yamlErr)
}

// mayBeJSON reports whether data could hold a JSON value: whether its first
// byte past JSON's whitespace can start one. Data that cannot is not tried as
// JSON, so a YAML document's error is not followed by the JSON parser's
// complaint about its first byte.
func mayBeJSON(data []byte) bool {
	data = bytes.TrimLeft(data, " \t\r\n")
	return len(data) != 0 && strings.IndexByte(`{["-0123456789tfn`, data[0]) >= 0
}

// unmarshalErrors wraps the error of each decode that was attempted.
func unmarshalErrors(jsonErr, yamlErr error) error {
	if jsonErr == nil {
		return fmt.Errorf("failed to unmarshal data: yaml error: %w", yamlErr)
	}
	return fmt.Errorf("failed to unmarshal data: json error: %w, yaml error: %w", jsonErr, yamlErr)
}

// UnmarshalFromData loads a document from swagger 2.0 bytes in either JSON or
// YAML. The v3 side has Loader for this; here the whole job is the decode.
func UnmarshalFromData(data []byte, doc *T) error {
	return unmarshal(data, doc)
}
