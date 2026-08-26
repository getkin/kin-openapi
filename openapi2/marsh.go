package openapi2

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/internal/yamlconv"
)

func unmarshalError(jsonUnmarshalErr error) error {
	if before, after, found := strings.Cut(jsonUnmarshalErr.Error(), "Bis"); found && before != "" && after != "" {
		before = strings.ReplaceAll(before, " Go struct ", " ")
		return fmt.Errorf("%s%s", before, strings.ReplaceAll(after, "Bis", ""))
	}
	return jsonUnmarshalErr
}

func unmarshal(data []byte, v any) error {
	var jsonErr, yamlErr error

	// See https://github.com/getkin/kin-openapi/issues/680
	if jsonErr = json.Unmarshal(data, v); jsonErr == nil {
		return nil
	}

	// YAML reaches these types through JSON, since they implement
	// UnmarshalJSON and not UnmarshalYAML.
	if yamlErr = yamlconv.Unmarshal(data, v); yamlErr == nil {
		return nil
	}

	// If both unmarshaling attempts fail, return a new error that includes both errors
	return fmt.Errorf("failed to unmarshal data: json error: %v, yaml error: %v", jsonErr, yamlErr)
}

// UnmarshalFromData loads a document from swagger 2.0 bytes in either JSON or
// YAML. The v3 side has Loader for this; here the whole job is the decode.
func UnmarshalFromData(data []byte, doc *T) error {
	return unmarshal(data, doc)
}
