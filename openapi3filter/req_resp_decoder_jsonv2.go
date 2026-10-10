//go:build go1.27 && goexperiment.jsonv2

// Requires Go 1.27's v2-backed encoding/json implementation,
// which calls UnmarshalJSONFrom and supports this option override.

package openapi3filter

import (
	"bytes"
	"encoding/json"
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
	"io"
	"sync"
)

// decodeJSONSlow decodes like json.Decoder with UseNumber, but also rejects
// objects with duplicate member names.
// See https://github.com/getkin/kin-openapi/issues/322
func decodeJSONSlow(body io.Reader) (any, error) {
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

// jsonDecoder holds a decoder and the body it reads, reused across calls.
type jsonDecoder struct {
	buf bytes.Buffer
	dec jsontext.Decoder
}

var jsonDecoders = sync.Pool{New: func() any { return new(jsonDecoder) }}

// maxPooledBody caps the buffers kept in jsonDecoders, as jsontext caps its
// own pools: a buffer stays as large as the largest body it held, and the pool
// keeps it allocated until garbage collection empties the pool.
const maxPooledBody = 64 << 10

// decodeJSON returns what decodeJSONSlow returns, reading the body once
// instead of twice: json.Decoder reads a value to its end before decoding it.
func decodeJSON(body io.Reader) (any, error) {
	d := jsonDecoders.Get().(*jsonDecoder)
	d.buf.Reset()
	if _, err := io.Copy(&d.buf, body); err != nil {
		return nil, err
	}
	data := d.buf.Bytes()

	// Like encoding/json, replace invalid UTF-8 with U+FFFD instead of failing.
	// decodeJSONValue detects duplicate names itself, faster than jsontext.
	d.dec.Reset(&d.buf, jsontext.AllowInvalidUTF8(true), jsontext.AllowDuplicateNames(true))
	value, err := decodeJSONValue(&d.dec)
	if err != nil {
		// Report malformed JSON and duplicate names with decodeJSONSlow's errors.
		if _, errSlow := decodeJSONSlow(bytes.NewReader(data)); errSlow != nil {
			err = errSlow
		}
	}
	if d.buf.Cap() <= maxPooledBody {
		jsonDecoders.Put(d)
	}
	return value, err
}

// decodeJSONValue reads the next JSON value into the types decodeJSONSlow uses.
func decodeJSONValue(dec *jsontext.Decoder) (any, error) {
	tok, err := dec.ReadToken()
	if err != nil {
		return nil, err
	}
	switch tok.Kind() {
	case '{':
		obj := make(map[string]any)
		for dec.PeekKind() != '}' {
			if tok, err = dec.ReadToken(); err != nil {
				return nil, err
			}
			name, n := tok.String(), len(obj)
			if obj[name], err = decodeJSONValue(dec); err != nil {
				return nil, err
			}
			if len(obj) == n {
				return nil, jsontext.ErrDuplicateName
			}
		}
		_, err = dec.ReadToken()
		return obj, err
	case '[':
		arr := []any{}
		for dec.PeekKind() != ']' {
			v, err := decodeJSONValue(dec)
			if err != nil {
				return nil, err
			}
			arr = append(arr, v)
		}
		_, err = dec.ReadToken()
		return arr, err
	case '"':
		return tok.String(), nil
	case '0':
		return json.Number(tok.String()), nil
	case 't', 'f':
		return tok.Bool(), nil
	}
	return nil, nil // null
}
