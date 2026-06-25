package openapi3

import (
	"errors"
	"strconv"
	"strings"
)

// SkipSubtree, when returned by a WalkSchemasFunc, tells WalkSchemas not to
// descend into the current schema's sub-schemas. The schema itself has already
// been visited. Any other non-nil error stops the walk and is returned by
// WalkSchemas. It mirrors filepath.SkipDir.
var SkipSubtree = errors.New("skip schema subtree")

// WalkSchemasFunc is called once for each schema visited by WalkSchemas.
//
// location is the JSON Pointer (RFC 6901) of the schema within the document,
// e.g. "/components/schemas/Pet/properties/tag" or
// "/paths/~1pets/get/responses/200/content/application~1json/schema". sr is
// never nil and sr.Value is never nil; the callback may modify sr.Value in
// place, so WalkSchemas serves transformers and not only read-only inspection.
//
// Returning SkipSubtree skips this schema's sub-schemas; returning any other
// error aborts the walk and is returned by WalkSchemas.
type WalkSchemasFunc func(location string, schema *SchemaRef) error

// WalkSchemas visits every schema reachable from the document exactly once, in
// document order, invoking fn for each. It follows resolved $ref targets
// (schema.Value) and guards against reference cycles, so each distinct *Schema
// is visited a single time regardless of how many references point at it.
//
// It covers schemas under components (schemas, parameters, headers, request
// bodies, responses, callbacks), the paths and their operations (parameters,
// request bodies, responses, headers, callbacks), and webhooks, then recurses
// through every sub-schema keyword: properties, items, allOf/anyOf/oneOf, not,
// additionalProperties, prefixItems, contains, patternProperties,
// dependentSchemas, propertyNames, if/then/else, and $defs.
//
// It is useful for validation, code generation, schema transformation,
// $ref/dependency analysis, and documentation: any consumer that needs to act
// on every schema in a document without re-deriving the (easy to get wrong)
// traversal itself.
func (doc *T) WalkSchemas(fn WalkSchemasFunc) error {
	if doc == nil {
		return nil
	}
	w := schemaWalker{fn: fn, seen: make(map[*Schema]struct{})}
	return w.document(doc)
}

type schemaWalker struct {
	fn   WalkSchemasFunc
	seen map[*Schema]struct{}
}

// escapeJSONPointerToken escapes a single JSON Pointer reference token per
// RFC 6901: '~' becomes '~0' and '/' becomes '~1'.
func escapeJSONPointerToken(s string) string {
	if !strings.ContainsAny(s, "~/") {
		return s
	}
	return strings.NewReplacer("~", "~0", "/", "~1").Replace(s)
}

func (w *schemaWalker) document(doc *T) error {
	if c := doc.Components; c != nil {
		for name, sr := range c.Schemas {
			if err := w.schemaRef("/components/schemas/"+escapeJSONPointerToken(name), sr); err != nil {
				return err
			}
		}
		for name, pr := range c.Parameters {
			if err := w.parameter("/components/parameters/"+escapeJSONPointerToken(name), pr); err != nil {
				return err
			}
		}
		for name, hr := range c.Headers {
			if err := w.header("/components/headers/"+escapeJSONPointerToken(name), hr); err != nil {
				return err
			}
		}
		for name, rbr := range c.RequestBodies {
			if rbr != nil && rbr.Value != nil {
				if err := w.content("/components/requestBodies/"+escapeJSONPointerToken(name)+"/content", rbr.Value.Content); err != nil {
					return err
				}
			}
		}
		for name, rr := range c.Responses {
			if err := w.response("/components/responses/"+escapeJSONPointerToken(name), rr); err != nil {
				return err
			}
		}
		for name, cbr := range c.Callbacks {
			if cbr != nil && cbr.Value != nil {
				if err := w.callback("/components/callbacks/"+escapeJSONPointerToken(name), cbr.Value); err != nil {
					return err
				}
			}
		}
	}
	if doc.Paths != nil {
		for path, item := range doc.Paths.Map() {
			if err := w.pathItem("/paths/"+escapeJSONPointerToken(path), item); err != nil {
				return err
			}
		}
	}
	for name, item := range doc.Webhooks {
		if err := w.pathItem("/webhooks/"+escapeJSONPointerToken(name), item); err != nil {
			return err
		}
	}
	return nil
}

func (w *schemaWalker) pathItem(loc string, item *PathItem) error {
	if item == nil {
		return nil
	}
	for i, pr := range item.Parameters {
		if err := w.parameter(loc+"/parameters/"+strconv.Itoa(i), pr); err != nil {
			return err
		}
	}
	for method, op := range item.Operations() {
		if err := w.operation(loc+"/"+strings.ToLower(method), op); err != nil {
			return err
		}
	}
	return nil
}

func (w *schemaWalker) operation(loc string, op *Operation) error {
	if op == nil {
		return nil
	}
	for i, pr := range op.Parameters {
		if err := w.parameter(loc+"/parameters/"+strconv.Itoa(i), pr); err != nil {
			return err
		}
	}
	if op.RequestBody != nil && op.RequestBody.Value != nil {
		if err := w.content(loc+"/requestBody/content", op.RequestBody.Value.Content); err != nil {
			return err
		}
	}
	if op.Responses != nil {
		for code, rr := range op.Responses.Map() {
			if err := w.response(loc+"/responses/"+escapeJSONPointerToken(code), rr); err != nil {
				return err
			}
		}
	}
	for name, cbr := range op.Callbacks {
		if cbr != nil && cbr.Value != nil {
			if err := w.callback(loc+"/callbacks/"+escapeJSONPointerToken(name), cbr.Value); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *schemaWalker) callback(loc string, cb *Callback) error {
	for expr, item := range cb.Map() {
		if err := w.pathItem(loc+"/"+escapeJSONPointerToken(expr), item); err != nil {
			return err
		}
	}
	return nil
}

func (w *schemaWalker) parameter(loc string, pr *ParameterRef) error {
	if pr == nil || pr.Value == nil {
		return nil
	}
	if err := w.schemaRef(loc+"/schema", pr.Value.Schema); err != nil {
		return err
	}
	return w.content(loc+"/content", pr.Value.Content)
}

func (w *schemaWalker) header(loc string, hr *HeaderRef) error {
	if hr == nil || hr.Value == nil {
		return nil
	}
	if err := w.schemaRef(loc+"/schema", hr.Value.Schema); err != nil {
		return err
	}
	return w.content(loc+"/content", hr.Value.Content)
}

func (w *schemaWalker) response(loc string, rr *ResponseRef) error {
	if rr == nil || rr.Value == nil {
		return nil
	}
	for name, hr := range rr.Value.Headers {
		if err := w.header(loc+"/headers/"+escapeJSONPointerToken(name), hr); err != nil {
			return err
		}
	}
	return w.content(loc+"/content", rr.Value.Content)
}

func (w *schemaWalker) content(loc string, content Content) error {
	for mediaType, media := range content {
		if media == nil {
			continue
		}
		if err := w.schemaRef(loc+"/"+escapeJSONPointerToken(mediaType)+"/schema", media.Schema); err != nil {
			return err
		}
	}
	return nil
}

func (w *schemaWalker) schemaRefs(loc string, refs SchemaRefs) error {
	for i, sub := range refs {
		if err := w.schemaRef(loc+"/"+strconv.Itoa(i), sub); err != nil {
			return err
		}
	}
	return nil
}

func (w *schemaWalker) schemaRef(loc string, sr *SchemaRef) error {
	if sr == nil || sr.Value == nil {
		return nil
	}
	s := sr.Value
	if _, ok := w.seen[s]; ok {
		// Already visited (shared $ref target or reference cycle).
		return nil
	}
	w.seen[s] = struct{}{}

	if err := w.fn(loc, sr); err != nil {
		if errors.Is(err, SkipSubtree) {
			return nil
		}
		return err
	}

	for name, sub := range s.Properties {
		if err := w.schemaRef(loc+"/properties/"+escapeJSONPointerToken(name), sub); err != nil {
			return err
		}
	}
	if err := w.schemaRef(loc+"/items", s.Items); err != nil {
		return err
	}
	if s.AdditionalProperties.Schema != nil {
		if err := w.schemaRef(loc+"/additionalProperties", s.AdditionalProperties.Schema); err != nil {
			return err
		}
	}
	if err := w.schemaRefs(loc+"/allOf", s.AllOf); err != nil {
		return err
	}
	if err := w.schemaRefs(loc+"/anyOf", s.AnyOf); err != nil {
		return err
	}
	if err := w.schemaRefs(loc+"/oneOf", s.OneOf); err != nil {
		return err
	}
	if err := w.schemaRef(loc+"/not", s.Not); err != nil {
		return err
	}
	if err := w.schemaRefs(loc+"/prefixItems", s.PrefixItems); err != nil {
		return err
	}
	if err := w.schemaRef(loc+"/contains", s.Contains); err != nil {
		return err
	}
	for name, sub := range s.PatternProperties {
		if err := w.schemaRef(loc+"/patternProperties/"+escapeJSONPointerToken(name), sub); err != nil {
			return err
		}
	}
	for name, sub := range s.DependentSchemas {
		if err := w.schemaRef(loc+"/dependentSchemas/"+escapeJSONPointerToken(name), sub); err != nil {
			return err
		}
	}
	if err := w.schemaRef(loc+"/propertyNames", s.PropertyNames); err != nil {
		return err
	}
	if err := w.schemaRef(loc+"/if", s.If); err != nil {
		return err
	}
	if err := w.schemaRef(loc+"/then", s.Then); err != nil {
		return err
	}
	if err := w.schemaRef(loc+"/else", s.Else); err != nil {
		return err
	}
	for name, sub := range s.Defs {
		if err := w.schemaRef(loc+"/$defs/"+escapeJSONPointerToken(name), sub); err != nil {
			return err
		}
	}
	return nil
}
