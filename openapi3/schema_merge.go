package openapi3

// Merge replaces objects under AllOf with a flattened equivalent
func Merge(schema Schema) *Schema {
	if !isListOfObjects(&schema) {
		return &schema
	}

	result := schema
	result.AllOf = SchemaRefs{schema.AllOf[0]}
	for _, schema := range schema.AllOf[1:] {
		add(&result, schema.Value)
	}

	return &result
}

func isListOfObjects(schema *Schema) bool {
	if schema == nil || schema.AllOf == nil {
		return false
	}

	for _, subSchema := range schema.AllOf {
		if subSchema.Value.Type != "object" {
			return false
		}
	}

	return true
}

func add(result, schema *Schema) {
	if schema == nil || schema.Properties == nil {
		return
	}

	for name, schema := range schema.Properties {
		result.AllOf[0].Value.Properties[name] = schema
	}
}