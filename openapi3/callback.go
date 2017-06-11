package openapi3

import "github.com/jban332/kinapi/jsoninfo"

// Callback is specified by OpenAPI/Swagger standard version 3.0.
type Callback struct {
	jsoninfo.RefProps
	jsoninfo.ExtensionProps
	Description string `json:"description,omitempty"`
}

func (value *Callback) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStructFields(value)
}

func (value *Callback) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStructFields(data, value)
}
