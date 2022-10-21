package openapi3

import "context"

// ValidationOption allows the modification of how the OpenAPI document is validated.
type ValidationOption func(options *ValidationOptions)

// ValidationOptions provides configuration for validating OpenAPI documents.
type ValidationOptions struct {
	schemaFormatValidationEnabled                    bool
	schemaPatternValidationDisabled                  bool
	examplesValidationDisabled                       bool
	examplesValidationAsReq, examplesValidationAsRes bool
}

type validationOptionsKey struct{}

// EnableSchemaFormatValidation makes Validate not return an error when validating documents that mention schema formats that are not defined by the OpenAPIv3 specification.
func EnableSchemaFormatValidation() ValidationOption {
	return func(options *ValidationOptions) {
		options.schemaFormatValidationEnabled = true
	}
}

// DisableSchemaPatternValidation makes Validate not return an error when validating patterns that are not supported by the Go regexp engine.
func DisableSchemaPatternValidation() ValidationOption {
	return func(options *ValidationOptions) {
		options.schemaPatternValidationDisabled = true
	}
}

// DisableExamplesValidation disables all example schema validation.
func DisableExamplesValidation() ValidationOption {
	return func(options *ValidationOptions) {
		options.examplesValidationDisabled = true
	}
}

// WithValidationOptions allows adding validation options to a context object that can be used when validationg any OpenAPI type.
func WithValidationOptions(ctx context.Context, options *ValidationOptions) context.Context {
	return context.WithValue(ctx, validationOptionsKey{}, options)
}

func getValidationOptions(ctx context.Context) *ValidationOptions {
	if options, ok := ctx.Value(validationOptionsKey{}).(*ValidationOptions); ok {
		return options
	}
	return &ValidationOptions{}
}
