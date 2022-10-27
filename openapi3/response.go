package openapi3

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/go-openapi/jsonpointer"

	"github.com/getkin/kin-openapi/jsoninfo"
)

// Responses is specified by OpenAPI/Swagger 3.0 standard.
// See https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md#responsesObject
type Responses map[string]*ResponseRef

var _ jsonpointer.JSONPointable = (*Responses)(nil)

func NewResponses() Responses {
	r := make(Responses)
	r["default"] = &ResponseRef{Value: NewResponse().WithDescription("")}
	return r
}

func (responses Responses) Default() *ResponseRef {
	return responses["default"]
}

func (responses Responses) Get(status int) *ResponseRef {
	return responses[strconv.FormatInt(int64(status), 10)]
}

// Validate returns an error if Responses does not comply with the OpenAPI spec.
func (responses Responses) Validate(ctx context.Context) error {
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

// JSONLookup implements github.com/go-openapi/jsonpointer#JSONPointable
func (responses Responses) JSONLookup(token string) (interface{}, error) {
	ref, ok := responses[token]
	if ok == false {
		return nil, fmt.Errorf("invalid token reference: %q", token)
	}

	if ref != nil && ref.Ref != "" {
		return &Ref{Ref: ref.Ref}, nil
	}
	return ref.Value, nil
}

// Response is specified by OpenAPI/Swagger 3.0 standard.
// See https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md#responseObject
type Response struct {
	ExtensionProps

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
func (response *Response) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStrictStruct(response)
}

// UnmarshalJSON sets Response to a copy of data.
func (response *Response) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStrictStruct(data, response)
}

// Validate returns an error if Response does not comply with the OpenAPI spec.
func (response *Response) Validate(ctx context.Context) error {
	if response.Description == nil {
		return errors.New("a short description of the response is required")
	}
	if vo := getValidationOptions(ctx); !vo.ExamplesValidationDisabled {
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
	return nil
}
