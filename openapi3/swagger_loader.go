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
	"reflect"
	"strconv"
	"strings"

	"github.com/ghodss/yaml"
)

func foundUnresolvedRef(ref string) error {
	return fmt.Errorf("Found unresolved ref: '%s'", ref)
}

func failedToResolveRefFragment(value string) error {
	return fmt.Errorf("Failed to resolve fragment in URI: '%s'", value)
}

func failedToResolveRefFragmentPart(value string, what string, reason string) error {
	return fmt.Errorf("Failed to resolve '%s' in fragment in URI: '%s' because: %s", what, value, reason)
}

type SwaggerLoader struct {
	IsExternalRefsAllowed  bool
	ClearResolvedRefs      bool
	SetMetadata            bool
	Context                context.Context
	LoadSwaggerFromURIFunc func(loader *SwaggerLoader, url *url.URL) (*Swagger, error)
	visited                map[interface{}]struct{}
	visitedFiles           map[string]struct{}
	loadedRemoteSchemas    map[url.URL]*Swagger
}

func WithAllowExternalRefs(allow bool) func(sl *SwaggerLoader) {
	return func(sl *SwaggerLoader) {
		sl.IsExternalRefsAllowed = allow
	}
}

func WithClearResolvedRefs(clear bool) func(sl *SwaggerLoader) {
	return func(sl *SwaggerLoader) {
		sl.ClearResolvedRefs = clear
	}
}

func WithSetMetadata(setMD bool) func(sl *SwaggerLoader) {
	return func(sl *SwaggerLoader) {
		sl.SetMetadata = setMD
	}
}

func WithURILoader(l func(loader *SwaggerLoader, url *url.URL) (*Swagger, error)) func(sl *SwaggerLoader) {
	return func(sl *SwaggerLoader) {
		sl.LoadSwaggerFromURIFunc = l
	}
}

func NewSwaggerLoader(options ...func(*SwaggerLoader)) *SwaggerLoader {
	sl := &SwaggerLoader{loadedRemoteSchemas: map[url.URL]*Swagger{}, SetMetadata: true}
	for _, o := range options {
		o(sl)
	}
	return sl
}

func (swaggerLoader *SwaggerLoader) reset() {
	swaggerLoader.visitedFiles = make(map[string]struct{})
}

func (swaggerLoader *SwaggerLoader) LoadSwaggerFromURI(location *url.URL) (*Swagger, error) {
	swaggerLoader.reset()
	return swaggerLoader.loadSwaggerFromURIInternal(location)
}

func (swaggerLoader *SwaggerLoader) loadSwaggerFromURIInternal(location *url.URL) (*Swagger, error) {
	f := swaggerLoader.LoadSwaggerFromURIFunc
	if f != nil {
		return f(swaggerLoader, location)
	}
	data, err := readURL(location)
	if err != nil {
		return nil, err
	}
	return swaggerLoader.loadSwaggerFromDataWithPathInternal(data, location)
}

// loadSingleElementFromURI read the data from ref and unmarshal to JSON to the
// passed element.
func (swaggerLoader *SwaggerLoader) loadSingleElementFromURI(ref string, rootPath *url.URL, element json.Unmarshaler) error {
	if !swaggerLoader.IsExternalRefsAllowed {
		return fmt.Errorf("encountered non-allowed external reference: '%s'", ref)
	}

	parsedURL, err := url.Parse(ref)
	if err != nil {
		return err
	}

	if parsedURL.Fragment != "" {
		return errors.New("references to files which contain more than one element definition are not supported")
	}

	resolvedPath, err := resolvePath(rootPath, parsedURL)
	if err != nil {
		return fmt.Errorf("could not resolve path: %v", err)
	}

	data, err := readURL(resolvedPath)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(data, element); err != nil {
		return err
	}

	return nil
}

func readURL(location *url.URL) ([]byte, error) {
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
	swaggerLoader.reset()
	return swaggerLoader.loadSwaggerFromFileInternal(path)
}

func (swaggerLoader *SwaggerLoader) loadSwaggerFromFileInternal(path string) (*Swagger, error) {
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
	return swaggerLoader.loadSwaggerFromDataWithPathInternal(data, &url.URL{
		Path: path,
	})
}

func (swaggerLoader *SwaggerLoader) LoadSwaggerFromData(data []byte) (*Swagger, error) {
	swaggerLoader.reset()
	return swaggerLoader.loadSwaggerFromDataInternal(data)
}

func (swaggerLoader *SwaggerLoader) loadSwaggerFromDataInternal(data []byte) (*Swagger, error) {
	swagger := &Swagger{}
	if err := yaml.Unmarshal(data, swagger); err != nil {
		return nil, err
	}
	return swagger, swaggerLoader.ResolveRefsIn(swagger, nil)
}

// LoadSwaggerFromDataWithPath takes the OpenApi spec data in bytes and a path where the resolver can find referred
// elements and returns a *Swagger with all resolved data or an error if unable to load data or resolve refs.
func (swaggerLoader *SwaggerLoader) LoadSwaggerFromDataWithPath(data []byte, path *url.URL) (*Swagger, error) {
	swaggerLoader.reset()
	return swaggerLoader.loadSwaggerFromDataWithPathInternal(data, path)
}

func (swaggerLoader *SwaggerLoader) loadSwaggerFromDataWithPathInternal(data []byte, path *url.URL) (*Swagger, error) {
	swagger := &Swagger{}
	if err := yaml.Unmarshal(data, swagger); err != nil {
		return nil, err
	}

	if swaggerLoader.SetMetadata {
		// mark each resource with its id and path
		if err := swaggerLoader.fixMetadata(swagger, path); err != nil {
			return nil, err
		}
	}

	if err := swaggerLoader.ResolveRefsIn(swagger, path); err != nil {
		return nil, err
	}

	if swaggerLoader.ClearResolvedRefs {
		ClearResolvedExternalRefs(swagger)
	}

	return swagger, nil
}

func (swaggerLoader *SwaggerLoader) ResolveRefsIn(swagger *Swagger, path *url.URL) (err error) {
	if swaggerLoader.visited == nil {
		swaggerLoader.visited = make(map[interface{}]struct{})
	}
	if swaggerLoader.visitedFiles == nil {
		swaggerLoader.reset()
	}

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
		if err = swaggerLoader.resolveSecuritySchemeRef(swagger, component, path); err != nil {
			return
		}
	}
	for _, component := range components.Examples {
		if err = swaggerLoader.resolveExampleRef(swagger, component, path); err != nil {
			return
		}
	}

	// Visit all operations
	for entrypoint, pathItem := range swagger.Paths {
		if pathItem == nil {
			continue
		}
		if err = swaggerLoader.resolvePathItemRef(swagger, entrypoint, pathItem, path); err != nil {
			return
		}
		for _, parameter := range pathItem.Parameters {
			fmt.Printf("before path param: %#v\n", parameter)
			if err = swaggerLoader.resolveParameterRef(swagger, parameter, path); err != nil {
				return
			}
		}
		for _, operation := range pathItem.Operations() {
			for _, parameter := range operation.Parameters {
				if err = swaggerLoader.resolveParameterRef(swagger, parameter, path); err != nil {
					return
				}

			}
			if requestBody := operation.RequestBody; requestBody != nil {
				if err = swaggerLoader.resolveRequestBodyRef(swagger, requestBody, path); err != nil {
					return
				}
			}
			for _, response := range operation.Responses {
				if err = swaggerLoader.resolveResponseRef(swagger, response, path); err != nil {
					return
				}
			}
		}
	}

	return
}

func (swaggerLoader *SwaggerLoader) fixMetadata(swagger *Swagger, path *url.URL) (err error) {
	comps := reflect.ValueOf(swagger.Components)
	const compPfx = "/components"
	var p url.URL

	if path != nil {
		p = *path
	}

	swagger.Metadata.ID = p.Path
	swagger.Metadata.Path = p

	for i := 0; i < comps.NumField(); i++ {

		fn := comps.Type().Field(i).Name
		f := comps.Field(i)
		if f.Kind() != reflect.Map {
			continue
		}

		p.Fragment = strings.ToLower(fmt.Sprintf("%s/%s", compPfx, fn))

		if err := swaggerLoader.fixComponentMetadata(f, swagger, p); err != nil {
			return err
		}
	}

	return nil
}

func (swaggerLoader *SwaggerLoader) fixComponentMetadata(comp reflect.Value, swagger *Swagger, p url.URL) error {
	if comp.Kind() != reflect.Map {
		return fmt.Errorf("expecting collection to be a map: %v", comp)
	}

	mi := comp.MapRange()
	for mi.Next() {
		k := mi.Key()
		v := mi.Value()

		if k.Kind() != reflect.String {
			return fmt.Errorf("expecting collection to be a map with string k: %v", k)
		}

		if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
			return fmt.Errorf("expecting collection to be a map with value that is a pointer to struct: (k: %s) %v", k, v)
		}

		// get `Ref` , Value` fields
		rf := v.Elem().FieldByName("Ref")
		v = v.Elem().FieldByName("Value")

		if !v.IsValid() {
			return fmt.Errorf("expecting collection to be a map value to be a pointer to Ref type with a `Value` field: (k: %s) %v", k, v)
		}

		// if this is a reference, value is nil. Ensure reference is not empty
		if v.IsNil() {
			if !rf.IsValid() || rf.Type().Kind() != reflect.String || rf.String() == "" {
				return fmt.Errorf("reference entry is either empty or not string: (k: %s) %v", k, rf)
			}

			continue
		}

		if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
			return fmt.Errorf("expecting collection element Value field to point to a struct: (k: %s) %v", k, v)
		}

		v = v.Elem().FieldByName("Metadata")
		if !v.IsValid() {
			return fmt.Errorf("expecting collection element `Value` field to point to a struct with `Metadata` field: (k: %s) %v", k, v)
		}

		_, ok := v.Interface().(Metadata)
		if !ok {
			return fmt.Errorf("expecting `Metadata` field to be of type Metadata: (k: %s) %v", k, v)
		}

		pp := p
		pp.Fragment = fmt.Sprintf("%s/%s", p.Fragment, k.String())
		v.FieldByName("ID").Set(k)
		v.FieldByName("Path").Set(reflect.ValueOf(pp))
	}

	return nil
}

func copyURL(basePath *url.URL) (*url.URL, error) {
	return url.Parse(basePath.String())
}

func join(basePath *url.URL, relativePath *url.URL) (*url.URL, error) {
	if basePath == nil {
		return relativePath, nil
	}
	newPath, err := copyURL(basePath)
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

func (swaggerLoader *SwaggerLoader) resolveComponent(swagger *Swagger, ref string, prefix string, path *url.URL) (
	cursor interface{},
	componentPath *url.URL,
	err error,
) {
	if swagger, ref, componentPath, err = swaggerLoader.resolveRefSwagger(swagger, ref, prefix, path); err != nil {
		return nil, nil, err
	}

	parsedURL, err := url.Parse(ref)
	if err != nil {
		return nil, nil, fmt.Errorf("Can't parse reference: '%s': %v", ref, parsedURL)
	}
	fragment := parsedURL.Fragment
	if !strings.HasPrefix(fragment, "/") {
		err := fmt.Errorf("expected fragment prefix '#/' in URI '%s'", ref)
		return nil, nil, err
	}

	cursor = swagger
	for _, pathPart := range strings.Split(fragment[1:], "/") {
		pathPart = unescapeRefString(pathPart)

		if cursor, err = drillIntoSwaggerField(cursor, pathPart); err != nil {
			return nil, nil, fmt.Errorf("Failed to resolve '%s' in fragment in URI: '%s': %v", ref, pathPart, err.Error())
		}
		if cursor == nil {
			return nil, nil, failedToResolveRefFragmentPart(ref, pathPart, "nil cursor")
		}
	}

	return cursor, componentPath, nil
}

func drillIntoSwaggerField(cursor interface{}, fieldName string) (interface{}, error) {
	switch val := reflect.Indirect(reflect.ValueOf(cursor)); val.Kind() {
	case reflect.Map:
		elementValue := val.MapIndex(reflect.ValueOf(fieldName))
		if !elementValue.IsValid() {
			return nil, fmt.Errorf("Map key not found: %v", fieldName)
		}
		return elementValue.Interface(), nil

	case reflect.Slice:
		i, err := strconv.ParseUint(fieldName, 10, 32)
		if err != nil {
			return nil, err
		}
		index := int(i)
		if index >= val.Len() {
			return nil, errors.New("slice index out of bounds")
		}
		return val.Index(index).Interface(), nil

	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			field := val.Type().Field(i)
			tagValue := field.Tag.Get("yaml")
			yamlKey := strings.Split(tagValue, ",")[0]
			if yamlKey == fieldName {
				return val.Field(i).Interface(), nil
			}
		}
		// if cursor if a "ref wrapper" struct (e.g. RequestBodyRef), try digging into its Value field
		_, ok := val.Type().FieldByName("Value")
		if ok {
			return drillIntoSwaggerField(val.FieldByName("Value").Interface(), fieldName) // recurse into .Value
		}
		// give up
		return nil, fmt.Errorf("Struct field not found: %v", fieldName)

	default:
		return nil, errors.New("not a map, slice nor struct")
	}
}

func (swaggerLoader *SwaggerLoader) resolveRefSwagger(swagger *Swagger, ref string, prefix string, path *url.URL) (*Swagger, string, *url.URL, error) {
	componentPath := path
	if !strings.HasPrefix(ref, "#") {
		if !swaggerLoader.IsExternalRefsAllowed {
			return nil, "", nil, fmt.Errorf("Encountered non-allowed external reference: '%s'", ref)
		}
		parsedURL, err := url.Parse(ref)
		if err != nil {
			return nil, "", nil, fmt.Errorf("Can't parse reference: '%s': %v", ref, parsedURL)
		}
		fragment := parsedURL.Fragment
		parsedURL.Fragment = ""

		resolvedPath, err := resolvePath(path, parsedURL)
		if err != nil {
			return nil, "", nil, fmt.Errorf("Error while resolving path: %v", err)
		}

		if swg2, ok := swaggerLoader.loadedRemoteSchemas[*parsedURL]; !ok {
			if swg2, err = swaggerLoader.loadSwaggerFromURIInternal(resolvedPath); err != nil {
				return nil, "", nil, fmt.Errorf("Error while resolving reference '%s': %v", ref, err)
			}
			swaggerLoader.loadedRemoteSchemas[*parsedURL] = swg2
		}

		swagger = swaggerLoader.loadedRemoteSchemas[*parsedURL]
		ref = fmt.Sprintf("#%s", fragment)
		if !strings.HasPrefix(ref, prefix) {
			err := fmt.Errorf("expected prefix '%s' in URI '%s'", prefix, ref)
			return nil, "", nil, err
		}
		id := ref[len(prefix):]
		if strings.IndexByte(id, '/') >= 0 {
			return nil, "", nil, failedToResolveRefFragmentPart(ref, id, "failed to strip local path")
		}
		componentPath = resolvedPath
	}
	return swagger, ref, componentPath, nil
}
func (swaggerLoader *SwaggerLoader) resolved(md Metadata) bool {
	_, ok := swaggerLoader.visited[md.Path.String()]
	return ok
}
func (swaggerLoader *SwaggerLoader) setResolved(md Metadata) {
	if k := md.Path.String(); k != "" {
		swaggerLoader.visited[k] = struct{}{}
	}
}

func (swaggerLoader *SwaggerLoader) resolveHeaderRef(swagger *Swagger, component *HeaderRef, path *url.URL) error {
	visited := swaggerLoader.visited
	if _, isVisited := visited[component]; isVisited {
		return nil
	}
	if !component.IsValid() || component.Resolved() {
		return nil
	}
	if component.IsValue() && swaggerLoader.resolved(component.Value.Metadata) {
		return nil
	}
	visited[component] = struct{}{}

	// Prevent infinite recursion
	if !component.IsValid() || component.Resolved() {
		return nil
	}
	if component.IsValue() && swaggerLoader.resolved(component.Value.Metadata) {
		return nil
	}

	// Resolve ref
	const prefix = "#/components/headers/"
	if component == nil {
		return errors.New("invalid header: value MUST be a JSON object")
	}
	if ref := component.Ref; ref != "" {
		if isSingleRefElement(ref) {
			var header Header
			if err := swaggerLoader.loadSingleElementFromURI(ref, path, &header); err != nil {
				return err
			}

			component.Value = &header
		} else {
			untypedResolved, componentPath, err := swaggerLoader.resolveComponent(swagger, ref, prefix, path)
			if err != nil {
				return err
			}
			resolved, ok := untypedResolved.(*HeaderRef)
			if !ok {
				return failedToResolveRefFragment(ref)
			}
			if err := swaggerLoader.resolveHeaderRef(swagger, resolved, componentPath); err != nil {
				return err
			}
			component.Value = resolved.Value
		}
	}

	swaggerLoader.setResolved(component.Value.Metadata)

	value := component.Value
	if value == nil {
		return nil
	}
	if schema := value.Schema; schema != nil {
		if err := swaggerLoader.resolveSchemaRef(swagger, schema, path); err != nil {
			return err
		}
	}
	return nil
}

func (swaggerLoader *SwaggerLoader) resolveParameterRef(swagger *Swagger, component *ParameterRef, path *url.URL) error {
	visited := swaggerLoader.visited
	if _, isVisited := visited[component]; isVisited {
		return nil
	}
	visited[component] = struct{}{}
	// Prevent infinite recursion
	if !component.IsValid() || component.Resolved() {
		return nil
	}
	if component.IsValue() && swaggerLoader.resolved(component.Value.Metadata) {
		return nil
	}

	// Resolve ref
	const prefix = "#/components/parameters/"
	if component == nil {
		return errors.New("invalid parameter: value MUST be a JSON object")
	}
	ref := component.Ref
	if ref != "" {
		if isSingleRefElement(ref) {
			var param Parameter
			if err := swaggerLoader.loadSingleElementFromURI(ref, path, &param); err != nil {
				return err
			}
			component.Value = &param
		} else {
			untypedResolved, componentPath, err := swaggerLoader.resolveComponent(swagger, ref, prefix, path)
			if err != nil {
				return err
			}
			resolved, ok := untypedResolved.(*ParameterRef)
			if !ok {
				return failedToResolveRefFragment(ref)
			}
			if err := swaggerLoader.resolveParameterRef(swagger, resolved, componentPath); err != nil {
				return err
			}
			component.Value = resolved.Value
		}
	}

	swaggerLoader.setResolved(component.Value.Metadata)

	value := component.Value
	if value == nil {
		return nil
	}

	refDocumentPath, err := referencedDocumentPath(path, ref)
	if err != nil {
		return err
	}

	if value.Content != nil && value.Schema != nil {
		return errors.New("Cannot contain both schema and content in a parameter")
	}
	for _, contentType := range value.Content {
		if schema := contentType.Schema; schema != nil {
			if err := swaggerLoader.resolveSchemaRef(swagger, schema, refDocumentPath); err != nil {
				return err
			}
		}
	}
	if schema := value.Schema; schema != nil {
		if err := swaggerLoader.resolveSchemaRef(swagger, schema, refDocumentPath); err != nil {
			return err
		}
	}
	return nil
}

func (swaggerLoader *SwaggerLoader) resolveRequestBodyRef(swagger *Swagger, component *RequestBodyRef, path *url.URL) error {
	visited := swaggerLoader.visited
	if _, isVisited := visited[component]; isVisited {
		return nil
	}
	visited[component] = struct{}{}
	// Prevent infinite recursion
	if !component.IsValid() || component.Resolved() {
		return nil
	}
	if component.IsValue() && swaggerLoader.resolved(component.Value.Metadata) {
		return nil
	}

	const prefix = "#/components/requestBodies/"
	if component == nil {
		return errors.New("invalid requestBody: value MUST be a JSON object")
	}
	if ref := component.Ref; ref != "" {
		if isSingleRefElement(ref) {
			var requestBody RequestBody
			if err := swaggerLoader.loadSingleElementFromURI(ref, path, &requestBody); err != nil {
				return err
			}

			component.Value = &requestBody
		} else {
			untypedResolved, componentPath, err := swaggerLoader.resolveComponent(swagger, ref, prefix, path)
			if err != nil {
				return err
			}
			resolved, ok := untypedResolved.(*RequestBodyRef)
			if !ok {
				return failedToResolveRefFragment(ref)
			}
			if err = swaggerLoader.resolveRequestBodyRef(swagger, resolved, componentPath); err != nil {
				return err
			}
			component.Value = resolved.Value
		}
	}

	swaggerLoader.setResolved(component.Value.Metadata)

	value := component.Value
	if value == nil {
		return nil
	}
	for _, contentType := range value.Content {
		for name, example := range contentType.Examples {
			if err := swaggerLoader.resolveExampleRef(swagger, example, path); err != nil {
				return err
			}
			contentType.Examples[name] = example
		}
		if schema := contentType.Schema; schema != nil {
			if err := swaggerLoader.resolveSchemaRef(swagger, schema, path); err != nil {
				return err
			}
		}
	}
	return nil
}

func (swaggerLoader *SwaggerLoader) resolveResponseRef(swagger *Swagger, component *ResponseRef, documentPath *url.URL) error {
	visited := swaggerLoader.visited
	if _, isVisited := visited[component]; isVisited {
		return nil
	}
	visited[component] = struct{}{}
	// Prevent infinite recursion
	if component == nil {
		return errors.New("invalid response: value MUST be a JSON object")
	}
	if !component.IsValid() || component.Resolved() {
		return nil
	}
	if component.IsValue() && swaggerLoader.resolved(component.Value.Metadata) {
		return nil
	}

	const prefix = "#/components/responses/"

	ref := component.Ref
	if ref != "" {

		if isSingleRefElement(ref) {
			var resp Response
			if err := swaggerLoader.loadSingleElementFromURI(ref, documentPath, &resp); err != nil {
				return err
			}

			component.Value = &resp
		} else {
			untypedResolved, componentPath, err := swaggerLoader.resolveComponent(swagger, ref, prefix, documentPath)
			if err != nil {
				return err
			}
			resolved, ok := untypedResolved.(*ResponseRef)
			if !ok {
				return failedToResolveRefFragment(ref)
			}
			if err := swaggerLoader.resolveResponseRef(swagger, resolved, componentPath); err != nil {
				return err
			}

			component.Value = resolved.Value
		}
	}
	refDocumentPath, err := referencedDocumentPath(documentPath, ref)
	if err != nil {
		return err
	}
	swaggerLoader.setResolved(component.Value.Metadata)

	value := component.Value
	if value == nil {
		return nil
	}
	for _, header := range value.Headers {
		if err := swaggerLoader.resolveHeaderRef(swagger, header, refDocumentPath); err != nil {
			return err
		}
	}
	for _, contentType := range value.Content {
		if contentType == nil {
			continue
		}
		for name, example := range contentType.Examples {
			if err := swaggerLoader.resolveExampleRef(swagger, example, refDocumentPath); err != nil {
				return err
			}
			contentType.Examples[name] = example
		}
		if schema := contentType.Schema; schema != nil {
			if err := swaggerLoader.resolveSchemaRef(swagger, schema, refDocumentPath); err != nil {
				return err
			}
			contentType.Schema = schema
		}
	}
	for _, link := range value.Links {
		if err := swaggerLoader.resolveLinkRef(swagger, link, refDocumentPath); err != nil {
			return err
		}
	}
	return nil
}

func (swaggerLoader *SwaggerLoader) resolveSchemaRef(swagger *Swagger, component *SchemaRef, documentPath *url.URL) error {
	visited := swaggerLoader.visited
	if _, isVisited := visited[component]; isVisited {
		return nil
	}
	visited[component] = struct{}{}
	// Prevent infinite recursion
	if !component.IsValid() || component.Resolved() {
		return nil
	}
	if component.IsValue() && swaggerLoader.resolved(component.Value.Metadata) {
		return nil
	}

	const prefix = "#/components/schemas/"
	if component == nil {
		return errors.New("invalid schema: value MUST be a JSON object")
	}
	ref := component.Ref
	if ref != "" {
		if isSingleRefElement(ref) {
			var schema Schema
			if err := swaggerLoader.loadSingleElementFromURI(ref, documentPath, &schema); err != nil {
				return err
			}
			component.Value = &schema
		} else {
			untypedResolved, componentPath, err := swaggerLoader.resolveComponent(swagger, ref, prefix, documentPath)
			if err != nil {
				return err
			}

			resolved, ok := untypedResolved.(*SchemaRef)
			if !ok {
				return failedToResolveRefFragment(ref)
			}
			if err := swaggerLoader.resolveSchemaRef(swagger, resolved, componentPath); err != nil {
				return err
			}
			component.Value = resolved.Value
		}

	}

	refDocumentPath, err := referencedDocumentPath(documentPath, ref)
	if err != nil {
		return err
	}
	swaggerLoader.setResolved(component.Value.Metadata)

	value := component.Value
	if value == nil {
		return nil
	}

	// ResolveRefs referred schemas
	if v := value.Items; v != nil {
		if err := swaggerLoader.resolveSchemaRef(swagger, v, refDocumentPath); err != nil {
			return err
		}
	}
	for _, v := range value.Properties {
		if err := swaggerLoader.resolveSchemaRef(swagger, v, refDocumentPath); err != nil {
			return err
		}
	}
	if v := value.AdditionalProperties; v != nil {
		if err := swaggerLoader.resolveSchemaRef(swagger, v, refDocumentPath); err != nil {
			return err
		}
	}
	if v := value.Not; v != nil {
		if err := swaggerLoader.resolveSchemaRef(swagger, v, refDocumentPath); err != nil {
			return err
		}
	}
	for _, v := range value.AllOf {
		if err := swaggerLoader.resolveSchemaRef(swagger, v, refDocumentPath); err != nil {
			return err
		}
	}
	for _, v := range value.AnyOf {
		if err := swaggerLoader.resolveSchemaRef(swagger, v, refDocumentPath); err != nil {
			return err
		}
	}
	for _, v := range value.OneOf {
		if err := swaggerLoader.resolveSchemaRef(swagger, v, refDocumentPath); err != nil {
			return err
		}
	}

	return nil
}

func (swaggerLoader *SwaggerLoader) resolveSecuritySchemeRef(swagger *Swagger, component *SecuritySchemeRef, path *url.URL) error {
	visited := swaggerLoader.visited
	if _, isVisited := visited[component]; isVisited {
		return nil
	}
	visited[component] = struct{}{}
	// Prevent infinite recursion
	if !component.IsValid() || component.Resolved() {
		return nil
	}
	if component.IsValue() && swaggerLoader.resolved(component.Value.Metadata) {
		return nil
	}

	const prefix = "#/components/securitySchemes/"
	if component == nil {
		return errors.New("invalid securityScheme: value MUST be a JSON object")
	}
	if ref := component.Ref; ref != "" {
		if isSingleRefElement(ref) {
			var scheme SecurityScheme
			if err := swaggerLoader.loadSingleElementFromURI(ref, path, &scheme); err != nil {
				return err
			}

			component.Value = &scheme
		} else {
			untypedResolved, componentPath, err := swaggerLoader.resolveComponent(swagger, ref, prefix, path)
			if err != nil {
				return err
			}
			resolved, ok := untypedResolved.(*SecuritySchemeRef)
			if !ok {
				return failedToResolveRefFragment(ref)
			}
			if err := swaggerLoader.resolveSecuritySchemeRef(swagger, resolved, componentPath); err != nil {
				return err
			}
			component.Value = resolved.Value
		}
	}
	return nil
}

func (swaggerLoader *SwaggerLoader) resolveExampleRef(swagger *Swagger, component *ExampleRef, path *url.URL) error {
	visited := swaggerLoader.visited
	if _, isVisited := visited[component]; isVisited {
		return nil
	}
	visited[component] = struct{}{}
	// Prevent infinite recursion
	if component == nil || !component.IsValid() || component.Resolved() {
		return nil
	}
	if component.IsValue() && swaggerLoader.resolved(component.Value.Metadata) {
		return nil
	}

	const prefix = "#/components/examples/"
	if component == nil {
		return errors.New("invalid example: value MUST be a JSON object")
	}
	if ref := component.Ref; ref != "" {
		if isSingleRefElement(ref) {
			var example Example
			if err := swaggerLoader.loadSingleElementFromURI(ref, path, &example); err != nil {
				return err
			}

			component.Value = &example
		} else {
			untypedResolved, componentPath, err := swaggerLoader.resolveComponent(swagger, ref, prefix, path)
			if err != nil {
				return err
			}
			resolved, ok := untypedResolved.(*ExampleRef)
			if !ok {
				return failedToResolveRefFragment(ref)
			}
			if err := swaggerLoader.resolveExampleRef(swagger, resolved, componentPath); err != nil {
				return err
			}
			component.Value = resolved.Value
		}
	}
	return nil
}

func (swaggerLoader *SwaggerLoader) resolveLinkRef(swagger *Swagger, component *LinkRef, path *url.URL) error {
	visited := swaggerLoader.visited
	if _, isVisited := visited[component]; isVisited {
		return nil
	}
	visited[component] = struct{}{}

	const prefix = "#/components/links/"
	if component == nil {
		return errors.New("invalid link: value MUST be a JSON object")
	}
	if ref := component.Ref; ref != "" {
		if isSingleRefElement(ref) {
			var link Link
			if err := swaggerLoader.loadSingleElementFromURI(ref, path, &link); err != nil {
				return err
			}

			component.Value = &link
		} else {
			untypedResolved, componentPath, err := swaggerLoader.resolveComponent(swagger, ref, prefix, path)
			if err != nil {
				return err
			}
			resolved, ok := untypedResolved.(*LinkRef)
			if !ok {
				return failedToResolveRefFragment(ref)
			}
			if err := swaggerLoader.resolveLinkRef(swagger, resolved, componentPath); err != nil {
				return err
			}
			component.Value = resolved.Value
		}
	}
	return nil
}

func (swaggerLoader *SwaggerLoader) resolvePathItemRef(swagger *Swagger, entrypoint string, pathItem *PathItem, documentPath *url.URL) (err error) {
	visited := swaggerLoader.visitedFiles
	key := "_"
	if documentPath != nil {
		key = documentPath.EscapedPath()
	}
	key += entrypoint
	if _, isVisited := visited[key]; isVisited {
		return nil
	}
	visited[key] = struct{}{}

	const prefix = "#/paths/"
	if pathItem == nil {
		return errors.New("invalid path item: value MUST be a JSON object")
	}
	ref := pathItem.Ref
	if ref != "" {
		if isSingleRefElement(ref) {
			var p PathItem
			if err := swaggerLoader.loadSingleElementFromURI(ref, documentPath, &p); err != nil {
				return err
			}
			*pathItem = p
		} else {
			if swagger, ref, documentPath, err = swaggerLoader.resolveRefSwagger(swagger, ref, prefix, documentPath); err != nil {
				return
			}

			if !strings.HasPrefix(ref, prefix) {
				err = fmt.Errorf("expected prefix '%s' in URI '%s'", prefix, ref)
				return
			}
			id := unescapeRefString(ref[len(prefix):])

			definitions := swagger.Paths
			if definitions == nil {
				return failedToResolveRefFragmentPart(ref, "paths", "nil definitions")
			}
			resolved := definitions[id]
			if resolved == nil {
				return failedToResolveRefFragmentPart(ref, id, "nothing found for id")
			}

			*pathItem = *resolved
		}
	}

	refDocumentPath, err := referencedDocumentPath(documentPath, ref)
	if err != nil {
		return err
	}

	for _, parameter := range pathItem.Parameters {
		if err = swaggerLoader.resolveParameterRef(swagger, parameter, refDocumentPath); err != nil {
			return
		}
	}
	for _, operation := range pathItem.Operations() {
		for _, parameter := range operation.Parameters {
			if err = swaggerLoader.resolveParameterRef(swagger, parameter, refDocumentPath); err != nil {
				return
			}
		}
		if requestBody := operation.RequestBody; requestBody != nil {
			if err = swaggerLoader.resolveRequestBodyRef(swagger, requestBody, refDocumentPath); err != nil {
				return
			}
		}
		for _, response := range operation.Responses {
			if err = swaggerLoader.resolveResponseRef(swagger, response, refDocumentPath); err != nil {
				return
			}
		}
	}

	return nil
}

func unescapeRefString(ref string) string {
	return strings.Replace(strings.Replace(ref, "~1", "/", -1), "~0", "~", -1)
}

func referencedDocumentPath(documentPath *url.URL, ref string) (*url.URL, error) {
	newDocumentPath := documentPath
	if documentPath != nil {
		refDirectory, err := url.Parse(path.Dir(ref))
		if err != nil {
			return nil, err
		}
		joinedDirectory := path.Join(path.Dir(documentPath.String()), refDirectory.String())
		if newDocumentPath, err = url.Parse(joinedDirectory + "/"); err != nil {
			return nil, err
		}
	}
	return newDocumentPath, nil
}
