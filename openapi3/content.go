package openapi3

import (
	"context"
	"strings"
)

// Content is specified by OpenAPI/Swagger 3.0 standard.
type Content map[string]*MediaType

func NewContent() Content {
	return make(map[string]*MediaType, 4)
}

func NewContentWithJSONSchema(schema *Schema) Content {
	return Content{
		"application/json": NewMediaType().WithSchema(schema),
	}
}
func NewContentWithJSONSchemaRef(schema *SchemaRef) Content {
	return Content{
		"application/json": NewMediaType().WithSchemaRef(schema),
	}
}

func (content Content) Get(mime string) *MediaType {
	if v := content[mime]; v != nil {
		return v
	}
	i := strings.IndexByte(mime, ';')
	if i < 0 {
		return nil
	}
	return content[mime[:i]]
}

func (content Content) Validate(c context.Context) error {
	for _, v := range content {
		// Validate MediaType
		if err := v.Validate(c); err != nil {
			return err
		}
	}
	return nil
}
