package openapi3

// Merge replaces objects under AllOf with a flattened equivalent
func Merge(schema Schema) *Schema {
	if !isListOfObjects(&schema) {
		return &schema
	}

	for _, allOfSchema := range schema.AllOf {
		add(&schema, allOfSchema.Value)
	}

	schema.AllOf = nil
	return &schema
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
		if result.Properties == nil {
			result.Properties = Schemas{}
		}
		result.Properties[name] = schema
	}

	result.Required = addRequired(result.Required, schema)
}

// addRequired combines required fields from schema into 'required'
func addRequired(required []string, schema *Schema) []string {
	requiredSet := make(map[string]struct{})
	for _, name := range schema.Required {
		requiredSet[name] = struct{}{}
	}
	for _, name := range required {
		requiredSet[name] = struct{}{}
	}

	requiredList := make([]string, len(requiredSet))
	i := 0
	for name := range requiredSet {
		requiredList[i] = name
		i++
	}

	return requiredList
}
