package openapi3

import (
	"context"
	"github.com/jban332/kinapi/jsoninfo"
	"strconv"
)

// Operation represents "operation" specified by" OpenAPI/Swagger 3.0 standard.
type Operation struct {
	jsoninfo.ExtensionProps

	// Optional ID.
	ID string `json:"id,omitempty"`

	// Optional summary what the operation does.
	Summary string `json:"summary,omitempty"`

	// Optional description what the operation does.
	Description string `json:"description,omitempty"`

	// Optional tags for documentation.
	Tags []string `json:"tags,omitempty"`

	// Optional security requirements.
	// If the pointer is nil, the security is the application default security.
	Security *SecurityRequirements `json:"security,omitempty"`

	// Optional parameters.
	Parameters Parameters `json:"parameters,omitempty"`

	// Optional body parameter.
	RequestBody *RequestBody `json:"body,omitempty"`

	// Optional responses.
	Responses Responses `json:"responses,omitempty"`
}

func NewOperation() *Operation {
	return &Operation{}
}

func (value *Operation) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStructFields(value)
}

func (value *Operation) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStructFields(data, value)
}

func (operation *Operation) AddParameter(p *Parameter) {
	operation.Parameters = append(operation.Parameters, p)
}

func (operation *Operation) AddResponse(status int, response *Response) {
	responses := operation.Responses
	if responses == nil {
		operation.Responses = NewResponses()
	}
	if status == 0 {
		responses["default"] = response
	} else {
		responses[strconv.FormatInt(int64(status), 10)] = response
	}
}

func (operation *Operation) ValidateOperation(c context.Context, pathItem *PathItem, method string) error {
	if v := operation.Parameters; v != nil {
		if err := v.Validate(c); err != nil {
			return err
		}
	}
	if v := operation.RequestBody; v != nil {
		if err := v.Validate(c); err != nil {
			return err
		}
	}
	if v := operation.Responses; v != nil {
		if err := v.Validate(c); err != nil {
			return err
		}
	}
	return nil
}
