package openapi3_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// Every version's Reference Object says "any properties added SHALL be
// ignored". From 3.1 on summary and description are fixed fields that override
// the referenced component's.
func TestReferenceObjectSiblingsAreIgnored(t *testing.T) {
	t.Parallel()
	for _, version := range []string{"3.0.3", "3.1.0", "3.2.0"} {
		t.Run(version, func(t *testing.T) {
			t.Parallel()
			loader := openapi3.NewLoader()
			doc, err := loader.LoadFromData(fmt.Appendf(nil, `{
  "openapi": %q,
  "info": {"title": "t", "version": "1"},
  "paths": {"/x": {"get": {"responses": {"200": {
    "$ref": "#/components/responses/R",
    "description": "local",
    "unknown": 1
  }}}}},
  "components": {"responses": {"R": {"description": "target"}}}
}`, version))
			require.NoError(t, err)
			require.NoError(t, doc.Validate(t.Context()))

			want := "local"
			if doc.IsOpenAPI30() {
				want = "target"
			}
			response := doc.Paths.Value("/x").Get.Responses.Value("200").Value
			require.Equal(t, want, *response.Description)
		})
	}
}

// InternalizeRefs replaces a component that is a $ref by what it refers to,
// which leaves out the fields that were next to the $ref.
func TestReferenceObjectSiblingsAreGoneOnceInternalized(t *testing.T) {
	t.Parallel()
	for _, version := range []string{"3.0.3", "3.1.0"} {
		t.Run(version, func(t *testing.T) {
			t.Parallel()
			loader := openapi3.NewLoader()
			doc, err := loader.LoadFromData(fmt.Appendf(nil, `{
  "openapi": %q,
  "info": {"title": "t", "version": "1"},
  "paths": {},
  "components": {
    "schemas": {
      "A": {"type": "string"},
      "B": {"$ref": "#/components/schemas/A", "minLength": 2}
    },
    "parameters": {
      "P": {"name": "p", "in": "query", "schema": {"type": "string"}},
      "Q": {"$ref": "#/components/parameters/P", "description": "local", "unknown": 1}
    }
  }
}`, version))
			require.NoError(t, err)
			require.NoError(t, doc.Validate(t.Context()))

			doc.InternalizeRefs(t.Context(), nil)
			require.NoError(t, doc.Validate(t.Context()))
		})
	}
}

func TestProhibitExtensionsWithRefInEveryVersion(t *testing.T) {
	t.Parallel()
	for _, version := range []string{"3.0.3", "3.1.0"} {
		t.Run(version, func(t *testing.T) {
			t.Parallel()
			loader := openapi3.NewLoader()
			doc, err := loader.LoadFromData(fmt.Appendf(nil, `{
  "openapi": %q,
  "info": {"title": "t", "version": "1"},
  "paths": {},
  "components": {"schemas": {
    "A": {"type": "string"},
    "B": {"$ref": "#/components/schemas/A", "x-order": 1}
  }}
}`, version))
			require.NoError(t, err)
			require.NoError(t, doc.Validate(t.Context()))
			require.ErrorContains(t, doc.Validate(t.Context(), openapi3.ProhibitExtensionsWithRef()), "extra sibling fields: [x-order]")
			require.NoError(t, doc.Validate(t.Context(), openapi3.ProhibitExtensionsWithRef(), openapi3.AllowExtraSiblingFields("x-order")))
		})
	}
}
