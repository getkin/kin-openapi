package openapi3

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"
)

func foundUnresolvedRef(ref string) error {
	return fmt.Errorf("Found unresolved ref: '%s'", ref)
}

func failedToResolveRefDefinitions(value string) error {
	return fmt.Errorf("Failed to resolve fragment in URI: '%s'", value)
}

func failedToResolveRefFragment(value string) error {
	return fmt.Errorf("Failed to resolve fragment in URI: '%s'", value)
}

func failedToResolveRefFragmentPart(value string, what string) error {
	return fmt.Errorf("Failed to resolve '%s' in fragment in URI: '%s'", what, value)
}

type SwaggerLoader struct {
	IsExternalRefsAllowed  bool
	Context                context.Context
	LoadSwaggerFromURIFunc func(loader *SwaggerLoader, url *url.URL) (*Swagger, error)
	visited                map[interface{}]struct{}
}

func NewSwaggerLoader() *SwaggerLoader {
	return &SwaggerLoader{}
}

func (swaggerLoader *SwaggerLoader) LoadSwaggerFromURI(location *url.URL) (*Swagger, error) {
	f := swaggerLoader.LoadSwaggerFromURIFunc
	if f != nil {
		return f(swaggerLoader, location)
	}
	if location.Scheme != "" || location.Host != "" || location.RawQuery != "" {
		return nil, fmt.Errorf("Unsupported URI: '%s'", location.String())
	}
	data, err := ioutil.ReadFile(location.Path)
	if err != nil {
		return nil, err
	}
	return swaggerLoader.LoadSwaggerFromData(data)
}

func (swaggerLoader *SwaggerLoader) LoadSwaggerFromFile(path string) (*Swagger, error) {
	f := swaggerLoader.LoadSwaggerFromURIFunc
	if f != nil {
		return f(swaggerLoader, &url.URL{
			Path: path,
		})
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return swaggerLoader.LoadSwaggerFromData(data)
}

func (swaggerLoader *SwaggerLoader) LoadSwaggerFromData(data []byte) (*Swagger, error) {
	swagger := &Swagger{}
	err := json.Unmarshal(data, swagger)
	if err != nil {
		return nil, err
	}
	return swagger, swaggerLoader.ResolveRefsIn(swagger)
}

func (resolver *SwaggerLoader) ResolveRefsIn(swagger *Swagger) error {
	resolver.visited = make(map[interface{}]struct{})

	// Visit all components
	if m := swagger.Components.Headers; m != nil {
		for _, component := range m {
			err := resolver.resolveHeaderRef(swagger, component)
			if err != nil {
				return err
			}
		}
	}
	if m := swagger.Components.Parameters; m != nil {
		for _, component := range m {
			err := resolver.resolveParameterRef(swagger, component)
			if err != nil {
				return err
			}
		}
	}
	if m := swagger.Components.RequestBodies; m != nil {
		for _, component := range m {
			err := resolver.resolveRequestBodyRef(swagger, component)
			if err != nil {
				return err
			}
		}
	}
	if m := swagger.Components.Responses; m != nil {
		for _, component := range m {
			err := resolver.resolveResponseRef(swagger, component)
			if err != nil {
				return err
			}
		}
	}
	if m := swagger.Components.Schemas; m != nil {
		for _, component := range m {
			err := resolver.resolveSchemaRef(swagger, component)
			if err != nil {
				return err
			}
		}
	}
	if m := swagger.Components.SecuritySchemes; m != nil {
		for _, component := range m {
			err := resolver.resolveSecuritySchemeRef(swagger, component)
			if err != nil {
				return err
			}
		}
	}

	// Visit all operations
	if paths := swagger.Paths; paths != nil {
		for _, pathItem := range paths {
			if pathItem == nil {
				continue
			}
			for _, operation := range pathItem.Operations() {
				if parameters := operation.Parameters; parameters != nil {
					for _, parameter := range parameters {
						err := resolver.resolveParameterRef(swagger, parameter)
						if err != nil {
							return err
						}
					}
				}
				if requestBody := operation.RequestBody; requestBody != nil {
					err := resolver.resolveRequestBodyRef(swagger, requestBody)
					if err != nil {
						return err
					}
				}
				if responses := operation.Responses; responses != nil {
					for _, response := range responses {
						err := resolver.resolveResponseRef(swagger, response)
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

func (resolver *SwaggerLoader) resolveComponent(swagger *Swagger, ref string, prefix string) (components *Components, id string, err error) {
	if !strings.HasPrefix(ref, "#") {
		if !resolver.IsExternalRefsAllowed {
			return nil, "", fmt.Errorf("Encountered non-allowed external reference: '%s'", ref)
		}
		parsedURL, err := url.Parse(ref)
		if err != nil {
			return nil, "", fmt.Errorf("Can't parse reference: '%s': %v", ref, parsedURL)
		}
		fragment := parsedURL.Fragment
		parsedURL.Fragment = ""
		swagger, err = resolver.LoadSwaggerFromURI(parsedURL)
		if err != nil {
			return nil, "", fmt.Errorf("Error while resolving reference '%s': %v", ref, err)
		}
		ref = fragment
	}
	if !strings.HasPrefix(ref, prefix) {
		return nil, "", failedToResolveRefFragment(ref)
	}
	id = ref[len(prefix):]
	if strings.IndexByte(id, '/') >= 0 {
		return nil, "", failedToResolveRefFragmentPart(ref, id)
	}
	return &swagger.Components, id, nil
}

func (resolver *SwaggerLoader) resolveHeaderRef(swagger *Swagger, component *HeaderRef) error {
	// Prevent infinite recursion
	visited := resolver.visited
	if _, isVisited := visited[component]; isVisited {
		return nil
	}
	visited[component] = struct{}{}

	// Resolve ref
	const prefix = "#/components/headers/"
	if ref := component.Ref; len(ref) > 0 {
		components, id, err := resolver.resolveComponent(swagger, ref, prefix)
		if err != nil {
			return err
		}
		definitions := components.Headers
		if definitions == nil {
			return failedToResolveRefFragment(ref)
		}
		resolved := definitions[id]
		if resolved == nil {
			return failedToResolveRefFragment(ref)
		}
		err = resolver.resolveHeaderRef(swagger, resolved)
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
		err := resolver.resolveSchemaRef(swagger, schema)
		if err != nil {
			return err
		}
	}
	return nil
}

func (resolver *SwaggerLoader) resolveParameterRef(swagger *Swagger, component *ParameterRef) error {
	// Prevent infinite recursion
	visited := resolver.visited
	if _, isVisited := visited[component]; isVisited {
		return nil
	}
	visited[component] = struct{}{}

	// Resolve ref
	const prefix = "#/components/parameters/"
	if ref := component.Ref; len(ref) > 0 {
		components, id, err := resolver.resolveComponent(swagger, ref, prefix)
		if err != nil {
			return err
		}
		definitions := components.Parameters
		if definitions == nil {
			return failedToResolveRefFragmentPart(ref, "parameters")
		}
		resolved := definitions[id]
		if resolved == nil {
			return failedToResolveRefFragmentPart(ref, id)
		}
		err = resolver.resolveParameterRef(swagger, resolved)
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
		err := resolver.resolveSchemaRef(swagger, schema)
		if err != nil {
			return err
		}
	}
	return nil
}

func (resolver *SwaggerLoader) resolveRequestBodyRef(swagger *Swagger, component *RequestBodyRef) error {
	// Prevent infinite recursion
	visited := resolver.visited
	if _, isVisited := visited[component]; isVisited {
		return nil
	}
	visited[component] = struct{}{}

	// Resolve ref
	const prefix = "#/components/requestBodies/"
	if ref := component.Ref; len(ref) > 0 {
		components, id, err := resolver.resolveComponent(swagger, ref, prefix)
		if err != nil {
			return err
		}
		definitions := components.RequestBodies
		if definitions == nil {
			return failedToResolveRefFragmentPart(ref, "requestBodies")
		}
		resolved := definitions[id]
		if resolved == nil {
			return failedToResolveRefFragmentPart(ref, id)
		}
		err = resolver.resolveRequestBodyRef(swagger, resolved)
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
				err := resolver.resolveSchemaRef(swagger, schema)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (resolver *SwaggerLoader) resolveResponseRef(swagger *Swagger, component *ResponseRef) error {
	// Prevent infinite recursion
	visited := resolver.visited
	if _, isVisited := visited[component]; isVisited {
		return nil
	}
	visited[component] = struct{}{}

	// Resolve ref
	const prefix = "#/components/responses/"
	if ref := component.Ref; len(ref) > 0 {
		components, id, err := resolver.resolveComponent(swagger, ref, prefix)
		if err != nil {
			return err
		}
		definitions := components.Responses
		if definitions == nil {
			return failedToResolveRefFragmentPart(ref, "responses")
		}
		resolved := definitions[id]
		if resolved == nil {
			return failedToResolveRefFragmentPart(ref, id)
		}
		err = resolver.resolveResponseRef(swagger, resolved)
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
				err := resolver.resolveSchemaRef(swagger, schema)
				if err != nil {
					return err
				}
				contentType.Schema = schema
			}
		}
	}
	return nil
}

func (resolver *SwaggerLoader) resolveSchemaRef(swagger *Swagger, component *SchemaRef) error {
	// Prevent infinite recursion
	visited := resolver.visited
	if _, isVisited := visited[component]; isVisited {
		return nil
	}
	visited[component] = struct{}{}

	// Resolve ref
	const prefix = "#/components/schemas/"
	if ref := component.Ref; len(ref) > 0 {
		components, id, err := resolver.resolveComponent(swagger, ref, prefix)
		if err != nil {
			return err
		}
		definitions := components.Schemas
		if definitions == nil {
			return failedToResolveRefFragmentPart(ref, "schemas")
		}
		resolved := definitions[id]
		if resolved == nil {
			return failedToResolveRefFragmentPart(ref, id)
		}
		err = resolver.resolveSchemaRef(swagger, resolved)
		if err != nil {
			return err
		}
		component.Value = resolved.Value
	}
	value := component.Value

	// ResolveRefs referred schemas
	if v := value.Items; v != nil {
		err := resolver.resolveSchemaRef(swagger, v)
		if err != nil {
			return err
		}
	}
	if m := value.Properties; m != nil {
		for _, v := range m {
			err := resolver.resolveSchemaRef(swagger, v)
			if err != nil {
				return err
			}
		}
	}
	if v := value.AdditionalProperties; v != nil {
		err := resolver.resolveSchemaRef(swagger, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (resolver *SwaggerLoader) resolveSecuritySchemeRef(swagger *Swagger, component *SecuritySchemeRef) error {
	// Prevent infinite recursion
	visited := resolver.visited
	if _, isVisited := visited[component]; isVisited {
		return nil
	}
	visited[component] = struct{}{}

	// Resolve ref
	const prefix = "#/components/securitySchemes/"
	if ref := component.Ref; len(ref) > 0 {
		components, id, err := resolver.resolveComponent(swagger, ref, prefix)
		if err != nil {
			return err
		}
		definitions := components.SecuritySchemes
		if definitions == nil {
			return failedToResolveRefFragmentPart(ref, "securitySchemes")
		}
		resolved := definitions[id]
		if resolved == nil {
			return failedToResolveRefFragmentPart(ref, id)
		}
		err = resolver.resolveSecuritySchemeRef(swagger, resolved)
		if err != nil {
			return err
		}
		component.Value = resolved.Value
	}
	return nil
}

func (resolver *SwaggerLoader) resolveExampleRef(swagger *Swagger, component *ExampleRef) error {
	const prefix = "#/components/examples"
	if ref := component.Ref; len(ref) > 0 {
		components, id, err := resolver.resolveComponent(swagger, ref, prefix)
		if err != nil {
			return err
		}
		definitions := components.Examples
		if definitions == nil {
			return failedToResolveRefFragmentPart(ref, "examples")
		}
		resolved := definitions[id]
		if resolved == nil {
			return failedToResolveRefFragmentPart(ref, id)
		}
		err = resolver.resolveExampleRef(swagger, resolved)
		if err != nil {
			return err
		}
		component.Value = resolved.Value
	}
	return nil
}
