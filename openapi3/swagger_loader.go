package openapi3

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/ghodss/yaml"
)

func foundUnresolvedRef(ref string) error {
	return fmt.Errorf("Found unresolved ref: '%s'", ref)
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
	data, err := readUrl(location)
	if err != nil {
		return nil, err
	}
	return swaggerLoader.LoadSwaggerFromData(data)
}

func readUrl(location *url.URL) ([]byte, error) {
	if location.Scheme != "" && location.Host != "" {
		resp, err := http.Get(location.String())
		if err != nil {
			return nil, err
		}
		data, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	if location.Scheme != "" || location.Host != "" || location.RawQuery != "" {
		return nil, fmt.Errorf("Unsupported URI: '%s'", location.String())
	}
	data, err := ioutil.ReadFile(location.Path)
	if err != nil {
		return nil, err
	}
	return data, nil
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
	return swaggerLoader.LoadSwaggerFromDataWithPath(data, &url.URL{
		Path: path,
	})
}

func (swaggerLoader *SwaggerLoader) LoadSwaggerFromData(data []byte) (*Swagger, error) {
	swagger := &Swagger{}
	if err := json.Unmarshal(data, swagger); err != nil {
		return nil, err
	}
	return swagger, swaggerLoader.ResolveRefsIn(swagger, nil)
}

func (swaggerLoader *SwaggerLoader) LoadSwaggerFromDataWithPath(data []byte, path *url.URL) (*Swagger, error) {
	swagger := &Swagger{}
	if err := json.Unmarshal(data, swagger); err != nil {
		return nil, err
	}
	return swagger, swaggerLoader.ResolveRefsIn(swagger, path)
}

func (swaggerLoader *SwaggerLoader) LoadSwaggerFromYAMLData(data []byte) (*Swagger, error) {
	swagger := &Swagger{}
	if err := yaml.Unmarshal(data, swagger); err != nil {
		return nil, err
	}
	return swagger, swaggerLoader.ResolveRefsIn(swagger, nil)
}

// TODO: Find better way to handle it
func (swaggerLoader *SwaggerLoader) ResolveRefsIn(swagger *Swagger, path *url.URL) (err error) {
	swaggerLoader.visited = make(map[interface{}]struct{})

	// Visit all components
	components := swagger.Components
	for _, component := range components.Headers {
		if err = swaggerLoader.resolveHeaderRef(swagger, component, path); err != nil {
			return
		}
	}
	for _, component := range components.Parameters {
		if err = swaggerLoader.resolveParameterRef(swagger, component, path); err != nil {
			return
		}
	}
	for _, component := range components.RequestBodies {
		if err = swaggerLoader.resolveRequestBodyRef(swagger, component, path); err != nil {
			return
		}
	}
	for _, component := range components.Responses {
		if err = swaggerLoader.resolveResponseRef(swagger, component, path); err != nil {
			return
		}
	}
	for _, component := range components.Schemas {
		if err = swaggerLoader.resolveSchemaRef(swagger, component, path); err != nil {
			return
		}
	}
	for _, component := range components.SecuritySchemes {
		if err = swaggerLoader.resolveSecuritySchemeRef(swagger, component); err != nil {
			return
		}
	}
	for _, component := range components.Examples {
		if err = swaggerLoader.resolveExampleRef(swagger, component, path); err != nil {
			return
		}
	}

	// Visit all operations
	for _, pathItem := range swagger.Paths {
		if pathItem == nil {
			continue
		}
		for _, parameter := range pathItem.Parameters {
			// TODO
			if err = swaggerLoader.resolveParameterRef(swagger, parameter, nil); err != nil {
				return
			}
		}
		for _, operation := range pathItem.Operations() {
			for _, parameter := range operation.Parameters {
				// TODO
				if err = swaggerLoader.resolveParameterRef(swagger, parameter, nil); err != nil {
					return
				}
			}
			if requestBody := operation.RequestBody; requestBody != nil {
				// TODO
				if err = swaggerLoader.resolveRequestBodyRef(swagger, requestBody, path); err != nil {
					return
				}
			}
			for _, response := range operation.Responses {
				// TODO
				if err = swaggerLoader.resolveResponseRef(swagger, response, nil); err != nil {
					return
				}
			}
		}
	}

	return
}

func join(basePath *url.URL, rest *url.URL) *url.URL {
	if basePath == nil {
		return rest
	}
	newPath := &*basePath
	newPath.Path = path.Join(path.Dir(newPath.Path), rest.Path)
	return newPath
}

func (swaggerLoader *SwaggerLoader) resolveComponent(swagger *Swagger, ref string, prefix string, path *url.URL) (
	components *Components,
	id string,
	err error,
) {
	if !strings.HasPrefix(ref, "#") {
		if !swaggerLoader.IsExternalRefsAllowed {
			return nil, "", fmt.Errorf("Encountered non-allowed external reference: '%s'", ref)
		}
		parsedURL, err := url.Parse(ref)
		if err != nil {
			return nil, "", fmt.Errorf("Can't parse reference: '%s': %v", ref, parsedURL)
		}
		fragment := parsedURL.Fragment
		parsedURL.Fragment = ""
		fullURL := join(path, parsedURL)
		if swagger, err = swaggerLoader.LoadSwaggerFromURI(fullURL); err != nil {
			return nil, "", fmt.Errorf("Error while resolving reference '%s': %v", ref, err)
		}
		ref = fmt.Sprintf("#%s", fragment)
	}
	if !strings.HasPrefix(ref, prefix) {
		err := fmt.Errorf("expected prefix '%s' in URI '%s'", prefix, ref)
		return nil, "", err
	}
	id = ref[len(prefix):]
	if strings.IndexByte(id, '/') >= 0 {
		return nil, "", failedToResolveRefFragmentPart(ref, id)
	}
	return &swagger.Components, id, nil
}

func (swaggerLoader *SwaggerLoader) resolveHeaderRef(swagger *Swagger, component *HeaderRef, path *url.URL) error {
	// Prevent infinite recursion
	visited := swaggerLoader.visited
	if _, isVisited := visited[component]; isVisited {
		return nil
	}
	visited[component] = struct{}{}

	// Resolve ref
	const prefix = "#/components/headers/"
	if ref := component.Ref; len(ref) > 0 {
		components, id, err := swaggerLoader.resolveComponent(swagger, ref, prefix, path)
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
		if err := swaggerLoader.resolveHeaderRef(swagger, resolved, path); err != nil {
			return err
		}
		component.Value = resolved.Value
	}
	value := component.Value
	if value == nil {
		return nil
	}
	if schema := value.Schema; schema != nil {
		// TODO
		if err := swaggerLoader.resolveSchemaRef(swagger, schema, nil); err != nil {
			return err
		}
	}
	return nil
}

func (swaggerLoader *SwaggerLoader) resolveParameterRef(swagger *Swagger, component *ParameterRef, path *url.URL) error {
	// Prevent infinite recursion
	visited := swaggerLoader.visited
	if _, isVisited := visited[component]; isVisited {
		return nil
	}
	visited[component] = struct{}{}

	// Resolve ref
	const prefix = "#/components/parameters/"
	if ref := component.Ref; len(ref) > 0 {
		// TODO
		components, id, err := swaggerLoader.resolveComponent(swagger, ref, prefix, path)
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
		if err := swaggerLoader.resolveParameterRef(swagger, resolved, path); err != nil {
			return err
		}
		component.Value = resolved.Value
	}
	value := component.Value
	if value == nil {
		return nil
	}
	if schema := value.Schema; schema != nil {
		// TODO
		if err := swaggerLoader.resolveSchemaRef(swagger, schema, nil); err != nil {
			return err
		}
	}
	return nil
}

func (swaggerLoader *SwaggerLoader) resolveRequestBodyRef(swagger *Swagger, component *RequestBodyRef, path *url.URL) error {
	// Prevent infinite recursion
	visited := swaggerLoader.visited
	if _, isVisited := visited[component]; isVisited {
		return nil
	}
	visited[component] = struct{}{}

	// Resolve ref
	const prefix = "#/components/requestBodies/"
	if ref := component.Ref; len(ref) > 0 {
		// TODO
		components, id, err := swaggerLoader.resolveComponent(swagger, ref, prefix, path)
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
		if err = swaggerLoader.resolveRequestBodyRef(swagger, resolved, path); err != nil {
			return err
		}
		component.Value = resolved.Value
	}
	value := component.Value
	if value == nil {
		return nil
	}
	for _, contentType := range value.Content {
		for name, example := range contentType.Examples {
			// TODO
			if err := swaggerLoader.resolveExampleRef(swagger, example, nil); err != nil {
				return err
			}
			contentType.Examples[name] = example
		}
		if schema := contentType.Schema; schema != nil {
			// TODO
			if err := swaggerLoader.resolveSchemaRef(swagger, schema, nil); err != nil {
				return err
			}
		}
	}
	return nil
}

func (swaggerLoader *SwaggerLoader) resolveResponseRef(swagger *Swagger, component *ResponseRef, path *url.URL) error {
	// Prevent infinite recursion
	visited := swaggerLoader.visited
	if _, isVisited := visited[component]; isVisited {
		return nil
	}
	visited[component] = struct{}{}

	// Resolve ref
	const prefix = "#/components/responses/"
	if ref := component.Ref; len(ref) > 0 {
		// TODO
		components, id, err := swaggerLoader.resolveComponent(swagger, ref, prefix, path)
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
		if err := swaggerLoader.resolveResponseRef(swagger, resolved, path); err != nil {
			return err
		}
		component.Value = resolved.Value
	}
	value := component.Value
	if value == nil {
		return nil
	}
	for _, header := range value.Headers {
		if err := swaggerLoader.resolveHeaderRef(swagger, header); err != nil {
			return err
		}
	}
	for _, contentType := range value.Content {
		if contentType == nil {
			continue
		}
		for name, example := range contentType.Examples {
			// TODO
			if err := swaggerLoader.resolveExampleRef(swagger, example, nil); err != nil {
				return err
			}
			contentType.Examples[name] = example
		}
		if schema := contentType.Schema; schema != nil {
			// TODO
			if err := swaggerLoader.resolveSchemaRef(swagger, schema, nil); err != nil {
				return err
			}
			contentType.Schema = schema
		}
	}
	return nil
}

func (swaggerLoader *SwaggerLoader) resolveSchemaRef(swagger *Swagger, component *SchemaRef, path *url.URL) error {
	// Prevent infinite recursion
	visited := swaggerLoader.visited
	if _, isVisited := visited[component]; isVisited {
		return nil
	}
	visited[component] = struct{}{}

	// Resolve ref
	const prefix = "#/components/schemas/"
	if ref := component.Ref; len(ref) > 0 {
		components, id, err := swaggerLoader.resolveComponent(swagger, ref, prefix, path)
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
		if err := swaggerLoader.resolveSchemaRef(swagger, resolved, path); err != nil {
			return err
		}
		component.Value = resolved.Value
	}
	value := component.Value
	if value == nil {
		return nil
	}

	// ResolveRefs referred schemas
	if v := value.Items; v != nil {
		if err := swaggerLoader.resolveSchemaRef(swagger, v, path); err != nil {
			return err
		}
	}
	for _, v := range value.Properties {
		if err := swaggerLoader.resolveSchemaRef(swagger, v, path); err != nil {
			return err
		}
	}
	if v := value.AdditionalProperties; v != nil {
		if err := swaggerLoader.resolveSchemaRef(swagger, v, path); err != nil {
			return err
		}
	}
	if v := value.Not; v != nil {
		if err := swaggerLoader.resolveSchemaRef(swagger, v, path); err != nil {
			return err
		}
	}
	for _, v := range value.AllOf {
		if err := swaggerLoader.resolveSchemaRef(swagger, v, path); err != nil {
			return err
		}
	}
	for _, v := range value.AnyOf {
		if err := swaggerLoader.resolveSchemaRef(swagger, v, path); err != nil {
			return err
		}
	}
	for _, v := range value.OneOf {
		if err := swaggerLoader.resolveSchemaRef(swagger, v, path); err != nil {
			return err
		}
	}

	return nil
}

func (swaggerLoader *SwaggerLoader) resolveSecuritySchemeRef(swagger *Swagger, component *SecuritySchemeRef) error {
	// Prevent infinite recursion
	visited := swaggerLoader.visited
	if _, isVisited := visited[component]; isVisited {
		return nil
	}
	visited[component] = struct{}{}

	// Resolve ref
	const prefix = "#/components/securitySchemes/"
	if ref := component.Ref; len(ref) > 0 {
		components, id, err := swaggerLoader.resolveComponent(swagger, ref, prefix, nil)
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
		if err := swaggerLoader.resolveSecuritySchemeRef(swagger, resolved); err != nil {
			return err
		}
		component.Value = resolved.Value
	}
	return nil
}

func (swaggerLoader *SwaggerLoader) resolveExampleRef(swagger *Swagger, component *ExampleRef, path *url.URL) error {
	// Prevent infinite recursion
	visited := swaggerLoader.visited
	if _, isVisited := visited[component]; isVisited {
		return nil
	}
	visited[component] = struct{}{}

	const prefix = "#/components/examples/"
	if ref := component.Ref; len(ref) > 0 {
		components, id, err := swaggerLoader.resolveComponent(swagger, ref, prefix, path)
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
		if err := swaggerLoader.resolveExampleRef(swagger, resolved, path); err != nil {
			return err
		}
		component.Value = resolved.Value
	}
	return nil
}
