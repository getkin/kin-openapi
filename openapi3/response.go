package openapi3

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/go-openapi/jsonpointer"
)

// Responses is specified by OpenAPI/Swagger 3.0 standard.
// See https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md#responses-object
type Responses map[string]*ResponseRef

var _ jsonpointer.JSONPointable = (*Responses)(nil)

// JSONLookup implements https://pkg.go.dev/github.com/go-openapi/jsonpointer#JSONPointable
func (responses Responses) JSONLookup(token string) (interface{}, error) {
	ref, ok := responses[token]
	if !ok {
		return nil, fmt.Errorf("invalid token reference: %q", token)
	}

	if ref != nil && ref.Ref != "" {
		return &Ref{Ref: ref.Ref}, nil
	}
	return ref.Value, nil
}

func NewResponses() Responses {
	r := make(Responses)
	r["default"] = &ResponseRef{Value: NewResponse().WithDescription("")}
	return r
}

// Default returns the default response
func (responses Responses) Default() *ResponseRef {
	return responses["default"]
}

// Status returns a ResponseRef for the given status
// If an exact match isn't initially found a patterned field is checked using
// the first digit to determine the range (eg: 201 to 2XX)
// See https://spec.openapis.org/oas/v3.0.3#patterned-fields-0
func (responses Responses) Status(status int) *ResponseRef {
	st := strconv.FormatInt(int64(status), 10)
	if rref, ok := responses[st]; ok {
		return rref
	}
	st = string(st[0]) + "XX" //fixme status $ 100
	switch st {
	case "1XX", "2XX", "3XX", "4XX", "5XX":
		return responses[st]
	default:
		return nil
	}
}

// FIXME Any HTTP status code can be used as the property name, but only one property per code, to describe the expected response for that HTTP status code. A Reference Object can link to a response that is defined in the OpenAPI Object's components/responses section. This field MUST be enclosed in quotation marks (for example, "200") for compatibility between JSON and YAML. To define a range of response codes, this field MAY contain the uppercase wildcard character X. For example, 2XX represents all response codes between [200-299]. Only the following range definitions are allowed: 1XX, 2XX, 3XX, 4XX, and 5XX. If a response is defined using an explicit code, the explicit code definition takes precedence over the range definition for that code.

// Validate returns an error if Responses does not comply with the OpenAPI spec.
func (responses Responses) Validate(ctx context.Context, opts ...ValidationOption) error {
	ctx = WithValidationOptions(ctx, opts...)

	if len(responses) == 0 {
		return errors.New("the responses object MUST contain at least one response code")
	}

	keys := make([]string, 0, len(responses))
	for key := range responses {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		v := responses[key]
		if err := v.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Response is specified by OpenAPI/Swagger 3.0 standard.
// See https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md#response-object
type Response struct {
	Extensions map[string]interface{} `json:"-" yaml:"-"`

	Description *string `json:"description,omitempty" yaml:"description,omitempty"`
	Headers     Headers `json:"headers,omitempty" yaml:"headers,omitempty"`
	Content     Content `json:"content,omitempty" yaml:"content,omitempty"`
	Links       Links   `json:"links,omitempty" yaml:"links,omitempty"`
}

func NewResponse() *Response {
	return &Response{}
}

func (response *Response) WithDescription(value string) *Response {
	response.Description = &value
	return response
}

func (response *Response) WithContent(content Content) *Response {
	response.Content = content
	return response
}

func (response *Response) WithJSONSchema(schema *Schema) *Response {
	response.Content = NewContentWithJSONSchema(schema)
	return response
}

func (response *Response) WithJSONSchemaRef(schema *SchemaRef) *Response {
	response.Content = NewContentWithJSONSchemaRef(schema)
	return response
}

// MarshalJSON returns the JSON encoding of Response.
func (response Response) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{}, 4+len(response.Extensions))
	for k, v := range response.Extensions {
		m[k] = v
	}
	if x := response.Description; x != nil {
		m["description"] = x
	}
	if x := response.Headers; len(x) != 0 {
		m["headers"] = x
	}
	if x := response.Content; len(x) != 0 {
		m["content"] = x
	}
	if x := response.Links; len(x) != 0 {
		m["links"] = x
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
	delete(x.Extensions, "description")
	delete(x.Extensions, "headers")
	delete(x.Extensions, "content")
	delete(x.Extensions, "links")
	if len(x.Extensions) == 0 {
		x.Extensions = nil
	}
	*response = Response(x)
	return nil
}

// Validate returns an error if Response does not comply with the OpenAPI spec.
func (response *Response) Validate(ctx context.Context, opts ...ValidationOption) error {
	ctx = WithValidationOptions(ctx, opts...)

	if response.Description == nil {
		return errors.New("a short description of the response is required")
	}
	if vo := getValidationOptions(ctx); !vo.examplesValidationDisabled {
		vo.examplesValidationAsReq, vo.examplesValidationAsRes = false, true
	}

	if content := response.Content; content != nil {
		if err := content.Validate(ctx); err != nil {
			return err
		}
	}

	headers := make([]string, 0, len(response.Headers))
	for name := range response.Headers {
		headers = append(headers, name)
	}
	sort.Strings(headers)
	for _, name := range headers {
		header := response.Headers[name]
		if err := header.Validate(ctx); err != nil {
			return err
		}
	}

	links := make([]string, 0, len(response.Links))
	for name := range response.Links {
		links = append(links, name)
	}
	sort.Strings(links)
	for _, name := range links {
		link := response.Links[name]
		if err := link.Validate(ctx); err != nil {
			return err
		}
	}

	return validateExtensions(ctx, response.Extensions)
}
