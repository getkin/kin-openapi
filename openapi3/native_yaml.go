package openapi3

//go:generate go run nativeyamlgenerator.go

import (
	"errors"
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
		if name := yamlKey(f); name != "-" {
			known[name] = struct{}{}
		}
	}
	knownYAMLFieldsCache.Store(t, known)
	return known
}

// yamlKey is the document key a struct field is written under.
func yamlKey(f reflect.StructField) string {
	name, _, _ := strings.Cut(f.Tag.Get("yaml"), ",")
	if name == "" {
		name = strings.ToLower(f.Name)
	}
	return name
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
func normalizeDecoded(v any) (any, error) {
	switch t := v.(type) {
	case int:
		return float64(t), nil
	case int64:
		return float64(t), nil
	case uint64:
		return float64(t), nil
	case map[string]any:
		for k, e := range t {
			n, err := normalizeDecoded(e)
			if err != nil {
				return nil, err
			}
			t[k] = n
		}
	case map[any]any:
		m := make(map[string]any, len(t))
		for k, e := range t {
			name, err := keyName(k)
			if err != nil {
				return nil, err
			}
			n, err := normalizeDecoded(e)
			if err != nil {
				return nil, err
			}
			m[name] = n
		}
		return m, nil
	case []any:
		for i, e := range t {
			n, err := normalizeDecoded(e)
			if err != nil {
				return nil, err
			}
			t[i] = n
		}
	}
	return v, nil
}

// keyName is the name a non-string key is written under. YAML resolves a key
// the way it resolves any scalar, and a document's names are strings, so a
// scalar key is spelled out.
//
// A null key spells nothing. It names no member of the document, and the json
// path has no key to decode it into, so a document that uses one is refused
// rather than given a made-up name.
func keyName(k any) (string, error) {
	if k == nil {
		return "", errors.New("mapping key is null, so it names nothing")
	}
	return fmt.Sprintf("%v", k), nil
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
//
// Interim: https://github.com/yaml/go-yaml/issues/451 asks the parser for a
// mapping's effective pairs, which is this function and the cycle guard below
// it. Reimplementing the rules here is also what the self-merge crasher came
// from.
func mappingPairs(node *yaml.Node) [][2]*yaml.Node {
	return mappingPairsOf(node, nil)
}

// mappingPairsOf is mappingPairs carrying the mappings whose merges are being
// expanded, so a mapping that merges itself through an alias contributes
// nothing instead of recursing until the stack ends. The set is allocated only
// for a mapping that has a merge key at all.
func mappingPairsOf(node *yaml.Node, expanding map[*yaml.Node]struct{}) [][2]*yaml.Node {
	pairs := make([][2]*yaml.Node, 0, len(node.Content)/2)
	var merged [][2]*yaml.Node
	seen := make(map[string]struct{}, len(node.Content)/2)

	for i := 0; i+1 < len(node.Content); i += 2 {
		k, v := node.Content[i], node.Content[i+1]
		if isMergeKey(k) {
			if expanding == nil {
				expanding = make(map[*yaml.Node]struct{}, 1)
			}
			expanding[node] = struct{}{}
			merged = append(merged, mergeSourcePairs(v, expanding)...)
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

// mappingValue returns the value node for key, or nil.
func mappingValue(node *yaml.Node, key string) *yaml.Node {
	_, v := mappingEntry(node, key)
	return v
}

// mappingEntry returns both halves of a mapping entry. Merge keys are applied,
// so a field a merge brought in is found here the same as one written in place.
func mappingEntry(node *yaml.Node, key string) (*yaml.Node, *yaml.Node) {
	if node.Kind != yaml.MappingNode {
		return nil, nil
	}
	for _, kv := range mappingPairs(node) {
		if kv[0].Value == key {
			return kv[0], kv[1]
		}
	}
	return nil, nil
}

// isMergeKey reports whether k is a merge key rather than an ordinary one.
// Only a plain << merges; quoted, it is a member actually named "<<", which
// the json path keeps and so must this.
func isMergeKey(k *yaml.Node) bool {
	if k.Tag == mergeKeyTag {
		return true
	}
	return k.Value == mergeKey && k.Style == 0 && (k.Tag == "" || k.Tag == strTag)
}

// mergeSourcePairs returns the pairs a merge key's value contributes: one
// mapping, or a sequence of them in precedence order. Anything else is not a
// valid merge source and contributes nothing.
func mergeSourcePairs(v *yaml.Node, expanding map[*yaml.Node]struct{}) [][2]*yaml.Node {
	if v.Kind == yaml.AliasNode {
		v = v.Alias
	}
	if v == nil {
		return nil
	}
	switch v.Kind {
	case yaml.MappingNode:
		if _, ok := expanding[v]; ok {
			return nil
		}
		return mappingPairsOf(v, expanding)
	case yaml.SequenceNode:
		var out [][2]*yaml.Node
		for _, e := range v.Content {
			out = append(out, mergeSourcePairs(e, expanding)...)
		}
		return out
	}
	return nil
}

const (
	mergeKey     = "<<"
	mergeKeyTag  = "!!merge"
	strTag       = "!!str"
	nullTag      = "!!null"
	timestampTag = "!!timestamp"
)

// decodeAny decodes a value node into an any and normalizes it. Every path that
// puts a decoded value behind an any must come through here, so none of them can
// omit a normalization by forgetting to call it.
func decodeAny(node *yaml.Node) (any, error) {
	var v any
	if err := node.Decode(&v); err != nil {
		return nil, err
	}
	return normalizeDecoded(v)
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
func normalizeAnyFields(out any) error {
	v := reflect.ValueOf(out)
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	for i := range v.NumField() {
		f := v.Field(i)
		if !f.CanSet() {
			continue
		}
		switch f.Kind() {
		case reflect.Interface:
			if f.IsNil() {
				continue
			}
			n, err := normalizeDecoded(f.Interface())
			if err != nil {
				return err
			}
			f.Set(reflect.ValueOf(n))
		case reflect.Slice, reflect.Map:
			// []any and map[string]any: normalise the elements in place.
			if f.Type().Elem().Kind() != reflect.Interface || f.IsNil() {
				continue
			}
			if f.Kind() == reflect.Slice {
				for j := range f.Len() {
					e := f.Index(j)
					if e.IsNil() {
						continue
					}
					n, err := normalizeDecoded(e.Interface())
					if err != nil {
						return err
					}
					e.Set(reflect.ValueOf(n))
				}
				continue
			}
			for _, k := range f.MapKeys() {
				e := f.MapIndex(k)
				if e.IsNil() {
					continue
				}
				n, err := normalizeDecoded(e.Interface())
				if err != nil {
					return err
				}
				f.SetMapIndex(k, reflect.ValueOf(n))
			}
		}
	}
	return nil
}

// decodeMapping decodes node into a method-less view of the target, supplied by
// the caller as a locally-declared shadow type, and returns the keys the target
// does not declare. T is the target itself, named so that an error reports the
// type the document was read into.
func decodeMapping[T, S any](node *yaml.Node, shadow *S) (map[string]any, error) {
	ext, err := decodeStructWithExtensions(node, shadow)
	if err != nil {
		return nil, targetTypeError[T, S](err)
	}
	return ext, nil
}

// targetTypeError names T where a decoder error names the shadow type S, which
// exists only to stop the decode recursing and means nothing to a reader.
func targetTypeError[T, S any](err error) error {
	var typeErr *yaml.TypeError
	if !errors.As(err, &typeErr) {
		return err
	}
	shadow, target := reflect.TypeFor[S]().String(), reflect.TypeFor[T]().String()
	for i, message := range typeErr.Errors {
		typeErr.Errors[i] = strings.ReplaceAll(message, shadow, target)
	}
	return typeErr
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
	if err := checkDeclaredFields(node, out); err != nil {
		return nil, err
	}
	ext, _, err := collectExtensions(node, knownYAMLFields(reflect.TypeOf(out).Elem()))
	if err != nil {
		return nil, err
	}
	if err := normalizeAnyFields(out); err != nil {
		return nil, err
	}
	return ext, nil
}

// checkDeclaredFields holds out's declared fields to the values a JSON
// document would carry, which the decoder is laxer than on two counts.
//
// A scalar's text goes into a string field whatever the scalar resolved to, so
// `title: 1` reads as "1" where the json path refuses the document. And a
// sequence element the target cannot hold is dropped, so `required: [a, null,
// b]` reads as two entries where the json path keeps the null's place as an
// empty string. A document that changes meaning with the decoder is worse than
// one that is refused, so a scalar that resolved to something other than a
// string is refused here, and a null in a list of strings keeps its place.
//
// Only out's own fields. A field whose type has an UnmarshalYAML of its own is
// held to this when the decoder reaches that type.
func checkDeclaredFields(node *yaml.Node, out any) error {
	if node.Kind != yaml.MappingNode {
		return nil
	}
	v := reflect.ValueOf(out).Elem()
	t := v.Type()
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		key := yamlKey(f)
		if key == "-" {
			continue
		}
		value := mappingValue(node, key)
		if value == nil {
			continue
		}
		if err := checkFieldValue(v.Field(i), value, key); err != nil {
			return err
		}
	}
	return nil
}

// checkFieldValue applies checkDeclaredFields to one field: the string slots
// are a string field, a list of strings, and a map whose values are written as
// text. Anything else is the decoder's business.
func checkFieldValue(field reflect.Value, node *yaml.Node, key string) error {
	for field.Kind() == reflect.Pointer {
		if field.IsNil() {
			return nil
		}
		field = field.Elem()
	}
	switch {
	case field.Kind() == reflect.String:
		return requireStringScalar(node, key)
	case field.Kind() == reflect.Slice && field.Type().Elem().Kind() == reflect.String:
		return setStringList(field, node, key)
	case field.Kind() == reflect.Map && writtenAsText(field.Type().Elem()):
		for _, kv := range mappingPairs(node) {
			if err := requireStringScalar(kv[1], kv[0].Value); err != nil {
				return err
			}
		}
	}
	return nil
}

// writtenAsText reports whether a value of type t is written as a scalar
// string, either because it is one or because it decodes from text.
func writtenAsText(t reflect.Type) bool {
	if t.Kind() == reflect.String {
		return true
	}
	return reflect.PointerTo(t).Implements(reflect.TypeFor[interface{ UnmarshalText(text []byte) error }]())
}

// setStringList sets a list of strings from node. The field may be any type
// whose underlying type is a list of strings, such as Types.
func setStringList(field reflect.Value, node *yaml.Node, key string) error {
	list, ok, err := stringList(node, key)
	if err != nil || !ok {
		return err
	}
	field.Set(reflect.ValueOf(list).Convert(field.Type()))
	return nil
}

// stringList is the list of strings node holds, keeping the place of a null as
// the empty string. The decoder drops such an element, since a string cannot
// hold nil, which shortens the list without saying so.
//
// ok is false for a node that is not a sequence: one value where a list is
// allowed, as in `type: string`, and anything the decode reports itself.
func stringList(node *yaml.Node, key string) ([]string, bool, error) {
	if node.Kind == yaml.AliasNode {
		node = node.Alias
	}
	if node.Kind == yaml.ScalarNode {
		return nil, false, requireStringScalar(node, key)
	}
	if node.Kind != yaml.SequenceNode {
		return nil, false, nil
	}

	list := make([]string, len(node.Content))
	for i, e := range node.Content {
		if e.Kind == yaml.AliasNode {
			e = e.Alias
		}
		if err := requireStringScalar(e, key); err != nil {
			return nil, false, err
		}
		if e.Kind != yaml.ScalarNode {
			return nil, false, nil
		}
		if e.Tag != nullTag {
			list[i] = e.Value
		}
	}
	return list, true, nil
}

// requireStringScalar refuses a scalar that resolved to something other than a
// string where a string belongs.
//
// A null is allowed: a string cannot hold nil, so the json path reads it as the
// empty string, and so does setStringList. A date-shaped scalar is allowed
// because it is a string in this document's data model, which is what
// stripTimestamp says.
func requireStringScalar(node *yaml.Node, key string) error {
	if node.Kind != yaml.ScalarNode {
		return nil
	}
	switch node.Tag {
	case "", strTag, nullTag, timestampTag:
		return nil
	}
	return fmt.Errorf("yaml: line %d: cannot unmarshal %s into string field %q", node.Line, node.Tag, key)
}

// prepareScalars walks the tree once, which is the one place every scalar is
// seen before the decode reads it.
func prepareScalars(n *yaml.Node) error {
	if n == nil {
		return nil
	}
	if n.Kind == yaml.ScalarNode {
		stripTimestamp(n)
		if err := checkFinite(n); err != nil {
			return err
		}
	}
	for _, c := range n.Content {
		if err := prepareScalars(c); err != nil {
			return err
		}
	}
	return nil
}

// stripTimestamp retags a date-shaped scalar as a string.
//
// YAML 1.1 resolves an untagged scalar such as 2020-06-11T16:32:50-03:00 to a
// timestamp, which would make an OpenAPI `example` of that shape decode to a
// time.Time: not a type the document's own data model has, and one that
// marshals to text that reads back as a string, so the document would not
// survive a round trip. A tag written out asks for the same type and cannot
// have it either, so it is retagged too, and the scalar keeps the text the
// document gave it rather than the time it names.
//
// Interim: https://github.com/yaml/go-yaml/issues/452 asks for an option to
// leave such a scalar a string, which removes this.
func stripTimestamp(n *yaml.Node) {
	if n.Tag == timestampTag {
		n.Tag = strTag
	}
}

// checkFinite rejects infinity and NaN, which YAML has and JSON does not.
// Such a document decodes, and a float field or an extension then holds a
// value that cannot be written back: marshaling fails wherever the value sits,
// naming neither the document nor the line. Refusing it here names both.
func checkFinite(n *yaml.Node) error {
	if n.Tag != "!!float" {
		return nil
	}
	switch strings.ToLower(strings.TrimLeft(n.Value, "+-")) {
	case ".inf", ".nan":
		return fmt.Errorf("line %d: %s has no JSON representation", n.Line, n.Value)
	}
	return nil
}
