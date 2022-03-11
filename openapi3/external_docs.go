package openapi3

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/getkin/kin-openapi/jsoninfo"
)

// ExternalDocs is specified by OpenAPI/Swagger standard version 3.
// See https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md#external-documentation-object
type ExternalDocs struct {
	ExtensionProps

	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	URL         string `json:"url,omitempty" yaml:"url,omitempty"`
}

func (e *ExternalDocs) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStrictStruct(e)
}

func (e *ExternalDocs) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStrictStruct(data, e)
}

func (e *ExternalDocs) Validate(ctx context.Context) error {
	if e.URL == "" {
		return errors.New("url is required")
	}
	if _, err := url.Parse(e.URL); err != nil {
		return fmt.Errorf("url is incorrect: %w", err)
	}
	return nil
}
