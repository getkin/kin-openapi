package openapi3

// UnmarshalYAML for the maplike collections, whose entries are components
// rather than declared fields, and for the union-typed values.

import (
	"strings"

	yaml "go.yaml.in/yaml/v3"
)

// unmarshalMaplikeYAML decodes a mapping whose x- keys are extensions and whose
// remaining entries are components, stamping each entry's origin from the key
// that heads it.
func unmarshalMaplikeYAML[V any](node *yaml.Node, ext *map[string]any, out *map[string]*V) error {
	if node.Kind != yaml.MappingNode {
		return node.Decode(out)
	}
	*ext = make(map[string]any)
	*out = make(map[string]*V, len(node.Content)/2)
	for _, kv := range mappingPairs(node) {
		k, v := kv[0].Value, kv[1]
		if strings.HasPrefix(k, "x-") {
			a, err := decodeAny(v)
			if err != nil {
				return err
			}
			(*ext)[k] = a
			continue
		}
		var vv V
		if err := v.Decode(&vv); err != nil {
			return err
		}
		(*out)[k] = &vv
	}
	return nil
}

func (responses *Responses) UnmarshalYAML(node *yaml.Node) error {
	var x Responses
	if err := unmarshalMaplikeYAML(node, &x.Extensions, &x.m); err != nil {
		return err
	}
	*responses = x
	return nil
}

func (callback *Callback) UnmarshalYAML(node *yaml.Node) error {
	var x Callback
	if err := unmarshalMaplikeYAML(node, &x.Extensions, &x.m); err != nil {
		return err
	}
	*callback = x
	return nil
}

func (paths *Paths) UnmarshalYAML(node *yaml.Node) error {
	var x Paths
	if err := unmarshalMaplikeYAML(node, &x.Extensions, &x.m); err != nil {
		return err
	}
	*paths = x
	return nil
}

// Header embeds Parameter and carries no fields of its own.
func (header *Header) UnmarshalYAML(node *yaml.Node) error {
	return header.Parameter.UnmarshalYAML(node)
}

// BoolSchema is `true`/`false` or a schema.
func (bs *BoolSchema) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		if node.Tag == "!!null" {
			return nil
		}
		var b bool
		if err := node.Decode(&b); err == nil {
			bs.Has = &b
			return nil
		}
	}
	var sr SchemaRef
	if err := node.Decode(&sr); err != nil {
		return err
	}
	bs.Schema = &sr
	return nil
}

// Schema is `true`/`false` or a mapping. JSON Schema 2020-12, which OpenAPI 3.1
// adopts, allows a boolean wherever a schema is expected.
//
// The mapping branch is what the generator emits for the other declared-field
// types; only the scalar case is particular to Schema.
func (schema *Schema) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode && node.Tag == "!!bool" {
		var b bool
		if err := node.Decode(&b); err != nil {
			return err
		}
		schema.Always = &b
		return nil
	}

	type bis Schema
	ext, err := decodeMapping(node, (*bis)(schema))
	if err != nil {
		return err
	}
	schema.Extensions = ext
	return nil
}

// ExclusiveBound is a bool in OAS 3.0, where it modifies minimum/maximum, or a
// number in 3.1, where it is the bound itself.
func (eb *ExclusiveBound) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.ScalarNode || node.Tag == "!!null" {
		return nil
	}
	var b bool
	if err := node.Decode(&b); err == nil {
		eb.Bool = &b
		return nil
	}
	var f float64
	if err := node.Decode(&f); err != nil {
		return err
	}
	eb.Value = &f
	return nil
}

// Operation distinguishes an omitted responses from an explicitly null one:
// the first is allowed in OAS 3.1 and later, the second never is. A null node
// decodes to an empty Responses, which is indistinguishable from `{}` without
// the flag.
func (operation *Operation) UnmarshalYAML(node *yaml.Node) error {
	type bis Operation
	ext, err := decodeMapping(node, (*bis)(operation))
	if err != nil {
		return err
	}
	operation.Extensions = ext
	if v := mappingValue(node, "responses"); v != nil && v.Tag == "!!null" {
		operation.Responses = &Responses{explicitlyNull: true}
	}
	return nil
}

func (doc *T) UnmarshalYAML(node *yaml.Node) error {
	type bis T
	ext, err := decodeMapping(node, (*bis)(doc))
	if err != nil {
		return err
	}
	doc.Extensions = ext
	return nil
}
