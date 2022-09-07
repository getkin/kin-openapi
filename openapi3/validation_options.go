package openapi3

import "context"

// ValidationOption allows the modification of how the OpenAPI document is validated.
type ValidationOption func(options *validationOptions)

type validationOptions struct {
	schemaFormatValidationEnabled   bool
	schemaPatternValidationDisabled bool
}

type validationOptionsKey struct{}

// EnableSchemaFormatValidation makes Validate not return an error when validating documents that mention schema formats that are not defined by the OpenAPIv3 specification.
func EnableSchemaFormatValidation() ValidationOption {
	return func(options *validationOptions) {
		options.schemaFormatValidationEnabled = true
	}
}

// DisableSchemaPatternValidation makes Validate not return an error when validating patterns that are not supported by the Go regexp engine.
func DisableSchemaPatternValidation() ValidationOption {
	return func(options *validationOptions) {
		options.schemaPatternValidationDisabled = true
	}
}

func withValidationOptions(ctx context.Context, options *validationOptions) context.Context {
	return context.WithValue(ctx, validationOptionsKey{}, options)
}

func getValidationOptions(ctx context.Context) *validationOptions {
	if options, ok := ctx.Value(validationOptionsKey{}).(*validationOptions); ok {
		return options
	}
	return &validationOptions{}
}
