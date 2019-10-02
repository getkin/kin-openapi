package openapi3

import (
	"context"

	"github.com/getkin/kin-openapi/jsoninfo"
)

// Link is specified by OpenAPI/Swagger standard version 3.0.
type Link struct {
	ExtensionProps
	Description string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Href        string                 `json:"href,omitempty" yaml:"href,omitempty"`
	OperationID string                 `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Headers     map[string]*Schema     `json:"headers,omitempty" yaml:"headers,omitempty"`
}

func (value *Link) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStrictStruct(value)
}

func (value *Link) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStrictStruct(data, value)
}

func (value *Link) Validate(c context.Context) error {
	return nil
}
