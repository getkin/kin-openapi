package openapi3

// newSchemaError starts building a SchemaError about value failing schema.
func newSchemaError(schema *Schema, value any) *SchemaError {
	return &SchemaError{
		Value:  value,
		Schema: schema,
	}
}

// newSchemaError starts building a SchemaError about value failing schema,
// carrying over the message customizer of settings, which may be nil.
func (settings *schemaValidationSettings) newSchemaError(schema *Schema, value any) *SchemaError {
	err := newSchemaError(schema, value)
	if settings != nil {
		err.customizeMessageError = settings.customizeMessageError
	}
	return err
}

// withPathKey appends key to the path of the value that failed validation.
func (err *SchemaError) withPathKey(key string) *SchemaError {
	err.reversePath = append(err.reversePath, key)
	return err
}

func (err *SchemaError) withField(field string) *SchemaError {
	err.SchemaField = field
	return err
}

func (err *SchemaError) withReason(reason string) *SchemaError {
	err.Reason = reason
	return err
}

func (err *SchemaError) withOrigin(origin error) *SchemaError {
	err.Origin = origin
	return err
}

// withCause records the error of the JSON Schema 2020-12 validator this error is converted from.
func (err *SchemaError) withCause(cause error) *SchemaError {
	err.cause = cause
	return err
}
