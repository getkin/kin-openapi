package openapi3

import (
	"context"
	"fmt"
	"net/url"

	"github.com/getkin/kin-openapi/jsoninfo"
)

// ExternalDocs is specified by OpenAPI/Swagger standard version 3.0.
type ExternalDocs struct {
	ExtensionProps

	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
}

func (e *ExternalDocs) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStrictStruct(e)
}

func (e *ExternalDocs) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStrictStruct(data, e)
}

func (e *ExternalDocs) Validate(ctx context.Context) error {
	if e.URL == "" {
		return fmt.Errorf("url is required")
	}
	if _, err := url.Parse(e.URL); err != nil {
		return fmt.Errorf("url is incorrect: %w", err)
	}
	return nil
}
