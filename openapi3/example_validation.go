package openapi3

func validateExampleValue(input interface{}, schema *Schema, vo *ValidationOptions) error {
	opts := make([]SchemaValidationOption, 0, 3)

	if vo.ExamplesValidation.AsReq {
		opts = append(opts, VisitAsRequest())
	} else if vo.ExamplesValidation.AsRes {
		opts = append(opts, VisitAsResponse())
	}
	opts = append(opts, MultiErrors())

	return schema.VisitJSON(input, opts...)
}
