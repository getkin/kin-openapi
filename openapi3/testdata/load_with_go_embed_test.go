//+build with_embed

package openapi3_test

import (
	"embed"
	"fmt"
	"net/url"

	"github.com/getkin/kin-openapi/openapi3"
)

//go:embed recursiveRef/*
var fs embed.FS

func Example() {
	loader := openapi3.NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.SwaggerLoader, uri *url.URL) ([]byte, error) {
		return fs.ReadFile(uri.Path)
	}

	doc, err := loader.LoadSwaggerFromFile("recursiveRef/openapi.yml")
	if err != nil {
		panic(err)
	}

	if err = doc.Validate(loader.Context); err != nil {
		panic(err)
	}

	fmt.Println(doc.Paths["/foo"].Get.Responses["200"].Value.Content["application/json"].Schema.Value.Properties["foo"].Value.Properties["bar"].Value.Type)
	// Output: array
}
