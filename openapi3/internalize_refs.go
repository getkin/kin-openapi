package openapi3

import (
	"context"
	"strconv"
	"strings"
)

func schemaNames(s Schemas) []string {
	out := make([]string, 0, len(s))
	for i := range s {
		out = append(out, i)
	}
	return out
}

func parametersMapNames(s ParametersMap) []string {
	out := make([]string, 0, len(s))
	for i := range s {
		out = append(out, i)
	}
	return out
}

// InternalizeRefs removes all references to external files from the spec and moves them
// to the components section.
//
// refNameResolver takes in references to returns a name to store the reference under locally.
// Some care should be taken to make sure the resolver does not return duplicate names for different
// references.
func (spec *T) InternalizeRefs(ctx context.Context, refNameResolver func(ref string) string) *T {
	isExternalRef := func(ref string) bool {
		return ref != "" && !strings.HasPrefix(ref, "#/components/")
	}

	addSchemaToSpec := func(s *SchemaRef) string {
		if s == nil || !isExternalRef(s.Ref) {
			return ""
		}

		name := refNameResolver(s.Ref)

		// I just added a more comprehensive check for this being a schema but people should really
		// make sure their resolver returns unique names
		val := 1
		for {
			// basic check, may need something more exhaustive...
			if existing, ok := spec.Components.Schemas[name]; ok {
				if existing.Value.Description != s.Value.Description || s.Value.Type != existing.Value.Type {
					name = refNameResolver(s.Ref) + strconv.Itoa(val)
					val++
					continue
				}
				newRef := "#/components/schemas/" + name
				if newRef == s.Ref {
					s.Ref = ""
				} else {
					s.Ref = newRef
				}
				return s.Ref
			}
			break
		}
		if spec.Components.Schemas == nil {
			spec.Components.Schemas = make(Schemas)
		}
		spec.Components.Schemas[name] = s.Value.NewRef()
		s.Ref = "#/components/schemas/" + name
		return s.Ref
	}

	addParameterToSpec := func(p *ParameterRef) string {
		if p == nil || !isExternalRef(p.Ref) {
			return ""
		}
		name := refNameResolver(p.Ref)
		if _, ok := spec.Components.Parameters[name]; ok {
			p.Ref = "#/components/parameters/" + name
			return p.Ref
		}

		if spec.Components.Parameters == nil {
			spec.Components.Parameters = make(ParametersMap)
		}
		spec.Components.Parameters[name] = &ParameterRef{Value: p.Value}
		p.Ref = "#/components/parameters/" + name
		return p.Ref
	}

	addHeaderToSpec := func(h *HeaderRef) string {
		if h == nil || !isExternalRef(h.Ref) {
			return ""
		}
		name := refNameResolver(h.Ref)
		if _, ok := spec.Components.Headers[name]; ok {
			h.Ref = "#/components/headers/" + name
			return h.Ref
		}
		if spec.Components.Headers == nil {
			spec.Components.Headers = make(Headers)
		}
		spec.Components.Headers[name] = &HeaderRef{Value: h.Value}
		h.Ref = "#/components/headers/" + name
		return h.Ref
	}

	addRequestBodyToSpec := func(r *RequestBodyRef) string {
		if r == nil || !isExternalRef(r.Ref) {
			return ""
		}
		name := refNameResolver(r.Ref)
		if _, ok := spec.Components.RequestBodies[name]; ok {
			r.Ref = "#/components/requestBodies/" + name
			return r.Ref
		}
		if spec.Components.RequestBodies == nil {
			spec.Components.RequestBodies = make(RequestBodies)
		}
		spec.Components.RequestBodies[name] = &RequestBodyRef{Value: r.Value}
		r.Ref = "#/components/requestBodies/" + name
		return r.Ref
	}

	addResponseToSpec := func(r *ResponseRef) string {
		if r == nil || !isExternalRef(r.Ref) {
			return ""
		}
		name := refNameResolver(r.Ref)
		if _, ok := spec.Components.Responses[name]; ok {
			r.Ref = "#/components/responses/" + name
			return r.Ref
		}
		if spec.Components.Responses == nil {
			spec.Components.Responses = make(Responses)
		}
		spec.Components.Responses[name] = &ResponseRef{Value: r.Value}
		r.Ref = "#/components/responses/" + name
		return r.Ref
	}

	addSecuritySchemeToSpec := func(ss *SecuritySchemeRef) string {
		if ss == nil || !isExternalRef(ss.Ref) {
			return ""
		}
		name := refNameResolver(ss.Ref)
		if _, ok := spec.Components.SecuritySchemes[name]; ok {
			ss.Ref = "#/components/securitySchemes/" + name
			return ss.Ref
		}
		if spec.Components.SecuritySchemes == nil {
			spec.Components.SecuritySchemes = make(SecuritySchemes)
		}
		spec.Components.SecuritySchemes[name] = &SecuritySchemeRef{Value: ss.Value}
		ss.Ref = "#/components/securitySchemes/" + name
		return ss.Ref
	}

	addExampleToSpec := func(e *ExampleRef) string {
		if e == nil || !isExternalRef(e.Ref) {
			return ""
		}
		name := refNameResolver(e.Ref)
		if _, ok := spec.Components.Examples[name]; ok {
			e.Ref = "#/components/examples/" + name
			return e.Ref
		}
		if spec.Components.Examples == nil {
			spec.Components.Examples = make(Examples)
		}
		spec.Components.Examples[name] = &ExampleRef{Value: e.Value}
		e.Ref = "#/components/examples/" + name
		return e.Ref
	}

	addLinkToSpec := func(l *LinkRef) string {
		if l == nil || !isExternalRef(l.Ref) {
			return ""
		}
		name := refNameResolver(l.Ref)
		if _, ok := spec.Components.Links[name]; ok {
			l.Ref = "#/components/links/" + name
			return l.Ref
		}
		if spec.Components.Links == nil {
			spec.Components.Links = make(Links)
		}
		spec.Components.Links[name] = &LinkRef{Value: l.Value}
		l.Ref = "#/components/links/" + name
		return l.Ref
	}

	addCallbackToSpec := func(c *CallbackRef) string {
		if c == nil || !isExternalRef(c.Ref) {
			return ""
		}
		name := refNameResolver(c.Ref)
		if _, ok := spec.Components.Callbacks[name]; ok {
			c.Ref = "#/components/callbacks/" + name
			return c.Ref
		}
		if spec.Components.Callbacks == nil {
			spec.Components.Callbacks = make(Callbacks)
		}
		spec.Components.Callbacks[name] = &CallbackRef{Value: c.Value}
		c.Ref = "#/components/callbacks/" + name
		return c.Ref
	}

	var derefSchema func(*Schema, []*Schema)
	derefSchema = func(s *Schema, stack []*Schema) {
		if len(stack) > 1000 {
			panic("unresolved circular reference")
		}
		if s == nil {
			return
		}
		for _, p := range stack {
			if p == s {
				return
			}
		}
		stack = append(stack, s)

		for _, list := range []SchemaRefs{s.AllOf, s.AnyOf, s.OneOf} {
			for _, s2 := range list {
				addSchemaToSpec(s2)
				if s2 != nil {
					derefSchema(s2.Value, stack)
				}
			}
		}
		for _, s2 := range s.Properties {
			addSchemaToSpec(s2)
			if s2 != nil {
				derefSchema(s2.Value, stack)
			}
		}
		for _, ref := range []*SchemaRef{s.Not, s.AdditionalProperties, s.Items} {
			addSchemaToSpec(ref)
			if ref != nil {
				derefSchema(ref.Value, stack)
			}
		}
	}
	var derefParameter func(Parameter)

	derefHeaders := func(hs Headers) {
		for _, h := range hs {
			addHeaderToSpec(h)
			derefParameter(h.Value.Parameter)
		}
	}

	derefExamples := func(es Examples) {
		for _, e := range es {
			addExampleToSpec(e)
		}
	}

	derefContent := func(c Content) {
		for _, mediatype := range c {
			addSchemaToSpec(mediatype.Schema)
			if mediatype.Schema != nil {
				derefSchema(mediatype.Schema.Value, nil)
			}
			derefExamples(mediatype.Examples)
			for _, e := range mediatype.Encoding {
				derefHeaders(e.Headers)
			}
		}
	}

	derefLinks := func(ls Links) {
		for _, l := range ls {
			addLinkToSpec(l)
		}
	}

	derefResponses := func(es Responses) {
		for _, e := range es {
			addResponseToSpec(e)
			if e.Value != nil {
				derefHeaders(e.Value.Headers)
				derefContent(e.Value.Content)
				derefLinks(e.Value.Links)
			}
		}
	}

	derefParameter = func(p Parameter) {
		addSchemaToSpec(p.Schema)
		derefContent(p.Content)
		if p.Schema != nil {
			derefSchema(p.Schema.Value, nil)
		}
	}

	derefRequestBody := func(r RequestBody) {
		derefContent(r.Content)
	}

	var derefPaths func(map[string]*PathItem)
	derefPaths = func(paths map[string]*PathItem) {
		for _, ops := range paths {
			// inline full operations
			ops.Ref = ""

			for _, op := range ops.Operations() {
				addRequestBodyToSpec(op.RequestBody)
				if op.RequestBody != nil && op.RequestBody.Value != nil {
					derefRequestBody(*op.RequestBody.Value)
				}
				for _, cb := range op.Callbacks {
					addCallbackToSpec(cb)
					if cb.Value != nil {
						derefPaths(*cb.Value)
					}
				}
				derefResponses(op.Responses)
				for _, param := range op.Parameters {
					addParameterToSpec(param)
					if param.Value != nil {
						derefParameter(*param.Value)
					}
				}
			}
		}
	}

	// Handle components section
	names := schemaNames(spec.Components.Schemas)
	for _, name := range names {
		schema := spec.Components.Schemas[name]
		addSchemaToSpec(schema)
		if schema != nil {
			schema.Ref = "" // always dereference the top level
			derefSchema(schema.Value, nil)
		}
	}
	names = parametersMapNames(spec.Components.Parameters)
	for _, name := range names {
		p := spec.Components.Parameters[name]
		addParameterToSpec(p)
		if p != nil && p.Value != nil {
			p.Ref = "" // always dereference the top level
			derefParameter(*p.Value)
		}
	}
	derefHeaders(spec.Components.Headers)
	for _, req := range spec.Components.RequestBodies {
		addRequestBodyToSpec(req)
		if req != nil && req.Value != nil {
			req.Ref = "" // always dereference the top level
			derefRequestBody(*req.Value)
		}
	}
	derefResponses(spec.Components.Responses)
	for _, ss := range spec.Components.SecuritySchemes {
		addSecuritySchemeToSpec(ss)
	}
	derefExamples(spec.Components.Examples)
	derefLinks(spec.Components.Links)
	for _, cb := range spec.Components.Callbacks {
		addCallbackToSpec(cb)
		if cb != nil && cb.Value != nil {
			cb.Ref = "" // always dereference the top level
			derefPaths(*cb.Value)
		}
	}

	derefPaths(spec.Paths)
	return spec
}
