package openapi3

import "github.com/santhosh-tekuri/jsonschema/v5"

const oasComponentSchema = `{
  "$id": "https://spec.openapis.org/oas/3.1/schema/2021-09-28",
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "components": {
      "type": "object",
      "properties": {
        "schemas": {
          "type": "object",
          "additionalProperties": {
            "$dynamicRef": "#meta"
          }
        }
      }
    }
  },
  "$defs": {
    "schema": {
      "$comment": "https://spec.openapis.org/oas/v3.1.0#schema-object",
      "$dynamicAnchor": "meta",
      "type": [
        "object",
        "boolean"
      ]
    }
  }
}
`

type componentsCompiler struct{}

func (componentsCompiler) Compile(ctx jsonschema.CompilerContext, m map[string]interface{}) (jsonschema.ExtSchema, error) {
	if c, ok := m["components"]; ok {
		resultSchemas := make(componentsSchema)
		if components, ok := c.(map[string]interface{}); ok {
			if s, ok := components["schemas"]; ok {
				if schemas, ok := s.(map[string]interface{}); ok {
					for name := range schemas {
						ss, err := ctx.Compile("components/schemas/"+name, false)
						if err != nil {
							return resultSchemas, err
						}
						resultSchemas[name] = ss
					}
				}
			}
		}
		return resultSchemas, nil
	}

	// nothing to compile, return nil
	return nil, nil
}

type componentsSchema map[string]*jsonschema.Schema

func (c componentsSchema) Validate(ctx jsonschema.ValidationContext, v interface{}) error {
	return nil
}

func (c componentsSchema) GetSchema(key string) *jsonschema.Schema {
	if s, ok := c[key]; ok {
		return s
	}

	return nil
}
