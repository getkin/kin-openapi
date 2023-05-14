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

	resultProperties := result.AllOf[0].Value.Properties
	for name, schema := range schema.Properties {
		resultProperties[name] = schema
	}

	result.AllOf[0].Value.Required = addRequired(result.AllOf[0].Value.Required, schema)
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
