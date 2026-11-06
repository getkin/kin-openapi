package openapi3

import (
	"context"
)

// Callback is specified by OpenAPI/Swagger standard version 3.
// See https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md#callback-object
type Callback struct {
	Extensions map[string]any `json:"-" yaml:"-"`
	Origin     *Origin        `json:"-" yaml:"-"`

	m map[string]*PathItem
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
	return callback.validate(ctx, nil)
}

// validate skips the callbacks in visiting, those being validated further up:
// an operation of a callback may name that same callback among its own
// callbacks.
func (callback *Callback) validate(ctx context.Context, visiting map[*Callback]struct{}) error {
	if _, ok := visiting[callback]; ok {
		return nil
	}
	if visiting == nil {
		visiting = make(map[*Callback]struct{})
	}
	visiting[callback] = struct{}{}
	defer delete(visiting, callback)

	for _, key := range callback.Keys() {
		v := callback.Value(key)
		if err := v.validate(ctx, visiting); err != nil {
			return err
		}
	}

	return validateExtensions(ctx, callback.Extensions, callback.Origin)
}

// validate is CallbackRef.Validate passing visiting on to Callback.validate.
func (x *CallbackRef) validate(ctx context.Context, visiting map[*Callback]struct{}) error {
	if x == nil {
		return newUnresolvedRef("", nil)
	}
	if err := x.validateExtras(ctx); err != nil {
		return err
	}
	if v := x.Value; v != nil {
		return v.validate(ctx, visiting)
	}
	return newUnresolvedRef(x.Ref, x.Origin)
}
