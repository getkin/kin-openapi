package openapi3

import (
	"context"
	"errors"
	"fmt"

	"github.com/getkin/kin-openapi/jsoninfo"
	"github.com/go-openapi/jsonpointer"
)

// Headers represents components' header mapping
type Headers map[string]*HeaderRef

var _ jsonpointer.JSONPointable = (*Headers)(nil)

func (hs Headers) JSONLookup(token string) (interface{}, error) {
	ref, ok := hs[token]
	if ref == nil || !ok {
		return nil, fmt.Errorf("object has no field %q", token)
	}

	if ref.Ref != "" {
		return &Ref{Ref: ref.Ref}, nil
	}
	return ref.Value, nil
}

// Header is specified by OpenAPI/Swagger 3.0 standard.
// See https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.0.md#headerObject
type Header struct {
	Parameter
}

var _ jsonpointer.JSONPointable = (*Header)(nil)

func (h *Header) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStrictStruct(data, h)
}

// SerializationMethod returns a header's serialization method.
func (h *Header) SerializationMethod() (*SerializationMethod, error) {
	style := h.Style
	if style == "" {
		style = SerializationSimple
	}
	explode := false
	if h.Explode != nil {
		explode = *h.Explode
	}
	return &SerializationMethod{Style: style, Explode: explode}, nil
}

// Validate goes through the receiver value and its descendants and errors on any non compliance to the OpenAPIv3 specification.
func (h *Header) Validate(ctx context.Context) error {
	if h.Name != "" {
		return errors.New("header 'name' MUST NOT be specified, it is given in the corresponding headers map")
	}
	if h.In != "" {
		return errors.New("header 'in' MUST NOT be specified, it is implicitly in header")
	}

	// Validate a parameter's serialization method.
	sm, err := h.SerializationMethod()
	if err != nil {
		return err
	}
	if smSupported := false ||
		sm.Style == SerializationSimple && !sm.Explode ||
		sm.Style == SerializationSimple && sm.Explode; !smSupported {
		e := fmt.Errorf("serialization method with style=%q and explode=%v is not supported by a header parameter", sm.Style, sm.Explode)
		return fmt.Errorf("header schema is invalid: %v", e)
	}

	if (h.Schema == nil) == (h.Content == nil) {
		e := fmt.Errorf("parameter must contain exactly one of content and schema: %v", h)
		return fmt.Errorf("header schema is invalid: %v", e)
	}
	if schema := h.Schema; schema != nil {
		if err := schema.Validate(ctx); err != nil {
			return fmt.Errorf("header schema is invalid: %v", err)
		}
	}

	if content := h.Content; content != nil {
		if err := content.Validate(ctx); err != nil {
			return fmt.Errorf("header content is invalid: %v", err)
		}
	}
	return nil
}

func (h Header) JSONLookup(token string) (interface{}, error) {
	switch token {
	case "schema":
		if h.Schema != nil {
			if h.Schema.Ref != "" {
				return &Ref{Ref: h.Schema.Ref}, nil
			}
			return h.Schema.Value, nil
		}
	case "name":
		return h.Name, nil
	case "in":
		return h.In, nil
	case "description":
		return h.Description, nil
	case "style":
		return h.Style, nil
	case "explode":
		return h.Explode, nil
	case "allowEmptyh":
		return h.AllowEmptyValue, nil
	case "allowReserved":
		return h.AllowReserved, nil
	case "deprecated":
		return h.Deprecated, nil
	case "required":
		return h.Required, nil
	case "example":
		return h.Example, nil
	case "examples":
		return h.Examples, nil
	case "content":
		return h.Content, nil
	}

	v, _, err := jsonpointer.GetForToken(h.ExtensionProps, token)
	return v, err
}
