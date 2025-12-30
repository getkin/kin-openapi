package openapi3

import "context"

func validateExampleValue(ctx context.Context, input any, schema *Schema) error {
	opts := make([]SchemaValidationOption, 0, 2)

	vo := getValidationOptions(ctx)
	if vo.examplesValidationAsReq {
		opts = append(opts, VisitAsRequest())
	} else if vo.examplesValidationAsRes {
		opts = append(opts, VisitAsResponse())
	}

	if vo.jsonSchema2020ValidationEnabled {
		opts = append(opts, EnableJSONSchema2020())
	}

	opts = append(opts, MultiErrors())

	return schema.VisitJSON(input, opts...)
}
