package openapi3

import (
	"context"
	"strconv"
)

func isInternalRef(ref string) bool {
	return ref != "" && ref[0] == '#'
}
func isExternalRef(ref string) bool {
	return ref != "" && ref[0] != '#'
}

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

// InternalizeSpecRefs removes all references to external files from the spec and moves them
// to the components section.
//
// refNameResolver takes in references to returns a name to store the reference under locally.
//
// Currently response and request bodies are just inlined rather than moved to the components.
func (spec *T) InternalizeSpecRefs(ctx context.Context, refNameResolver func(ref string) string) *T {
	addSchemaToSpec := func(s *SchemaRef) string {
		name := refNameResolver(s.Ref)

		val := 1
		for {
			// basic check, may need something more exhaustive...
			if existing, ok := spec.Components.Schemas[name]; ok {
				if existing.Value.Description != s.Value.Description || s.Value.Type != existing.Value.Type {
					name = refNameResolver(s.Ref) + strconv.Itoa(val)
					val++
					continue
				}
				s.Ref = "#/components/schemas/" + name
				return s.Ref
			}
			break
		}

		spec.Components.Schemas[name] = s.Value.NewRef()
		s.Ref = "#/components/schemas/" + name
		return s.Ref
	}
	tryAddSchemaRef := func(s *SchemaRef) {
		if s != nil && isExternalRef(s.Ref) {
			addSchemaToSpec(s)
		}
	}

	addParameterToSpec := func(p *ParameterRef) string {
		name := refNameResolver(p.Ref)
		if _, ok := spec.Components.Parameters[name]; ok {
			p.Ref = "#/components/parameters/" + name
			return p.Ref
		}

		spec.Components.Parameters[name] = &ParameterRef{Value: p.Value}
		p.Ref = "#/components/parameters/" + name
		return p.Ref
	}

	var derefSchema func(*Schema, []*Schema)
	derefSchema = func(s *Schema, stack []*Schema) {
		if len(stack) > 1000 {
			panic("unresolved circular reference")
		}
		for _, p := range stack {
			if p == s {
				return
			}
		}
		stack = append(stack, s)

		for _, list := range []SchemaRefs{s.AllOf, s.AnyOf, s.OneOf} {
			for _, s2 := range list {
				tryAddSchemaRef(s2)
				derefSchema(s2.Value, stack)
			}
		}
		for _, s2 := range s.Properties {
			tryAddSchemaRef(s2)
			derefSchema(s2.Value, stack)
		}
		for _, ref := range []*SchemaRef{s.Not, s.AdditionalProperties, s.Items} {
			if ref != nil {
				tryAddSchemaRef(ref)
				derefSchema(ref.Value, stack)
			}
		}
	}

	derefParameter := func(p *Parameter) {
		tryAddSchemaRef(p.Schema)
		if p.Schema.Value != nil {
			derefSchema(p.Schema.Value, nil)
		}
	}

	derefContent := func(c Content) {
		for _, mediatype := range c {
			tryAddSchemaRef(mediatype.Schema)
			if mediatype.Schema != nil {
				derefSchema(mediatype.Schema.Value, nil)
			}
		}
	}

	// inline all component references
	names := schemaNames(spec.Components.Schemas)
	for _, name := range names {
		schema := spec.Components.Schemas[name]
		derefSchema(schema.Value, nil)
	}
	names = parametersMapNames(spec.Components.Parameters)
	for _, name := range names {
		p := spec.Components.Parameters[name]
		derefParameter(p.Value)
	}

	for _, ops := range spec.Paths {
		// inline full operations
		ops.Ref = ""

		for _, op := range ops.Operations() {
			if op.RequestBody != nil {
				op.RequestBody.Ref = ""
				derefContent(op.RequestBody.Value.Content)
			}
			for _, res := range op.Responses {
				res.Ref = ""
				derefContent(res.Value.Content)
			}
			for _, param := range op.Parameters {
				// don't inline, add to file refs
				if isExternalRef(param.Ref) {
					addParameterToSpec(param)
				}
				derefParameter(param.Value)
			}
		}
	}

	return spec
}
