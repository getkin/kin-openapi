package openapi3

import (
	"context"
	"strings"
)

// Content is specified by OpenAPI/Swagger 3.0 standard.
type Content map[string]*MediaType

func NewContent() Content {
	return make(Content)
}

func NewContentWithSchema(schema *Schema, consumes []string) Content {
	if len(consumes) == 0 {
		return Content{
			"*/*": NewMediaType().WithSchema(schema),
		}
	}
	content := make(map[string]*MediaType, len(consumes))
	for _, mediaType := range consumes {
		content[mediaType] = NewMediaType().WithSchema(schema)
	}
	return content
}

func NewContentWithSchemaRef(schema *SchemaRef, consumes []string) Content {
	if len(consumes) == 0 {
		return Content{
			"*/*": NewMediaType().WithSchemaRef(schema),
		}
	}
	content := make(map[string]*MediaType, len(consumes))
	for _, mediaType := range consumes {
		content[mediaType] = NewMediaType().WithSchemaRef(schema)
	}
	return content
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

func NewContentWithFormDataSchema(schema *Schema) Content {
	return Content{
		"multipart/form-data": NewMediaType().WithSchema(schema),
	}
}

func NewContentWithFormDataSchemaRef(schema *SchemaRef) Content {
	return Content{
		"multipart/form-data": NewMediaType().WithSchemaRef(schema),
	}
}

// splitMediaType separates the type/subtype portion of a mime from its
// parameters, which keep the leading ';' and their original case: RFC 9110
// section 5.6.6 leaves parameter values case-sensitive unless the parameter
// itself says otherwise.
func splitMediaType(mime string) (mediaType, parameters string) {
	if i := strings.IndexByte(mime, ';'); i >= 0 {
		return mime[:i], mime[i:]
	}
	return mime, ""
}

func (content Content) get(mime string) *MediaType {
	if v := content[mime]; v != nil {
		return v
	}

	mediaType, parameters := splitMediaType(mime)
	if !strings.ContainsRune(mediaType, '/') {
		return nil
	}

	// Compare against each declaration. The lexicographically smallest match is
	// taken so that a document declaring several case variants of one media
	// type still resolves deterministically.
	match := ""
	for candidate := range content {
		if match != "" && candidate >= match {
			continue
		}
		candidateType, candidateParameters := splitMediaType(candidate)
		if parameters == candidateParameters && strings.EqualFold(mediaType, candidateType) {
			match = candidate
		}
	}
	if match == "" {
		return nil
	}
	return content[match]
}

func (content Content) Get(mime string) *MediaType {
	// If the mime is empty then short-circuit to the wildcard.
	// We do this here so that we catch only the specific case of
	// and empty mime rather than a present, but invalid, mime type.
	if mime == "" {
		return content["*/*"]
	}
	// Start by making the most specific match possible
	// by using the mime type in full.
	if v := content.get(mime); v != nil {
		return v
	}
	// If an exact match is not found then we strip all
	// metadata from the mime type and only use the x/y
	// portion. Without metadata the full mime type is
	// preserved for later wildcard searches, and retrying
	// it here would repeat the search above.
	mime, parameters := splitMediaType(mime)
	if parameters != "" {
		if v := content.get(mime); v != nil {
			return v
		}
	}
	// If the x/y pattern has no specific match then we
	// try the x/* pattern.
	i := strings.IndexByte(mime, '/')
	if i < 0 {
		// In the case that the given mime type is not valid because it is
		// missing the subtype we return nil so that this does not accidentally
		// resolve with the wildcard.
		return nil
	}
	if v := content.get(mime[:i] + "/*"); v != nil {
		return v
	}
	// Finally, the most generic match of */* is returned
	// as a catch-all.
	return content["*/*"]
}

// Validate returns an error if Content does not comply with the OpenAPI spec.
func (content Content) Validate(ctx context.Context, opts ...ValidationOption) error {
	ctx = WithValidationOptions(ctx, opts...)

	for _, k := range componentNames(content) {
		if err := content[k].Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}
