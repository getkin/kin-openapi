//go:build go1.27 && goexperiment.jsonv2

// Requires Go 1.27's v2-backed encoding/json implementation,
// which calls UnmarshalJSONFrom.

package openapi3

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
	x := Responses{
		Extensions: make(map[string]any),
		m:          make(map[string]*ResponseRef),
	}
	if dec.PeekKind() != jsontext.KindBeginObject {
		var m map[string]any // Reports non-object values as UnmarshalJSON does
		if err := jsonv2.UnmarshalDecode(dec, &m); err != nil {
			return err
		}
		*responses = x
		return nil
	}
	if _, err := dec.ReadToken(); err != nil {
		return err
	}
	for dec.PeekKind() != jsontext.KindEndObject {
		tok, err := dec.ReadToken()
		if err != nil {
			return err
		}
		k := tok.String()
		if strings.HasPrefix(k, "x-") {
			var v any
			if err := jsonv2.UnmarshalDecode(dec, &v); err != nil {
				return err
			}
			x.Extensions[k] = v
			continue
		}
		var v ResponseRef
		if err := jsonv2.UnmarshalDecode(dec, &v); err != nil {
			return err
		}
		x.m[k] = &v
	}
	if _, err := dec.ReadToken(); err != nil {
		return err
	}
	*responses = x
	return nil
}

var _ jsonv2.UnmarshalerFrom = (*Callback)(nil)

// UnmarshalJSONFrom sets Callback to the next JSON value read from dec.
// Decoding each member from dec (rather than from a copy of its bytes)
// keeps the member's key in the path of any error.
func (callback *Callback) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	x := Callback{
		Extensions: make(map[string]any),
		m:          make(map[string]*PathItem),
	}
	if dec.PeekKind() != jsontext.KindBeginObject {
		var m map[string]any // Reports non-object values as UnmarshalJSON does
		if err := jsonv2.UnmarshalDecode(dec, &m); err != nil {
			return err
		}
		*callback = x
		return nil
	}
	if _, err := dec.ReadToken(); err != nil {
		return err
	}
	for dec.PeekKind() != jsontext.KindEndObject {
		tok, err := dec.ReadToken()
		if err != nil {
			return err
		}
		k := tok.String()
		if strings.HasPrefix(k, "x-") {
			var v any
			if err := jsonv2.UnmarshalDecode(dec, &v); err != nil {
				return err
			}
			x.Extensions[k] = v
			continue
		}
		var v PathItem
		if err := jsonv2.UnmarshalDecode(dec, &v); err != nil {
			return err
		}
		x.m[k] = &v
	}
	if _, err := dec.ReadToken(); err != nil {
		return err
	}
	*callback = x
	return nil
}

var _ jsonv2.UnmarshalerFrom = (*Paths)(nil)

// UnmarshalJSONFrom sets Paths to the next JSON value read from dec.
// Decoding each member from dec (rather than from a copy of its bytes)
// keeps the member's key in the path of any error.
func (paths *Paths) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	x := Paths{
		Extensions: make(map[string]any),
		m:          make(map[string]*PathItem),
	}
	if dec.PeekKind() != jsontext.KindBeginObject {
		var m map[string]any // Reports non-object values as UnmarshalJSON does
		if err := jsonv2.UnmarshalDecode(dec, &m); err != nil {
			return err
		}
		*paths = x
		return nil
	}
	if _, err := dec.ReadToken(); err != nil {
		return err
	}
	for dec.PeekKind() != jsontext.KindEndObject {
		tok, err := dec.ReadToken()
		if err != nil {
			return err
		}
		k := tok.String()
		if strings.HasPrefix(k, "x-") {
			var v any
			if err := jsonv2.UnmarshalDecode(dec, &v); err != nil {
				return err
			}
			x.Extensions[k] = v
			continue
		}
		var v PathItem
		if err := jsonv2.UnmarshalDecode(dec, &v); err != nil {
			return err
		}
		x.m[k] = &v
	}
	if _, err := dec.ReadToken(); err != nil {
		return err
	}
	*paths = x
	return nil
}
