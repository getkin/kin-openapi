package openapi3

import (
	"fmt"
	"strings"
)

func newRefErr(value string) error {
	return fmt.Errorf("Failed to resolve reference '%s'", value)
}

type resolver struct {
	Swagger *Swagger
	visited map[interface{}]struct{}
}

func (resolver *resolver) ResolveComponent(ref string, prefix string) (components *Components, id string, err error) {
	if !strings.HasPrefix(ref, prefix) {
		return nil, "", newRefErr(ref)
	}
	id = ref[len(prefix):]
	if strings.IndexByte(id, '/') >= 0 {
		return nil, "", newRefErr(ref)
	}
	return &resolver.Swagger.Components, id, nil
}

func (resolver *resolver) ResolveRefs() error {
	resolver.visited = make(map[interface{}]struct{})

	// Visit all components
	if m := resolver.Swagger.Components.Headers; m != nil {
		for _, component := range m {
			err := resolver.resolveHeaderRef(component)
			if err != nil {
				return err
			}
		}
	}
	if m := resolver.Swagger.Components.Parameters; m != nil {
		for _, component := range m {
			err := resolver.resolveParameterRef(component)
			if err != nil {
				return err
			}
		}
	}
	if m := resolver.Swagger.Components.RequestBodies; m != nil {
		for _, component := range m {
			err := resolver.resolveRequestBodyRef(component)
			if err != nil {
				return err
			}
		}
	}
	if m := resolver.Swagger.Components.Responses; m != nil {
		for _, component := range m {
			err := resolver.resolveResponseRef(component)
			if err != nil {
				return err
			}
		}
	}
	if m := resolver.Swagger.Components.Schemas; m != nil {
		for _, component := range m {
			err := resolver.resolveSchemaRef(component)
			if err != nil {
				return err
			}
		}
	}
	if m := resolver.Swagger.Components.SecuritySchemes; m != nil {
		for _, component := range m {
			err := resolver.resolveSecuritySchemeRef(component)
			if err != nil {
				return err
			}
		}
	}

	// Visit all operations
	if paths := resolver.Swagger.Paths; paths != nil {
		for _, pathItem := range paths {
			if pathItem == nil {
				continue
			}
			for _, operation := range pathItem.Operations() {
				if parameters := operation.Parameters; parameters != nil {
					for _, parameter := range parameters {
						err := resolver.resolveParameterRef(parameter)
						if err != nil {
							return err
						}
					}
				}
				if requestBody := operation.RequestBody; requestBody != nil {
					err := resolver.resolveRequestBodyRef(requestBody)
					if err != nil {
						return err
					}
				}
				if responses := operation.Responses; responses != nil {
					for _, response := range responses {
						err := resolver.resolveResponseRef(response)
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}
	return nil
}

func (resolver *resolver) resolveHeaderRef(component *HeaderRef) error {
	// Prevent infinite recursion
	visited := resolver.visited
	if _, isVisited := visited[component]; isVisited {
		return nil
	}
	visited[component] = struct{}{}

	// Resolve ref
	const prefix = "#/components/headers"
	if ref := component.Ref; len(ref) > 0 {
		components, id, err := resolver.ResolveComponent(ref, prefix)
		if err != nil {
			return err
		}
		definitions := components.Headers
		if definitions == nil {
			return newRefErr(ref)
		}
		resolved := definitions[id]
		if resolved == nil {
			return newRefErr(ref)
		}
		err = resolver.resolveHeaderRef(resolved)
		if err != nil {
			return err
		}
		component.Value = resolved.Value
	}
	value := component.Value
	if value == nil {
		return nil
	}
	if schema := value.Schema; schema != nil {
		err := resolver.resolveSchemaRef(schema)
		if err != nil {
			return err
		}
	}
	return nil
}

func (resolver *resolver) resolveParameterRef(component *ParameterRef) error {
	// Prevent infinite recursion
	visited := resolver.visited
	if _, isVisited := visited[component]; isVisited {
		return nil
	}
	visited[component] = struct{}{}

	// Resolve ref
	const prefix = "#/components/parameters"
	if ref := component.Ref; len(ref) > 0 {
		components, id, err := resolver.ResolveComponent(ref, prefix)
		if err != nil {
			return err
		}
		definitions := components.Parameters
		if definitions == nil {
			return newRefErr(ref)
		}
		resolved := definitions[id]
		if resolved == nil {
			return newRefErr(ref)
		}
		err = resolver.resolveParameterRef(resolved)
		if err != nil {
			return err
		}
		component.Value = resolved.Value
	}
	value := component.Value
	if value == nil {
		return nil
	}
	if schema := value.Schema; schema != nil {
		err := resolver.resolveSchemaRef(schema)
		if err != nil {
			return err
		}
	}
	return nil
}

func (resolver *resolver) resolveRequestBodyRef(component *RequestBodyRef) error {
	// Prevent infinite recursion
	visited := resolver.visited
	if _, isVisited := visited[component]; isVisited {
		return nil
	}
	visited[component] = struct{}{}

	// Resolve ref
	const prefix = "#/components/requestBodies"
	if ref := component.Ref; len(ref) > 0 {
		components, id, err := resolver.ResolveComponent(ref, prefix)
		if err != nil {
			return err
		}
		definitions := components.RequestBodies
		if definitions == nil {
			return newRefErr(ref)
		}
		resolved := definitions[id]
		if resolved == nil {
			return newRefErr(ref)
		}
		err = resolver.resolveRequestBodyRef(resolved)
		if err != nil {
			return err
		}
		component.Value = resolved.Value
	}
	value := component.Value
	if value == nil {
		return nil
	}
	if content := value.Content; content != nil {
		for _, contentType := range content {
			if schema := contentType.Schema; schema != nil {
				err := resolver.resolveSchemaRef(schema)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (resolver *resolver) resolveResponseRef(component *ResponseRef) error {
	// Prevent infinite recursion
	visited := resolver.visited
	if _, isVisited := visited[component]; isVisited {
		return nil
	}
	visited[component] = struct{}{}

	// Resolve ref
	const prefix = "#/components/responses"
	if ref := component.Ref; len(ref) > 0 {
		components, id, err := resolver.ResolveComponent(ref, prefix)
		if err != nil {
			return err
		}
		definitions := components.Responses
		if definitions == nil {
			return newRefErr(ref)
		}
		resolved := definitions[id]
		if resolved == nil {
			return newRefErr(ref)
		}
		err = resolver.resolveResponseRef(resolved)
		if err != nil {
			return err
		}
		component.Value = resolved.Value
	}
	value := component.Value
	if value == nil {
		return nil
	}
	if content := value.Content; content != nil {
		for _, contentType := range content {
			if schema := contentType.Schema; schema != nil {
				err := resolver.resolveSchemaRef(schema)
				if err != nil {
					return err
				}
				contentType.Schema = schema
			}
		}
	}
	return nil
}

func (resolver *resolver) resolveSchemaRef(component *SchemaRef) error {
	// Prevent infinite recursion
	visited := resolver.visited
	if _, isVisited := visited[component]; isVisited {
		return nil
	}
	visited[component] = struct{}{}

	// Resolve ref
	const prefix = "#/components/schemas"
	if ref := component.Ref; len(ref) > 0 {
		components, id, err := resolver.ResolveComponent(ref, prefix)
		if err != nil {
			return err
		}
		definitions := components.Schemas
		if definitions == nil {
			return newRefErr(ref)
		}
		resolved := definitions[id]
		if resolved == nil {
			return newRefErr(ref)
		}
		err = resolver.resolveSchemaRef(resolved)
		if err != nil {
			return err
		}
	}
	value := component.Value

	// ResolveRefs referred schemas
	if v := value.Items; v != nil {
		err := resolver.resolveSchemaRef(v)
		if err != nil {
			return err
		}
	}
	if m := value.Properties; m != nil {
		for _, v := range m {
			err := resolver.resolveSchemaRef(v)
			if err != nil {
				return err
			}
		}
	}
	if v := value.AdditionalProperties; v != nil {
		err := resolver.resolveSchemaRef(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (resolver *resolver) resolveSecuritySchemeRef(component *SecuritySchemeRef) error {
	// Prevent infinite recursion
	visited := resolver.visited
	if _, isVisited := visited[component]; isVisited {
		return nil
	}
	visited[component] = struct{}{}

	// Resolve ref
	const prefix = "#/components/securitySchemes"
	if ref := component.Ref; len(ref) > 0 {
		components, id, err := resolver.ResolveComponent(ref, prefix)
		if err != nil {
			return err
		}
		definitions := components.SecuritySchemes
		if definitions == nil {
			return newRefErr(ref)
		}
		resolved := definitions[id]
		if resolved == nil {
			return newRefErr(ref)
		}
		err = resolver.resolveSecuritySchemeRef(resolved)
		if err != nil {
			return err
		}
		component.Value = resolved.Value
	}
	return nil
}

func (resolver *resolver) resolveExampleRef(component *ExampleRef) error {
	const prefix = "#/components/examples"
	if ref := component.Ref; len(ref) > 0 {
		components, id, err := resolver.ResolveComponent(ref, prefix)
		if err != nil {
			return err
		}
		definitions := components.Examples
		if definitions == nil {
			return newRefErr(ref)
		}
		resolved := definitions[id]
		if resolved == nil {
			return newRefErr(ref)
		}
		err = resolver.resolveExampleRef(resolved)
		if err != nil {
			return err
		}
		component.Value = resolved.Value
	}
	return nil
}
