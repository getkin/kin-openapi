// Package yamlconv converts between YAML and types that describe themselves
// with json tags and MarshalJSON/UnmarshalJSON methods.
//
// YAML reaches such a type through JSON, so its methods run and extensions are
// carried. Two things have to be reconciled first, both of which YAML resolves
// and JSON cannot represent.
package yamlconv

import (
	"encoding/json"

	yaml "go.yaml.in/yaml/v3"
)

// PrepareForJSON retags the scalars that would not survive the conversion.
//
// A date-shaped scalar resolves to a timestamp, which has no JSON form. A
// non-string mapping key -- the unquoted 200: written in most specs for a
// status code -- decodes to a map[any]any that json.Marshal rejects. Both
// become strings. An explicitly tagged value is left alone, being a deliberate
// request for that type.
func PrepareForJSON(n *yaml.Node) {
	if n == nil {
		return
	}
	if n.Kind == yaml.ScalarNode && n.Tag == "!!timestamp" && n.Style != yaml.TaggedStyle {
		n.Tag = "!!str"
	}
	if n.Kind == yaml.MappingNode {
		for i := 0; i < len(n.Content); i += 2 {
			if k := n.Content[i]; k.Kind == yaml.ScalarNode && k.Tag != "!!str" {
				k.Tag = "!!str"
			}
		}
	}
	for _, c := range n.Content {
		PrepareForJSON(c)
	}
}

// Unmarshal decodes YAML into v via JSON, so v's UnmarshalJSON runs.
func Unmarshal(data []byte, v any) error {
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return err
	}
	PrepareForJSON(&root)
	var generic any
	if err := root.Decode(&generic); err != nil {
		return err
	}
	j, err := json.Marshal(generic)
	if err != nil {
		return err
	}
	return json.Unmarshal(j, v)
}

// Marshal renders v as YAML via JSON, so v's MarshalJSON runs.
func Marshal(v any) ([]byte, error) {
	j, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var generic any
	if err := json.Unmarshal(j, &generic); err != nil {
		return nil, err
	}
	return yaml.Marshal(generic)
}
