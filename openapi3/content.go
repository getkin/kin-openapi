package openapi3

import (
	"context"
	"iter"
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

// Get returns the MediaType whose key matches mime, or nil.
//
// Keys and mime are compared case-insensitively (RFC 9110, section 8.3.1).
// A key may be a comma-separated list of media ranges, any of which matches.
// The most specific match wins, tried in this order:
//
//  1. mime in full, parameters included (e.g. "application/json;charset=utf-8")
//  2. mime without its parameters (e.g. "application/json")
//  3. its type wildcard (e.g. "application/*")
//  4. the full wildcard "*/*"
//
// An empty mime matches only "*/*". A mime without a subtype matches nothing.
// Content.Validate rejects keys that would make a match ambiguous.
func (content Content) Get(mime string) *MediaType {
	if mime == "" {
		return content.lookup("*/*")
	}
	mime = strings.ToLower(strings.TrimSpace(mime))
	if v := content.lookup(mime); v != nil {
		return v
	}
	base, _, _ := strings.Cut(mime, ";")
	base = strings.TrimSpace(base)
	if v := content.lookup(base); v != nil {
		return v
	}
	typ, _, ok := strings.Cut(base, "/")
	if !ok {
		// Not a valid media type: do not let it resolve to a wildcard.
		return nil
	}
	if v := content.lookup(typ + "/*"); v != nil {
		return v
	}
	return content.lookup("*/*")
}

// lookup returns the MediaType whose key lists the lowercased media range want.
func (content Content) lookup(want string) *MediaType {
	if v := content[want]; v != nil {
		return v
	}
	// Map iteration order does not matter here: Content.Validate rejects keys
	// that overlap, so at most one key lists want. On a Content that does not
	// validate, which of the overlapping keys is returned is unspecified.
	for k, v := range content {
		for entry := range mediaRanges(k) {
			if entry == want {
				return v
			}
		}
	}
	return nil
}

// mediaRanges yields the non-empty entries of a comma-separated list of media
// ranges, trimmed and lowercased: "Image/PNG, image/*" yields "image/png" then "image/*".
func mediaRanges(s string) iter.Seq[string] {
	return func(yield func(string) bool) {
		for entry := range strings.SplitSeq(s, ",") {
			if entry = strings.ToLower(strings.TrimSpace(entry)); entry == "" {
				continue
			}
			if !yield(entry) {
				return
			}
		}
	}
}

// Validate returns an error if Content does not comply with the OpenAPI spec.
func (content Content) Validate(ctx context.Context, opts ...ValidationOption) error {
	ctx = WithValidationOptions(ctx, opts...)

	// No two keys may list the same media range (compared case-insensitively),
	// otherwise Get could not tell which one a media type matches.
	seen := make(map[string]string, len(content))
	for _, k := range componentNames(content) {
		for entry := range mediaRanges(k) {
			if other, ok := seen[entry]; ok && other != k {
				var origin *Origin
				if mt := content[k]; mt != nil {
					origin = mt.Origin
				}
				return newContentKeysOverlap(other, k, entry, origin)
			}
			seen[entry] = k
		}
	}

	for _, k := range componentNames(content) {
		if err := content[k].Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}
