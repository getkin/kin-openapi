package openapi3

// Shared by the generated $ref wrapper UnmarshalYAML methods in refs.go.

import (
	"strings"

	yaml "go.yaml.in/yaml/v3"
)

// refYAMLFields are the keys the wrapper declares itself, so they are not
// siblings. summary and description count only for wrappers that carry them:
// SchemaRef does not, and there they are siblings like any other key, which is
// what the JSON path records.

// unmarshalRefYAML fills the reference half of a wrapper and reports whether
// the node was a reference. When false the caller decodes the value.
func unmarshalRefYAML(node *yaml.Node, ref *string, summary, description **string, extensions *map[string]any, extra *[]string, siblingExtras *map[string]any) bool {
	refNode := mappingValue(node, "$ref")
	if refNode != nil && refNode.Kind == yaml.AliasNode && refNode.Alias != nil {
		refNode = refNode.Alias
	}
	// A $ref is a string. A null, a number or a collection is not a reference,
	// so the object decodes as an ordinary value, which is what the json path
	// does with a $ref that will not decode into a string.
	if refNode == nil || refNode.Kind != yaml.ScalarNode || refNode.Tag != "!!str" || refNode.Value == "" {
		return false
	}
	*ref = refNode.Value

	known := map[string]struct{}{"$ref": {}}
	if summary != nil {
		known["summary"] = struct{}{}
	}
	if description != nil {
		known["description"] = struct{}{}
	}

	for _, kv := range mappingPairs(node) {
		switch k, v := kv[0].Value, kv[1]; {
		case k == "summary" && summary != nil:
			var str string
			if v.Decode(&str) == nil {
				*summary = &str
			}
		case k == "description" && description != nil:
			var str string
			if v.Decode(&str) == nil {
				*description = &str
			}
		}
	}

	// Every undeclared sibling is recorded, not only the x- ones: Validate reads
	// these names, and a name it never sees is a name it cannot report. The x-
	// ones are extensions, the rest are siblings of the $ref, which 3.1 lets a
	// schema apply alongside the resolved reference.
	siblings, names, err := collectExtensions(node, known)
	if err != nil {
		return true
	}
	*extra = names
	for k, v := range siblings {
		if !strings.HasPrefix(k, "x-") {
			if *siblingExtras == nil {
				*siblingExtras = make(map[string]any, len(siblings))
			}
			(*siblingExtras)[k] = v
			continue
		}
		if *extensions == nil {
			*extensions = make(map[string]any)
		}
		(*extensions)[k] = v
	}
	return true
}
