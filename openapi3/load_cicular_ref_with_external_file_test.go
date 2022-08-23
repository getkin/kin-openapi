//go:build go1.16
// +build go1.16

package openapi3_test

import (
	"embed"
	"fmt"
	"net/url"
	"testing"

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

	doc, err := loader.LoadFromFile("testdata/circularRef/base.yml")
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%v\n", doc)
}
