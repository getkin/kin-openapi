package openapi3

import (
	"context"
	"fmt"

	"github.com/getkin/kin-openapi/jsoninfo"
	"github.com/go-openapi/jsonpointer"
)

type Headers map[string]*HeaderRef

var _ jsonpointer.JSONPointable = (*Headers)(nil)

func (h Headers) JSONLookup(token string) (interface{}, error) {
	ref, ok := h[token]
	if ref == nil || !ok {
		return nil, fmt.Errorf("object has no field %q", token)
	}

	if ref.Ref != "" {
		return &Ref{Ref: ref.Ref}, nil
	}
	return ref.Value, nil
}

type Header struct {
	ExtensionProps

	// Optional description. Should use CommonMark syntax.
	Description string      `json:"description,omitempty" yaml:"description,omitempty"`
	Deprecated  bool        `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	Required    bool        `json:"required,omitempty" yaml:"required,omitempty"`
	Schema      *SchemaRef  `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example     interface{} `json:"example,omitempty" yaml:"example,omitempty"`
	Examples    Examples    `json:"examples,omitempty" yaml:"examples,omitempty"`
	Content     Content     `json:"content,omitempty" yaml:"content,omitempty"`
}

var _ jsonpointer.JSONPointable = (*Header)(nil)

func (value *Header) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStrictStruct(data, value)
}

func (value *Header) Validate(ctx context.Context) error {
	if v := value.Schema; v != nil {
		if err := v.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (value Header) JSONLookup(token string) (interface{}, error) {
	switch token {
	case "schema":
		if value.Schema != nil {
			if value.Schema.Ref != "" {
				return &Ref{Ref: value.Schema.Ref}, nil
			}
			return value.Schema.Value, nil
		}
	case "description":
		return value.Description, nil
	case "deprecated":
		return value.Deprecated, nil
	case "required":
		return value.Required, nil
	case "example":
		return value.Example, nil
	case "examples":
		return value.Examples, nil
	case "content":
		return value.Content, nil
	}

	v, _, err := jsonpointer.GetForToken(value.ExtensionProps, token)
	return v, err
}
