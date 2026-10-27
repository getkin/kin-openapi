package openapi3filter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
)

// OrderedJSONBodyDecoder decodes a JSON formatted body like JSONBodyDecoder,
// but returns each object as an *openapi3.OrderedObject. Errors about the
// members of an object are then reported in the order they appear in the body.
// Register it with RegisterBodyDecoder to use it.
func OrderedJSONBodyDecoder(body io.Reader, header http.Header, schema *openapi3.SchemaRef, encFn EncodingFn) (any, error) {
	dec := json.NewDecoder(body)
	dec.UseNumber()
	value, err := decodeOrderedJSON(dec, 0)
	if err != nil {
		return nil, &ParseError{Kind: KindInvalidFormat, Cause: err}
	}
	return value, nil
}

// maxOrderedJSONDepth is the nesting limit of encoding/json.
const maxOrderedJSONDepth = 10000

func decodeOrderedJSON(dec *json.Decoder, depth int) (any, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	delim, ok := tok.(json.Delim)
	if !ok {
		return tok, nil
	}
	if depth++; depth > maxOrderedJSONDepth {
		return nil, fmt.Errorf("exceeded max depth")
	}
	switch delim {
	case '{':
		obj := openapi3.NewOrderedObject()
		for dec.More() {
			name, err := dec.Token()
			if err != nil {
				return nil, err
			}
			if _, ok := obj.Get(name.(string)); ok && rejectDuplicateNames {
				return nil, fmt.Errorf("duplicate object member name %q", name)
			}
			value, err := decodeOrderedJSON(dec, depth)
			if err != nil {
				return nil, err
			}
			obj.Set(name.(string), value)
		}
		_, err := dec.Token()
		return obj, err
	case '[':
		items := []any{}
		for dec.More() {
			item, err := decodeOrderedJSON(dec, depth)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		}
		_, err := dec.Token()
		return items, err
	}
	return nil, fmt.Errorf("unexpected delimiter %q", delim)
}
