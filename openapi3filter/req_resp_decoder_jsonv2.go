//go:build go1.27 && goexperiment.jsonv2

// Build-tagged go1.27 because it uses encoding/json/jsontext, which comes with Go 1.27's encoding/json/v2.

package openapi3filter

import (
	"bytes"
	"encoding/json"
	"encoding/json/jsontext"
	"io"
	"sync"
)

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

// decodeJSON decodes like json.Decoder with UseNumber, reading the body once
// instead of twice (json.Decoder reads a value to its end before decoding it),
// and rejects objects with duplicate member names.
// See https://github.com/getkin/kin-openapi/issues/322
func decodeJSON(body io.Reader) (any, error) {
	d := jsonDecoders.Get().(*jsonDecoder)
	d.buf.Reset()
	if _, err := io.Copy(&d.buf, body); err != nil {
		return nil, err
	}

	// Like encoding/json, replace invalid UTF-8 with U+FFFD instead of failing.
	// decodeJSONValue detects duplicate names itself, faster than jsontext.
	d.dec.Reset(&d.buf, jsontext.AllowInvalidUTF8(true), jsontext.AllowDuplicateNames(true))
	value, err := decodeJSONValue(&d.dec)
	if d.buf.Cap() <= maxPooledBody {
		jsonDecoders.Put(d)
	}
	return value, err
}

// decodeJSONValue reads the next JSON value into the types json.Decoder uses.
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
				return nil, &jsontext.SyntacticError{JSONPointer: dec.StackPointer(), Err: jsontext.ErrDuplicateName}
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
