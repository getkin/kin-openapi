package openapi3

import (
	"context"
	"sort"
)

// Callback is specified by OpenAPI/Swagger standard version 3.
// See https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md#callback-object
type Callback map[string]*PathItem

// Validate returns an error if Callback does not comply with the OpenAPI spec.
func (callback Callback) Validate(ctx context.Context, opts ...ValidationOption) error {
	ctx = WithValidationOptions(ctx, opts...)

	keys := make([]string, 0, len(callback))
	for key := range callback {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		v := callback[key]
		if err := v.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}
