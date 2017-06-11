package openapi3

import (
	"context"
	"github.com/jban332/kinapi/jsoninfo"
	"strings"
)

// Content is specified by OpenAPI/Swagger 3.0 standard.
type Content map[string]*ContentType

func NewContent() Content {
	return make(map[string]*ContentType, 4)
}

func NewContentWithJSONSchema(schema *Schema) Content {
	return Content{
		"application/json": &ContentType{
			Schema: schema,
		},
	}
}

func (ct Content) Get(mime string) *ContentType {
	if v := ct[mime]; v != nil {
		return v
	}
	i := strings.IndexByte(mime, ';')
	if i < 0 {
		return nil
	}
	return ct[mime[:i]]
}

func (content Content) Validate(c context.Context) error {
	for _, v := range content {
		// Validate ContentType
		if err := v.Validate(c); err != nil {
			return err
		}
	}
	return nil
}

// ContentType is specified by OpenAPI/Swagger 3.0 standard.
type ContentType struct {
	jsoninfo.RefProps
	jsoninfo.ExtensionProps
	Ref         string        `json:"$ref,omitempty"`
	Description string        `json:"description,omitempty"`
	Schema      *Schema       `json:"schema,omitempty"`
	Examples    []interface{} `json:"examples,omitempty"`
}

func NewContentType() *ContentType {
	return &ContentType{}
}

func (value *ContentType) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStructFields(value)
}

func (value *ContentType) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStructFields(data, value)
}

func (ct *ContentType) Validate(c context.Context) error {
	if schema := ct.Schema; schema != nil {
		if err := schema.Validate(c); err != nil {
			return err
		}
	}
	return nil
}
