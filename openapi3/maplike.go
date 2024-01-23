package openapi3

import (
	"encoding/json"
	"strings"

	"github.com/go-openapi/jsonpointer"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// NewResponsesWithCapacity builds a responses object of the given capacity.
func NewResponsesWithCapacity(cap int) *Responses {
	if cap == 0 {
		return &Responses{om: orderedmap.New[string, *ResponseRef]()}
	}
	return &Responses{om: orderedmap.New[string, *ResponseRef](cap)}
}

// Value returns the responses for key or nil
func (responses *Responses) Value(key string) *ResponseRef {
	if responses.Len() == 0 {
		return nil
	}
	return responses.om.Value(key)
}

// Set adds or replaces key 'key' of 'responses' with 'value'.
// Note: 'responses' MUST be non-nil
func (responses *Responses) Set(key string, value *ResponseRef) {
	if responses.om == nil {
		responses.om = NewResponsesWithCapacity(0).om
	}
	_, _ = responses.om.Set(key, value)
}

// Len returns the amount of keys in responses excluding responses.Extensions.
func (responses *Responses) Len() int {
	if responses == nil || responses.om == nil {
		return 0
	}
	return responses.om.Len()
}

// Map returns responses as a 'map'.
// Note: iteration on Go maps is not ordered.
func (responses *Responses) Map() (m map[string]*ResponseRef) {
	if responses == nil || responses.om == nil {
		return make(map[string]*ResponseRef)
	}
	m = make(map[string]*ResponseRef, responses.Len())
	for pair := responses.Iter(); pair != nil; pair = pair.Next() {
		m[pair.Key] = pair.Value
	}
	return
}

type responsesKV orderedmap.Pair[string, *ResponseRef] //FIXME: pub?
// Iter returns a pointer to the first pair, in insertion order.
func (responses *Responses) Iter() *responsesKV {
	if responses.Len() == 0 {
		return nil
	}
	return (*responsesKV)(responses.om.Oldest())
}

// Next returns a pointer to the next pair, in insertion order.
func (pair *responsesKV) Next() *responsesKV {
	ompair := (*orderedmap.Pair[string, *ResponseRef])(pair)
	return (*responsesKV)(ompair.Next())
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
func (responses *Responses) MarshalJSON() ([]byte, error) {
	om := orderedmap.New[string, interface{}](responses.Len() + len(responses.Extensions))
	for pair := responses.Iter(); pair != nil; pair = pair.Next() {
		om.Set(pair.Key, pair.Value)
	}
	for k, v := range responses.Extensions {
		om.Set(k, v)
	}
	return om.MarshalJSON()
}

// UnmarshalJSON sets Responses to a copy of data.
func (responses *Responses) UnmarshalJSON(data []byte) (err error) {
	om := orderedmap.New[string, interface{}]()
	if err = json.Unmarshal(data, &om); err != nil {
		return
	}

	x := NewResponsesWithCapacity(om.Len())
	x.Extensions = make(map[string]interface{})

	for pair := om.Oldest(); pair != nil; pair = pair.Next() {
		k, v := pair.Key, pair.Value
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
		x.Set(k, &vv)
	}
	*responses = *x
	return
}

// NewCallbackWithCapacity builds a callback object of the given capacity.
func NewCallbackWithCapacity(cap int) *Callback {
	if cap == 0 {
		return &Callback{om: orderedmap.New[string, *PathItem]()}
	}
	return &Callback{om: orderedmap.New[string, *PathItem](cap)}
}

// Value returns the callback for key or nil
func (callback *Callback) Value(key string) *PathItem {
	if callback.Len() == 0 {
		return nil
	}
	return callback.om.Value(key)
}

// Set adds or replaces key 'key' of 'callback' with 'value'.
// Note: 'callback' MUST be non-nil
func (callback *Callback) Set(key string, value *PathItem) {
	if callback.om == nil {
		callback.om = NewCallbackWithCapacity(0).om
	}
	_, _ = callback.om.Set(key, value)
}

// Len returns the amount of keys in callback excluding callback.Extensions.
func (callback *Callback) Len() int {
	if callback == nil || callback.om == nil {
		return 0
	}
	return callback.om.Len()
}

// Map returns callback as a 'map'.
// Note: iteration on Go maps is not ordered.
func (callback *Callback) Map() (m map[string]*PathItem) {
	if callback == nil || callback.om == nil {
		return make(map[string]*PathItem)
	}
	m = make(map[string]*PathItem, callback.Len())
	for pair := callback.Iter(); pair != nil; pair = pair.Next() {
		m[pair.Key] = pair.Value
	}
	return
}

type callbackKV orderedmap.Pair[string, *PathItem] //FIXME: pub?
// Iter returns a pointer to the first pair, in insertion order.
func (callback *Callback) Iter() *callbackKV {
	if callback.Len() == 0 {
		return nil
	}
	return (*callbackKV)(callback.om.Oldest())
}

// Next returns a pointer to the next pair, in insertion order.
func (pair *callbackKV) Next() *callbackKV {
	ompair := (*orderedmap.Pair[string, *PathItem])(pair)
	return (*callbackKV)(ompair.Next())
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
func (callback *Callback) MarshalJSON() ([]byte, error) {
	om := orderedmap.New[string, interface{}](callback.Len() + len(callback.Extensions))
	for pair := callback.Iter(); pair != nil; pair = pair.Next() {
		om.Set(pair.Key, pair.Value)
	}
	for k, v := range callback.Extensions {
		om.Set(k, v)
	}
	return om.MarshalJSON()
}

// UnmarshalJSON sets Callback to a copy of data.
func (callback *Callback) UnmarshalJSON(data []byte) (err error) {
	om := orderedmap.New[string, interface{}]()
	if err = json.Unmarshal(data, &om); err != nil {
		return
	}

	x := NewCallbackWithCapacity(om.Len())
	x.Extensions = make(map[string]interface{})

	for pair := om.Oldest(); pair != nil; pair = pair.Next() {
		k, v := pair.Key, pair.Value
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
		x.Set(k, &vv)
	}
	*callback = *x
	return
}

// NewPathsWithCapacity builds a paths object of the given capacity.
func NewPathsWithCapacity(cap int) *Paths {
	if cap == 0 {
		return &Paths{om: orderedmap.New[string, *PathItem]()}
	}
	return &Paths{om: orderedmap.New[string, *PathItem](cap)}
}

// Value returns the paths for key or nil
func (paths *Paths) Value(key string) *PathItem {
	if paths.Len() == 0 {
		return nil
	}
	return paths.om.Value(key)
}

// Set adds or replaces key 'key' of 'paths' with 'value'.
// Note: 'paths' MUST be non-nil
func (paths *Paths) Set(key string, value *PathItem) {
	if paths.om == nil {
		paths.om = NewPathsWithCapacity(0).om
	}
	_, _ = paths.om.Set(key, value)
}

// Len returns the amount of keys in paths excluding paths.Extensions.
func (paths *Paths) Len() int {
	if paths == nil || paths.om == nil {
		return 0
	}
	return paths.om.Len()
}

// Map returns paths as a 'map'.
// Note: iteration on Go maps is not ordered.
func (paths *Paths) Map() (m map[string]*PathItem) {
	if paths == nil || paths.om == nil {
		return make(map[string]*PathItem)
	}
	m = make(map[string]*PathItem, paths.Len())
	for pair := paths.Iter(); pair != nil; pair = pair.Next() {
		m[pair.Key] = pair.Value
	}
	return
}

type pathsKV orderedmap.Pair[string, *PathItem] //FIXME: pub?
// Iter returns a pointer to the first pair, in insertion order.
func (paths *Paths) Iter() *pathsKV {
	if paths.Len() == 0 {
		return nil
	}
	return (*pathsKV)(paths.om.Oldest())
}

// Next returns a pointer to the next pair, in insertion order.
func (pair *pathsKV) Next() *pathsKV {
	ompair := (*orderedmap.Pair[string, *PathItem])(pair)
	return (*pathsKV)(ompair.Next())
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
func (paths *Paths) MarshalJSON() ([]byte, error) {
	om := orderedmap.New[string, interface{}](paths.Len() + len(paths.Extensions))
	for pair := paths.Iter(); pair != nil; pair = pair.Next() {
		om.Set(pair.Key, pair.Value)
	}
	for k, v := range paths.Extensions {
		om.Set(k, v)
	}
	return om.MarshalJSON()
}

// UnmarshalJSON sets Paths to a copy of data.
func (paths *Paths) UnmarshalJSON(data []byte) (err error) {
	om := orderedmap.New[string, interface{}]()
	if err = json.Unmarshal(data, &om); err != nil {
		return
	}

	x := NewPathsWithCapacity(om.Len())
	x.Extensions = make(map[string]interface{})

	for pair := om.Oldest(); pair != nil; pair = pair.Next() {
		k, v := pair.Key, pair.Value
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
		x.Set(k, &vv)
	}
	*paths = *x
	return
}
