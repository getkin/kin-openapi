package openapi3

import (
	"github.com/jban332/kinapi/jsoninfo"
)

// Link is specified by OpenAPI/Swagger standard version 3.0.
type Link struct {
	jsoninfo.RefProps
	jsoninfo.ExtensionProps
	LinkProperties
}

type LinkProperties struct {
	Description string                     `json:"description,omitempty"`
	Href        string                     `json:"href,omitempty"`
	OperationID string                     `json:"operationId,omitempty"`
	Parameters  map[string]interface{}     `json:"parameters,omitempty"`
	Headers     map[string]*Schema `json:"headers,omitempty"`
}

func (value *Link) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStructFields(value)
}

func (value *Link) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStructFields(data, value)
}
