package openapi3

//go:generate go run nativeyamlgenerator.go

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	yaml "go.yaml.in/yaml/v3"
)

var knownYAMLFieldsCache sync.Map // reflect.Type -> map[string]struct{}

// knownYAMLFields returns the yaml keys a struct type declares, skipping "-".
func knownYAMLFields(t reflect.Type) map[string]struct{} {
	if v, ok := knownYAMLFieldsCache.Load(t); ok {
		return v.(map[string]struct{})
	}
	known := make(map[string]struct{}, t.NumField())
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		name, _, _ := strings.Cut(f.Tag.Get("yaml"), ",")
		if name == "" {
			name = strings.ToLower(f.Name)
		}
		if name == "-" {
			continue
		}
		known[name] = struct{}{}
	}
	knownYAMLFieldsCache.Store(t, known)
	return known
}

// normalizeDecoded gives a decoded any the shape the JSON round trip used to
// guarantee, so a consumer's type switch does not depend on YAML notation:
//
//   - integers widen to float64. JSON has one number type, so a value reaching
//     an any-typed field carried a float64 whichever notation the source used.
//     YAML resolves 42, 0x2A and 4_2 to an int.
//   - non-string map keys become strings. YAML allows any scalar as a key, so
//     an unquoted 1: becomes map[any]any, which json.Marshal rejects outright.
//
// Applied only to any-typed values. A declared integer field keeps its own
// type and its full range, which a blanket conversion would cost beyond 2^53.
func normalizeDecoded(v any) any {
	switch t := v.(type) {
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case uint64:
		return float64(t)
	case map[string]any:
		for k, e := range t {
			t[k] = normalizeDecoded(e)
		}
	case map[any]any:
		m := make(map[string]any, len(t))
		for k, e := range t {
			m[fmt.Sprintf("%v", k)] = normalizeDecoded(e)
		}
		return m
	case []any:
		for i, e := range t {
			t[i] = normalizeDecoded(e)
		}
	}
	return v
}

// mappingPairs returns node's effective key/value node pairs, applying YAML
// merge keys.
//
// Only Decode applies a merge, and these helpers read node.Content directly, so
// a << key arrives here unexpanded. Ignoring it drops whatever the merge brought
// in; treating it as an ordinary key records "<<" as an extension on a struct
// and as a member on a collection. Both corrupt the document, so the merge is
// resolved here instead, giving every raw-key pass the same view Decode has.
//
// Merge semantics: an explicit key wins over a merged one, and among merged
// sources an earlier one wins, per the YAML merge key spec.
func mappingPairs(node *yaml.Node) [][2]*yaml.Node {
	pairs := make([][2]*yaml.Node, 0, len(node.Content)/2)
	var merged [][2]*yaml.Node
	seen := make(map[string]struct{}, len(node.Content)/2)

	for i := 0; i+1 < len(node.Content); i += 2 {
		k, v := node.Content[i], node.Content[i+1]
		if k.Tag == mergeKeyTag || k.Value == mergeKey {
			merged = append(merged, mergeSourcePairs(v)...)
			continue
		}
		pairs = append(pairs, [2]*yaml.Node{k, v})
		seen[k.Value] = struct{}{}
	}

	for _, kv := range merged {
		if _, ok := seen[kv[0].Value]; ok {
			continue
		}
		seen[kv[0].Value] = struct{}{}
		pairs = append(pairs, kv)
	}
	return pairs
}

// mergeSourcePairs returns the pairs a merge key's value contributes: one
// mapping, or a sequence of them in precedence order. Anything else is not a
// valid merge source and contributes nothing.
func mergeSourcePairs(v *yaml.Node) [][2]*yaml.Node {
	if v.Kind == yaml.AliasNode {
		v = v.Alias
	}
	if v == nil {
		return nil
	}
	switch v.Kind {
	case yaml.MappingNode:
		return mappingPairs(v)
	case yaml.SequenceNode:
		var out [][2]*yaml.Node
		for _, e := range v.Content {
			out = append(out, mergeSourcePairs(e)...)
		}
		return out
	}
	return nil
}

const (
	mergeKey    = "<<"
	mergeKeyTag = "!!merge"
)

// decodeAny decodes a value node into an any and normalizes it. Every path that
// puts a decoded value behind an any must come through here, so none of them can
// omit a normalization by forgetting to call it.
func decodeAny(node *yaml.Node) (any, error) {
	var v any
	if err := node.Decode(&v); err != nil {
		return nil, err
	}
	return normalizeDecoded(v), nil
}

// collectExtensions returns the mapping keys of node that known does not
// declare, decoded and normalized, along with their names in document order.
// The names are every undeclared sibling, including x- ones; callers that keep
// only extensions filter the map, and callers that validate siblings read the
// names.
func collectExtensions(node *yaml.Node, known map[string]struct{}) (map[string]any, []string, error) {
	if node.Kind != yaml.MappingNode {
		return nil, nil, nil
	}
	var ext map[string]any
	var names []string
	for _, kv := range mappingPairs(node) {
		key := kv[0].Value
		if _, ok := known[key]; ok {
			continue
		}
		v, err := decodeAny(kv[1])
		if err != nil {
			return nil, nil, err
		}
		if ext == nil {
			ext = make(map[string]any)
		}
		ext[key] = v
		names = append(names, key)
	}
	return ext, names, nil
}

// normalizeAnyFields applies normalizeNumbers to a struct's any-typed fields,
// which the decoder fills directly: Example and Default, but also Enum, whose
// element type is any. Missing the slice case left an integer example as a
// float64 and the enum it must match as an int, so a schema failed against its
// own allowed values.
func normalizeAnyFields(out any) {
	v := reflect.ValueOf(out)
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return
	}
	for i := range v.NumField() {
		f := v.Field(i)
		if !f.CanSet() {
			continue
		}
		switch f.Kind() {
		case reflect.Interface:
			if !f.IsNil() {
				f.Set(reflect.ValueOf(normalizeDecoded(f.Interface())))
			}
		case reflect.Slice, reflect.Map:
			// []any and map[string]any: normalise the elements in place.
			if f.Type().Elem().Kind() != reflect.Interface || f.IsNil() {
				continue
			}
			if f.Kind() == reflect.Slice {
				for j := range f.Len() {
					e := f.Index(j)
					if !e.IsNil() {
						e.Set(reflect.ValueOf(normalizeDecoded(e.Interface())))
					}
				}
				continue
			}
			for _, k := range f.MapKeys() {
				e := f.MapIndex(k)
				if !e.IsNil() {
					f.SetMapIndex(k, reflect.ValueOf(normalizeDecoded(e.Interface())))
				}
			}
		}
	}
}

// decodeMapping decodes node into a method-less view of the target, supplied by
// the caller as a locally-declared shadow type, and returns the keys the target
// does not declare.
func decodeMapping[S any](node *yaml.Node, shadow *S) (map[string]any, error) {
	return decodeStructWithExtensions(node, shadow)
}

// decodeStructWithExtensions decodes node into out and returns the mapping keys
// out does not declare, which are the extensions. Returns nil rather than an
// empty map when there are none.
//
// The declared set comes from out's yaml tags, so adding a field to a struct is
// enough to stop it being collected as an extension.
func decodeStructWithExtensions(node *yaml.Node, out any) (map[string]any, error) {
	if err := node.Decode(out); err != nil {
		return nil, err
	}
	if node.Kind != yaml.MappingNode {
		return nil, nil
	}
	ext, _, err := collectExtensions(node, knownYAMLFields(reflect.TypeOf(out).Elem()))
	if err != nil {
		return nil, err
	}
	normalizeAnyFields(out)
	return ext, nil
}

// stripTimestamps retags implicitly-resolved date-shaped scalars as strings.
//
// YAML 1.1 resolves an untagged scalar such as 2020-06-11T16:32:50-03:00 to a
// timestamp, which would make an OpenAPI `example` of that shape decode to a
// time.Time and fail validation as an unhandled type. An explicit !!timestamp
// tag is a deliberate request for a time.Time and is left alone.
func stripTimestamps(n *yaml.Node) {
	if n == nil {
		return
	}
	if n.Kind == yaml.ScalarNode && n.Tag == "!!timestamp" && n.Style != yaml.TaggedStyle {
		n.Tag = "!!str"
	}
	for _, c := range n.Content {
		stripTimestamps(c)
	}
}
