package openapi3

// ExternalDocs is specified by OpenAPI/Swagger standard version 3.0.
type ExternalDocs struct {
	ExtensionProps

	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
}
