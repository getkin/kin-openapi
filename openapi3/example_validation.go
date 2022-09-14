package openapi3

import (
	"context"
)

func ValidateExampleValue(ctx context.Context, input interface{}, schema *Schema) error {
	opts := make([]SchemaValidationOption, 0, 3) // 3 potential opts here
	opts = append(opts, VisitAsResponse())
	opts = append(opts, VisitAsRequest())
	opts = append(opts, MultiErrors())

	// Validate input with the schema
	if err := schema.VisitJSON(input, opts...); err != nil {
		return err
	}

	return nil
}
