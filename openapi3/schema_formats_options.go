package openapi3

// SchemaFormatOption allows the modification of how the OpenAPI document is validated.
type SchemaFormatOption func(options *SchemaFormatOptions)

// SchemaFormatOptions provides configuration for validating OpenAPI documents.
type SchemaFormatOptions struct {
	fromOpenAPIMinorVersion uint64
}

// FromOpenAPIMinorVersion allows to declare a string format available only at some minor OpenAPI version
func FromOpenAPIMinorVersion(fromMinorVersion uint64) SchemaFormatOption {
	return func(options *SchemaFormatOptions) {
		options.fromOpenAPIMinorVersion = fromMinorVersion
	}
}
