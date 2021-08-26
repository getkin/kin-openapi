package openapi3

type schemaJSON = map[string]interface{}
type schemasJSON = map[string]schemaJSON

func (s *SchemaRef) fromOpenAPISchema(settings *schemaValidationSettings) (schema schemaJSON) {
	if ref := s.Ref; ref != "" {
		return schemaJSON{"$ref": ref}
	}
	return s.Value.fromOpenAPISchema(settings)
}

func (s *Schema) fromOpenAPISchema(settings *schemaValidationSettings) (schema schemaJSON) {
	schema = make(schemaJSON)

	if sEnum := s.Enum; len(sEnum) != 0 {
		schema["enum"] = sEnum
	}

	if sMinLength := s.MinLength; sMinLength != 0 {
		schema["minLength"] = sMinLength
	}
	if sMaxLength := s.MaxLength; nil != sMaxLength {
		schema["maxLength"] = *sMaxLength
	}

	if sFormat := s.Format; sFormat != "" {
		schema["format"] = sFormat
	}
	if sPattern := s.Pattern; sPattern != "" {
		schema["pattern"] = sPattern
	}

	if nil != s.Min {
		schema["minimum"] = *s.Min
	}
	if nil != s.Max {
		schema["maximum"] = *s.Max
	}
	if sExMin := s.ExclusiveMin; sExMin {
		schema["exclusiveMinimum"] = sExMin
	}
	if sExMax := s.ExclusiveMax; sExMax {
		schema["exclusiveMaximum"] = sExMax
	}
	if nil != s.MultipleOf {
		schema["multipleOf"] = *s.MultipleOf
	}

	if sUniq := s.UniqueItems; sUniq {
		schema["uniqueItems"] = sUniq
	}
	if sMinItems := s.MinItems; sMinItems != 0 {
		schema["minItems"] = sMinItems
	}
	if nil != s.MaxItems {
		schema["maxItems"] = *s.MaxItems
	}
	if sItems := s.Items; nil != sItems {
		if sItems.Value != nil && sItems.Value.IsEmpty() {
			schema["items"] = []schemaJSON{}
		} else {
			schema["items"] = []schemaJSON{sItems.fromOpenAPISchema(settings)}
		}
	}

	if sMinProps := s.MinProps; sMinProps != 0 {
		schema["minProperties"] = sMinProps
	}
	if nil != s.MaxProps {
		schema["maxProperties"] = *s.MaxProps
	}

	if sRequired := s.Required; len(sRequired) != 0 {
		required := make([]string, 0, len(sRequired))
		for _, propName := range sRequired {
			prop := s.Properties[propName]
			switch {
			case settings.asreq && prop != nil && prop.Value.ReadOnly:
			case settings.asrep && prop != nil && prop.Value.WriteOnly:
			default:
				required = append(required, propName)
			}
		}
		schema["required"] = required
	}

	if count := len(s.Properties); count != 0 {
		properties := make(schemasJSON, count)
		for propName, prop := range s.Properties {
			properties[propName] = prop.fromOpenAPISchema(settings)
		}
		schema["properties"] = properties
	}

	if sAddProps := s.AdditionalPropertiesAllowed; sAddProps != nil {
		// TODO: complete handling
		schema["additionalProperties"] = sAddProps
	}

	if sAllOf := s.AllOf; len(sAllOf) != 0 {
		allOf := make([]schemaJSON, 0, len(sAllOf))
		for _, sOf := range sAllOf {
			allOf = append(allOf, sOf.fromOpenAPISchema(settings))
		}
		schema["allOf"] = allOf
	}
	if sAnyOf := s.AnyOf; len(sAnyOf) != 0 {
		anyOf := make([]schemaJSON, 0, len(sAnyOf))
		for _, sOf := range sAnyOf {
			anyOf = append(anyOf, sOf.fromOpenAPISchema(settings))
		}
		schema["anyOf"] = anyOf
	}
	if sOneOf := s.OneOf; len(sOneOf) != 0 {
		oneOf := make([]schemaJSON, 0, len(sOneOf))
		for _, sOf := range sOneOf {
			oneOf = append(oneOf, sOf.fromOpenAPISchema(settings))
		}
		schema["oneOf"] = oneOf
	}

	if sType := s.Type; sType != "" {
		schema["type"] = []string{s.Type}
	}

	if sNot := s.Not; sNot != nil {
		schema["not"] = sNot.fromOpenAPISchema(settings)
	}

	if s.IsEmpty() {
		schema = schemaJSON{"not": schemaJSON{"type": "null"}}
	}

	if s.Nullable {
		schema = schemaJSON{"anyOf": []schemaJSON{
			{"type": "null"},
			schema,
		}}
	}

	schema["$schema"] = "http://json-schema.org/draft-04/schema#"
	//FIXME
	//https://github.com/openapi-contrib/openapi-schema-to-json-schema/blob/45c080c38027c30652263b4cc44cd3534f5ccc1b/lib/converters/schema.js
	//https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#schemaObject
	return
}
