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
	value, err := decodeOrderedJSON(dec)
	if err != nil {
		return nil, &ParseError{Kind: KindInvalidFormat, Cause: err}
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, &ParseError{Kind: KindInvalidFormat, Cause: fmt.Errorf("unexpected data after the top-level value")}
	}
	return value, nil
}

func decodeOrderedJSON(dec *json.Decoder) (any, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	delim, ok := tok.(json.Delim)
	if !ok {
		return tok, nil
	}
	switch delim {
	case '{':
		obj := openapi3.NewOrderedObject()
		for dec.More() {
			name, err := dec.Token()
			if err != nil {
				return nil, err
			}
			value, err := decodeOrderedJSON(dec)
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
			item, err := decodeOrderedJSON(dec)
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
