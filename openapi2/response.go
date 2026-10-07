package openapi2

import (
	"encoding/json"
	"maps"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

type Response struct {
	Extensions map[string]any `json:"-" yaml:"-"`

	Ref string `json:"$ref,omitempty" yaml:"$ref,omitempty"`

	Description string             `json:"description,omitempty" yaml:"description,omitempty"`
	Schema      *SchemaRef         `json:"schema,omitempty" yaml:"schema,omitempty"`
	Headers     map[string]*Header `json:"headers,omitempty" yaml:"headers,omitempty"`
	Examples    map[string]any     `json:"examples,omitempty" yaml:"examples,omitempty"`
}

// MarshalJSON returns the JSON encoding of Response.
func (response Response) MarshalJSON() ([]byte, error) {
	if ref := response.Ref; ref != "" {
		return json.Marshal(openapi3.Ref{Ref: ref})
	}

	m := make(map[string]any, 4+len(response.Extensions))
	maps.Copy(m, response.Extensions)
	if x := response.Description; x != "" {
		m["description"] = x
	}
	if x := response.Schema; x != nil {
		m["schema"] = x
	}
	if x := response.Headers; len(x) != 0 {
		m["headers"] = x
	}
	if x := response.Examples; len(x) != 0 {
		m["examples"] = x
	}
	return json.Marshal(m)
}

// UnmarshalJSON sets Response to a copy of data.
func (response *Response) UnmarshalJSON(data []byte) error {
	type ResponseBis Response
	var x ResponseBis
	if err := json.Unmarshal(data, &x); err != nil {
		return unmarshalError(err)
	}
	_ = json.Unmarshal(data, &x.Extensions)
	delete(x.Extensions, "$ref")
	delete(x.Extensions, "description")
	delete(x.Extensions, "schema")
	delete(x.Extensions, "headers")
	delete(x.Extensions, "examples")
	if len(x.Extensions) == 0 {
		x.Extensions = nil
	}
	*response = Response(x)
	return nil
}

// Responses is specified by OpenAPI/Swagger 2.0 standard.
// See https://github.com/OAI/OpenAPI-Specification/blob/main/versions/2.0.md#responses-object
type Responses struct {
	Extensions map[string]any `json:"-" yaml:"-"`

	m map[string]*Response
}

// NewResponsesWithCapacity builds a responses object of the given capacity.
func NewResponsesWithCapacity(cap int) *Responses {
	return &Responses{m: make(map[string]*Response, cap)}
}

// Value returns the response for key or nil
func (responses *Responses) Value(key string) *Response {
	if responses.Len() == 0 {
		return nil
	}
	return responses.m[key]
}

// Set adds or replaces key 'key' of 'responses' with 'value'.
// Note: 'responses' MUST be non-nil
func (responses *Responses) Set(key string, value *Response) {
	if responses.m == nil {
		responses.m = make(map[string]*Response)
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

// Delete removes the entry associated with key 'key' from 'responses'.
func (responses *Responses) Delete(key string) {
	if responses != nil {
		delete(responses.m, key)
	}
}

// Map returns responses as a 'map'.
// Note: iteration on Go maps is not ordered.
func (responses *Responses) Map() map[string]*Response {
	m := make(map[string]*Response, responses.Len())
	if responses != nil {
		maps.Copy(m, responses.m)
	}
	return m
}

// MarshalJSON returns the JSON encoding of Responses.
func (responses Responses) MarshalJSON() ([]byte, error) {
	m := make(map[string]any, len(responses.m)+len(responses.Extensions))
	maps.Copy(m, responses.Extensions)
	for k, v := range responses.m {
		m[k] = v
	}
	return json.Marshal(m)
}

// UnmarshalJSON sets Responses to a copy of data.
func (responses *Responses) UnmarshalJSON(data []byte) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return unmarshalError(err)
	}
	x := Responses{m: make(map[string]*Response, len(m))}
	for k, v := range m {
		if strings.HasPrefix(k, "x-") {
			if x.Extensions == nil {
				x.Extensions = make(map[string]any)
			}
			var ext any
			if err := json.Unmarshal(v, &ext); err != nil {
				return unmarshalError(err)
			}
			x.Extensions[k] = ext
			continue
		}
		var response *Response
		if err := json.Unmarshal(v, &response); err != nil {
			return err
		}
		x.m[k] = response
	}
	*responses = x
	return nil
}
