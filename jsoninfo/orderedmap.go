package jsoninfo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
)

// ExtractObjectKeys extracts the keys of an object in a JSON string. The keys
// are returned in the order they appear in the JSON string.
func ExtractObjectKeys(b []byte) ([]string, error) {
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
