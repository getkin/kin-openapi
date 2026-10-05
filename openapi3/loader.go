package openapi3

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

// IncludeOrigin specifies whether to include the origin of the OpenAPI elements.
// Deprecated: set Loader.IncludeOrigin instead. This global is read by NewLoader
// for backward compatibility but is not safe for concurrent use.
var IncludeOrigin = false

func failedToResolveRefFragmentPart(value, what string) error {
	return fmt.Errorf("failed to resolve %q in fragment in URI: %q", what, value)
}

// Loader helps deserialize an OpenAPIv3 document
type Loader struct {
	// IsExternalRefsAllowed enables visiting other files. Enforced only when
	// ReadFromURIFunc is nil; a custom ReadFromURIFunc bypasses this flag and
	// owns the access policy itself — see ReadFromURIFunc.
	IsExternalRefsAllowed bool

	// IncludeOrigin enables recording the file/line/column of each OpenAPI element.
	// Prefer this over the package-level IncludeOrigin global, which is not safe for
	// concurrent use.
	IncludeOrigin bool

	// ReadFromURIFunc overrides how the loader reads a referenced file or URL.
	//
	// SECURITY: when a custom ReadFromURIFunc is set, IsExternalRefsAllowed is
	// NOT enforced — this function alone decides which locations may be read. A
	// func that reads whatever URI it is handed (e.g. by delegating to
	// DefaultReadFromURI) resolves external $refs even when IsExternalRefsAllowed
	// is false, which on untrusted documents enables local file reads
	// (`$ref: "/etc/passwd"`) and SSRF (`$ref: "http://169.254.169.254/..."`).
	// A custom func must apply its own scheme/host allowlist, or re-check
	// IsExternalRefsAllowed, before reading.
	ReadFromURIFunc ReadFromURIFunc

	// JoinFunc allows overriding how relative $ref paths are resolved against
	// a base path. When set, it is called instead of the default join logic
	// that uses path.Dir and path.Join. This is useful when loading specs from
	// non-filesystem sources (e.g. git objects, remote archives) where the base
	// path follows a different convention than filesystem paths.
	JoinFunc func(basePath *url.URL, relativePath *url.URL) *url.URL

	Context context.Context

	rootDir      string
	rootLocation string

	visitedPathItemRefs map[string]struct{}

	visitedDocuments map[string]*T

	// originTrees retains each loaded document's origin tree, keyed by the
	// document itself so insert and lookup cannot disagree, populated when
	// IncludeOrigin is set. resolveComponent uses it to re-attach origins to
	// components that lose them in the generic-map path, without re-reading
	// or re-parsing the file.
	originTrees map[*T]*originTree

	visitedRefs map[string]struct{}
	visitedPath []string
	backtrack   map[string][]func(value any)

	// schemaResolution is shared by nested ResolveRefsIn calls while one root
	// document is being resolved. It keeps the root dialect stable across raw
	// external schema documents and owns the graph used to close schema cycles.
	schemaResolution *schemaResolutionContext
}

// NewLoader returns an empty Loader
func NewLoader() *Loader {
	return &Loader{
		Context:       context.Background(),
		IncludeOrigin: IncludeOrigin,
	}
}

func (loader *Loader) resetVisitedPathItemRefs() {
	loader.visitedPathItemRefs = make(map[string]struct{})
	loader.visitedRefs = make(map[string]struct{})
	loader.visitedPath = nil
	loader.backtrack = make(map[string][]func(value any))
}

// LoadFromURI loads a spec from a remote URL
func (loader *Loader) LoadFromURI(location *url.URL) (*T, error) {
	loader.resetVisitedPathItemRefs()
	return loader.loadFromURIInternal(location)
}

// LoadFromFile loads a spec from a local file path
func (loader *Loader) LoadFromFile(location string) (*T, error) {
	loader.rootDir = path.Dir(location)
	return loader.LoadFromURI(&url.URL{Path: filepath.ToSlash(location)})
}

func (loader *Loader) loadFromURIInternal(location *url.URL) (*T, error) {
	data, err := loader.readURL(location)
	if err != nil {
		return nil, err
	}
	return loader.loadFromDataWithPathInternal(data, location)
}

func (loader *Loader) allowsExternalRefs(ref string) (err error) {
	if !loader.IsExternalRefsAllowed {
		err = fmt.Errorf("encountered disallowed external reference: %q", ref)
	}
	return
}

func (loader *Loader) loadSingleElementFromURI(ref string, rootPath *url.URL, element any) (*url.URL, error) {
	// IsExternalRefsAllowed is enforced here only when no custom ReadFromURIFunc
	// is installed; otherwise the custom func owns the access policy (see the
	// SECURITY note on the ReadFromURIFunc field).
	if loader.ReadFromURIFunc == nil {
		if err := loader.allowsExternalRefs(ref); err != nil {
			return nil, err
		}
	}

	resolvedPath, err := loader.resolvePathWithRef(ref, rootPath)
	if err != nil {
		return nil, err
	}
	if frag := resolvedPath.Fragment; frag != "" {
		return nil, fmt.Errorf("unexpected ref fragment %q", frag)
	}

	data, err := loader.readURL(resolvedPath)
	if err != nil {
		return nil, err
	}
	if _, err := unmarshal(data, element, loader.IncludeOrigin, resolvedPath); err != nil {
		return nil, err
	}

	return resolvedPath, nil
}

// rememberOriginTree retains doc's origin tree for attachOriginToResolved.
// tree is nil when IncludeOrigin is off or the data took the json path.
//
// The tree is kept only for a document a $ref can reach into untyped, which is
// what attachOriginToResolved exists to re-origin. In practice that means a
// file of shared fragments, whose top level is the fragment name itself rather
// than the fields of an OpenAPI Object:
//
//	User:            # a $ref to "./schemas.yaml#/User" lands here, untyped
//	  type: object
//
// Anything OpenAPI defines a field for resolves through typed structures and
// keeps its origins on the way, so its tree could never be read. That includes
// a referenced document that is itself an OpenAPI Object: a $ref to
// "#/components/schemas/User" needs no tree. (A top-level x- extension is
// undefined by the same rule, so a document carrying one keeps its tree too,
// whether or not anything ever points at it.)
//
// Worth the condition: on a 22 MB spec the retained tree was a third of
// everything the loader held.
func (loader *Loader) rememberOriginTree(doc *T, tree *originTree) {
	if tree == nil || len(doc.Extensions) == 0 {
		return
	}
	if loader.originTrees == nil {
		loader.originTrees = make(map[*T]*originTree)
	}
	loader.originTrees[doc] = tree
}

func (loader *Loader) readURL(location *url.URL) ([]byte, error) {
	if f := loader.ReadFromURIFunc; f != nil {
		return f(loader, location)
	}
	return DefaultReadFromURI(loader, location)
}

// LoadFromStdin loads a spec from stdin
func (loader *Loader) LoadFromStdin() (*T, error) {
	return loader.LoadFromIoReader(os.Stdin)
}

// LoadFromStdin loads a spec from io.Reader
func (loader *Loader) LoadFromIoReader(reader io.Reader) (*T, error) {
	if reader == nil {
		return nil, fmt.Errorf("invalid reader: %v", reader)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return loader.LoadFromData(data)
}

// LoadFromData loads a spec from a byte array
func (loader *Loader) LoadFromData(data []byte) (*T, error) {
	loader.resetVisitedPathItemRefs()
	doc := &T{}
	tree, err := unmarshal(data, doc, loader.IncludeOrigin, nil)
	if err != nil {
		return nil, err
	}
	loader.rememberOriginTree(doc, tree)
	if err := loader.ResolveRefsIn(doc, nil); err != nil {
		return nil, err
	}
	return doc, nil
}

// LoadFromDataWithPath takes the OpenAPI document data in bytes and a path where the resolver can find referred
// elements and returns a *T with all resolved data or an error if unable to load data or resolve refs.
func (loader *Loader) LoadFromDataWithPath(data []byte, location *url.URL) (*T, error) {
	loader.resetVisitedPathItemRefs()
	return loader.loadFromDataWithPathInternal(data, location)
}

func (loader *Loader) loadFromDataWithPathInternal(data []byte, location *url.URL) (*T, error) {
	if loader.visitedDocuments == nil {
		loader.visitedDocuments = make(map[string]*T)
		loader.rootLocation = location.Path
	}
	uri := location.String()
	if doc, ok := loader.visitedDocuments[uri]; ok {
		return doc, nil
	}

	doc := &T{}
	loader.visitedDocuments[uri] = doc

	tree, err := unmarshal(data, doc, loader.IncludeOrigin, location)
	if err != nil {
		return nil, err
	}
	loader.rememberOriginTree(doc, tree)

	doc.url = copyURI(location)

	if err := loader.ResolveRefsIn(doc, location); err != nil {
		return nil, err
	}

	return doc, nil
}

// ResolveRefsIn expands references if for instance spec was just unmarshaled
func (loader *Loader) ResolveRefsIn(doc *T, location *url.URL) (err error) {
	if loader.Context == nil {
		loader.Context = context.Background()
	}

	if loader.visitedPathItemRefs == nil {
		loader.resetVisitedPathItemRefs()
	}

	schemaResolution, rootSchemaResolution := loader.beginSchemaResolution(doc)
	if rootSchemaResolution {
		defer func() { loader.schemaResolution = nil }()
	}

	if components := doc.Components; components != nil {
		for _, name := range componentNames(components.Headers) {
			component := components.Headers[name]
			if err = loader.resolveHeaderRef(doc, component, location); err != nil {
				return
			}
		}
		for _, name := range componentNames(components.Parameters) {
			component := components.Parameters[name]
			if err = loader.resolveParameterRef(doc, component, location); err != nil {
				return
			}
		}
		for _, name := range componentNames(components.RequestBodies) {
			component := components.RequestBodies[name]
			if err = loader.resolveRequestBodyRef(doc, component, location); err != nil {
				return
			}
		}
		for _, name := range componentNames(components.Responses) {
			component := components.Responses[name]
			if err = loader.resolveResponseRef(doc, component, location); err != nil {
				return
			}
		}
		for _, name := range componentNames(components.Schemas) {
			component := components.Schemas[name]
			if err = loader.resolveSchemaRef(doc, component, location, schemaResolution); err != nil {
				return
			}
		}
		for _, name := range componentNames(components.SecuritySchemes) {
			component := components.SecuritySchemes[name]
			if err = loader.resolveSecuritySchemeRef(doc, component, location); err != nil {
				return
			}
		}
		for _, name := range componentNames(components.Examples) {
			component := components.Examples[name]
			if err = loader.resolveExampleRef(doc, component, location); err != nil {
				return
			}
		}
		for _, name := range componentNames(components.Callbacks) {
			component := components.Callbacks[name]
			if err = loader.resolveCallbackRef(doc, component, location); err != nil {
				return
			}
		}
	}

	// Visit all operations
	pathItems := doc.Paths.Map()
	for _, name := range componentNames(pathItems) {
		pathItem := pathItems[name]
		if pathItem == nil {
			continue
		}
		if err = loader.resolvePathItemRef(doc, pathItem, location); err != nil {
			return
		}
	}

	for _, name := range componentNames(doc.Webhooks) {
		if pathItem := doc.Webhooks[name]; pathItem != nil {
			if err = loader.resolvePathItemRef(doc, pathItem, location); err != nil {
				return
			}
		}
	}

	return
}

func defaultJoin(basePath *url.URL, relativePath *url.URL) *url.URL {
	if basePath == nil {
		return relativePath
	}
	newPath := *basePath
	newPath.Path = path.Join(path.Dir(newPath.Path), relativePath.Path)
	return &newPath
}

func (loader *Loader) resolvePath(basePath *url.URL, componentPath *url.URL) *url.URL {
	if is_file(componentPath) {
		// support absolute paths
		if filepath.IsAbs(componentPath.Path) {
			return componentPath
		}
		if loader.JoinFunc != nil {
			return loader.JoinFunc(basePath, componentPath)
		}
		return defaultJoin(basePath, componentPath)
	}
	return componentPath
}

func (loader *Loader) resolvePathWithRef(ref string, rootPath *url.URL) (*url.URL, error) {
	parsedURL, err := url.Parse(ref)
	if err != nil {
		return nil, fmt.Errorf("cannot parse reference: %q: %w", ref, err)
	}

	resolvedPath := loader.resolvePath(rootPath, parsedURL)
	resolvedPath.Fragment = parsedURL.Fragment
	return resolvedPath, nil
}

func (loader *Loader) resolveRefPath(ref string, path *url.URL) (*url.URL, error) {
	if ref != "" && ref[0] != '#' && loader.ReadFromURIFunc == nil {
		if err := loader.allowsExternalRefs(ref); err != nil {
			return nil, err
		}
	}

	return loader.resolveRefPathUnchecked(ref, path)
}

func (loader *Loader) resolveRefPathUnchecked(ref string, path *url.URL) (*url.URL, error) {
	if ref != "" && ref[0] == '#' {
		path = copyURI(path)
		// Resolving internal refs of a doc loaded from memory
		// has no path, so just set the Fragment.
		if path == nil {
			path = new(url.URL)
		}

		path.Fragment = strings.TrimPrefix(ref, "#")
		return path, nil
	}

	resolvedPath, err := loader.resolvePathWithRef(ref, path)
	if err != nil {
		return nil, err
	}

	return resolvedPath, nil
}

func isSingleRefElement(ref string) bool {
	return !strings.Contains(ref, "#")
}

func (loader *Loader) visitRef(ref string) {
	if loader.visitedRefs == nil {
		loader.visitedRefs = make(map[string]struct{})
		loader.backtrack = make(map[string][]func(value any))
	}
	loader.visitedPath = append(loader.visitedPath, ref)
	loader.visitedRefs[ref] = struct{}{}
}

func (loader *Loader) unvisitRef(ref string, value any) {
	if value != nil {
		for _, fn := range loader.backtrack[ref] {
			fn(value)
		}
	}
	delete(loader.visitedRefs, ref)
	delete(loader.backtrack, ref)
	loader.visitedPath = loader.visitedPath[:len(loader.visitedPath)-1]
}

func (loader *Loader) shouldVisitRef(ref string, fn func(value any)) bool {
	if _, ok := loader.visitedRefs[ref]; ok {
		loader.backtrack[ref] = append(loader.backtrack[ref], fn)
		return false
	}
	return true
}

type schemaResolutionContext struct {
	isOpenAPI31OrLater bool
	refs               map[string]*schemaResolutionNode
}

type schemaResolutionNode struct {
	resolving   bool
	value       *Schema
	refPath     *url.URL
	placeholder bool
}

func (loader *Loader) beginSchemaResolution(doc *T) (*schemaResolutionContext, bool) {
	if resolution := loader.schemaResolution; resolution != nil {
		return resolution, false
	}
	resolution := &schemaResolutionContext{
		isOpenAPI31OrLater: doc.IsOpenAPI31OrLater(),
		refs:               make(map[string]*schemaResolutionNode),
	}
	loader.schemaResolution = resolution
	return resolution, true
}

func schemaResolutionKey(ref string, documentPath *url.URL) string {
	document := copyURI(documentPath)
	if document == nil {
		document = new(url.URL)
	}
	document.Fragment = ""
	return document.String() + "\x00" + ref
}

func (loader *Loader) resolveComponent(doc *T, ref string, path *url.URL, resolved any) (
	componentDoc *T,
	componentPath *url.URL,
	err error,
) {
	if componentDoc, ref, componentPath, err = loader.resolveRefAndDocument(doc, ref, path); err != nil {
		return nil, nil, err
	}

	parsedURL, err := url.Parse(ref)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot parse reference: %q: %v", ref, parsedURL)
	}
	fragment := parsedURL.Fragment
	if fragment == "" {
		fragment = "/"
	}
	if fragment[0] != '/' {
		return nil, nil, fmt.Errorf("expected fragment prefix '#/' in URI %q", ref)
	}

	drill := func(cursor any) (any, error) {
		for pathPart := range strings.SplitSeq(fragment[1:], "/") {
			pathPart = unescapeRefString(pathPart)
			attempted := false

			switch c := cursor.(type) {
			// Special case of T
			// See issue856: a ref to doc => we assume that doc is a T => things live in T.Extensions
			case *T:
				if pathPart == "" {
					cursor = c.Extensions
					attempted = true
				}

			// Special case due to multijson
			case *SchemaRef:
				if pathPart == "additionalProperties" {
					if s := c.Value; s != nil {
						if ap := s.AdditionalProperties.Has; ap != nil {
							cursor = *ap
						} else {
							cursor = s.AdditionalProperties.Schema
						}
					}
					attempted = true
				}

			case *Responses:
				cursor = c.m // m map[string]*ResponseRef
			case *Callback:
				cursor = c.m // m map[string]*PathItem
			case *Paths:
				cursor = c.m // m map[string]*PathItem
			}

			if !attempted {
				if cursor, err = drillIntoField(cursor, pathPart); err != nil {
					e := failedToResolveRefFragmentPart(ref, pathPart)
					return nil, fmt.Errorf("%s: %w", e, err)
				}
			}

			if cursor == nil {
				return nil, failedToResolveRefFragmentPart(ref, pathPart)
			}
		}
		return cursor, nil
	}
	var cursor any
	if cursor, err = drill(componentDoc); err != nil {
		if path == nil {
			return nil, nil, err
		}
		var err2 error
		data, err2 := loader.readURL(path)
		if err2 != nil {
			return nil, nil, err
		}
		if _, err2 = unmarshal(data, &cursor, loader.IncludeOrigin, path); err2 != nil {
			return nil, nil, err
		}
		if cursor, err2 = drill(cursor); err2 != nil || cursor == nil {
			return nil, nil, err
		}
		err = nil
	}

	setPathRef := func(target any) {
		if i, ok := target.(interface {
			setRefPath(*url.URL)
		}); ok {
			pathRef := copyURI(componentPath)
			// Resolving internal refs of a doc loaded from memory
			// has no path, so just set the Fragment.
			if pathRef == nil {
				pathRef = new(url.URL)
			}
			pathRef.Fragment = fragment

			i.setRefPath(pathRef)
		}
	}

	switch {
	case reflect.TypeOf(cursor) == reflect.TypeOf(resolved):
		setPathRef(cursor)

		reflect.ValueOf(resolved).Elem().Set(reflect.ValueOf(cursor).Elem())
		return componentDoc, componentPath, nil

	case reflect.TypeOf(cursor) == reflect.TypeFor[map[string]any]():
		codec := func(got, expect any) error {
			enc, err := json.Marshal(got)
			if err != nil {
				return err
			}
			if err = json.Unmarshal(enc, expect); err != nil {
				return err
			}

			setPathRef(expect)
			return nil
		}
		if err := codec(cursor, resolved); err != nil {
			return nil, nil, fmt.Errorf("bad data in %q (expecting %s)", ref, readableType(resolved))
		}
		// The value came from a generic map in T.Extensions (a $ref to an
		// arbitrary top-level key), so the json round-trip above stripped its
		// origins. Re-attach them from the file, best-effort.
		loader.attachOriginToResolved(resolved, componentDoc, fragment)
		return componentDoc, componentPath, nil

	default:
		return nil, nil, fmt.Errorf("bad data in %q (expecting %s)", ref, readableType(resolved))
	}
}

// attachOriginToResolved re-attaches source origins to a component resolved
// through the generic-map path: a $ref to a schema under an arbitrary top-level
// key lands in T.Extensions, and the json round-trip in resolveComponent strips
// its origin. It walks the document's retained origin tree (see originTrees,
// populated when the document was first unmarshaled: no re-read, no re-parse)
// down to the ref fragment and applies that subtree, so the object carries the
// same origins a typed resolution would, with the original file's line numbers.
// Best-effort: a missing tree or fragment leaves the object without origins.
func (loader *Loader) attachOriginToResolved(resolved any, componentDoc *T, fragment string) {
	if !loader.IncludeOrigin {
		return
	}
	tree := loader.originTrees[componentDoc]
	if tree == nil {
		return
	}
	for part := range strings.SplitSeq(strings.Trim(fragment, "/"), "/") {
		if part == "" {
			continue
		}
		if tree = tree.Fields[unescapeRefString(part)]; tree == nil {
			return
		}
	}
	applyOrigins(resolved, tree)
}

func readableType(x any) string {
	switch x.(type) {
	case *Callback:
		return "callback object"
	case *CallbackRef:
		return "ref to callback object"
	case *ExampleRef:
		return "ref to example object"
	case *HeaderRef:
		return "ref to header object"
	case *LinkRef:
		return "ref to link object"
	case *ParameterRef:
		return "ref to parameter object"
	case *PathItem:
		return "pathItem object"
	case *RequestBodyRef:
		return "ref to requestBody object"
	case *ResponseRef:
		return "ref to response object"
	case *SchemaRef:
		return "ref to schema object"
	case *SecuritySchemeRef:
		return "ref to securityScheme object"
	default:
		panic(fmt.Sprintf("unreachable %T", x))
	}
}

func drillIntoField(cursor any, fieldName string) (any, error) {
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
		for i := range val.NumField() {
			hasFields = true
			if yamlTag := val.Type().Field(i).Tag.Get("yaml"); yamlTag != "-" {
				if tagName, _, _ := strings.Cut(yamlTag, ","); tagName != "" {
					if fieldName == tagName {
						return val.Field(i).Interface(), nil
					}
				}
			}
		}

		// if cursor is a "ref wrapper" struct (e.g. RequestBodyRef),
		if _, ok := val.Type().FieldByName("Value"); ok {
			// try digging into its Value field
			return drillIntoField(val.FieldByName("Value").Interface(), fieldName)
		}
		if hasFields {
			if ff := val.Type().Field(0); ff.PkgPath == "" && ff.Name == "Extensions" {
				extensions := val.Field(0).Interface().(map[string]any)
				if enc, ok := extensions[fieldName]; ok {
					return enc, nil
				}
			}
		}
		return nil, fmt.Errorf("struct field %q not found", fieldName)

	default:
		return nil, errors.New("not a map, slice nor struct")
	}
}

func (loader *Loader) resolveRefAndDocument(doc *T, ref string, path *url.URL) (*T, string, *url.URL, error) {
	if ref != "" && ref[0] == '#' {
		return doc, ref, path, nil
	}

	fragment, resolvedPath, err := loader.resolveRef(ref, path)
	if err != nil {
		return nil, "", nil, err
	}

	if doc, err = loader.loadFromURIInternal(resolvedPath); err != nil {
		return nil, "", nil, fmt.Errorf("error resolving reference %q: %w", ref, err)
	}

	return doc, fragment, resolvedPath, nil
}

func (loader *Loader) resolveRef(ref string, path *url.URL) (string, *url.URL, error) {
	resolvedPathRef, err := loader.resolveRefPath(ref, path)
	if err != nil {
		return "", nil, err
	}

	fragment := "#" + resolvedPathRef.Fragment
	resolvedPathRef.Fragment = ""
	return fragment, resolvedPathRef, nil
}

var (
	errMUSTCallback       = errors.New("invalid callback: value MUST be an object")
	errMUSTExample        = errors.New("invalid example: value MUST be an object")
	errMUSTHeader         = errors.New("invalid header: value MUST be an object")
	errMUSTLink           = errors.New("invalid link: value MUST be an object")
	errMUSTParameter      = errors.New("invalid parameter: value MUST be an object")
	errMUSTPathItem       = errors.New("invalid path item: value MUST be an object")
	errMUSTRequestBody    = errors.New("invalid requestBody: value MUST be an object")
	errMUSTResponse       = errors.New("invalid response: value MUST be an object")
	errMUSTSchema         = errors.New("invalid schema: value MUST be an object")
	errMUSTSecurityScheme = errors.New("invalid securityScheme: value MUST be an object")
)

func applyHeaderRefMetadata(value *Header, component *HeaderRef, isOpenAPI31OrLater bool) {
	if !isOpenAPI31OrLater || value == nil || component.Description == nil {
		return
	}

	value.Description = *component.Description
}

func (loader *Loader) resolveHeaderRef(doc *T, component *HeaderRef, documentPath *url.URL) (err error) {
	isOpenAPI31OrLater := doc.IsOpenAPI31OrLater()

	if component.isEmpty() {
		return errMUSTHeader
	}

	if ref := component.Ref; ref != "" {
		if component.Value != nil {
			return nil
		}
		if !loader.shouldVisitRef(ref, func(value any) {
			component.Value = value.(*Header)
			applyHeaderRefMetadata(component.Value, component, isOpenAPI31OrLater)
			refPath, _ := loader.resolveRefPath(ref, documentPath)
			component.setRefPath(refPath)
		}) {
			return nil
		}
		loader.visitRef(ref)
		if isSingleRefElement(ref) {
			var header Header
			if documentPath, err = loader.loadSingleElementFromURI(ref, documentPath, &header); err != nil {
				return err
			}
			component.Value = &header
			component.setRefPath(documentPath)
		} else {
			var resolved HeaderRef
			doc, componentPath, err := loader.resolveComponent(doc, ref, documentPath, &resolved)
			if err != nil {
				return err
			}
			if err := loader.resolveHeaderRef(doc, &resolved, componentPath); err != nil {
				if err == errMUSTHeader {
					return nil
				}
				return err
			}
			component.Value = resolved.Value
			component.setRefPath(resolved.RefPath())
		}
		defer loader.unvisitRef(ref, component.Value)
		applyHeaderRefMetadata(component.Value, component, isOpenAPI31OrLater)
	}
	value := component.Value
	if value == nil {
		return nil
	}

	if schema := value.Schema; schema != nil {
		if err := loader.resolveSchemaRef(doc, schema, documentPath, loader.schemaResolution); err != nil {
			return err
		}
	}
	for _, k := range componentNames(value.Examples) {
		if err := loader.resolveExampleRef(doc, value.Examples[k], documentPath); err != nil {
			return err
		}
	}
	return nil
}

func applyParameterRefMetadata(value *Parameter, component *ParameterRef, isOpenAPI31OrLater bool) {
	if !isOpenAPI31OrLater || value == nil || component.Description == nil {
		return
	}

	value.Description = *component.Description
}

func (loader *Loader) resolveParameterRef(doc *T, component *ParameterRef, documentPath *url.URL) (err error) {
	isOpenAPI31OrLater := doc.IsOpenAPI31OrLater()

	if component.isEmpty() {
		return errMUSTParameter
	}

	if ref := component.Ref; ref != "" {
		if component.Value != nil {
			return nil
		}
		if !loader.shouldVisitRef(ref, func(value any) {
			component.Value = value.(*Parameter)
			applyParameterRefMetadata(component.Value, component, isOpenAPI31OrLater)
			refPath, _ := loader.resolveRefPath(ref, documentPath)
			component.setRefPath(refPath)
		}) {
			return nil
		}
		loader.visitRef(ref)
		if isSingleRefElement(ref) {
			var param Parameter
			if documentPath, err = loader.loadSingleElementFromURI(ref, documentPath, &param); err != nil {
				return err
			}
			component.Value = &param
			component.setRefPath(documentPath)
		} else {
			var resolved ParameterRef
			doc, componentPath, err := loader.resolveComponent(doc, ref, documentPath, &resolved)
			if err != nil {
				return err
			}
			if err := loader.resolveParameterRef(doc, &resolved, componentPath); err != nil {
				if err == errMUSTParameter {
					return nil
				}
				return err
			}
			component.Value = resolved.Value
			component.setRefPath(resolved.RefPath())
		}
		defer loader.unvisitRef(ref, component.Value)
		applyParameterRefMetadata(component.Value, component, isOpenAPI31OrLater)
	}
	value := component.Value
	if value == nil {
		return nil
	}

	if value.Content != nil && value.Schema != nil {
		return errors.New("cannot contain both schema and content in a parameter")
	}
	for _, name := range componentNames(value.Content) {
		if err := loader.resolveMediaTypeRefs(doc, value.Content[name], documentPath); err != nil {
			return err
		}
	}
	if schema := value.Schema; schema != nil {
		if err := loader.resolveSchemaRef(doc, schema, documentPath, loader.schemaResolution); err != nil {
			return err
		}
	}
	for _, k := range componentNames(value.Examples) {
		if err := loader.resolveExampleRef(doc, value.Examples[k], documentPath); err != nil {
			return err
		}
	}
	return nil
}

func applyRequestBodyRefMetadata(value *RequestBody, component *RequestBodyRef, isOpenAPI31OrLater bool) {
	if !isOpenAPI31OrLater || value == nil || component.Description == nil {
		return
	}

	value.Description = *component.Description
}

func (loader *Loader) resolveRequestBodyRef(doc *T, component *RequestBodyRef, documentPath *url.URL) (err error) {
	isOpenAPI31OrLater := doc.IsOpenAPI31OrLater()

	if component.isEmpty() {
		return errMUSTRequestBody
	}

	if ref := component.Ref; ref != "" {
		if component.Value != nil {
			return nil
		}
		if !loader.shouldVisitRef(ref, func(value any) {
			component.Value = value.(*RequestBody)
			applyRequestBodyRefMetadata(component.Value, component, isOpenAPI31OrLater)
			refPath, _ := loader.resolveRefPath(ref, documentPath)
			component.setRefPath(refPath)
		}) {
			return nil
		}
		loader.visitRef(ref)
		if isSingleRefElement(ref) {
			var requestBody RequestBody
			if documentPath, err = loader.loadSingleElementFromURI(ref, documentPath, &requestBody); err != nil {
				return err
			}
			component.Value = &requestBody
			component.setRefPath(documentPath)
		} else {
			var resolved RequestBodyRef
			doc, componentPath, err := loader.resolveComponent(doc, ref, documentPath, &resolved)
			if err != nil {
				return err
			}
			if err = loader.resolveRequestBodyRef(doc, &resolved, componentPath); err != nil {
				if err == errMUSTRequestBody {
					return nil
				}
				return err
			}
			component.Value = resolved.Value
			component.setRefPath(resolved.RefPath())
		}
		defer loader.unvisitRef(ref, component.Value)
		applyRequestBodyRefMetadata(component.Value, component, isOpenAPI31OrLater)
	}
	value := component.Value
	if value == nil {
		return nil
	}

	for _, name := range componentNames(value.Content) {
		if err := loader.resolveMediaTypeRefs(doc, value.Content[name], documentPath); err != nil {
			return err
		}
	}
	return nil
}

func applyResponseRefMetadata(value *Response, component *ResponseRef, isOpenAPI31OrLater bool) {
	if !isOpenAPI31OrLater || value == nil || component.Description == nil {
		return
	}

	value.Description = component.Description
}

func (loader *Loader) resolveResponseRef(doc *T, component *ResponseRef, documentPath *url.URL) (err error) {
	isOpenAPI31OrLater := doc.IsOpenAPI31OrLater()

	if component.isEmpty() {
		return errMUSTResponse
	}

	if ref := component.Ref; ref != "" {
		if component.Value != nil {
			return nil
		}
		if !loader.shouldVisitRef(ref, func(value any) {
			component.Value = value.(*Response)
			applyResponseRefMetadata(component.Value, component, isOpenAPI31OrLater)
			refPath, _ := loader.resolveRefPath(ref, documentPath)
			component.setRefPath(refPath)
		}) {
			return nil
		}
		loader.visitRef(ref)
		if isSingleRefElement(ref) {
			var resp Response
			if documentPath, err = loader.loadSingleElementFromURI(ref, documentPath, &resp); err != nil {
				return err
			}
			component.Value = &resp
			component.setRefPath(documentPath)
		} else {
			var resolved ResponseRef
			doc, componentPath, err := loader.resolveComponent(doc, ref, documentPath, &resolved)
			if err != nil {
				return err
			}
			if err := loader.resolveResponseRef(doc, &resolved, componentPath); err != nil {
				if err == errMUSTResponse {
					return nil
				}
				return err
			}
			component.Value = resolved.Value
			component.setRefPath(resolved.RefPath())
		}
		defer loader.unvisitRef(ref, component.Value)
		applyResponseRefMetadata(component.Value, component, isOpenAPI31OrLater)
	}
	value := component.Value
	if value == nil {
		return nil
	}

	for _, name := range componentNames(value.Headers) {
		header := value.Headers[name]
		if err := loader.resolveHeaderRef(doc, header, documentPath); err != nil {
			return err
		}
	}
	for _, name := range componentNames(value.Content) {
		if err := loader.resolveMediaTypeRefs(doc, value.Content[name], documentPath); err != nil {
			return err
		}
	}
	for _, name := range componentNames(value.Links) {
		link := value.Links[name]
		if err := loader.resolveLinkRef(doc, link, documentPath); err != nil {
			return err
		}
	}
	return nil
}

func (loader *Loader) resolveMediaTypeRefs(doc *T, mediaType *MediaType, documentPath *url.URL) (err error) {
	if mediaType == nil {
		return
	}
	if schema := mediaType.Schema; schema != nil {
		if err = loader.resolveSchemaRef(doc, schema, documentPath, loader.schemaResolution); err != nil {
			return
		}
	}
	if itemSchema := mediaType.ItemSchema; itemSchema != nil {
		if err = loader.resolveSchemaRef(doc, itemSchema, documentPath, loader.schemaResolution); err != nil {
			return
		}
	}
	for _, name := range componentNames(mediaType.Examples) {
		example := mediaType.Examples[name]
		if err = loader.resolveExampleRef(doc, example, documentPath); err != nil {
			return
		}
		mediaType.Examples[name] = example
	}
	return
}

func (loader *Loader) resolveSchemaRef(
	doc *T,
	component *SchemaRef,
	documentPath *url.URL,
	resolution *schemaResolutionContext,
) (err error) {
	if component.isEmpty() {
		return errMUSTSchema
	}

	if ref := component.Ref; ref != "" {
		if err := prepareSchemaRefSiblings(component, resolution.isOpenAPI31OrLater); err != nil {
			return err
		}
		siblingDocumentPath := documentPath
		resolvedRefPath, err := loader.resolveRefPathUnchecked(ref, documentPath)
		if err != nil {
			return err
		}
		if isSingleRefElement(ref) {
			// Publish a whole-file reference's own location before traversing the
			// loaded schema. A recursive child may point back at this SchemaRef and
			// must not replace its external identity with the child's terminal path.
			component.setRefPath(resolvedRefPath)
		}
		if component.sibling != nil && resolution.isOpenAPI31OrLater {
			// The local schema is already the effective Value. Its synthetic allOf
			// edge resolves the target through the ordinary SchemaRef path, so cycles
			// do not need a separate field-merging or SCC-materialization algorithm.
			component.setRefPath(resolvedRefPath)
			return loader.resolveSchemaChildren(doc, component.Value, siblingDocumentPath, resolution)
		}
		if component.Value != nil {
			// The reference is already resolved, so only retain its location metadata.
			// Do not apply the external-reference access policy when no read is needed.
			component.setRefPath(resolvedRefPath)
			if err := loader.resolveSchemaChildren(doc, component.sibling, siblingDocumentPath, resolution); err != nil {
				return err
			}
			return nil
		}

		// The sibling is part of the serialized source tree even in OpenAPI 3.0,
		// where it does not contribute to the effective Value. Resolve its child
		// refs so transformations such as InternalizeRefs can still reach them.
		if err := loader.resolveSchemaChildren(doc, component.sibling, siblingDocumentPath, resolution); err != nil {
			return err
		}

		target, targetRefPath, err := loader.resolveSchemaTarget(doc, ref, documentPath, resolvedRefPath, resolution)
		if err != nil {
			if err == errMUSTSchema {
				return nil
			}
			return err
		}
		component.setRefPath(targetRefPath)
		component.Value = target
		if component.siblingTarget != nil {
			component.siblingTarget.Value = target
			component.siblingTarget.setRefPath(targetRefPath)
		}
		return nil
	}
	return loader.resolveSchemaChildren(doc, component.Value, documentPath, resolution)
}

func (loader *Loader) resolveSchemaTarget(
	doc *T,
	ref string,
	documentPath *url.URL,
	resolvedRefPath *url.URL,
	resolution *schemaResolutionContext,
) (_ *Schema, _ *url.URL, err error) {
	key := schemaResolutionKey(ref, documentPath)
	if node := resolution.refs[key]; node != nil {
		if node.resolving {
			if node.value == nil {
				// A reference-only cycle with no concrete schema is equivalent to an
				// unconstrained schema. In OAS 3.0 this deliberately excludes ignored
				// siblings; OAS 3.1 sibling schemas are published before recursion and
				// therefore reach this branch only when the whole cycle has none.
				node.value = new(Schema)
				node.placeholder = true
			}
		}
		return node.value, node.refPath, nil
	}

	node := &schemaResolutionNode{resolving: true, refPath: resolvedRefPath}
	resolution.refs[key] = node
	defer func() {
		delete(resolution.refs, key)
	}()

	var target *Schema
	var resolved SchemaRef
	if isSingleRefElement(ref) {
		var schema Schema
		targetPath, loadErr := loader.loadSingleElementFromURI(ref, documentPath, &schema)
		if loadErr != nil {
			return nil, nil, loadErr
		}
		node.refPath = targetPath
		target = &schema
		node.value = target
		if err := loader.resolveSchemaChildren(doc, target, targetPath, resolution); err != nil {
			return nil, nil, err
		}
	} else {
		componentDoc, componentPath, resolveErr := loader.resolveComponent(doc, ref, documentPath, &resolved)
		if resolveErr != nil {
			return nil, nil, resolveErr
		}
		if resolved.Ref == ref && resolved.Value == nil && resolved.sibling == nil &&
			!isDirectComponentSchemaPath(resolvedRefPath) {
			// resolveComponent deliberately leaves unresolved descendant lookups
			// represented by their original ref. Do not mistake that no-progress
			// result for a reference-only SCC and turn it into an empty schema.
			node.resolving = false
			return nil, resolved.RefPath(), nil
		}
		if err := prepareSchemaRefSiblings(&resolved, resolution.isOpenAPI31OrLater); err != nil {
			return nil, nil, err
		}
		// Publish a concrete schema before descending into its children. Ordinary
		// self-references and OAS 3.1 sibling wrappers can then retain the original
		// shared pointer; placeholders are needed only for cycles with no schema.
		if resolved.Ref == "" || resolved.sibling != nil && resolution.isOpenAPI31OrLater {
			node.value = resolved.Value
		}
		if err := loader.resolveSchemaRef(componentDoc, &resolved, componentPath, resolution); err != nil {
			return nil, nil, err
		}
		target = resolved.Value
		node.refPath = resolved.RefPath()
	}

	if node.placeholder {
		if target != nil && target != node.value {
			*node.value = *target
		}
		target = node.value
	} else {
		node.value = target
	}
	node.resolving = false
	return target, node.refPath, nil
}

func isDirectComponentSchemaPath(refPath *url.URL) bool {
	if refPath == nil {
		return false
	}
	parts := strings.Split(strings.TrimPrefix(refPath.Fragment, "/"), "/")
	return len(parts) == 3 && parts[0] == "components" && parts[1] == "schemas" && parts[2] != ""
}

func prepareSchemaRefSiblings(component *SchemaRef, isOpenAPI31OrLater bool) error {
	if component == nil || component.Ref == "" {
		return nil
	}
	if isOpenAPI31OrLater && component.siblingErr != nil {
		return unmarshalError(component.siblingErr)
	}
	if component.sibling == nil {
		return nil
	}

	if component.siblingTarget == nil {
		for _, item := range component.sibling.AllOf {
			if item != nil && item.syntheticSiblingTarget {
				component.siblingTarget = item
				break
			}
		}
	}
	if component.siblingTarget == nil {
		component.siblingTarget = &SchemaRef{
			Ref:                    component.Ref,
			Origin:                 component.Origin,
			syntheticSiblingTarget: true,
		}
	}
	if !isOpenAPI31OrLater {
		return nil
	}

	// Publish the local schema before resolving its target. A recursive alias can
	// then point back to a real effective schema rather than forcing the resolver
	// to guess which sibling fields belong in a cycle placeholder.
	component.Value = component.sibling
	if slices.Contains(component.sibling.AllOf, component.siblingTarget) {
		return nil
	}
	component.sibling.AllOf = append(component.sibling.AllOf, component.siblingTarget)
	return nil
}

func (loader *Loader) resolveSchemaChildren(
	doc *T,
	value *Schema,
	documentPath *url.URL,
	resolution *schemaResolutionContext,
) error {
	if value == nil {
		return nil
	}
	if err := forEachSchemaRef(value, func(_ string, ref *SchemaRef) error {
		if ref == nil {
			return nil
		}
		return loader.resolveSchemaRef(doc, ref, documentPath, resolution)
	}); err != nil {
		return err
	}

	// Discriminator mapping refs are a special case since they are not full
	// ref objects but are plain strings that reference schema objects.
	// Plain schema names like "Dog" are not references and are left alone.
	if value.Discriminator != nil {
		inExternalDoc := documentPath != nil && documentPath.Path != "" && documentPath.Path != loader.rootLocation
		for _, k := range componentNames(value.Discriminator.Mapping) {
			v := value.Discriminator.Mapping[k]
			if !strings.Contains(v.Ref, "/") {
				continue
			}
			// A document-local mapping needs resolving only when it lives in
			// another document, because InternalizeRefs then has to rewrite it
			// against that document. One in the root document is already
			// written against the document it will end up in.
			if strings.HasPrefix(v.Ref, "#") && !inExternalDoc {
				continue
			}
			if err := loader.resolveSchemaRef(doc, (*SchemaRef)(&v), documentPath, resolution); err != nil {
				return err
			}
			value.Discriminator.Mapping[k] = v
		}
	}
	return nil
}

func applySecuritySchemeRefMetadata(value *SecurityScheme, component *SecuritySchemeRef, isOpenAPI31OrLater bool) {
	if !isOpenAPI31OrLater || value == nil || component.Description == nil {
		return
	}

	value.Description = *component.Description
}

func (loader *Loader) resolveSecuritySchemeRef(doc *T, component *SecuritySchemeRef, documentPath *url.URL) (err error) {
	isOpenAPI31OrLater := doc.IsOpenAPI31OrLater()

	if component.isEmpty() {
		return errMUSTSecurityScheme
	}

	if ref := component.Ref; ref != "" {
		if component.Value != nil {
			return nil
		}
		if !loader.shouldVisitRef(ref, func(value any) {
			component.Value = value.(*SecurityScheme)
			applySecuritySchemeRefMetadata(component.Value, component, isOpenAPI31OrLater)
			refPath, _ := loader.resolveRefPath(ref, documentPath)
			component.setRefPath(refPath)
		}) {
			return nil
		}
		loader.visitRef(ref)
		if isSingleRefElement(ref) {
			var scheme SecurityScheme
			if _, err = loader.loadSingleElementFromURI(ref, documentPath, &scheme); err != nil {
				return err
			}
			component.Value = &scheme
			component.setRefPath(documentPath)
		} else {
			var resolved SecuritySchemeRef
			doc, componentPath, err := loader.resolveComponent(doc, ref, documentPath, &resolved)
			if err != nil {
				return err
			}
			if err := loader.resolveSecuritySchemeRef(doc, &resolved, componentPath); err != nil {
				if err == errMUSTSecurityScheme {
					return nil
				}
				return err
			}
			component.Value = resolved.Value
			component.setRefPath(resolved.RefPath())
		}
		defer loader.unvisitRef(ref, component.Value)
		applySecuritySchemeRefMetadata(component.Value, component, isOpenAPI31OrLater)
	}
	return nil
}

func applyExampleRefMetadata(value *Example, component *ExampleRef, isOpenAPI31OrLater bool) {
	if !isOpenAPI31OrLater || value == nil || (component.Summary == nil && component.Description == nil) {
		return
	}

	if component.Summary != nil {
		value.Summary = *component.Summary
	}
	if component.Description != nil {
		value.Description = *component.Description
	}
}

func (loader *Loader) resolveExampleRef(doc *T, component *ExampleRef, documentPath *url.URL) (err error) {
	isOpenAPI31OrLater := doc.IsOpenAPI31OrLater()

	if component.isEmpty() {
		return errMUSTExample
	}

	if ref := component.Ref; ref != "" {
		if component.Value != nil {
			return nil
		}
		if !loader.shouldVisitRef(ref, func(value any) {
			component.Value = value.(*Example)
			applyExampleRefMetadata(component.Value, component, isOpenAPI31OrLater)
			refPath, _ := loader.resolveRefPath(ref, documentPath)
			component.setRefPath(refPath)
		}) {
			return nil
		}
		loader.visitRef(ref)
		if isSingleRefElement(ref) {
			var example Example
			if _, err = loader.loadSingleElementFromURI(ref, documentPath, &example); err != nil {
				return err
			}
			component.Value = &example
			component.setRefPath(documentPath)
		} else {
			var resolved ExampleRef
			doc, componentPath, err := loader.resolveComponent(doc, ref, documentPath, &resolved)
			if err != nil {
				return err
			}
			if err := loader.resolveExampleRef(doc, &resolved, componentPath); err != nil {
				if err == errMUSTExample {
					return nil
				}
				return err
			}
			component.Value = resolved.Value
			component.setRefPath(resolved.RefPath())
		}
		defer loader.unvisitRef(ref, component.Value)
		applyExampleRefMetadata(component.Value, component, isOpenAPI31OrLater)
	}
	return nil
}

func (loader *Loader) resolveCallbackRef(doc *T, component *CallbackRef, documentPath *url.URL) (err error) {
	if component.isEmpty() {
		return errMUSTCallback
	}

	if ref := component.Ref; ref != "" {
		if component.Value != nil {
			return nil
		}
		if !loader.shouldVisitRef(ref, func(value any) {
			component.Value = value.(*Callback)
			refPath, _ := loader.resolveRefPath(ref, documentPath)
			component.setRefPath(refPath)
		}) {
			return nil
		}
		loader.visitRef(ref)
		if isSingleRefElement(ref) {
			var resolved Callback
			if documentPath, err = loader.loadSingleElementFromURI(ref, documentPath, &resolved); err != nil {
				return err
			}
			component.Value = &resolved
			component.setRefPath(documentPath)
		} else {
			var resolved CallbackRef
			doc, componentPath, err := loader.resolveComponent(doc, ref, documentPath, &resolved)
			if err != nil {
				return err
			}
			if err = loader.resolveCallbackRef(doc, &resolved, componentPath); err != nil {
				if err == errMUSTCallback {
					return nil
				}
				return err
			}
			component.Value = resolved.Value
			component.setRefPath(resolved.RefPath())
		}
		defer loader.unvisitRef(ref, component.Value)
	}
	value := component.Value
	if value == nil {
		return nil
	}

	pathItems := value.Map()
	for _, name := range componentNames(pathItems) {
		pathItem := pathItems[name]
		if err = loader.resolvePathItemRef(doc, pathItem, documentPath); err != nil {
			return err
		}
	}
	return nil
}

func applyLinkRefMetadata(value *Link, component *LinkRef, isOpenAPI31OrLater bool) {
	if !isOpenAPI31OrLater || value == nil || component.Description == nil {
		return
	}

	value.Description = *component.Description
}

func (loader *Loader) resolveLinkRef(doc *T, component *LinkRef, documentPath *url.URL) (err error) {
	isOpenAPI31OrLater := doc.IsOpenAPI31OrLater()

	if component.isEmpty() {
		return errMUSTLink
	}

	if ref := component.Ref; ref != "" {
		if component.Value != nil {
			return nil
		}
		if !loader.shouldVisitRef(ref, func(value any) {
			component.Value = value.(*Link)
			applyLinkRefMetadata(component.Value, component, isOpenAPI31OrLater)
			refPath, _ := loader.resolveRefPath(ref, documentPath)
			component.setRefPath(refPath)
		}) {
			return nil
		}
		loader.visitRef(ref)
		if isSingleRefElement(ref) {
			var link Link
			if _, err = loader.loadSingleElementFromURI(ref, documentPath, &link); err != nil {
				return err
			}
			component.Value = &link
			component.setRefPath(documentPath)
		} else {
			var resolved LinkRef
			doc, componentPath, err := loader.resolveComponent(doc, ref, documentPath, &resolved)
			if err != nil {
				return err
			}
			if err := loader.resolveLinkRef(doc, &resolved, componentPath); err != nil {
				if err == errMUSTLink {
					return nil
				}
				return err
			}
			component.Value = resolved.Value
			component.setRefPath(resolved.RefPath())
		}
		defer loader.unvisitRef(ref, component.Value)
		applyLinkRefMetadata(component.Value, component, isOpenAPI31OrLater)
	}
	return nil
}

func (loader *Loader) resolvePathItemRef(doc *T, pathItem *PathItem, documentPath *url.URL) (err error) {
	if pathItem == nil {
		err = errMUSTPathItem
		return
	}

	if ref := pathItem.Ref; ref != "" {
		if !pathItem.isEmpty() {
			return
		}
		if !loader.shouldVisitRef(ref, func(value any) {
			*pathItem = *value.(*PathItem)
		}) {
			return nil
		}
		loader.visitRef(ref)
		if isSingleRefElement(ref) {
			var p PathItem
			if documentPath, err = loader.loadSingleElementFromURI(ref, documentPath, &p); err != nil {
				return
			}
			*pathItem = p
		} else {
			var resolved PathItem
			if doc, documentPath, err = loader.resolveComponent(doc, ref, documentPath, &resolved); err != nil {
				if err == errMUSTPathItem {
					return nil
				}
				return
			}
			*pathItem = resolved
		}
		pathItem.Ref = ref
		defer loader.unvisitRef(ref, pathItem)
	}

	for _, parameter := range pathItem.Parameters {
		if err = loader.resolveParameterRef(doc, parameter, documentPath); err != nil {
			return
		}
	}
	operations := pathItem.Operations()
	for _, name := range componentNames(operations) {
		operation := operations[name]
		for _, parameter := range operation.Parameters {
			if err = loader.resolveParameterRef(doc, parameter, documentPath); err != nil {
				return
			}
		}
		if requestBody := operation.RequestBody; requestBody != nil {
			if err = loader.resolveRequestBodyRef(doc, requestBody, documentPath); err != nil {
				return
			}
		}
		responses := operation.Responses.Map()
		for _, name := range componentNames(responses) {
			response := responses[name]
			if err = loader.resolveResponseRef(doc, response, documentPath); err != nil {
				return
			}
		}
		for _, name := range componentNames(operation.Callbacks) {
			callback := operation.Callbacks[name]
			if err = loader.resolveCallbackRef(doc, callback, documentPath); err != nil {
				return
			}
		}
	}
	return
}

func unescapeRefString(ref string) string {
	return strings.ReplaceAll(strings.ReplaceAll(ref, "~1", "/"), "~0", "~")
}
