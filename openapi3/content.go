package openapi3

import (
	"context"
	"github.com/ronniedada/kin-openapi/jsoninfo"
	"strings"
)

// Content is specified by OpenAPI/Swagger 3.0 standard.
type Content map[string]*ContentType

func NewContent() Content {
	return make(map[string]*ContentType, 4)
}

func NewContentWithJSONSchema(schema *Schema) Content {
	return Content{
		"application/json": NewContentType().WithSchema(schema),
	}
}
func NewContentWithJSONSchemaRef(schema *SchemaRef) Content {
	return Content{
		"application/json": NewContentType().WithSchemaRef(schema),
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
	ExtensionProps
	Description string       `json:"description,omitempty"`
	Schema      *SchemaRef   `json:"schema,omitempty"`
	Examples    []ExampleRef `json:"examples,omitempty"`
}

func NewContentType() *ContentType {
	return &ContentType{}
}

func (contentType *ContentType) WithSchema(schema *Schema) *ContentType {
	if schema == nil {
		contentType.Schema = nil
	} else {
		contentType.Schema = &SchemaRef{
			Value: schema,
		}
	}
	return contentType
}

func (contentType *ContentType) WithSchemaRef(schema *SchemaRef) *ContentType {
	contentType.Schema = schema
	return contentType
}

func (contentType *ContentType) WithExample(value interface{}) *ContentType {
	contentType.Examples = append(contentType.Examples, ExampleRef{
		Value: &value,
	})
	return contentType
}

func (value *ContentType) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStrictStruct(value)
}

func (value *ContentType) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStrictStruct(data, value)
}

func (ct *ContentType) Validate(c context.Context) error {
	if ct == nil {
		return nil
	}
	if schema := ct.Schema; schema != nil {
		if err := schema.Validate(c); err != nil {
			return err
		}
	}
	return nil
}
