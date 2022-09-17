package openapi3

import "context"

func validateExampleValue(ctx context.Context, input interface{}, schema *Schema) error {
	opts := make([]SchemaValidationOption, 0, 3)

	if vo := getValidationOptions(ctx); vo.ExamplesValidation.AsReq {
		opts = append(opts, VisitAsRequest())
	} else if vo.ExamplesValidation.AsRes {
		opts = append(opts, VisitAsResponse())
	}
	opts = append(opts, MultiErrors())

	return schema.VisitJSON(input, opts...)
}
