package openapi2

import (
	"encoding/json"
	"fmt"
	"maps"
	"strings"
)

// Both the Paths and Responses objects mix typed entries with arbitrary x- values.
type mapWithExtensions[V any] struct {
	Extensions map[string]any `json:"-" yaml:"-"`

	m map[string]*V
}

func (obj mapWithExtensions[V]) MarshalJSON() ([]byte, error) {
	m := make(map[string]any, len(obj.m)+len(obj.Extensions))
	maps.Copy(m, obj.Extensions)
	for k, v := range obj.m {
		m[k] = v
	}
	return json.Marshal(m)
}

func (obj *mapWithExtensions[V]) unmarshalJSON(data []byte, member string) error {
	var m map[string]json.RawMessage
	if err := unmarshalJSON(data, &m); err != nil {
		return unmarshalError(err)
	}
	x := mapWithExtensions[V]{m: make(map[string]*V, len(m))}
	for k, v := range m {
		if strings.HasPrefix(k, "x-") {
			if x.Extensions == nil {
				x.Extensions = make(map[string]any)
			}
			var ext any
			if err := unmarshalJSON(v, &ext); err != nil {
				return unmarshalError(err)
			}
			x.Extensions[k] = ext
			continue
		}
		var value *V
		if err := unmarshalJSON(v, &value); err != nil {
			// Decoding the member on its own loses its key from the error: add it back.
			return fmt.Errorf("%s %q: %w", member, k, err)
		}
		x.m[k] = value
	}
	*obj = x
	return nil
}

// pathsObject is the wire representation of the two maps on T.
type pathsObject struct{ mapWithExtensions[PathItem] }

func (paths *pathsObject) UnmarshalJSON(data []byte) error {
	return paths.unmarshalJSON(data, "path")
}
