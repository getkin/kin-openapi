package openapi3

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/ghodss/yaml"
)

func foundUnresolvedRef(ref string) error {
	return fmt.Errorf("found unresolved ref: %q", ref)
}

func failedToResolveRefFragmentPart(value, what string) error {
	return fmt.Errorf("failed to resolve %q in fragment in URI: %q", what, value)
}

// SwaggerLoader helps deserialize a Swagger object
type SwaggerLoader struct {
	// IsExternalRefsAllowed enables visiting other files
	IsExternalRefsAllowed bool

	// ReadFromURIFunc allows overriding the any file/URL reading func
	ReadFromURIFunc func(loader *SwaggerLoader, url *url.URL) ([]byte, error)

	Context context.Context

	visitedPathItemRefs map[string]struct{}

	visitedDocuments map[string]*Swagger

	// visitedCallback       map[*Callback]struct{}
	visitedExample        map[*Example]struct{}
	visitedHeader         map[*Header]struct{}
	visitedLink           map[*Link]struct{}
	visitedParameter      map[*Parameter]struct{}
	visitedRequestBody    map[*RequestBody]struct{}
	visitedResponse       map[*Response]struct{}
	visitedSchema         map[*Schema]struct{}
	visitedSecurityScheme map[*SecurityScheme]struct{}
}

// NewSwaggerLoader returns an empty SwaggerLoader
func NewSwaggerLoader() *SwaggerLoader {
	return &SwaggerLoader{}
}

func (swaggerLoader *SwaggerLoader) resetVisitedPathItemRefs() {
	swaggerLoader.visitedPathItemRefs = make(map[string]struct{})
}

// LoadSwaggerFromURI loads a spec from a remote URL
func (swaggerLoader *SwaggerLoader) LoadSwaggerFromURI(location *url.URL) (*Swagger, error) {
	swaggerLoader.resetVisitedPathItemRefs()
	return swaggerLoader.loadSwaggerFromURIInternal(location)
}

// LoadSwaggerFromFile loads a spec from a local file path
func (swaggerLoader *SwaggerLoader) LoadSwaggerFromFile(location string) (*Swagger, error) {
	return swaggerLoader.LoadSwaggerFromURI(&url.URL{Path: filepath.ToSlash(location)})
}

func (swaggerLoader *SwaggerLoader) loadSwaggerFromURIInternal(location *url.URL) (*Swagger, error) {
	data, err := swaggerLoader.readURL(location)
	if err != nil {
		return nil, err
	}
	return swaggerLoader.loadSwaggerFromDataWithPathInternal(data, location)
}

func (swaggerLoader *SwaggerLoader) allowsExternalRefs(ref string) (err error) {
	if !swaggerLoader.IsExternalRefsAllowed {
		err = fmt.Errorf("encountered disallowed external reference: %q", ref)
	}
	return
}

// loadSingleElementFromURI reads the data from ref and unmarshals to the passed element.
func (swaggerLoader *SwaggerLoader) loadSingleElementFromURI(ref string, rootPath *url.URL, element interface{}) (*url.URL, error) {
	if err := swaggerLoader.allowsExternalRefs(ref); err != nil {
		return nil, err
	}

	parsedURL, err := url.Parse(ref)
	if err != nil {
		return nil, err
	}
	if fragment := parsedURL.Fragment; fragment != "" {
		return nil, fmt.Errorf("unexpected ref fragment %q", fragment)
	}

	resolvedPath, err := resolvePath(rootPath, parsedURL)
	if err != nil {
		return nil, fmt.Errorf("could not resolve path: %v", err)
	}

	data, err := swaggerLoader.readURL(resolvedPath)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, element); err != nil {
		return nil, err
	}

	return resolvedPath, nil
}

func (swaggerLoader *SwaggerLoader) readURL(location *url.URL) ([]byte, error) {
	if f := swaggerLoader.ReadFromURIFunc; f != nil {
		return f(swaggerLoader, location)
	}

	if location.Scheme != "" && location.Host != "" {
		resp, err := http.Get(location.String())
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode > 399 {
			return nil, fmt.Errorf("request returned status code %d", resp.StatusCode)
		}
		return ioutil.ReadAll(resp.Body)
	}
	if location.Scheme != "" || location.Host != "" || location.RawQuery != "" {
		return nil, fmt.Errorf("unsupported URI: %q", location.String())
	}
	return ioutil.ReadFile(location.Path)
}

// LoadSwaggerFromData loads a spec from a byte array
func (swaggerLoader *SwaggerLoader) LoadSwaggerFromData(data []byte) (*Swagger, error) {
	swaggerLoader.resetVisitedPathItemRefs()
	doc := &Swagger{}
	if err := yaml.Unmarshal(data, doc); err != nil {
		return nil, err
	}
	if err := swaggerLoader.ResolveRefsIn(doc, nil); err != nil {
		return nil, err
	}
	return doc, nil
}

// LoadSwaggerFromDataWithPath takes the OpenApi spec data in bytes and a path where the resolver can find referred
// elements and returns a *Swagger with all resolved data or an error if unable to load data or resolve refs.
func (swaggerLoader *SwaggerLoader) LoadSwaggerFromDataWithPath(data []byte, location *url.URL) (*Swagger, error) {
	swaggerLoader.resetVisitedPathItemRefs()
	return swaggerLoader.loadSwaggerFromDataWithPathInternal(data, location)
}

func (swaggerLoader *SwaggerLoader) loadSwaggerFromDataWithPathInternal(data []byte, location *url.URL) (*Swagger, error) {
	if swaggerLoader.visitedDocuments == nil {
		swaggerLoader.visitedDocuments = make(map[string]*Swagger)
	}
	uri := location.String()
	if doc, ok := swaggerLoader.visitedDocuments[uri]; ok {
		return doc, nil
	}

	swagger := &Swagger{}
	swaggerLoader.visitedDocuments[uri] = swagger

	if err := yaml.Unmarshal(data, swagger); err != nil {
		return nil, err
	}
	if err := swaggerLoader.ResolveRefsIn(swagger, location); err != nil {
		return nil, err
	}

	return swagger, nil
}

// ResolveRefsIn expands references if for instance spec was just unmarshalled
func (swaggerLoader *SwaggerLoader) ResolveRefsIn(swagger *Swagger, location *url.URL) (err error) {
	if swaggerLoader.visitedPathItemRefs == nil {
		swaggerLoader.resetVisitedPathItemRefs()
	}

	// Visit all components
	components := swagger.Components
	for _, component := range components.Headers {
		if err = swaggerLoader.resolveHeaderRef(swagger, component, location); err != nil {
			return
		}
	}
	for _, component := range components.Parameters {
		if err = swaggerLoader.resolveParameterRef(swagger, component, location); err != nil {
			return
		}
	}
	for _, component := range components.RequestBodies {
		if err = swaggerLoader.resolveRequestBodyRef(swagger, component, location); err != nil {
			return
		}
	}
	for _, component := range components.Responses {
		if err = swaggerLoader.resolveResponseRef(swagger, component, location); err != nil {
			return
		}
	}
	for _, component := range components.Schemas {
		if err = swaggerLoader.resolveSchemaRef(swagger, component, location); err != nil {
			return
		}
	}
	for _, component := range components.SecuritySchemes {
		if err = swaggerLoader.resolveSecuritySchemeRef(swagger, component, location); err != nil {
			return
		}
	}
	for _, component := range components.Examples {
		if err = swaggerLoader.resolveExampleRef(swagger, component, location); err != nil {
			return
		}
	}
	for _, component := range components.Callbacks {
		if err = swaggerLoader.resolveCallbackRef(swagger, component, location); err != nil {
			return
		}
	}

	// Visit all operations
	for entrypoint, pathItem := range swagger.Paths {
		if pathItem == nil {
			continue
		}
		if err = swaggerLoader.resolvePathItemRef(swagger, entrypoint, pathItem, location); err != nil {
			return
		}
	}

	return
}

func join(basePath *url.URL, relativePath *url.URL) (*url.URL, error) {
	if basePath == nil {
		return relativePath, nil
	}
	newPath, err := url.Parse(basePath.String())
	if err != nil {
		return nil, fmt.Errorf("cannot copy path: %q", basePath.String())
	}
	newPath.Path = path.Join(path.Dir(newPath.Path), relativePath.Path)
	return newPath, nil
}

func resolvePath(basePath *url.URL, componentPath *url.URL) (*url.URL, error) {
	if componentPath.Scheme == "" && componentPath.Host == "" {
		// support absolute paths
		if componentPath.Path[0] == '/' {
			return componentPath, nil
		}
		return join(basePath, componentPath)
	}
	return componentPath, nil
}

func isSingleRefElement(ref string) bool {
	return !strings.Contains(ref, "#")
}

func (swaggerLoader *SwaggerLoader) resolveComponent(
	swagger *Swagger,
	ref string,
	path *url.URL,
	resolved interface{},
) (
	componentPath *url.URL,
	err error,
) {
	if swagger, ref, componentPath, err = swaggerLoader.resolveRefSwagger(swagger, ref, path); err != nil {
		return nil, err
	}

	parsedURL, err := url.Parse(ref)
	if err != nil {
		return nil, fmt.Errorf("cannot parse reference: %q: %v", ref, parsedURL)
	}
	fragment := parsedURL.Fragment
	if !strings.HasPrefix(fragment, "/") {
		return nil, fmt.Errorf("expected fragment prefix '#/' in URI %q", ref)
	}

	drill := func(cursor interface{}) (interface{}, error) {
		for _, pathPart := range strings.Split(fragment[1:], "/") {
			pathPart = unescapeRefString(pathPart)

			if cursor, err = drillIntoSwaggerField(cursor, pathPart); err != nil {
				e := failedToResolveRefFragmentPart(ref, pathPart)
				return nil, fmt.Errorf("%s: %s", e.Error(), err.Error())
			}
			if cursor == nil {
				return nil, failedToResolveRefFragmentPart(ref, pathPart)
			}
		}
		return cursor, nil
	}
	var cursor interface{}
	if cursor, err = drill(swagger); err != nil {
		var err2 error
		data, err2 := swaggerLoader.readURL(path)
		if err2 != nil {
			return nil, err
		}
		if err2 = yaml.Unmarshal(data, &cursor); err2 != nil {
			return nil, err
		}
		if cursor, err2 = drill(cursor); err2 != nil || cursor == nil {
			return nil, err
		}
		err = nil
	}

	switch {
	case reflect.TypeOf(cursor) == reflect.TypeOf(resolved):
		reflect.ValueOf(resolved).Elem().Set(reflect.ValueOf(cursor).Elem())
		return componentPath, nil

	case reflect.TypeOf(cursor) == reflect.TypeOf(map[string]interface{}{}):
		codec := func(got, expect interface{}) error {
			enc, err := json.Marshal(got)
			if err != nil {
				return err
			}
			if err = json.Unmarshal(enc, expect); err != nil {
				return err
			}
			return nil
		}
		if err := codec(cursor, resolved); err != nil {
			return nil, fmt.Errorf("bad data in %q", ref)
		}
		return componentPath, nil

	default:
		return nil, fmt.Errorf("bad data in %q", ref)
	}
}

func drillIntoSwaggerField(cursor interface{}, fieldName string) (interface{}, error) {
	switch val := reflect.Indirect(reflect.ValueOf(cursor)); val.Kind() {
	case reflect.Map:
		elementValue := val.MapIndex(reflect.ValueOf(fieldName))
		if !elementValue.IsValid() {
			return nil, fmt.Errorf("map key %q not found", fieldName)
		}
		return elementValue.Interface(), nil

	case reflect.Slice:
		i, err := strconv.ParseUint(fieldName, 10, 32)
		if err != nil {
			return nil, err
		}
		index := int(i)
		if 0 > index || index >= val.Len() {
			return nil, errors.New("slice index out of bounds")
		}
		return val.Index(index).Interface(), nil

	case reflect.Struct:
		hasFields := false
		for i := 0; i < val.NumField(); i++ {
			hasFields = true
			field := val.Type().Field(i)
			tagValue := field.Tag.Get("yaml")
			yamlKey := strings.Split(tagValue, ",")[0]
			if yamlKey == fieldName {
				return val.Field(i).Interface(), nil
			}
		}
		// if cursor is a "ref wrapper" struct (e.g. RequestBodyRef),
		if _, ok := val.Type().FieldByName("Value"); ok {
			// try digging into its Value field
			return drillIntoSwaggerField(val.FieldByName("Value").Interface(), fieldName)
		}
		if hasFields {
			if ff := val.Type().Field(0); ff.PkgPath == "" && ff.Name == "ExtensionProps" {
				extensions := val.Field(0).Interface().(ExtensionProps).Extensions
				if enc, ok := extensions[fieldName]; ok {
					var dec interface{}
					if err := json.Unmarshal(enc.(json.RawMessage), &dec); err != nil {
						return nil, err
					}
					return dec, nil
				}
			}
		}
		return nil, fmt.Errorf("struct field %q not found", fieldName)

	default:
		return nil, errors.New("not a map, slice nor struct")
	}
}

func (swaggerLoader *SwaggerLoader) resolveRefSwagger(swagger *Swagger, ref string, path *url.URL) (*Swagger, string, *url.URL, error) {
	if ref != "" && ref[0] == '#' {
		return swagger, ref, path, nil
	}

	if err := swaggerLoader.allowsExternalRefs(ref); err != nil {
		return nil, "", nil, err
	}

	parsedURL, err := url.Parse(ref)
	if err != nil {
		return nil, "", nil, fmt.Errorf("cannot parse reference: %q: %v", ref, parsedURL)
	}
	fragment := parsedURL.Fragment
	parsedURL.Fragment = ""

	var resolvedPath *url.URL
	if resolvedPath, err = resolvePath(path, parsedURL); err != nil {
		return nil, "", nil, fmt.Errorf("error resolving path: %v", err)
	}

	if swagger, err = swaggerLoader.loadSwaggerFromURIInternal(resolvedPath); err != nil {
		return nil, "", nil, fmt.Errorf("error resolving reference %q: %v", ref, err)
	}

	return swagger, "#" + fragment, resolvedPath, nil
}

func (swaggerLoader *SwaggerLoader) resolveHeaderRef(swagger *Swagger, component *HeaderRef, documentPath *url.URL) (err error) {
	if component != nil && component.Value != nil {
		if swaggerLoader.visitedHeader == nil {
			swaggerLoader.visitedHeader = make(map[*Header]struct{})
		}
		if _, ok := swaggerLoader.visitedHeader[component.Value]; ok {
			return nil
		}
		swaggerLoader.visitedHeader[component.Value] = struct{}{}
	}

	if component == nil {
		return errors.New("invalid header: value MUST be an object")
	}
	if ref := component.Ref; ref != "" {
		if isSingleRefElement(ref) {
			var header Header
			if documentPath, err = swaggerLoader.loadSingleElementFromURI(ref, documentPath, &header); err != nil {
				return err
			}
			component.Value = &header
		} else {
			var resolved HeaderRef
			componentPath, err := swaggerLoader.resolveComponent(swagger, ref, documentPath, &resolved)
			if err != nil {
				return err
			}
			if err := swaggerLoader.resolveHeaderRef(swagger, &resolved, componentPath); err != nil {
				return err
			}
			component.Value = resolved.Value
		}
	}
	value := component.Value
	if value == nil {
		return nil
	}

	if schema := value.Schema; schema != nil {
		if err := swaggerLoader.resolveSchemaRef(swagger, schema, documentPath); err != nil {
			return err
		}
	}
	return nil
}

func (swaggerLoader *SwaggerLoader) resolveParameterRef(swagger *Swagger, component *ParameterRef, documentPath *url.URL) (err error) {
	if component != nil && component.Value != nil {
		if swaggerLoader.visitedParameter == nil {
			swaggerLoader.visitedParameter = make(map[*Parameter]struct{})
		}
		if _, ok := swaggerLoader.visitedParameter[component.Value]; ok {
			return nil
		}
		swaggerLoader.visitedParameter[component.Value] = struct{}{}
	}

	if component == nil {
		return errors.New("invalid parameter: value MUST be an object")
	}
	ref := component.Ref
	if ref != "" {
		if isSingleRefElement(ref) {
			var param Parameter
			if documentPath, err = swaggerLoader.loadSingleElementFromURI(ref, documentPath, &param); err != nil {
				return err
			}
			component.Value = &param
		} else {
			var resolved ParameterRef
			componentPath, err := swaggerLoader.resolveComponent(swagger, ref, documentPath, &resolved)
			if err != nil {
				return err
			}
			if err := swaggerLoader.resolveParameterRef(swagger, &resolved, componentPath); err != nil {
				return err
			}
			component.Value = resolved.Value
		}
	}
	value := component.Value
	if value == nil {
		return nil
	}

	if value.Content != nil && value.Schema != nil {
		return errors.New("cannot contain both schema and content in a parameter")
	}
	for _, contentType := range value.Content {
		if schema := contentType.Schema; schema != nil {
			if err := swaggerLoader.resolveSchemaRef(swagger, schema, documentPath); err != nil {
				return err
			}
		}
	}
	if schema := value.Schema; schema != nil {
		if err := swaggerLoader.resolveSchemaRef(swagger, schema, documentPath); err != nil {
			return err
		}
	}
	return nil
}

func (swaggerLoader *SwaggerLoader) resolveRequestBodyRef(swagger *Swagger, component *RequestBodyRef, documentPath *url.URL) (err error) {
	if component != nil && component.Value != nil {
		if swaggerLoader.visitedRequestBody == nil {
			swaggerLoader.visitedRequestBody = make(map[*RequestBody]struct{})
		}
		if _, ok := swaggerLoader.visitedRequestBody[component.Value]; ok {
			return nil
		}
		swaggerLoader.visitedRequestBody[component.Value] = struct{}{}
	}

	if component == nil {
		return errors.New("invalid requestBody: value MUST be an object")
	}
	if ref := component.Ref; ref != "" {
		if isSingleRefElement(ref) {
			var requestBody RequestBody
			if documentPath, err = swaggerLoader.loadSingleElementFromURI(ref, documentPath, &requestBody); err != nil {
				return err
			}
			component.Value = &requestBody
		} else {
			var resolved RequestBodyRef
			componentPath, err := swaggerLoader.resolveComponent(swagger, ref, documentPath, &resolved)
			if err != nil {
				return err
			}
			if err = swaggerLoader.resolveRequestBodyRef(swagger, &resolved, componentPath); err != nil {
				return err
			}
			component.Value = resolved.Value
		}
	}
	value := component.Value
	if value == nil {
		return nil
	}

	for _, contentType := range value.Content {
		for name, example := range contentType.Examples {
			if err := swaggerLoader.resolveExampleRef(swagger, example, documentPath); err != nil {
				return err
			}
			contentType.Examples[name] = example
		}
		if schema := contentType.Schema; schema != nil {
			if err := swaggerLoader.resolveSchemaRef(swagger, schema, documentPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func (swaggerLoader *SwaggerLoader) resolveResponseRef(swagger *Swagger, component *ResponseRef, documentPath *url.URL) (err error) {
	if component != nil && component.Value != nil {
		if swaggerLoader.visitedResponse == nil {
			swaggerLoader.visitedResponse = make(map[*Response]struct{})
		}
		if _, ok := swaggerLoader.visitedResponse[component.Value]; ok {
			return nil
		}
		swaggerLoader.visitedResponse[component.Value] = struct{}{}
	}

	if component == nil {
		return errors.New("invalid response: value MUST be an object")
	}
	ref := component.Ref
	if ref != "" {
		if isSingleRefElement(ref) {
			var resp Response
			if documentPath, err = swaggerLoader.loadSingleElementFromURI(ref, documentPath, &resp); err != nil {
				return err
			}
			component.Value = &resp
		} else {
			var resolved ResponseRef
			componentPath, err := swaggerLoader.resolveComponent(swagger, ref, documentPath, &resolved)
			if err != nil {
				return err
			}
			if err := swaggerLoader.resolveResponseRef(swagger, &resolved, componentPath); err != nil {
				return err
			}
			component.Value = resolved.Value
		}
	}
	value := component.Value
	if value == nil {
		return nil
	}

	for _, header := range value.Headers {
		if err := swaggerLoader.resolveHeaderRef(swagger, header, documentPath); err != nil {
			return err
		}
	}
	for _, contentType := range value.Content {
		if contentType == nil {
			continue
		}
		for name, example := range contentType.Examples {
			if err := swaggerLoader.resolveExampleRef(swagger, example, documentPath); err != nil {
				return err
			}
			contentType.Examples[name] = example
		}
		if schema := contentType.Schema; schema != nil {
			if err := swaggerLoader.resolveSchemaRef(swagger, schema, documentPath); err != nil {
				return err
			}
			contentType.Schema = schema
		}
	}
	for _, link := range value.Links {
		if err := swaggerLoader.resolveLinkRef(swagger, link, documentPath); err != nil {
			return err
		}
	}
	return nil
}

func (swaggerLoader *SwaggerLoader) resolveSchemaRef(swagger *Swagger, component *SchemaRef, documentPath *url.URL) (err error) {
	if component != nil && component.Value != nil {
		if swaggerLoader.visitedSchema == nil {
			swaggerLoader.visitedSchema = make(map[*Schema]struct{})
		}
		if _, ok := swaggerLoader.visitedSchema[component.Value]; ok {
			return nil
		}
		swaggerLoader.visitedSchema[component.Value] = struct{}{}
	}

	if component == nil {
		return errors.New("invalid schema: value MUST be an object")
	}
	ref := component.Ref
	if ref != "" {
		if isSingleRefElement(ref) {
			var schema Schema
			if documentPath, err = swaggerLoader.loadSingleElementFromURI(ref, documentPath, &schema); err != nil {
				return err
			}
			component.Value = &schema
		} else {
			var resolved SchemaRef
			componentPath, err := swaggerLoader.resolveComponent(swagger, ref, documentPath, &resolved)
			if err != nil {
				return err
			}
			if err := swaggerLoader.resolveSchemaRef(swagger, &resolved, componentPath); err != nil {
				return err
			}
			component.Value = resolved.Value
		}
	}
	value := component.Value
	if value == nil {
		return nil
	}

	// ResolveRefs referred schemas
	if v := value.Items; v != nil {
		if err := swaggerLoader.resolveSchemaRef(swagger, v, documentPath); err != nil {
			return err
		}
	}
	for _, v := range value.Properties {
		if err := swaggerLoader.resolveSchemaRef(swagger, v, documentPath); err != nil {
			return err
		}
	}
	if v := value.AdditionalProperties; v != nil {
		if err := swaggerLoader.resolveSchemaRef(swagger, v, documentPath); err != nil {
			return err
		}
	}
	if v := value.Not; v != nil {
		if err := swaggerLoader.resolveSchemaRef(swagger, v, documentPath); err != nil {
			return err
		}
	}
	for _, v := range value.AllOf {
		if err := swaggerLoader.resolveSchemaRef(swagger, v, documentPath); err != nil {
			return err
		}
	}
	for _, v := range value.AnyOf {
		if err := swaggerLoader.resolveSchemaRef(swagger, v, documentPath); err != nil {
			return err
		}
	}
	for _, v := range value.OneOf {
		if err := swaggerLoader.resolveSchemaRef(swagger, v, documentPath); err != nil {
			return err
		}
	}
	return nil
}

func (swaggerLoader *SwaggerLoader) resolveSecuritySchemeRef(swagger *Swagger, component *SecuritySchemeRef, documentPath *url.URL) (err error) {
	if component != nil && component.Value != nil {
		if swaggerLoader.visitedSecurityScheme == nil {
			swaggerLoader.visitedSecurityScheme = make(map[*SecurityScheme]struct{})
		}
		if _, ok := swaggerLoader.visitedSecurityScheme[component.Value]; ok {
			return nil
		}
		swaggerLoader.visitedSecurityScheme[component.Value] = struct{}{}
	}

	if component == nil {
		return errors.New("invalid securityScheme: value MUST be an object")
	}
	if ref := component.Ref; ref != "" {
		if isSingleRefElement(ref) {
			var scheme SecurityScheme
			if documentPath, err = swaggerLoader.loadSingleElementFromURI(ref, documentPath, &scheme); err != nil {
				return err
			}
			component.Value = &scheme
		} else {
			var resolved SecuritySchemeRef
			componentPath, err := swaggerLoader.resolveComponent(swagger, ref, documentPath, &resolved)
			if err != nil {
				return err
			}
			if err := swaggerLoader.resolveSecuritySchemeRef(swagger, &resolved, componentPath); err != nil {
				return err
			}
			component.Value = resolved.Value
		}
	}
	return nil
}

func (swaggerLoader *SwaggerLoader) resolveExampleRef(swagger *Swagger, component *ExampleRef, documentPath *url.URL) (err error) {
	if component != nil && component.Value != nil {
		if swaggerLoader.visitedExample == nil {
			swaggerLoader.visitedExample = make(map[*Example]struct{})
		}
		if _, ok := swaggerLoader.visitedExample[component.Value]; ok {
			return nil
		}
		swaggerLoader.visitedExample[component.Value] = struct{}{}
	}

	if component == nil {
		return errors.New("invalid example: value MUST be an object")
	}
	if ref := component.Ref; ref != "" {
		if isSingleRefElement(ref) {
			var example Example
			if documentPath, err = swaggerLoader.loadSingleElementFromURI(ref, documentPath, &example); err != nil {
				return err
			}
			component.Value = &example
		} else {
			var resolved ExampleRef
			componentPath, err := swaggerLoader.resolveComponent(swagger, ref, documentPath, &resolved)
			if err != nil {
				return err
			}
			if err := swaggerLoader.resolveExampleRef(swagger, &resolved, componentPath); err != nil {
				return err
			}
			component.Value = resolved.Value
		}
	}
	return nil
}

func (swaggerLoader *SwaggerLoader) resolveCallbackRef(swagger *Swagger, component *CallbackRef, documentPath *url.URL) (err error) {
	// if component != nil && component.Value != nil {
	// 	if swaggerLoader.visitedCallback == nil {
	// 		swaggerLoader.visitedCallback = make(map[*Callback]struct{})
	// 	}
	// 	if _, ok := swaggerLoader.visitedCallback[component.Value]; ok {
	// 		return nil
	// 	}
	// 	swaggerLoader.visitedCallback[component.Value] = struct{}{}
	// }

	if component == nil {
		return errors.New("invalid callback: value MUST be an object")
	}
	if ref := component.Ref; ref != "" {
		if isSingleRefElement(ref) {
			var resolved Callback
			if documentPath, err = swaggerLoader.loadSingleElementFromURI(ref, documentPath, &resolved); err != nil {
				return err
			}
			component.Value = &resolved
		} else {
			var resolved CallbackRef
			componentPath, err := swaggerLoader.resolveComponent(swagger, ref, documentPath, &resolved)
			if err != nil {
				return err
			}
			if err := swaggerLoader.resolveCallbackRef(swagger, &resolved, componentPath); err != nil {
				return err
			}
			component.Value = resolved.Value
		}
	}
	value := component.Value
	if value == nil {
		return nil
	}

	for entrypoint, pathItem := range *value {
		entrypoint, pathItem := entrypoint, pathItem
		err = func() (err error) {
			key := "-"
			if documentPath != nil {
				key = documentPath.EscapedPath()
			}
			key += entrypoint
			if _, ok := swaggerLoader.visitedPathItemRefs[key]; ok {
				return nil
			}
			swaggerLoader.visitedPathItemRefs[key] = struct{}{}

			if pathItem == nil {
				return errors.New("invalid path item: value MUST be an object")
			}
			ref := pathItem.Ref
			if ref != "" {
				if isSingleRefElement(ref) {
					var p PathItem
					if documentPath, err = swaggerLoader.loadSingleElementFromURI(ref, documentPath, &p); err != nil {
						return err
					}
					*pathItem = p
				} else {
					if swagger, ref, documentPath, err = swaggerLoader.resolveRefSwagger(swagger, ref, documentPath); err != nil {
						return
					}

					rest := strings.TrimPrefix(ref, "#/components/callbacks/")
					if rest == ref {
						return fmt.Errorf(`expected prefix "#/components/callbacks/" in URI %q`, ref)
					}
					id := unescapeRefString(rest)

					definitions := swagger.Components.Callbacks
					if definitions == nil {
						return failedToResolveRefFragmentPart(ref, "callbacks")
					}
					resolved := definitions[id]
					if resolved == nil {
						return failedToResolveRefFragmentPart(ref, id)
					}

					for _, p := range *resolved.Value {
						*pathItem = *p
						break
					}
				}
			}
			return swaggerLoader.resolvePathItemRefContinued(swagger, pathItem, documentPath)
		}()
		if err != nil {
			return err
		}
	}
	return nil
}

func (swaggerLoader *SwaggerLoader) resolveLinkRef(swagger *Swagger, component *LinkRef, documentPath *url.URL) (err error) {
	if component != nil && component.Value != nil {
		if swaggerLoader.visitedLink == nil {
			swaggerLoader.visitedLink = make(map[*Link]struct{})
		}
		if _, ok := swaggerLoader.visitedLink[component.Value]; ok {
			return nil
		}
		swaggerLoader.visitedLink[component.Value] = struct{}{}
	}

	if component == nil {
		return errors.New("invalid link: value MUST be an object")
	}
	if ref := component.Ref; ref != "" {
		if isSingleRefElement(ref) {
			var link Link
			if documentPath, err = swaggerLoader.loadSingleElementFromURI(ref, documentPath, &link); err != nil {
				return err
			}
			component.Value = &link
		} else {
			var resolved LinkRef
			componentPath, err := swaggerLoader.resolveComponent(swagger, ref, documentPath, &resolved)
			if err != nil {
				return err
			}
			if err := swaggerLoader.resolveLinkRef(swagger, &resolved, componentPath); err != nil {
				return err
			}
			component.Value = resolved.Value
		}
	}
	return nil
}

func (swaggerLoader *SwaggerLoader) resolvePathItemRef(swagger *Swagger, entrypoint string, pathItem *PathItem, documentPath *url.URL) (err error) {
	key := "_"
	if documentPath != nil {
		key = documentPath.EscapedPath()
	}
	key += entrypoint
	if _, ok := swaggerLoader.visitedPathItemRefs[key]; ok {
		return nil
	}
	swaggerLoader.visitedPathItemRefs[key] = struct{}{}

	if pathItem == nil {
		return errors.New("invalid path item: value MUST be an object")
	}
	ref := pathItem.Ref
	if ref != "" {
		if isSingleRefElement(ref) {
			var p PathItem
			if documentPath, err = swaggerLoader.loadSingleElementFromURI(ref, documentPath, &p); err != nil {
				return err
			}
			*pathItem = p
		} else {
			if swagger, ref, documentPath, err = swaggerLoader.resolveRefSwagger(swagger, ref, documentPath); err != nil {
				return
			}

			rest := strings.TrimPrefix(ref, "#/paths/")
			if rest == ref {
				return fmt.Errorf(`expected prefix "#/paths/" in URI %q`, ref)
			}
			id := unescapeRefString(rest)

			definitions := swagger.Paths
			if definitions == nil {
				return failedToResolveRefFragmentPart(ref, "paths")
			}
			resolved := definitions[id]
			if resolved == nil {
				return failedToResolveRefFragmentPart(ref, id)
			}

			*pathItem = *resolved
		}
	}
	return swaggerLoader.resolvePathItemRefContinued(swagger, pathItem, documentPath)
}

func (swaggerLoader *SwaggerLoader) resolvePathItemRefContinued(swagger *Swagger, pathItem *PathItem, documentPath *url.URL) (err error) {
	for _, parameter := range pathItem.Parameters {
		if err = swaggerLoader.resolveParameterRef(swagger, parameter, documentPath); err != nil {
			return
		}
	}
	for _, operation := range pathItem.Operations() {
		for _, parameter := range operation.Parameters {
			if err = swaggerLoader.resolveParameterRef(swagger, parameter, documentPath); err != nil {
				return
			}
		}
		if requestBody := operation.RequestBody; requestBody != nil {
			if err = swaggerLoader.resolveRequestBodyRef(swagger, requestBody, documentPath); err != nil {
				return
			}
		}
		for _, response := range operation.Responses {
			if err = swaggerLoader.resolveResponseRef(swagger, response, documentPath); err != nil {
				return
			}
		}
		for _, callback := range operation.Callbacks {
			if err = swaggerLoader.resolveCallbackRef(swagger, callback, documentPath); err != nil {
				return
			}
		}
	}
	return
}

func unescapeRefString(ref string) string {
	return strings.Replace(strings.Replace(ref, "~1", "/", -1), "~0", "~", -1)
}
