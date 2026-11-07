package openapi3

import (
	"context"
	"path"
	"strings"
)

// RefNameResolver maps a component to a name that is used as it's internalized name.
//
// The function should avoid name collisions (i.e. be a injective mapping).
// It must only contain characters valid for fixed field names: [IdentifierRegExp].
type RefNameResolver func(*T, ComponentRef) string

// DefaultRefResolver is a default implementation of refNameResolver for the
// InternalizeRefs function.
//
// The external reference is internalized to (hopefully) a unique name. If
// the external reference matches (by path) to another reference in the root
// document then the name of that component is used.
//
// The transformation involves:
//   - Cutting the "#/components/<type>" part.
//   - Cutting the file extensions (.yaml/.json) from documents.
//   - Trimming the common directory with the root spec.
//   - Replace invalid characters with underscores.
//
// This is an injective mapping over a "reasonable" amount of the possible openapi
// spec domain space but is not perfect. There might be edge cases.
func DefaultRefNameResolver(doc *T, ref ComponentRef) string {
	if ref.RefString() == "" || ref.RefPath() == nil {
		panic("unable to resolve reference to name")
	}

	name := ref.RefPath()

	// If referring to a component in the root spec, no need to internalize just use
	// the existing component.
	// XXX(percivalalb): since this function call is iterating over components behind the
	// scenes during an internalization call it actually starts interating over
	// new & replaced internalized components. This might caused some edge cases,
	// haven't found one yet but this might need to actually be used on a frozen copy
	// of doc.
	if nameInRoot, found := ReferencesComponentInRootDocument(doc, ref); found {
		nameInRoot = strings.TrimPrefix(nameInRoot, "#")

		rootCompURI := copyURI(doc.url)
		rootCompURI.Fragment = nameInRoot
		name = rootCompURI
	}

	filePath, componentPath := name.Path, name.Fragment

	// Cut out the "#/components/<type>" to make the names shorter.
	// XXX(percivalalb): This might cause collisions but is worth the brevity.
	if b, a, ok := strings.Cut(componentPath, path.Join("components", ref.CollectionName(), "")); ok {
		componentPath = path.Join(b, a)
	}

	if filePath != "" {
		// If the path is the same as the root doc, just remove.
		if doc.url != nil && filePath == doc.url.Path {
			filePath = ""
		}

		// Remove the path extension to make this JSON/YAML agnostic. Only
		// the last one is stripped: earlier dots belong to the file name
		// itself (e.g. "Lib.AccountId.yaml" must stay "Lib.AccountId"),
		// otherwise two files like "Lib.AccountId.yaml" and
		// "Lib.PageSize.yaml" would collapse to the same name.
		if ext := path.Ext(filePath); len(ext) > 0 {
			filePath = strings.TrimSuffix(filePath, ext)
		}

		// Trim the common prefix with the root doc path.
		if doc.url != nil && filePath != "" {
			// path.Dir of an absolute path ends at "/", never at ".".
			for commonDir := path.Dir(doc.url.Path); /*no common prefix*/ commonDir != "." && commonDir != "/"; commonDir = path.Dir(commonDir) {
				if p, found := cutDirectories(filePath, commonDir); found {
					filePath = p
					break
				}
			}
		}
	}

	var internalizedName string

	// Trim .'s & slashes from start e.g. otherwise ./doc.yaml would end up as __doc
	if filePath != "" {
		internalizedName = strings.TrimLeft(filePath, "./")
	}

	if componentPath != "" {
		if internalizedName != "" {
			internalizedName += "_"
		}

		internalizedName += strings.TrimLeft(componentPath, "./")
	}

	// Replace invalid characters in component fixed field names.
	internalizedName = InvalidIdentifierCharRegExp.ReplaceAllString(internalizedName, "_")

	// Nothing is left of a ref to the directory itself, such as "/" or "./".
	if internalizedName == "" {
		internalizedName = ref.CollectionName()
	}

	return internalizedName
}

// cutDirectories removes the given directories from the start of the path if
// the path is a child.
func cutDirectories(p, dirs string) (string, bool) {
	if dirs == "" || p == "" {
		return p, false
	}

	p = strings.TrimRight(p, "/")
	dirs = strings.TrimRight(dirs, "/")

	var sb strings.Builder
	sb.Grow(len(ParameterInHeader))
	for segments := range strings.SplitSeq(p, "/") {
		sb.WriteString(segments)

		if sb.String() == p {
			return strings.TrimPrefix(p, dirs), true
		}

		sb.WriteRune('/')
	}

	return p, false
}

// isExternalRef reports whether ref has to be moved into the root document.
// The add*ToSpec functions also leave alone any ref without a RefPath: the
// loader did not follow it (its target was empty), so there is nothing to
// move and the resolver could not name it.
func isExternalRef(ref string, parentIsExternal bool) bool {
	return ref != "" && (!strings.HasPrefix(ref, "#/components/") || parentIsExternal)
}

func (doc *T) addSchemaToSpec(s *SchemaRef, refNameResolver RefNameResolver, parentIsExternal bool) bool {
	if s == nil || !isExternalRef(s.Ref, parentIsExternal) || s.RefPath() == nil {
		return false
	}

	name := refNameResolver(doc, s)
	if doc.Components != nil {
		if _, ok := doc.Components.Schemas[name]; ok {
			s.Ref = "#/components/schemas/" + name
			return true
		}
	}

	if doc.Components == nil {
		doc.Components = &Components{}
	}
	if doc.Components.Schemas == nil {
		doc.Components.Schemas = make(Schemas)
	}
	// A $ref with siblings has a copy of its target as Value, with the siblings
	// applied: the component is the target itself.
	value := s.Value
	if s.target != nil {
		value = s.target
	}
	doc.Components.Schemas[name] = value.NewRef()
	s.Ref = "#/components/schemas/" + name
	return true
}

func (doc *T) addParameterToSpec(p *ParameterRef, refNameResolver RefNameResolver, parentIsExternal bool) bool {
	if p == nil || !isExternalRef(p.Ref, parentIsExternal) || p.RefPath() == nil {
		return false
	}
	name := refNameResolver(doc, p)
	if doc.Components != nil {
		if _, ok := doc.Components.Parameters[name]; ok {
			p.Ref = "#/components/parameters/" + name
			return true
		}
	}

	if doc.Components == nil {
		doc.Components = &Components{}
	}
	if doc.Components.Parameters == nil {
		doc.Components.Parameters = make(ParametersMap)
	}
	doc.Components.Parameters[name] = &ParameterRef{Value: p.Value}
	p.Ref = "#/components/parameters/" + name
	return true
}

func (doc *T) addHeaderToSpec(h *HeaderRef, refNameResolver RefNameResolver, parentIsExternal bool) bool {
	if h == nil || !isExternalRef(h.Ref, parentIsExternal) || h.RefPath() == nil {
		return false
	}
	name := refNameResolver(doc, h)
	if doc.Components != nil {
		if _, ok := doc.Components.Headers[name]; ok {
			h.Ref = "#/components/headers/" + name
			return true
		}
	}

	if doc.Components == nil {
		doc.Components = &Components{}
	}
	if doc.Components.Headers == nil {
		doc.Components.Headers = make(Headers)
	}
	doc.Components.Headers[name] = &HeaderRef{Value: h.Value}
	h.Ref = "#/components/headers/" + name
	return true
}

func (doc *T) addRequestBodyToSpec(r *RequestBodyRef, refNameResolver RefNameResolver, parentIsExternal bool) bool {
	if r == nil || !isExternalRef(r.Ref, parentIsExternal) || r.RefPath() == nil {
		return false
	}
	name := refNameResolver(doc, r)
	if doc.Components != nil {
		if _, ok := doc.Components.RequestBodies[name]; ok {
			r.Ref = "#/components/requestBodies/" + name
			return true
		}
	}

	if doc.Components == nil {
		doc.Components = &Components{}
	}
	if doc.Components.RequestBodies == nil {
		doc.Components.RequestBodies = make(RequestBodies)
	}
	doc.Components.RequestBodies[name] = &RequestBodyRef{Value: r.Value}
	r.Ref = "#/components/requestBodies/" + name
	return true
}

func (doc *T) addResponseToSpec(r *ResponseRef, refNameResolver RefNameResolver, parentIsExternal bool) bool {
	if r == nil || !isExternalRef(r.Ref, parentIsExternal) || r.RefPath() == nil {
		return false
	}
	name := refNameResolver(doc, r)
	if doc.Components != nil {
		if _, ok := doc.Components.Responses[name]; ok {
			r.Ref = "#/components/responses/" + name
			return true
		}
	}

	if doc.Components == nil {
		doc.Components = &Components{}
	}
	if doc.Components.Responses == nil {
		doc.Components.Responses = make(ResponseBodies)
	}
	doc.Components.Responses[name] = &ResponseRef{Value: r.Value}
	r.Ref = "#/components/responses/" + name
	return true
}

func (doc *T) addSecuritySchemeToSpec(ss *SecuritySchemeRef, refNameResolver RefNameResolver, parentIsExternal bool) {
	if ss == nil || !isExternalRef(ss.Ref, parentIsExternal) || ss.RefPath() == nil {
		return
	}
	name := refNameResolver(doc, ss)
	if doc.Components != nil {
		if _, ok := doc.Components.SecuritySchemes[name]; ok {
			ss.Ref = "#/components/securitySchemes/" + name
			return
		}
	}

	if doc.Components == nil {
		doc.Components = &Components{}
	}
	if doc.Components.SecuritySchemes == nil {
		doc.Components.SecuritySchemes = make(SecuritySchemes)
	}
	doc.Components.SecuritySchemes[name] = &SecuritySchemeRef{Value: ss.Value}
	ss.Ref = "#/components/securitySchemes/" + name

}

func (doc *T) addExampleToSpec(e *ExampleRef, refNameResolver RefNameResolver, parentIsExternal bool) {
	if e == nil || !isExternalRef(e.Ref, parentIsExternal) || e.RefPath() == nil {
		return
	}
	name := refNameResolver(doc, e)
	if doc.Components != nil {
		if _, ok := doc.Components.Examples[name]; ok {
			e.Ref = "#/components/examples/" + name
			return
		}
	}

	if doc.Components == nil {
		doc.Components = &Components{}
	}
	if doc.Components.Examples == nil {
		doc.Components.Examples = make(Examples)
	}
	doc.Components.Examples[name] = &ExampleRef{Value: e.Value}
	e.Ref = "#/components/examples/" + name

}

func (doc *T) addLinkToSpec(l *LinkRef, refNameResolver RefNameResolver, parentIsExternal bool) {
	if l == nil || !isExternalRef(l.Ref, parentIsExternal) || l.RefPath() == nil {
		return
	}
	name := refNameResolver(doc, l)
	if doc.Components != nil {
		if _, ok := doc.Components.Links[name]; ok {
			l.Ref = "#/components/links/" + name
			return
		}
	}

	if doc.Components == nil {
		doc.Components = &Components{}
	}
	if doc.Components.Links == nil {
		doc.Components.Links = make(Links)
	}
	doc.Components.Links[name] = &LinkRef{Value: l.Value}
	l.Ref = "#/components/links/" + name

}

func (doc *T) addCallbackToSpec(c *CallbackRef, refNameResolver RefNameResolver, parentIsExternal bool) bool {
	if c == nil || !isExternalRef(c.Ref, parentIsExternal) || c.RefPath() == nil {
		return false
	}
	name := refNameResolver(doc, c)

	if doc.Components == nil {
		doc.Components = &Components{}
	}
	if doc.Components.Callbacks == nil {
		doc.Components.Callbacks = make(Callbacks)
	}
	c.Ref = "#/components/callbacks/" + name
	doc.Components.Callbacks[name] = &CallbackRef{Value: c.Value}
	return true
}

// derefSchemaRef moves s into the root document when it points elsewhere,
// then does the same for the refs within the schema it holds.
func (doc *T) derefSchemaRef(s *SchemaRef, refNameResolver RefNameResolver, parentIsExternal bool) (isExternal bool) {
	isExternal = doc.addSchemaToSpec(s, refNameResolver, parentIsExternal)
	if s == nil {
		return
	}
	if s.target == nil {
		doc.derefSchema(s.Value, refNameResolver, isExternal || parentIsExternal)
		return
	}
	// A $ref with siblings has a copy of its target as Value, with the siblings
	// applied. Each schema in it is either the target's, which lives where the
	// $ref points, or the siblings', which live next to the $ref, in the
	// document of its parent. The siblings are what gets written out next to
	// the $ref, even those Value left out, such as a conjunct that constrains
	// nothing.
	doc.derefSchema(s.target, refNameResolver, isExternal || parentIsExternal)
	doc.derefSchema(s.sibling, refNameResolver, parentIsExternal)
	return
}

func (doc *T) derefSchema(s *Schema, refNameResolver RefNameResolver, parentIsExternal bool) {
	if s == nil || doc.isVisitedSchema(s) {
		return
	}

	for _, list := range []SchemaRefs{s.AllOf, s.AnyOf, s.OneOf, s.PrefixItems} {
		for _, s2 := range list {
			doc.derefSchemaRef(s2, refNameResolver, parentIsExternal)
		}
	}

	// Discriminator mapping values are special cases since they are not full
	// ref objects but are string references to schema objects.
	if s.Discriminator != nil {
		for _, k := range componentNames(s.Discriminator.Mapping) {
			mapRef := s.Discriminator.Mapping[k]
			s2 := (*SchemaRef)(&mapRef)
			// The mapping value may be a plain schema name rather than a
			// reference, and a reference the loader could not follow was left
			// as written. Neither has a resolved path, so neither can be named
			// by the resolver: leave them exactly as they are.
			if s2.RefPath() == nil {
				continue
			}
			doc.derefSchemaRef(s2, refNameResolver, parentIsExternal)
			s.Discriminator.Mapping[k] = MappingRef(*s2)
		}
	}

	for _, schemas := range []Schemas{s.Properties, s.PatternProperties, s.DependentSchemas, s.Defs} {
		for _, name := range componentNames(schemas) {
			doc.derefSchemaRef(schemas[name], refNameResolver, parentIsExternal)
		}
	}
	for _, ref := range []*SchemaRef{
		s.Not, s.AdditionalProperties.Schema, s.Items,
		s.Contains, s.PropertyNames, s.If, s.Then, s.Else, s.ContentSchema,
		s.UnevaluatedItems.Schema, s.UnevaluatedProperties.Schema,
	} {
		doc.derefSchemaRef(ref, refNameResolver, parentIsExternal)
	}
}

func (doc *T) derefHeaders(hs Headers, refNameResolver RefNameResolver, parentIsExternal bool) {
	for _, name := range componentNames(hs) {
		h := hs[name]
		isExternal := doc.addHeaderToSpec(h, refNameResolver, parentIsExternal)
		if h == nil || h.Value == nil || doc.isVisitedHeader(h.Value) {
			continue
		}
		doc.derefParameter(h.Value.Parameter, refNameResolver, parentIsExternal || isExternal)
	}
}

func (doc *T) derefExamples(es Examples, refNameResolver RefNameResolver, parentIsExternal bool) {
	for _, name := range componentNames(es) {
		e := es[name]
		doc.addExampleToSpec(e, refNameResolver, parentIsExternal)
	}
}

func (doc *T) derefContent(c Content, refNameResolver RefNameResolver, parentIsExternal bool) {
	for _, name := range componentNames(c) {
		mediatype := c[name]
		if mediatype == nil {
			continue
		}
		doc.derefSchemaRef(mediatype.Schema, refNameResolver, parentIsExternal)
		doc.derefExamples(mediatype.Examples, refNameResolver, parentIsExternal)
		for _, name := range componentNames(mediatype.Encoding) {
			e := mediatype.Encoding[name]
			if e == nil {
				continue
			}
			doc.derefHeaders(e.Headers, refNameResolver, parentIsExternal)
		}
	}
}

func (doc *T) derefLinks(ls Links, refNameResolver RefNameResolver, parentIsExternal bool) {
	for _, name := range componentNames(ls) {
		l := ls[name]
		doc.addLinkToSpec(l, refNameResolver, parentIsExternal)
	}
}

func (doc *T) derefResponse(r *ResponseRef, refNameResolver RefNameResolver, parentIsExternal bool) {
	if r == nil {
		return
	}
	isExternal := doc.addResponseToSpec(r, refNameResolver, parentIsExternal)
	if v := r.Value; v != nil {
		doc.derefHeaders(v.Headers, refNameResolver, isExternal || parentIsExternal)
		doc.derefContent(v.Content, refNameResolver, isExternal || parentIsExternal)
		doc.derefLinks(v.Links, refNameResolver, isExternal || parentIsExternal)
	}
}

func (doc *T) derefResponses(rs *Responses, refNameResolver RefNameResolver, parentIsExternal bool) {
	doc.derefResponseBodies(rs.Map(), refNameResolver, parentIsExternal)
}

func (doc *T) derefResponseBodies(es ResponseBodies, refNameResolver RefNameResolver, parentIsExternal bool) {
	for _, name := range componentNames(es) {
		e := es[name]
		doc.derefResponse(e, refNameResolver, parentIsExternal)
	}
}

func (doc *T) derefParameter(p Parameter, refNameResolver RefNameResolver, parentIsExternal bool) {
	doc.derefExamples(p.Examples, refNameResolver, parentIsExternal)
	doc.derefContent(p.Content, refNameResolver, parentIsExternal)
	doc.derefSchemaRef(p.Schema, refNameResolver, parentIsExternal)
}

func (doc *T) derefRequestBody(r RequestBody, refNameResolver RefNameResolver, parentIsExternal bool) {
	doc.derefContent(r.Content, refNameResolver, parentIsExternal)
}

func (doc *T) derefPaths(paths map[string]*PathItem, refNameResolver RefNameResolver, parentIsExternal bool) {
	for _, name := range componentNames(paths) {
		ops := paths[name]
		if ops == nil {
			continue
		}
		pathIsExternal := isExternalRef(ops.Ref, parentIsExternal)
		// inline full operations
		ops.Ref = ""

		for _, param := range ops.Parameters {
			isExternal := doc.addParameterToSpec(param, refNameResolver, pathIsExternal)
			if param.Value != nil {
				doc.derefParameter(*param.Value, refNameResolver, pathIsExternal || isExternal)
			}
		}

		opsWithMethod := ops.Operations()
		for _, name := range componentNames(opsWithMethod) {
			op := opsWithMethod[name]
			isExternal := doc.addRequestBodyToSpec(op.RequestBody, refNameResolver, pathIsExternal)
			if op.RequestBody != nil && op.RequestBody.Value != nil {
				doc.derefRequestBody(*op.RequestBody.Value, refNameResolver, pathIsExternal || isExternal)
			}
			for _, name := range componentNames(op.Callbacks) {
				cb := op.Callbacks[name]
				isExternal := doc.addCallbackToSpec(cb, refNameResolver, pathIsExternal)
				if cb.Value != nil {
					cbValue := (*cb.Value).Map()
					doc.derefPaths(cbValue, refNameResolver, pathIsExternal || isExternal)
				}
			}
			doc.derefResponses(op.Responses, refNameResolver, pathIsExternal)
			for _, param := range op.Parameters {
				isExternal := doc.addParameterToSpec(param, refNameResolver, pathIsExternal)
				if param.Value != nil {
					doc.derefParameter(*param.Value, refNameResolver, pathIsExternal || isExternal)
				}
			}
		}
	}
}

// InternalizeRefs removes all references to external files from the spec and moves them
// to the components section.
//
// refNameResolver takes in references to returns a name to store the reference under locally.
// It MUST return a unique name for each reference type.
// A default implementation is provided that will suffice for most use cases. See the function
// documentation for more details.
//
// Example:
//
//	doc.InternalizeRefs(context.Background(), nil)
func (doc *T) InternalizeRefs(ctx context.Context, refNameResolver func(*T, ComponentRef) string) {
	doc.resetVisited()

	if refNameResolver == nil {
		refNameResolver = DefaultRefNameResolver
	}

	if components := doc.Components; components != nil {
		for _, name := range componentNames(components.Schemas) {
			schema := components.Schemas[name]
			doc.derefSchemaRef(schema, refNameResolver, false)
			if schema != nil && schema.Value != nil {
				schema.Ref = "" // always dereference the top level
			}
		}
		for _, name := range componentNames(components.Parameters) {
			p := components.Parameters[name]
			isExternal := doc.addParameterToSpec(p, refNameResolver, false)
			if p != nil && p.Value != nil {
				p.Ref = "" // always dereference the top level
				doc.derefParameter(*p.Value, refNameResolver, isExternal)
			}
		}
		doc.derefHeaders(components.Headers, refNameResolver, false)
		for _, name := range componentNames(components.Headers) {
			if h := components.Headers[name]; h != nil && h.Value != nil {
				h.Ref = "" // always dereference the top level
			}
		}
		for _, name := range componentNames(components.RequestBodies) {
			req := components.RequestBodies[name]
			isExternal := doc.addRequestBodyToSpec(req, refNameResolver, false)
			if req != nil && req.Value != nil {
				req.Ref = "" // always dereference the top level
				doc.derefRequestBody(*req.Value, refNameResolver, isExternal)
			}
		}
		doc.derefResponseBodies(components.Responses, refNameResolver, false)
		for _, name := range componentNames(components.Responses) {
			if r := components.Responses[name]; r != nil && r.Value != nil {
				r.Ref = "" // always dereference the top level
			}
		}
		for _, name := range componentNames(components.SecuritySchemes) {
			ss := components.SecuritySchemes[name]
			doc.addSecuritySchemeToSpec(ss, refNameResolver, false)
			if ss != nil && ss.Value != nil {
				ss.Ref = "" // always dereference the top level
			}
		}
		doc.derefExamples(components.Examples, refNameResolver, false)
		for _, name := range componentNames(components.Examples) {
			if e := components.Examples[name]; e != nil && e.Value != nil {
				e.Ref = "" // always dereference the top level
			}
		}
		doc.derefLinks(components.Links, refNameResolver, false)
		for _, name := range componentNames(components.Links) {
			if l := components.Links[name]; l != nil && l.Value != nil {
				l.Ref = "" // always dereference the top level
			}
		}

		for _, name := range componentNames(components.Callbacks) {
			cb := components.Callbacks[name]
			isExternal := doc.addCallbackToSpec(cb, refNameResolver, false)
			if cb != nil && cb.Value != nil {
				cb.Ref = "" // always dereference the top level
				cbValue := (*cb.Value).Map()
				doc.derefPaths(cbValue, refNameResolver, isExternal)
			}
		}
	}

	doc.derefPaths(doc.Paths.Map(), refNameResolver, false)
	doc.derefPaths(doc.Webhooks, refNameResolver, false)
}
