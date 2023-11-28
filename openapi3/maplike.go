package openapi3

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/go-openapi/jsonpointer"
)

// Value returns the responses for key or nil
func (responses *Responses) Value(key string) *ResponseRef {
	if responses.Len() == 0 {
		return nil
	}
	return responses.m[key]
}

// Set adds or replaces key 'key' of 'responses' with 'value'.
// Note: 'responses' MUST be non-nil
func (responses *Responses) Set(key string, value *ResponseRef) {
	if responses.m == nil {
		responses.m = make(map[string]*ResponseRef)
	}
	responses.m[key] = value
}

// Len returns the amount of keys in responses excluding responses.Extensions.
func (responses *Responses) Len() int {
	if responses == nil {
		return 0
	}
	return len(responses.m)
}

// Map returns responses as a 'map'.
// Note: iteration on Go maps is not ordered.
func (responses *Responses) Map() map[string]*ResponseRef {
	if responses.Len() == 0 {
		return nil
	}
	return responses.m
}

var _ jsonpointer.JSONPointable = (*Responses)(nil)

// JSONLookup implements https://github.com/go-openapi/jsonpointer#JSONPointable
func (responses Responses) JSONLookup(token string) (interface{}, error) {
	if v := responses.Value(token); v == nil {
		vv, _, err := jsonpointer.GetForToken(responses.Extensions, token)
		return vv, err
	} else if ref := v.Ref; ref != "" {
		return &Ref{Ref: ref}, nil
	} else {
		var vv *Response = v.Value
		return vv, nil
	}
}

// MarshalJSON returns the JSON encoding of Responses.
func (responses Responses) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{}, responses.Len()+len(responses.Extensions))
	for k, v := range responses.Extensions {
		m[k] = v
	}
	for k, v := range responses.Map() {
		m[k] = v
	}
	return json.Marshal(m)
}

// UnmarshalJSON sets Responses to a copy of data.
func (responses *Responses) UnmarshalJSON(data []byte) (err error) {
	var m map[string]interface{}
	if err = json.Unmarshal(data, &m); err != nil {
		return
	}

	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)

	x := Responses{
		Extensions: make(map[string]interface{}),
		m:          make(map[string]*ResponseRef, len(m)),
	}

	for _, k := range ks {
		v := m[k]
		if strings.HasPrefix(k, "x-") {
			x.Extensions[k] = v
			continue
		}

		var data []byte
		if data, err = json.Marshal(v); err != nil {
			return
		}
		var vv ResponseRef
		if err = vv.UnmarshalJSON(data); err != nil {
			return
		}
		x.m[k] = &vv
	}
	*responses = x
	return
}

// Value returns the callback for key or nil
func (callback *Callback) Value(key string) *PathItem {
	if callback.Len() == 0 {
		return nil
	}
	return callback.m[key]
}

// Set adds or replaces key 'key' of 'callback' with 'value'.
// Note: 'callback' MUST be non-nil
func (callback *Callback) Set(key string, value *PathItem) {
	if callback.m == nil {
		callback.m = make(map[string]*PathItem)
	}
	callback.m[key] = value
}

// Len returns the amount of keys in callback excluding callback.Extensions.
func (callback *Callback) Len() int {
	if callback == nil {
		return 0
	}
	return len(callback.m)
}

// Map returns callback as a 'map'.
// Note: iteration on Go maps is not ordered.
func (callback *Callback) Map() map[string]*PathItem {
	if callback.Len() == 0 {
		return nil
	}
	return callback.m
}

var _ jsonpointer.JSONPointable = (*Callback)(nil)

// JSONLookup implements https://github.com/go-openapi/jsonpointer#JSONPointable
func (callback Callback) JSONLookup(token string) (interface{}, error) {
	if v := callback.Value(token); v == nil {
		vv, _, err := jsonpointer.GetForToken(callback.Extensions, token)
		return vv, err
	} else if ref := v.Ref; ref != "" {
		return &Ref{Ref: ref}, nil
	} else {
		var vv *PathItem = v
		return vv, nil
	}
}

// MarshalJSON returns the JSON encoding of Callback.
func (callback Callback) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{}, callback.Len()+len(callback.Extensions))
	for k, v := range callback.Extensions {
		m[k] = v
	}
	for k, v := range callback.Map() {
		m[k] = v
	}
	return json.Marshal(m)
}

// UnmarshalJSON sets Callback to a copy of data.
func (callback *Callback) UnmarshalJSON(data []byte) (err error) {
	var m map[string]interface{}
	if err = json.Unmarshal(data, &m); err != nil {
		return
	}

	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)

	x := Callback{
		Extensions: make(map[string]interface{}),
		m:          make(map[string]*PathItem, len(m)),
	}

	for _, k := range ks {
		v := m[k]
		if strings.HasPrefix(k, "x-") {
			x.Extensions[k] = v
			continue
		}

		var data []byte
		if data, err = json.Marshal(v); err != nil {
			return
		}
		var vv PathItem
		if err = vv.UnmarshalJSON(data); err != nil {
			return
		}
		x.m[k] = &vv
	}
	*callback = x
	return
}

// Value returns the paths for key or nil
func (paths *Paths) Value(key string) *PathItem {
	if paths.Len() == 0 {
		return nil
	}
	return paths.m[key]
}

// Set adds or replaces key 'key' of 'paths' with 'value'.
// Note: 'paths' MUST be non-nil
func (paths *Paths) Set(key string, value *PathItem) {
	if paths.m == nil {
		paths.m = make(map[string]*PathItem)
	}
	paths.m[key] = value
}

// Len returns the amount of keys in paths excluding paths.Extensions.
func (paths *Paths) Len() int {
	if paths == nil {
		return 0
	}
	return len(paths.m)
}

// Map returns paths as a 'map'.
// Note: iteration on Go maps is not ordered.
func (paths *Paths) Map() map[string]*PathItem {
	if paths.Len() == 0 {
		return nil
	}
	return paths.m
}

var _ jsonpointer.JSONPointable = (*Paths)(nil)

// JSONLookup implements https://github.com/go-openapi/jsonpointer#JSONPointable
func (paths Paths) JSONLookup(token string) (interface{}, error) {
	if v := paths.Value(token); v == nil {
		vv, _, err := jsonpointer.GetForToken(paths.Extensions, token)
		return vv, err
	} else if ref := v.Ref; ref != "" {
		return &Ref{Ref: ref}, nil
	} else {
		var vv *PathItem = v
		return vv, nil
	}
}

// MarshalJSON returns the JSON encoding of Paths.
func (paths Paths) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{}, paths.Len()+len(paths.Extensions))
	for k, v := range paths.Extensions {
		m[k] = v
	}
	for k, v := range paths.Map() {
		m[k] = v
	}
	return json.Marshal(m)
}

// UnmarshalJSON sets Paths to a copy of data.
func (paths *Paths) UnmarshalJSON(data []byte) (err error) {
	var m map[string]interface{}
	if err = json.Unmarshal(data, &m); err != nil {
		return
	}

	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)

	x := Paths{
		Extensions: make(map[string]interface{}),
		m:          make(map[string]*PathItem, len(m)),
	}

	for _, k := range ks {
		v := m[k]
		if strings.HasPrefix(k, "x-") {
			x.Extensions[k] = v
			continue
		}

		var data []byte
		if data, err = json.Marshal(v); err != nil {
			return
		}
		var vv PathItem
		if err = vv.UnmarshalJSON(data); err != nil {
			return
		}
		x.m[k] = &vv
	}
	*paths = x
	return
}
