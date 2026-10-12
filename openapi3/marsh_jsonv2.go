//go:build go1.27 && goexperiment.jsonv2

package openapi3

import jsonv2 "encoding/json/v2"

// unmarshalJSON decodes data into v with encoding/json/v2 when available.
//
// Unlike encoding/json, it rejects duplicate object member names and invalid UTF-8,
// and its errors locate the faulty value from the root of the document,
// across the UnmarshalJSON methods it calls.
func unmarshalJSON(data []byte, v any) error {
	return jsonv2.Unmarshal(data, v)
}
