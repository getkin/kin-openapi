package openapi3

import (
	"context"
	"sort"

	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// Callback is specified by OpenAPI/Swagger standard version 3.
// See https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md#callback-object
type Callback struct {
	Extensions map[string]interface{} `json:"-" yaml:"-"`

	om *orderedmap.OrderedMap[string, *PathItem]
}

// NewCallback builds a Callback object with path items in insertion order.
func NewCallback(opts ...NewCallbackOption) *Callback {
	Callback := NewCallbackWithCapacity(len(opts))
	for _, opt := range opts {
		opt(Callback)
	}
	return Callback
}

// NewCallbackOption describes options to NewCallback func
type NewCallbackOption func(*Callback)

// WithCallback adds Callback as an option to NewCallback
func WithCallback(cb string, pathItem *PathItem) NewCallbackOption {
	return func(callback *Callback) {
		if p := pathItem; p != nil && cb != "" {
			callback.Set(cb, p)
		}
	}
}

// Validate returns an error if Callback does not comply with the OpenAPI spec.
func (callback *Callback) Validate(ctx context.Context, opts ...ValidationOption) error {
	ctx = WithValidationOptions(ctx, opts...)

	keys := make([]string, 0, callback.Len())
	for key := range callback.Map() {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		v := callback.Value(key)
		if err := v.Validate(ctx); err != nil {
			return err
		}
	}

	return validateExtensions(ctx, callback.Extensions)
}
