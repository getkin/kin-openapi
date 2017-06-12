package openapi3

import (
	"context"
	"github.com/jban332/kin-openapi/jsoninfo"
)

type Header struct {
	jsoninfo.ExtensionProps

	// Optional description. Should use CommonMark syntax.
	Description string `json:"description,omitempty"`

	// Optional schema
	Schema *SchemaRef `json:"schema,omitempty"`
}

func (value *Header) Validate(c context.Context) error {
	if v := value.Schema; v != nil {
		err := v.Validate(c)
		if err != nil {
			return err
		}
	}
	return nil
}
