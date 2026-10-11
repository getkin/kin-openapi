//go:build !go1.27 || !goexperiment.jsonv2

package openapi3filter

import (
	"encoding/json"
	"io"
)

// decodeJSON rejects duplicate member names only with encoding/json/v2.
func decodeJSON(body io.Reader) (any, error) {
	var value any
	dec := json.NewDecoder(body)
	dec.UseNumber()
	if err := dec.Decode(&value); err != nil {
		return nil, err
	}
	return value, nil
}
