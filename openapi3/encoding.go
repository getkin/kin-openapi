package openapi3

import (
	"context"

	"github.com/getkin/kin-openapi/jsoninfo"
)

// Encoding is specified by OpenAPI/Swagger 3.0 standard.
type Encoding struct {
	ExtensionProps

	ContentType   string                `json:"contentType,omitempty"`
	Headers       map[string]*HeaderRef `json:"headers,omitempty"`
	Style         string                `json:"style,omitempty"`
	Explode       bool                  `json:"explode,omitempty"`
	AllowReserved bool                  `json:"allowReserved,omitempty"`
}

func NewEncoding() *Encoding {
	return &Encoding{}
}

func (encoding *Encoding) WithHeader(name string, header *Header) *Encoding {
	return encoding.WithHeaderRef(name, &HeaderRef{
		Value: header,
	})
}

func (encoding *Encoding) WithHeaderRef(name string, ref *HeaderRef) *Encoding {
	headers := encoding.Headers
	if headers == nil {
		headers = make(map[string]*HeaderRef)
		encoding.Headers = headers
	}
	headers[name] = ref
	return encoding
}

func (encoding *Encoding) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStrictStruct(encoding)
}

func (encoding *Encoding) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStrictStruct(data, encoding)
}

func (encoding *Encoding) Validate(c context.Context) error {
	if encoding == nil {
		return nil
	}
	for k, v := range encoding.Headers {
		if err := ValidateIdentifier(k); err != nil {
			return nil
		}
		if err := v.Validate(c); err != nil {
			return nil
		}
	}
	return nil
}
