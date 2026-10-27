package openapi3

import (
	"bytes"
	"encoding/json"
	"maps"
	"reflect"
	"slices"
)

// OrderedObject is a JSON object that keeps the order of its members.
// Schema validation accepts it wherever it accepts a map[string]any,
// and reports the errors of its members in that order.
type OrderedObject struct {
	names   []string
	members map[string]any
}

// NewOrderedObject returns an empty OrderedObject.
func NewOrderedObject() *OrderedObject {
	return &OrderedObject{members: make(map[string]any)}
}

// Set sets the member name to value, adding it after the others if it is new.
func (o *OrderedObject) Set(name string, value any) {
	if _, ok := o.members[name]; !ok {
		o.names = append(o.names, name)
	}
	o.members[name] = value
}

// Get returns the member name.
func (o *OrderedObject) Get(name string) (any, bool) {
	value, ok := o.members[name]
	return value, ok
}

// Len returns the number of members.
func (o *OrderedObject) Len() int {
	return len(o.names)
}

// Names returns the names of the members in order.
func (o *OrderedObject) Names() []string {
	return slices.Clone(o.names)
}

// MarshalJSON encodes the members in order.
func (o *OrderedObject) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, name := range o.names {
		if i > 0 {
			buf.WriteByte(',')
		}
		key, err := json.Marshal(name)
		if err != nil {
			return nil, err
		}
		value, err := json.Marshal(o.members[name])
		if err != nil {
			return nil, err
		}
		buf.Write(key)
		buf.WriteByte(':')
		buf.Write(value)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// unordered returns value with every OrderedObject in it replaced by a map.
// It copies only what contains an OrderedObject.
func unordered(value any) any {
	switch value := value.(type) {
	case *OrderedObject:
		m := make(map[string]any, len(value.members))
		for name, member := range value.members {
			m[name] = unordered(member)
		}
		return m
	case []any:
		var s []any
		for i, item := range value {
			if u := unordered(item); s != nil || !sameJSONValue(u, item) {
				if s == nil {
					s = slices.Clone(value[:i])
				}
				s = append(s, u)
			}
		}
		if s == nil {
			return value
		}
		return s
	case map[string]any:
		var m map[string]any
		for name, member := range value {
			if u := unordered(member); !sameJSONValue(u, member) {
				if m == nil {
					m = maps.Clone(value)
				}
				m[name] = u
			}
		}
		if m == nil {
			return value
		}
		return m
	}
	return value
}

// sameJSONValue reports whether unordered returned value unchanged.
func sameJSONValue(u, value any) bool {
	switch value := value.(type) {
	case *OrderedObject:
		return false
	case []any:
		u := u.([]any)
		return len(value) == 0 || &u[0] == &value[0]
	case map[string]any:
		return reflect.ValueOf(u).Pointer() == reflect.ValueOf(value).Pointer()
	}
	return true
}
