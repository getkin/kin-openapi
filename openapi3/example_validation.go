package openapi3

func validateExampleValue(input interface{}, schema *Schema, validationOpts *ValidationOptions) error {
	opts := make([]SchemaValidationOption, 0, 3)

	if validationOpts.ExamplesValidation.AsReq {
		opts = append(opts, VisitAsRequest())
	} else if validationOpts.ExamplesValidation.AsRes {
		opts = append(opts, VisitAsResponse())
	}
	opts = append(opts, MultiErrors())

	return schema.VisitJSON(input, opts...)
}
