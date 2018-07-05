package openapi3_test

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestLoadYAML(t *testing.T) {
	spec := []byte(`
openapi: 3.0.0
info:
  title: An API
  version: v1

components:
  schemas:
    NewItem:
      required: [name]
      properties:
        name: {type: string}
        tag: {type: string}
    ErrorModel:
      type: object
      required: [code, message]
      properties:
        code: {type: integer}
        message: {type: string}

paths:
  /items:
    put:
      description: ''
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/NewItem'
      responses:
        default: &defaultResponse # a YAML ref
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorModel'
`)

	loader := openapi3.NewSwaggerLoader()
	doc, err := loader.LoadSwaggerFromYAMLData(spec)
	require.NoError(t, err)
	require.Equal(t, "An API", doc.Info.Title)
	require.Equal(t, 2, len(doc.Components.Schemas))
	require.Equal(t, 1, len(doc.Paths))
	def := doc.Paths["/items"].Put.Responses.Default().Value
	require.Equal(t, "unexpected error", def.Description)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
}

func ExampleSwaggerLoader() {
	source := `{"info":{"description":"An API"}}`
	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData([]byte(source))
	if err != nil {
		panic(err)
	}
	fmt.Print(swagger.Info.Description)
	// Output:
	// An API
}

func TestResolveSchemaRef(t *testing.T) {
	source := []byte(`{"info":{"description":"An API"},"components":{"schemas":{"B":{"type":"string"},"A":{"allOf":[{"$ref":"#/components/schemas/B"}]}}}}`)
	loader := openapi3.NewSwaggerLoader()
	doc, err := loader.LoadSwaggerFromData(source)
	require.NoError(t, err)
	err = doc.Validate(loader.Context)

	require.NoError(t, err)
	refAVisited := doc.Components.Schemas["A"].Value.AllOf[0]
	require.Equal(t, "#/components/schemas/B", refAVisited.Ref)
	require.NotNil(t, refAVisited.Value)
}

type sourceExample struct {
	Location *url.URL
	Spec     []byte
}

type multipleSourceSwaggerLoaderExample struct {
	Sources []*sourceExample
}

func (l *multipleSourceSwaggerLoaderExample) LoadSwaggerFromURI(
	loader *openapi3.SwaggerLoader,
	location *url.URL,
) (*openapi3.Swagger, error) {
	source := l.resolveSourceFromURI(location)
	if source == nil {
		return nil, fmt.Errorf("Unsupported URI: '%s'", location.String())
	}
	return loader.LoadSwaggerFromData(source.Spec)
}

func (l *multipleSourceSwaggerLoaderExample) resolveSourceFromURI(location *url.URL) *sourceExample {
	locationString := location.String()
	for _, v := range l.Sources {
		if v.Location.String() == locationString {
			return v
		}
	}
	return nil
}

func TestResolveSchemaExternalRef(t *testing.T) {
	rootLocation := &url.URL{Scheme: "http", Host: "example.com", Path: "spec.json"}
	externalLocation := &url.URL{Scheme: "http", Host: "example.com", Path: "external.json"}
	rootSpec := []byte(fmt.Sprintf(
		`{"info":{"description":"An API"},"components":{"schemas":{"Root":{"allOf":[{"$ref":"%s#/components/schemas/External"}]}}}}`,
		externalLocation.String(),
	))
	externalSpec := []byte(`{"info":{"description":"External Spec"},"components":{"schemas":{"External":{"type":"string"}}}}`)
	multipleSourceLoader := &multipleSourceSwaggerLoaderExample{
		Sources: []*sourceExample{
			&sourceExample{
				Location: rootLocation,
				Spec:     rootSpec,
			},
			&sourceExample{
				Location: externalLocation,
				Spec:     externalSpec,
			},
		},
	}
	loader := &openapi3.SwaggerLoader{
		IsExternalRefsAllowed:  true,
		LoadSwaggerFromURIFunc: multipleSourceLoader.LoadSwaggerFromURI,
	}
	doc, err := loader.LoadSwaggerFromURI(rootLocation)
	require.NoError(t, err)
	err = doc.Validate(loader.Context)

	require.NoError(t, err)
	refRootVisited := doc.Components.Schemas["Root"].Value.AllOf[0]
	require.Equal(t, fmt.Sprintf("%s#/components/schemas/External", externalLocation.String()), refRootVisited.Ref)
	require.NotNil(t, refRootVisited.Value)
}
