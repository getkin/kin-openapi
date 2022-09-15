package openapi3

import (
	"context"
)

func validateExampleValue(ctx context.Context, input interface{}, schema *Schema) error {
	opts := make([]SchemaValidationOption, 0, 1)
	opts = append(opts, MultiErrors())

	// Validate input with the schema
	if err := schema.VisitJSON(input, opts...); err != nil {
		return err
	}

	return nil
}
