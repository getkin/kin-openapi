//go:build go1.27 && goexperiment.jsonv2

// Requires Go 1.27's v2-backed encoding/json implementation,
// which calls UnmarshalJSONFrom and supports this option override.

package openapi3filter

import (
	"encoding/json"
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
	"io"
)

// decodeJSON decodes like json.Decoder with UseNumber, but also rejects
// objects with duplicate member names.
// See https://github.com/getkin/kin-openapi/issues/322
func decodeJSON(body io.Reader) (any, error) {
	var value any
	dec := json.NewDecoder(body)
	dec.UseNumber()
	if err := dec.Decode(&duplicateNameRejecter{&value}); err != nil {
		return nil, err
	}
	return value, nil
}

// duplicateNameRejecter is called by dec.Decode through jsonv2.UnmarshalerFrom.
type duplicateNameRejecter struct{ out *any }

var _ jsonv2.UnmarshalerFrom = (*duplicateNameRejecter)(nil)

func (r *duplicateNameRejecter) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	return jsonv2.UnmarshalDecode(dec, r.out, jsontext.AllowDuplicateNames(false))
}
