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
func unmarshalRefYAML(node *yaml.Node, ref *string, summary, description **string, extensions *map[string]any, extra *[]string) bool {
	refNode := mappingValue(node, "$ref")
	if refNode == nil || refNode.Value == "" {
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

	// Every undeclared sibling is recorded, not only the x- ones. Dropping the
	// rest here is what let a $ref carrying an unknown key validate: Validate
	// reads these names, and a name it never sees is a name it cannot report.
	siblings, names, err := collectExtensions(node, known)
	if err != nil {
		return true
	}
	*extra = names
	for k, v := range siblings {
		if !strings.HasPrefix(k, "x-") {
			continue
		}
		if *extensions == nil {
			*extensions = make(map[string]any)
		}
		(*extensions)[k] = v
	}
	return true
}
