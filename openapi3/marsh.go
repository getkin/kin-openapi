package openapi3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/invopop/yaml"
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

	// UnmarshalStrict(data, v) TODO: investigate how ymlv3 handles duplicate map keys
	if yamlErr = yaml.Unmarshal(data, v); yamlErr == nil {
		return nil
	}

	// If both unmarshaling attempts fail, return a new error that includes both errors
	return fmt.Errorf("failed to unmarshal data: json error: %v, yaml error: %v", jsonErr, yamlErr)
}

// extractObjectKeys extracts the keys of an object in a JSON string. The keys
// are returned in the order they appear in the JSON string.
func extractObjectKeys(b []byte) ([]string, error) {
	if !bytes.HasPrefix(b, []byte{'{'}) {
		return nil, fmt.Errorf("expected '{' at start of JSON object")
	}

	dec := json.NewDecoder(bytes.NewReader(b))
	var keys []string

	for dec.More() {
		// Read prop name
		t, err := dec.Token()
		if err != nil {
			log.Printf("Err: %v", err)
			break
		}

		name, ok := t.(string)
		if !ok {
			continue // May be a delimeter
		}

		keys = append(keys, name)

		var whatever nullMessage
		dec.Decode(&whatever)
	}

	return keys, nil
}

// nullMessage implements json.Unmarshaler and does nothing with the given
// value.
type nullMessage struct{}

func (*nullMessage) UnmarshalJSON(data []byte) error { return nil }
