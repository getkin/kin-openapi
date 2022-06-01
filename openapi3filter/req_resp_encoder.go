package openapi3filter

import (
	"encoding/json"
	"fmt"
)

func encodeBody(body interface{}, mediaType string) ([]byte, error) {
	encoder, ok := bodyEncoders[mediaType]
	if !ok {
		return nil, &ParseError{
			Kind:   KindUnsupportedFormat,
			Reason: fmt.Sprintf("%s %q", prefixUnsupportedCT, mediaType),
		}
	}
	return encoder(body)
}

type bodyEncoder func(body interface{}) ([]byte, error)

var bodyEncoders = map[string]bodyEncoder{
	"application/json": json.Marshal,
}
