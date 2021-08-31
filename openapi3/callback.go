package openapi3

import (
	"context"
	"fmt"

	"github.com/go-openapi/jsonpointer"
)

// Callbacks represents components' callback mapping
type Callbacks map[string]*CallbackRef

var _ jsonpointer.JSONPointable = (*Callbacks)(nil)

func (c Callbacks) JSONLookup(token string) (interface{}, error) {
	ref, ok := c[token]
	if ref == nil || !ok {
		return nil, fmt.Errorf("object has no field %q", token)
	}

	if ref.Ref != "" {
		return &Ref{Ref: ref.Ref}, nil
	}
	return ref.Value, nil
}

// Callback is specified by OpenAPI/Swagger standard version 3.0.
type Callback map[string]*PathItem

// Validate goes through the receiver value and its descendants and errors on any non compliance to the OpenAPIv3 specification.
func (cb Callback) Validate(ctx context.Context) error {
	for _, v := range cb {
		if err := v.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}
