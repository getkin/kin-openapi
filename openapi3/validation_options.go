package openapi3

import "context"

// ValidationOption allows the modification of how the OpenAPI document is validated.
type ValidationOption func(options *ValidationOptions)

// ValidationOptions provides configuration for validating OpenAPI documents.
type ValidationOptions struct {
	SchemaFormatValidationEnabled                    bool
	SchemaPatternValidationDisabled                  bool
	ExamplesValidationDisabled                       bool
	examplesValidationAsReq, examplesValidationAsRes bool
}

type validationOptionsKey struct{}

// EnableSchemaFormatValidation makes Validate not return an error when validating documents that mention schema formats that are not defined by the OpenAPIv3 specification.
// By default, schema format validation is disabled.
func EnableSchemaFormatValidation() ValidationOption {
	return func(options *ValidationOptions) {
		options.SchemaFormatValidationEnabled = true
	}
}

// DisableSchemaFormatValidation does the opposite of EnableSchemaFormatValidation.
// By default, schema format validation is disabled.
func DisableSchemaFormatValidation() ValidationOption {
	return func(options *ValidationOptions) {
		options.SchemaFormatValidationEnabled = false
	}
}

// EnableSchemaPatternValidation does the opposite of DisableSchemaPatternValidation.
// By default, schema pattern validation is enabled.
func EnableSchemaPatternValidation() ValidationOption {
	return func(options *ValidationOptions) {
		options.SchemaPatternValidationDisabled = false
	}
}

// DisableSchemaPatternValidation makes Validate not return an error when validating patterns that are not supported by the Go regexp engine.
func DisableSchemaPatternValidation() ValidationOption {
	return func(options *ValidationOptions) {
		options.SchemaPatternValidationDisabled = true
	}
}

// EnableExamplesValidation does the opposite of DisableExamplesValidation.
// By default, all schema examples are validated.
func EnableExamplesValidation() ValidationOption {
	return func(options *ValidationOptions) {
		options.ExamplesValidationDisabled = false
	}
}

// DisableExamplesValidation disables all example schema validation.
// By default, all schema examples are validated.
func DisableExamplesValidation() ValidationOption {
	return func(options *ValidationOptions) {
		options.ExamplesValidationDisabled = true
	}
}

// WithValidationOptions allows adding validation options to a context object that can be used when validationg any OpenAPI type.
func WithValidationOptions(ctx context.Context, opts ...ValidationOption) context.Context {
	if len(opts) == 0 {
		return ctx
	}
	options := &ValidationOptions{}
	for _, opt := range opts {
		opt(options)
	}
	return context.WithValue(ctx, validationOptionsKey{}, options)
}

func getValidationOptions(ctx context.Context) *ValidationOptions {
	if options, ok := ctx.Value(validationOptionsKey{}).(*ValidationOptions); ok {
		return options
	}
	return &ValidationOptions{}
}
