//go:build go1.16
// +build go1.16

package openapi3_test

import (
	"embed"
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
	if err != nil {
		t.Error(err)
	}

	foo := &openapi3.SchemaRef{
		Ref: "",
		Value: &openapi3.Schema{
			ExtensionProps: openapi3.ExtensionProps{Extensions: map[string]interface{}{}},
			Properties: map[string]*openapi3.SchemaRef{
				"foo2": { // reference to an external file
					Ref: "other.yml#/components/schemas/Foo2",
					Value: &openapi3.Schema{
						ExtensionProps: openapi3.ExtensionProps{Extensions: map[string]interface{}{}},
						Properties: map[string]*openapi3.SchemaRef{
							"id": {
								Value: &openapi3.Schema{
									Type:           "string",
									ExtensionProps: openapi3.ExtensionProps{Extensions: map[string]interface{}{}},
								}},
						},
					},
				},
			},
		},
	}
	bar := &openapi3.SchemaRef{
		Ref: "",
		Value: &openapi3.Schema{
			ExtensionProps: openapi3.ExtensionProps{Extensions: map[string]interface{}{}},
			Properties:     map[string]*openapi3.SchemaRef{},
		},
	}
	// circular reference
	bar.Value.Properties["foo"] = &openapi3.SchemaRef{Ref: "#/components/schemas/Foo", Value: foo.Value}
	foo.Value.Properties["bar"] = &openapi3.SchemaRef{Ref: "#/components/schemas/Bar", Value: bar.Value}

	want := &openapi3.T{
		OpenAPI: "3.0.3",
		Info: &openapi3.Info{
			Title:   "Recursive cyclic refs example",
			Version: "1.0",

			ExtensionProps: openapi3.ExtensionProps{Extensions: map[string]interface{}{}},
		},
		Components: openapi3.Components{
			ExtensionProps: openapi3.ExtensionProps{Extensions: map[string]interface{}{}},
			Schemas: map[string]*openapi3.SchemaRef{
				"Foo": foo,
				"Bar": bar,
			},
		},
		ExtensionProps: openapi3.ExtensionProps{Extensions: map[string]interface{}{}},
	}

	require.Equal(t, want, got)
}
