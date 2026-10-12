//go:build !go1.27 || !goexperiment.jsonv2

package openapi3

import "encoding/json"

// unmarshalJSON decodes data into v with encoding/json/v2 when available.
func unmarshalJSON(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
