//go:build go1.16
// +build go1.16

package openapi3_test

import (
	"embed"
	"encoding/json"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

//go:embed testdata/circularRef/*
var circularResSpecs embed.FS

func TestLoadCircularRefFromFile(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, uri *url.URL) ([]byte, error) {
		return circularResSpecs.ReadFile(uri.Path)
	}

	got, err := loader.LoadFromFile("testdata/circularRef/base.yml")
	require.NoError(t, err)

	foo := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Properties: map[string]*openapi3.SchemaRef{
				"foo2": {
					Ref: "other.yml#/components/schemas/Foo2", // reference to an external file
					Value: &openapi3.Schema{
						Properties: map[string]*openapi3.SchemaRef{
							"id": {
								Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
						},
					},
				},
			},
		},
	}
	bar := &openapi3.SchemaRef{Value: &openapi3.Schema{Properties: make(map[string]*openapi3.SchemaRef)}}
	// circular reference
	bar.Value.Properties["foo"] = &openapi3.SchemaRef{Ref: "#/components/schemas/Foo", Value: foo.Value}
	foo.Value.Properties["bar"] = &openapi3.SchemaRef{Ref: "#/components/schemas/Bar", Value: bar.Value}

	bazNestedRef := &openapi3.SchemaRef{Ref: "./baz.yml#/BazNested"}
	array := openapi3.NewArraySchema()
	array.Items = bazNestedRef
	bazNested := &openapi3.Schema{Properties: map[string]*openapi3.SchemaRef{
		"bazArray": {
			Value: &openapi3.Schema{
				Items: bazNestedRef,
			},
		},
		"baz": bazNestedRef,
	}}
	bazNestedRef.Value = bazNested

	want := &openapi3.T{
		OpenAPI: "3.0.3",
		Info: &openapi3.Info{
			Title:   "Recursive cyclic refs example",
			Version: "1.0",
		},
		Components: &openapi3.Components{
			Schemas: openapi3.Schemas{
				"Foo": foo,
				"Bar": bar,
				"Baz": bazNestedRef,
			},
		},
	}

	jsoner := func(doc *openapi3.T) string {
		data, err := json.Marshal(doc)
		require.NoError(t, err)
		return string(data)
	}
	require.JSONEq(t, jsoner(want), jsoner(got))
}
