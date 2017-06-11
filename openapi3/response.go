package openapi3

import (
	"context"
	"fmt"
	"github.com/jban332/kinapi/jsoninfo"
)

// Responses is specified by OpenAPI/Swagger 3.0 standard.
type Responses map[string]*Response

func NewResponses() Responses {
	return make(Responses, 8)
}

func (responses Responses) Default() *Response {
	return responses["default"]
}

func (responses Responses) Get(status int) *Response {
	return responses[fmt.Sprint(status)]
}

func (all Responses) Validate(c context.Context) error {
	for _, resp := range all {
		err := resp.Validate(c)
		if err != nil {
			return err
		}
	}
	return nil
}

// Response is specified by OpenAPI/Swagger 3.0 standard.
type Response struct {
	jsoninfo.RefProps
	jsoninfo.ExtensionProps

	Description string  `json:"description,omitempty"`
	Content     Content `json:"content,omitempty"`
}

func NewResponse() *Response {
	return &Response{}
}

func (response *Response) WithDescription(value string) *Response {
	response.Description = value
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

func (value *Response) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStructFields(value)
}

func (value *Response) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStructFields(data, value)
}

func (response *Response) Validate(c context.Context) error {
	if content := response.Content; content != nil {
		if err := content.Validate(c); err != nil {
			return err
		}
	}
	return nil
}
