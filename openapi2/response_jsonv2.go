//go:build go1.27 && goexperiment.jsonv2

// Requires Go 1.27's v2-backed encoding/json implementation,
// which calls UnmarshalJSONFrom.

package openapi2

import (
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
	"strings"
)

var _ jsonv2.UnmarshalerFrom = (*Responses)(nil)

// UnmarshalJSONFrom sets Responses to the next JSON value read from dec.
// Decoding each member from dec (rather than from a copy of its bytes)
// keeps the member's key in the path of any error.
func (responses *Responses) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	if dec.PeekKind() != jsontext.KindBeginObject {
		var m map[string]jsontext.Value // Reports non-object values as UnmarshalJSON does
		if err := jsonv2.UnmarshalDecode(dec, &m); err != nil {
			return err
		}
		*responses = Responses{m: make(map[string]*Response)}
		return nil
	}
	if _, err := dec.ReadToken(); err != nil {
		return err
	}
	x := Responses{m: make(map[string]*Response)}
	for dec.PeekKind() != jsontext.KindEndObject {
		tok, err := dec.ReadToken()
		if err != nil {
			return err
		}
		k := tok.String()
		if strings.HasPrefix(k, "x-") {
			if x.Extensions == nil {
				x.Extensions = make(map[string]any)
			}
			var ext any
			if err := jsonv2.UnmarshalDecode(dec, &ext); err != nil {
				return err
			}
			x.Extensions[k] = ext
			continue
		}
		var response *Response
		if err := jsonv2.UnmarshalDecode(dec, &response); err != nil {
			return err
		}
		x.m[k] = response
	}
	if _, err := dec.ReadToken(); err != nil {
		return err
	}
	*responses = x
	return nil
}
