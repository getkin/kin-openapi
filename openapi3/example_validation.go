package openapi3

func validateExampleValue(input interface{}, schema *Schema) error {
	return schema.VisitJSON(input, MultiErrors())
}
