package openapi3

// Origin records where each element of a document came from: the position of
// the key that heads a collection, of each of its fields, and of the scalar
// items in its sequence-valued fields.
//
// The positions are read from the nodes as the document decodes. A node knows
// where it starts, so the only piece it cannot supply is Key -- the key above
// it belongs to the parent, which stamps it.

import (
	"encoding/json"
	"reflect"
	"slices"
	"sort"
	"strings"
	"sync"

	yaml "go.yaml.in/yaml/v3"
)

var originPtrType = reflect.TypeFor[*Origin]()

// Origin contains the origin of a collection.
// Key is the location of the collection itself.
// Fields holds the location of each scalar field in the collection.
// Sequences is a map of the location of each item in sequence-valued fields.
//
// Sequences stays a map although Fields is a slice, which is deliberate.
// FieldLocations drops the map because Location.Name already carries the key,
// so the map was storing information the value repeated. Here Location.Name
// holds the *item's* value (an enum member, a required property) while the key
// is the *field's* name ("enum", "required", "tags"), so a slice would need a
// wrapper type invented to hold it. The memory argument is also much weaker:
// only a collection with a sequence-valued field allocates one at all, which
// measured at 5% of collections on a large spec, and a nil map is free.
type Origin struct {
	Key       *Location             `json:"key,omitempty" yaml:"key,omitempty"`
	Fields    FieldLocations        `json:"fields,omitempty" yaml:"fields,omitempty"`
	Sequences map[string][]Location `json:"sequences,omitempty" yaml:"sequences,omitempty"`
}

// FieldLocations holds the locations of a collection's scalar fields, in the
// order they appear in the document.
//
// It is a slice rather than a map[string]Location because a collection carries
// only a handful of fields, while a Go map allocates a whole bucket per
// collection whatever it holds. On a large document that overhead dominated
// the retained size of a parsed spec. Each Location already carries its Name,
// so the lookup key costs nothing extra here.
type FieldLocations []Location

// Get returns the location of the named field, or the zero Location when the
// field has none. Use Lookup to tell an absent field from a zero location.
func (f FieldLocations) Get(name string) Location {
	loc, _ := f.Lookup(name)
	return loc
}

// Lookup returns the location of the named field and whether it was found.
// The scan is linear: collections have few fields, and a linear scan over a
// contiguous slice beats a map lookup at these sizes.
//
// Deliberately a hand-written loop rather than slices.IndexFunc: the closure
// does not inline, so IndexFunc pays a call per element. Measured on 3/6/12
// fields it is 5-100% slower on a hit and 2-3x slower on a miss, and misses
// are the common case here (most fields carry no recorded location).
func (f FieldLocations) Lookup(name string) (Location, bool) {
	for i := range f {
		if f[i].Name == name {
			return f[i], true
		}
	}
	return Location{}, false
}

// MarshalJSON keeps the serialized shape a name-keyed object, as it was when
// this was a map, so the change is invisible to anything reading the output.
func (f FieldLocations) MarshalJSON() ([]byte, error) {
	m := make(map[string]Location, len(f))
	for _, loc := range f {
		m[loc.Name] = loc
	}
	return json.Marshal(m)
}

// UnmarshalJSON reads the name-keyed object written by MarshalJSON. Entries are
// sorted by name, since a JSON object carries no order to restore.
func (f *FieldLocations) UnmarshalJSON(data []byte) error {
	var m map[string]Location
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	out := make(FieldLocations, 0, len(m))
	for name, loc := range m {
		if loc.Name == "" {
			loc.Name = name
		}
		out = append(out, loc)
	}
	slices.SortFunc(out, func(a, b Location) int { return strings.Compare(a.Name, b.Name) })
	*f = out
	return nil
}

// Location is a struct that contains the location of a field.
type Location struct {
	File   string `json:"file,omitempty" yaml:"file,omitempty"`
	Line   int    `json:"line,omitempty" yaml:"line,omitempty"`
	Column int    `json:"column,omitempty" yaml:"column,omitempty"`
	Name   string `json:"name,omitempty" yaml:"name,omitempty"`

	// EndLine and EndColumn mark the end of the block this location heads (set
	// only on Origin.Key). For an operation or schema this spans the whole
	// block, so a consumer can extract the entire element from its source.
	// Both are zero when the underlying YAML carried no end information.
	EndLine   int `json:"endLine,omitempty" yaml:"endLine,omitempty"`
	EndColumn int `json:"endColumn,omitempty" yaml:"endColumn,omitempty"`
}

// originTree is a decoded document's node tree together with what the origins
// read off it were stamped against. A $ref resolved later decodes a subtree of
// this, and must stamp the file the subtree came from rather than whichever
// document happened to be decoded last.
type originTree struct {
	node *yaml.Node
	file string
	ends *endIndex
}

// originMu guards the three package-level variables below for the length of a
// decode. They exist because UnmarshalYAML receives a node and nothing else,
// with no way to carry per-decode state through the call, and a lock is what
// makes them safe to hold that way.
//
// The cost is that decodes serialise even when each has its own Loader, which
// TestIssue741 does. Correctness first: without this the three race, and the
// path this replaces passed the file as an argument and did not.
//
// A file recorded on the node itself would remove the need, since the callback
// already receives the node, but go-yaml's Node has no field for it.
var originMu sync.Mutex

// originFileVar is the file stamped into origins for the decode in progress.
// UnmarshalYAML receives a node and nothing else, so the file cannot be passed
// through the call.
var originFileVar string

// originEnabledVar mirrors the includeOrigin argument unmarshal receives, which
// comes from the Loader. The package-level IncludeOrigin only seeds NewLoader,
// so a caller that set it on its Loader alone would be missed.
var originEnabledVar bool

// Shared machinery for the UnmarshalYAML methods: extension collection, and
// origins read from the node being decoded.
//
// Origins record where an element starts, not where it ends. A consumer that
// needs the extent of a block derives it from the next key or sequence item at
// the same or shallower indentation.

func nativeOriginFile() string { return originFileVar }

// mappingValue returns the value node for key, or nil.
func mappingValue(node *yaml.Node, key string) *yaml.Node {
	_, v := mappingEntry(node, key)
	return v
}

// mappingEntry returns both halves of a mapping entry, the key being what a
// value's origin records as its own location. Merge keys are applied, so a
// field a merge brought in is found here the same as one written in place.
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

// originFromNode builds the origin data a mapping can see for itself: where
// each of its field keys is, and where the scalar items of its sequence-valued
// fields are.
//
// Origin.Key is not set here -- it is the location of the key heading this
// mapping in its parent, which a node does not know. See setChildOriginKeys.
func originFromNode(node *yaml.Node, file string) *Origin {
	// Origins are opt-in: without this every decode pays for them.
	if !originEnabledVar {
		return nil
	}
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	o := &Origin{}
	for i := 0; i+1 < len(node.Content); i += 2 {
		k, v := node.Content[i], node.Content[i+1]
		if o.Fields == nil {
			o.Fields = make(FieldLocations, 0, len(node.Content)/2)
		}
		o.Fields = append(o.Fields, Location{File: file, Line: k.Line, Column: k.Column, Name: k.Value})

		if v.Kind != yaml.SequenceNode {
			continue
		}
		var locs []Location
		for _, item := range v.Content {
			if item.Kind == yaml.ScalarNode {
				locs = append(locs, Location{File: file, Line: item.Line, Column: item.Column, Name: item.Value})
			}
		}
		if len(locs) > 0 {
			if o.Sequences == nil {
				o.Sequences = make(map[string][]Location)
			}
			o.Sequences[k.Value] = locs
		}
	}
	if o.Fields == nil && o.Sequences == nil {
		return nil
	}
	return o
}

// setChildOriginKeys sets Origin.Key on the immediate children of a mapping,
// from the key node heading each one.
//
// This is the only origin data a node cannot supply for itself: UnmarshalYAML
// receives the value node, and Key is the position of the key above it. Each
// child sets its own children's keys in turn, so one level per call covers the
// tree.
func setChildOriginKeys(node *yaml.Node, container any, file string) {
	if !originEnabledVar {
		return
	}
	if node == nil || node.Kind != yaml.MappingNode {
		return
	}
	v := reflect.ValueOf(container)
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		keyNode, valNode := node.Content[i], node.Content[i+1]
		child := childByKey(v, keyNode.Value)
		if !child.IsValid() {
			continue
		}
		setOriginKey(child, keyNode, valNode, file)
		recordScalarMapKeys(v, child, keyNode, valNode, file)

		switch c := deref(child); c.Kind() {
		case reflect.Map:
			// A map-valued field (Content, Headers, Links) holds children of
			// its own, keyed in valNode. The generic map decoder gives them no
			// hook of their own, so descend.
			if c.CanInterface() {
				setChildOriginKeys(valNode, c.Interface(), file)
			}
		case reflect.Slice:
			// A sequence item has no key above it, so it takes its own first
			// key as its Key.
			if valNode.Kind != yaml.SequenceNode {
				continue
			}
			for j := 0; j < len(valNode.Content) && j < c.Len(); j++ {
				item := valNode.Content[j]
				if item.Kind == yaml.MappingNode && len(item.Content) > 0 {
					setOriginKey(c.Index(j), item.Content[0], item, file)
				}
			}
		}
	}
}

func deref(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return v
		}
		v = v.Elem()
	}
	return v
}

// childByKey finds the struct field or map entry a mapping key decoded into.
func childByKey(v reflect.Value, key string) reflect.Value {
	switch v.Kind() {
	case reflect.Map:
		if v.IsNil() {
			return reflect.Value{}
		}
		return v.MapIndex(reflect.ValueOf(key))
	case reflect.Struct:
		t := v.Type()
		for i := range t.NumField() {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			if name, _, _ := strings.Cut(f.Tag.Get("yaml"), ","); name == key {
				return v.Field(i)
			}
		}
	}
	return reflect.Value{}
}

// setOriginKey stamps Key on a child carrying an *Origin, from the key's own
// position. The extent of what the key heads is the consumer's to derive.
func setOriginKey(child reflect.Value, keyNode, valNode *yaml.Node, file string) {
	if !originEnabledVar {
		return
	}
	keyNode, valNode = resolveAlias(keyNode, valNode)
	for child.Kind() == reflect.Pointer || child.Kind() == reflect.Interface {
		if child.IsNil() {
			return
		}
		child = child.Elem()
	}
	if child.Kind() != reflect.Struct {
		return
	}
	f := child.FieldByName("Origin")
	if !f.IsValid() || f.Type() != originPtrType || !f.CanSet() {
		// No origin of its own; it may still wrap something that has one.
		descendToWrapped(child, keyNode, valNode, file)
		return
	}
	if f.IsNil() {
		f.Set(reflect.ValueOf(&Origin{}))
	}
	key := withEnd(Location{
		File:   file,
		Line:   keyNode.Line,
		Column: keyNode.Column,
		Name:   keyNode.Value,
	}, valNode)
	f.Interface().(*Origin).Key = &key
	// A wrapper and the thing it holds occupy the same node, so both carry
	// that node's origin. Value is the $ref wrappers; Schema is BoolSchema,
	// which holds either a bool or a schema.
	descendToWrapped(child, keyNode, valNode, file)
}

// descendToWrapped stamps the thing a wrapper holds, which occupies the same
// node. Value is the $ref wrappers; Schema is BoolSchema, which holds either a
// bool or a schema.
func descendToWrapped(child reflect.Value, keyNode, valNode *yaml.Node, file string) {
	if child.Kind() != reflect.Struct {
		return
	}
	for _, name := range [...]string{"Value", "Schema"} {
		if inner := child.FieldByName(name); inner.IsValid() {
			setOriginKey(inner, keyNode, valNode, file)
		}
	}
}

// stampRootOrigin gives a document root the position of the document itself.
//
// Origin.Key is normally the key heading a mapping in its parent, stamped by
// that parent. A root has none -- an externally $ref'd file may be a bare
// schema -- so it takes the root node's own position and an empty name.
// Applied only when nothing has already set Key, so a type that supplies its
// own keeps it.
func stampRootOrigin(v any, node *yaml.Node) {
	if !originEnabledVar || node == nil {
		return
	}
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return
	}
	f := rv.FieldByName("Origin")
	if !f.IsValid() || f.Type() != originPtrType || !f.CanSet() || f.IsNil() {
		return
	}
	o := f.Interface().(*Origin)
	if o.Key != nil {
		return
	}
	key := withEnd(Location{File: nativeOriginFile(), Line: node.Line, Column: node.Column}, node)
	o.Key = &key
}

// recordScalarMapKeys records where each key of a scalar-valued map sits.
//
// A map[string]string -- scopes on an OAuth flow, say -- decodes to a plain Go
// map with nowhere to hang an Origin of its own, so its keys are recorded on
// the enclosing struct's Origin under the field name, sorted by key.
func recordScalarMapKeys(container, child reflect.Value, keyNode, valNode *yaml.Node, file string) {
	if child.Kind() != reflect.Map || valNode == nil || valNode.Kind != yaml.MappingNode {
		return
	}
	// A map of structs or pointers carries origins on its values instead.
	switch child.Type().Elem().Kind() {
	case reflect.Struct, reflect.Pointer, reflect.Interface, reflect.Map, reflect.Slice:
		return
	}
	f := container.FieldByName("Origin")
	if !f.IsValid() || f.Type() != originPtrType || !f.CanSet() {
		return
	}
	if f.IsNil() {
		f.Set(reflect.ValueOf(&Origin{}))
	}
	var locs []Location
	for i := 0; i+1 < len(valNode.Content); i += 2 {
		k := valNode.Content[i]
		locs = append(locs, Location{File: file, Line: k.Line, Column: k.Column, Name: k.Value})
	}
	if len(locs) == 0 {
		return
	}
	sort.Slice(locs, func(i, j int) bool { return locs[i].Name < locs[j].Name })
	o := f.Interface().(*Origin)
	if o.Sequences == nil {
		o.Sequences = make(map[string][]Location)
	}
	o.Sequences[keyNode.Value] = locs
}
