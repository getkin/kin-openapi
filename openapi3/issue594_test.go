package openapi3_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestIssue594(t *testing.T) {
	uri, err := url.Parse("https://raw.githubusercontent.com/sendgrid/sendgrid-oai/c3aaa432b769faa47285166aca17c7ed2ea71787/oai_v3_stoplight.json")
	require.NoError(t, err)

	sl := openapi3.NewLoader()
	var doc *openapi3.T
	if false {
		doc, err = sl.LoadFromURI(uri)
	} else {
		doc, err = sl.LoadFromFile("testdata/oai_v3_stoplight.json")
	}
	require.NoError(t, err)

	doc.Info.Version = "1.2.3"
	doc.Paths.Value("/marketing/contacts/search/emails").Post = nil
	doc.Components.Schemas["full-segment"].Value.Example = nil

	// This document is OpenAPI 3.1.0 yet relies on nullable, which since 3.1 is
	// an unknown keyword: it no longer lets an example be null.
	err = doc.Validate(sl.Context)
	require.ErrorContains(t, err, `invalid example: validation failed due to: error at "/hard_bounces": at '/hard_bounces': got null, want integer`)

	// Spelled the 3.1 way, the document is valid.
	err = doc.WalkSchemas(func(_ string, ref *openapi3.SchemaRef) error {
		if s := ref.Value; s.Nullable {
			s.Nullable = false
			if !s.Type.IsEmpty() && !s.Type.IncludesNull() {
				types := append(openapi3.Types{}, *s.Type...)
				types = append(types, openapi3.TypeNull)
				s.Type = &types
			}
		}
		return nil
	})
	require.NoError(t, err)
	err = doc.Validate(sl.Context)
	require.NoError(t, err)
}
