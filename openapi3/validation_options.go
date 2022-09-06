package openapi3

import "context"

type ValidationOption func(options *validationOptions)

type validationOptions struct {
	schemaFormatValidationDisabled  bool
	schemaPatternValidationDisabled bool
}

type contextKey string

const (
	contextKeyValidationOptions contextKey = "validationOptions"
)

func DisableSchemaFormatValidation() ValidationOption {
	return func(options *validationOptions) {
		options.schemaFormatValidationDisabled = true
	}
}

func DisableSchemaPatternValidation() ValidationOption {
	return func(options *validationOptions) {
		options.schemaPatternValidationDisabled = true
	}
}

func withValidationOptions(ctx context.Context, options *validationOptions) context.Context {
	return context.WithValue(ctx, contextKeyValidationOptions, options)
}

func getValidationOptions(ctx context.Context) *validationOptions {
	if options, ok := ctx.Value(contextKeyValidationOptions).(*validationOptions); ok {
		return options
	}
	return &validationOptions{}
}
