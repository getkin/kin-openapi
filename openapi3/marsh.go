package openapi3

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"

	"github.com/oasdiff/yaml"
)

func unmarshalError(jsonUnmarshalErr error) error {
	if before, after, found := strings.Cut(jsonUnmarshalErr.Error(), "Bis"); found && before != "" {
		before = strings.ReplaceAll(before, " Go struct ", " ")
		before = strings.ReplaceAll(before, " Go ", " ")
		return fmt.Errorf("%s%s", before, strings.ReplaceAll(after, "Bis", ""))
	}
	return jsonUnmarshalErr
}

// unmarshal decodes data into v. It returns the document origin tree when
// includeOrigin is set and the data took the yaml path (json input carries no
// origins), so the caller can retain it (see Loader.originTrees).
func unmarshal(data []byte, v any, includeOrigin bool, location *url.URL) (*originTree, error) {
	var jsonErr, yamlErr error

	// See https://github.com/getkin/kin-openapi/issues/680
	if mayBeJSON(data) {
		if jsonErr = unmarshalJSON(data, v); jsonErr == nil {
			return nil, nil
		}
	}

	// UnmarshalStrict(data, v) TODO: investigate how ymlv3 handles duplicate map keys
	var file string
	if location != nil {
		file = location.String()
	}
	if tree, err := yaml.Unmarshal(data, v, yaml.DecodeOpts{
		Origin:            yaml.OriginOpt{Enabled: includeOrigin, File: file},
		DisableTimestamps: true,
	}); err == nil {
		applyOrigins(v, tree)
		return tree, nil
	} else {
		yamlErr = err
	}

	// Every attempt failed: return what each one said
	return nil, unmarshalErrors(jsonErr, yamlErr)
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
