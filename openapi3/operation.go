package openapi3

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/getkin/kin-openapi/jsoninfo"
	"github.com/go-openapi/jsonpointer"
)

// Operation represents "operation" specified by" OpenAPI/Swagger 3.0 standard.
// See https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md#operation-object
type Operation struct {
	ExtensionProps
	RequestBody  *RequestBodyRef       `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Servers      *Servers              `json:"servers,omitempty" yaml:"servers,omitempty"`
	Security     *SecurityRequirements `json:"security,omitempty" yaml:"security,omitempty"`
	Callbacks    Callbacks             `json:"callbacks,omitempty" yaml:"callbacks,omitempty"`
	Responses    Responses             `json:"responses" yaml:"responses"`
	ExternalDocs *ExternalDocs         `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	OperationID  string                `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Description  string                `json:"description,omitempty" yaml:"description,omitempty"`
	Summary      string                `json:"summary,omitempty" yaml:"summary,omitempty"`
	Tags         []string              `json:"tags,omitempty" yaml:"tags,omitempty"`
	Parameters   Parameters            `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Deprecated   bool                  `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
}

var _ jsonpointer.JSONPointable = (*Operation)(nil)

func NewOperation() *Operation {
	return &Operation{}
}

// MarshalJSON returns the JSON encoding of Operation.
func (operation *Operation) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStrictStruct(operation)
}

// UnmarshalJSON sets Operation to a copy of data.
func (operation *Operation) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStrictStruct(data, operation)
}

// JSONLookup implements github.com/go-openapi/jsonpointer#JSONPointable
func (operation Operation) JSONLookup(token string) (interface{}, error) {
	switch token {
	case "requestBody":
		if operation.RequestBody != nil {
			if operation.RequestBody.Ref != "" {
				return &Ref{Ref: operation.RequestBody.Ref}, nil
			}
			return operation.RequestBody.Value, nil
		}
	case "tags":
		return operation.Tags, nil
	case "summary":
		return operation.Summary, nil
	case "description":
		return operation.Description, nil
	case "operationID":
		return operation.OperationID, nil
	case "parameters":
		return operation.Parameters, nil
	case "responses":
		return operation.Responses, nil
	case "callbacks":
		return operation.Callbacks, nil
	case "deprecated":
		return operation.Deprecated, nil
	case "security":
		return operation.Security, nil
	case "servers":
		return operation.Servers, nil
	case "externalDocs":
		return operation.ExternalDocs, nil
	}

	v, _, err := jsonpointer.GetForToken(operation.ExtensionProps, token)
	return v, err
}

func (operation *Operation) AddParameter(p *Parameter) {
	operation.Parameters = append(operation.Parameters, &ParameterRef{
		Value: p,
	})
}

func (operation *Operation) AddResponse(status int, response *Response) {
	responses := operation.Responses
	if responses == nil {
		responses = NewResponses()
		operation.Responses = responses
	}
	code := "default"
	if status != 0 {
		code = strconv.FormatInt(int64(status), 10)
	}
	responses[code] = &ResponseRef{
		Value: response,
	}
}

// Validate returns an error if Operation does not comply with the OpenAPI spec.
func (operation *Operation) Validate(ctx context.Context) error {
	if v := operation.Parameters; v != nil {
		if err := v.Validate(ctx); err != nil {
			return err
		}
	}
	if v := operation.RequestBody; v != nil {
		if err := v.Validate(ctx); err != nil {
			return err
		}
	}
	if v := operation.Responses; v != nil {
		if err := v.Validate(ctx); err != nil {
			return err
		}
	} else {
		return errors.New("value of responses must be an object")
	}
	if v := operation.ExternalDocs; v != nil {
		if err := v.Validate(ctx); err != nil {
			return fmt.Errorf("invalid external docs: %w", err)
		}
	}
	return nil
}
