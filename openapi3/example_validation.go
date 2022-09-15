package openapi3

import (
	"context"
)

func validateExampleValue(ctx context.Context, input interface{}, schema *Schema) error {
	return schema.VisitJSON(input, MultiErrors())
}
