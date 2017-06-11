package openapi3

import (
	"strings"
)

func (swagger *Swagger) resolveRefs() error {
	// Fill required properties
	if swagger.OpenAPI == "" {
		swagger.OpenAPI = "3.0"
	}

	// Prevent infinite recursion
	swagger.visitedSchemas = make(map[*Schema]struct{})
	defer func() {
		swagger.visitedSchemas = nil
	}()

	// Visit all operations
	if paths := swagger.Paths; paths != nil {
		for _, pathItem := range paths {
			if pathItem == nil {
				continue
			}
			for _, endpoint := range pathItem.Operations() {
				if parameters := endpoint.Parameters; parameters != nil {
					for k, parameter := range parameters {
						parameters[k] = swagger.resolveRefsForParameter(parameter)
					}
				}
				if requestBody := endpoint.RequestBody; requestBody != nil {
					endpoint.RequestBody = swagger.resolveRefsForRequestBody(requestBody)
				}
				if responses := endpoint.Responses; responses != nil {
					for k, response := range responses {
						responses[k] = swagger.resolveRefsForResponse(response)
					}
				}
			}
		}
	}

	// Visit all components
	if m := swagger.Components.Headers; m == nil {
		swagger.Components.Headers = make(map[string]*Parameter)
	} else {
		for _, component := range m {
			swagger.resolveRefsForParameter(component)
		}
	}
	if m := swagger.Components.Parameters; m == nil {
		swagger.Components.Parameters = make(map[string]*Parameter)
	} else {
		for _, component := range m {
			swagger.resolveRefsForParameter(component)
		}
	}
	if m := swagger.Components.RequestBodies; m == nil {
		swagger.Components.RequestBodies = make(map[string]*RequestBody)
	} else {
		for _, component := range m {
			swagger.resolveRefsForRequestBody(component)
		}
	}
	if m := swagger.Components.Responses; m == nil {
		swagger.Components.Responses = make(map[string]*Response)
	} else {
		for _, component := range m {
			swagger.resolveRefsForResponse(component)
		}
	}
	if m := swagger.Components.Schemas; m == nil {
		swagger.Components.Schemas = make(map[string]*Schema)
	} else {
		for _, component := range m {
			swagger.resolveRefsForSchema(component)
		}
	}
	if m := swagger.Components.SecuritySchemes; m == nil {
		swagger.Components.SecuritySchemes = make(map[string]*SecurityScheme)
	}
	return nil
}

func trimRefPrefix(component string, prefix string) string {
	if !strings.HasPrefix(component, prefix) {
		return ""
	}
	id := component[len(prefix):]
	if strings.IndexByte(id, '/') >= 0 {
		return ""
	}
	return id
}

func (swagger *Swagger) resolveRefsForParameter(component *Parameter) *Parameter {
	if ref := component.Ref; len(ref) > 0 {
		id := trimRefPrefix(ref, "#/components/parameters/")
		if id != "" {
			resolved := swagger.Components.Parameters[id]
			if resolved != nil {
				resolved.Ref = ref
				return resolved
			}
		}
	}
	if schema := component.Schema; schema != nil {
		component.Schema = swagger.resolveRefsForSchema(schema)
	}
	return component
}

func (swagger *Swagger) resolveRefsForRequestBody(component *RequestBody) *RequestBody {
	if ref := component.Ref; len(ref) > 0 {
		id := trimRefPrefix(ref, "#/components/requestBodies/")
		if id != "" {
			resolved := swagger.Components.RequestBodies[id]
			if resolved != nil {
				resolved.Ref = ref
				return resolved
			}
		}
	}
	if content := component.Content; content != nil {
		for _, contentType := range content {
			if schema := contentType.Schema; schema != nil {
				contentType.Schema = swagger.resolveRefsForSchema(schema)
			}
		}
	}
	return component
}

func (swagger *Swagger) resolveRefsForResponse(component *Response) *Response {
	if ref := component.Ref; len(ref) > 0 {
		id := trimRefPrefix(ref, "#/components/responses/")
		if id != "" {
			resolved := swagger.Components.Responses[id]
			if resolved != nil {
				resolved.Ref = ref
				return resolved
			}
		}
	}
	if content := component.Content; content != nil {
		for _, contentType := range content {
			if schema := contentType.Schema; schema != nil {
				contentType.Schema = swagger.resolveRefsForSchema(schema)
			}
		}
	}
	return component
}

func (swagger *Swagger) resolveRefsForSchema(component *Schema) *Schema {
	if ref := component.Ref; len(ref) > 0 {
		id := trimRefPrefix(ref, "#/components/schemas/")
		if id != "" {
			resolved := swagger.Components.Schemas[id]
			if resolved != nil {
				resolved.Ref = ref
				return resolved
			}
		}
	}

	// Prevent infinite recursion
	if _, isResolved := swagger.visitedSchemas[component]; isResolved {
		return component
	}
	swagger.visitedSchemas[component] = struct{}{}

	// ResolveRefs referred schemas
	if v := component.Items; v != nil {
		component.Items = swagger.resolveRefsForSchema(v)
	}
	if m := component.Properties; m != nil {
		for k, v := range m {
			m[k] = swagger.resolveRefsForSchema(v)
		}
	}
	if v := component.AdditionalKeys; v != nil {
		component.AdditionalKeys = swagger.resolveRefsForSchema(v)
	}
	if v := component.AdditionalProperties; v != nil {
		component.AdditionalProperties = swagger.resolveRefsForSchema(v)
	}
	return component
}

func (swagger *Swagger) resolveRefsForExample(component string) interface{} {
	id := trimRefPrefix(component, "#/components/examples/")
	if id == "" {
		return nil
	}
	return swagger.Components.Examples[id]
}
