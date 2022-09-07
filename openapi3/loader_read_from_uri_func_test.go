package openapi3

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoaderReadFromURIFunc(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *Loader, url *url.URL) ([]byte, error) {
		return ioutil.ReadFile(filepath.Join("testdata", url.Path))
	}
	doc, err := loader.LoadFromFile("recursiveRef/openapi.yml")
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.NoError(t, doc.Validate(loader.Context))
	require.Equal(t, "bar", doc.Paths["/foo"].Get.Responses.Get(200).Value.Content.Get("application/json").Schema.Value.Properties["foo2"].Value.Properties["foo"].Value.Properties["bar"].Value.Example)
}

type multipleSourceLoaderExample struct {
	Sources map[string][]byte
}

func (l *multipleSourceLoaderExample) LoadFromURI(
	loader *Loader,
	location *url.URL,
) ([]byte, error) {
	source := l.resolveSourceFromURI(location)
	if source == nil {
		return nil, fmt.Errorf("unsupported URI: %q", location.String())
	}
	return source, nil
}

func (l *multipleSourceLoaderExample) resolveSourceFromURI(location fmt.Stringer) []byte {
	return l.Sources[location.String()]
}

func TestResolveSchemaExternalRef(t *testing.T) {
	rootLocation := &url.URL{Scheme: "http", Host: "example.com", Path: "spec.json"}
	externalLocation := &url.URL{Scheme: "http", Host: "example.com", Path: "external.json"}
	rootSpec := []byte(fmt.Sprintf(
		`{"openapi":"3.0.0","info":{"title":"MyAPI","version":"0.1","description":"An API"},"paths":{},"components":{"schemas":{"Root":{"allOf":[{"$ref":"%s#/components/schemas/External"}]}}}}`,
		externalLocation.String(),
	))
	externalSpec := []byte(`{"openapi":"3.0.0","info":{"title":"MyAPI","version":"0.1","description":"External Spec"},"paths":{},"components":{"schemas":{"External":{"type":"string"}}}}`)
	multipleSourceLoader := &multipleSourceLoaderExample{
		Sources: map[string][]byte{
			rootLocation.String():     rootSpec,
			externalLocation.String(): externalSpec,
		},
	}
	loader := &Loader{
		IsExternalRefsAllowed: true,
		ReadFromURIFunc:       multipleSourceLoader.LoadFromURI,
	}

	doc, err := loader.LoadFromURI(rootLocation)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	refRootVisited := doc.Components.Schemas["Root"].Value.AllOf[0]
	require.Equal(t, fmt.Sprintf("%s#/components/schemas/External", externalLocation.String()), refRootVisited.Ref)
	require.NotNil(t, refRootVisited.Value)
}
