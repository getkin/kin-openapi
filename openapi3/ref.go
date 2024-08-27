package openapi3

import (
	"context"
)

//go:generate go run refsgenerator.go

// Ref is specified by OpenAPI/Swagger 3.0 standard.
// See https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md#reference-object
type Ref struct {
	Ref        string         `json:"$ref" yaml:"$ref"`
	Extensions map[string]any `json:"extensions,omitempty" yaml:"$ref,omitempty"`
}

// Validate returns an error if Extensions does not comply with the OpenAPI spec.
func (e *Ref) Validate(ctx context.Context, opts ...ValidationOption) error {
	ctx = WithValidationOptions(ctx, opts...)
	return validateExtensions(ctx, e.Extensions)
}
