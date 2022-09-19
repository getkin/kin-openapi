package openapi3

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const addr = "localhost:7965"

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

	loader := NewLoader()
	doc, err := loader.LoadFromData(spec)
	require.NoError(t, err)
	require.Equal(t, "An API", doc.Info.Title)
	require.Equal(t, 2, len(doc.Components.Schemas))
	require.Equal(t, 1, len(doc.Paths))
	def := doc.Paths["/items"].Put.Responses.Default().Value
	desc := "unexpected error"
	require.Equal(t, &desc, def.Description)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
}

func ExampleLoader() {
	const source = `{"info":{"description":"An API"}}`
	doc, err := NewLoader().LoadFromData([]byte(source))
	if err != nil {
		panic(err)
	}
	fmt.Print(doc.Info.Description)
	// Output: An API
}

func TestResolveSchemaRef(t *testing.T) {
	source := []byte(`{"openapi":"3.0.0","info":{"title":"MyAPI","version":"0.1",description":"An API"},"paths":{},"components":{"schemas":{"B":{"type":"string"},"A":{"allOf":[{"$ref":"#/components/schemas/B"}]}}}}`)
	loader := NewLoader()
	doc, err := loader.LoadFromData(source)
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	refAVisited := doc.Components.Schemas["A"].Value.AllOf[0]
	require.Equal(t, "#/components/schemas/B", refAVisited.Ref)
	require.NotNil(t, refAVisited.Value)
}

func TestResolveSchemaRefWithNullSchemaRef(t *testing.T) {
	source := []byte(`{"openapi":"3.0.0","info":{"title":"MyAPI","version":"0.1","description":"An API"},"paths":{"/foo":{"post":{"requestBody":{"content":{"application/json":{"schema":null}}}}}}}`)
	loader := NewLoader()
	doc, err := loader.LoadFromData(source)
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.EqualError(t, err, `invalid paths: found unresolved ref: ""`)
}

func TestResolveResponseExampleRef(t *testing.T) {
	source := []byte(`
openapi: 3.0.1
info:
  title: My API
  version: 1.0.0
components:
  examples:
    test:
      value:
        error: false
paths:
  /:
    get:
      responses:
        200:
          description: A test response
          content:
            application/json:
              examples:
                test:
                  $ref: '#/components/examples/test'`)
	loader := NewLoader()
	doc, err := loader.LoadFromData(source)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	example := doc.Paths["/"].Get.Responses.Get(200).Value.Content.Get("application/json").Examples["test"]
	require.NotNil(t, example.Value)
	require.Equal(t, example.Value.Value.(map[string]interface{})["error"].(bool), false)
}

func TestLoadErrorOnRefMisuse(t *testing.T) {
	spec := []byte(`
openapi: '3.0.0'
servers: [{url: /}]
info:
  title: Some API
  version: '1'
components:
  schemas:
    Thing: {type: string}
paths:
  /items:
    put:
      description: ''
      requestBody:
        # Uses a schema ref instead of a requestBody ref.
        $ref: '#/components/schemas/Thing'
      responses:
        '201':
          description: ''
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Thing'
`)

	loader := NewLoader()
	_, err := loader.LoadFromData(spec)
	require.Error(t, err)
}

func TestLoadPathParamRef(t *testing.T) {
	spec := []byte(`
openapi: '3.0.0'
info:
  title: ''
  version: '1'
components:
  parameters:
    testParam:
      name: test
      in: query
      schema:
        type: string
paths:
  '/':
    parameters:
      - $ref: '#/components/parameters/testParam'
    get:
      responses:
        '200':
          description: Test call.
`)

	loader := NewLoader()
	doc, err := loader.LoadFromData(spec)
	require.NoError(t, err)

	require.NotNil(t, doc.Paths["/"].Parameters[0].Value)
}

func TestLoadRequestExampleRef(t *testing.T) {
	spec := []byte(`
openapi: '3.0.0'
info:
  title: ''
  version: '1'
components:
  examples:
    test:
      value:
        hello: world
paths:
  '/':
    post:
      requestBody:
        content:
          application/json:
            examples:
              test:
                $ref: "#/components/examples/test"
      responses:
        '200':
          description: Test call.
`)

	loader := NewLoader()
	doc, err := loader.LoadFromData(spec)
	require.NoError(t, err)

	require.NotNil(t, doc.Paths["/"].Post.RequestBody.Value.Content.Get("application/json").Examples["test"])
}

func createTestServer(t *testing.T, handler http.Handler) *httptest.Server {
	ts := httptest.NewUnstartedServer(handler)
	l, err := net.Listen("tcp", addr)
	require.NoError(t, err)
	ts.Listener.Close()
	ts.Listener = l
	return ts
}

func TestLoadFromRemoteURL(t *testing.T) {
	fs := http.FileServer(http.Dir("testdata"))
	ts := createTestServer(t, fs)
	ts.Start()
	defer ts.Close()

	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	url, err := url.Parse("http://" + addr + "/test.openapi.json")
	require.NoError(t, err)

	doc, err := loader.LoadFromURI(url)
	require.NoError(t, err)

	require.Equal(t, "string", doc.Components.Schemas["TestSchema"].Value.Type)
}

func TestLoadWithReferenceInReference(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromFile("testdata/refInRef/openapi.json")
	require.NoError(t, err)
	require.NotNil(t, doc)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
	require.Equal(t, "string", doc.Paths["/api/test/ref/in/ref"].Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["definition_reference"].Value.Type)
}

func TestLoadFileWithExternalSchemaRef(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromFile("testdata/testref.openapi.json")
	require.NoError(t, err)
	require.NotNil(t, doc.Components.Schemas["AnotherTestSchema"].Value.Type)
}

func TestLoadFileWithExternalSchemaRefSingleComponent(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromFile("testdata/testrefsinglecomponent.openapi.json")
	require.NoError(t, err)

	require.NotNil(t, doc.Components.Responses["SomeResponse"])
	desc := "this is a single response definition"
	require.Equal(t, &desc, doc.Components.Responses["SomeResponse"].Value.Description)
}

func TestLoadRequestResponseHeaderRef(t *testing.T) {
	spec := []byte(`
{
    "openapi": "3.0.0",
    "info": {
        "title": "",
        "version": "1"
    },
    "paths": {
      "/test": {
        "post": {
          "responses": {
            "default": {
              "description": "test",
              "headers": {
                "X-TEST-HEADER": {
                  "$ref": "#/components/headers/TestHeader"
                }
              }
            }
          }
        }
      }
    },
    "components": {
      "headers": {
        "TestHeader": {
          "description": "testheader"
        }
      }
    }
}`)

	loader := NewLoader()
	doc, err := loader.LoadFromData(spec)
	require.NoError(t, err)

	require.NotNil(t, doc.Paths["/test"].Post.Responses["default"].Value.Headers["X-TEST-HEADER"].Value.Description)
	require.Equal(t, "testheader", doc.Paths["/test"].Post.Responses["default"].Value.Headers["X-TEST-HEADER"].Value.Description)
}

func TestLoadFromDataWithExternalRequestResponseHeaderRemoteRef(t *testing.T) {
	spec := []byte(`
{
    "openapi": "3.0.0",
    "info": {
        "title": "",
        "version": "1"
    },
    "paths": {
        "/test": {
            "post": {
                "responses": {
                    "default": {
                        "description": "test",
                        "headers": {
                            "X-TEST-HEADER": {
                                "$ref": "http://` + addr + `/components.openapi.json#/components/headers/CustomTestHeader"
                            }
                        }
                    }
                }
            }
        }
    }
}`)

	fs := http.FileServer(http.Dir("testdata"))
	ts := createTestServer(t, fs)
	ts.Start()
	defer ts.Close()

	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromDataWithPath(spec, &url.URL{Path: "testdata/testfilename.openapi.json"})
	require.NoError(t, err)

	require.NotNil(t, doc.Paths["/test"].Post.Responses["default"].Value.Headers["X-TEST-HEADER"].Value.Description)
	require.Equal(t, "description", doc.Paths["/test"].Post.Responses["default"].Value.Headers["X-TEST-HEADER"].Value.Description)
}

func TestLoadYamlFile(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromFile("testdata/test.openapi.yml")
	require.NoError(t, err)

	require.Equal(t, "OAI Specification in YAML", doc.Info.Title)
}

func TestLoadYamlFileWithExternalSchemaRef(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromFile("testdata/testref.openapi.yml")
	require.NoError(t, err)

	require.NotNil(t, doc.Components.Schemas["AnotherTestSchema"].Value.Type)
}

func TestLoadYamlFileWithExternalPathRef(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromFile("testdata/pathref.openapi.yml")
	require.NoError(t, err)

	require.NotNil(t, doc.Paths["/test"].Get.Responses["200"].Value.Content["application/json"].Schema.Value.Type)
	require.Equal(t, "string", doc.Paths["/test"].Get.Responses["200"].Value.Content["application/json"].Schema.Value.Type)
}

func TestResolveResponseLinkRef(t *testing.T) {
	source := []byte(`
openapi: 3.0.1
info:
  title: My API
  version: 1.0.0
components:
  links:
    Father:
        description: link to to the father
        operationId: getUserById
        parameters:
          "id": "$response.body#/fatherId"
paths:
  /users/{id}:
    get:
      operationId: getUserById,
      parameters:
        - name: id,
          in: path
          required: true
          schema:
            type: string
      responses:
        200:
          description: A test response
          content:
            application/json:
          links:
            father:
              $ref: '#/components/links/Father'
`)
	loader := NewLoader()
	doc, err := loader.LoadFromData(source)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	response := doc.Paths[`/users/{id}`].Get.Responses.Get(200).Value
	link := response.Links[`father`].Value
	require.NotNil(t, link)
	require.Equal(t, "getUserById", link.OperationID)
	require.Equal(t, "link to to the father", link.Description)
}

func TestLinksFromOAISpec(t *testing.T) {
	loader := NewLoader()
	doc, err := loader.LoadFromFile("testdata/link-example.yaml")
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
	response := doc.Paths[`/2.0/repositories/{username}/{slug}`].Get.Responses.Get(200).Value
	link := response.Links[`repositoryPullRequests`].Value
	require.Equal(t, map[string]interface{}{
		"username": "$response.body#/owner/username",
		"slug":     "$response.body#/slug",
	}, link.Parameters)
}

func TestResolveNonComponentsRef(t *testing.T) {
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
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorModel'
    post:
      description: ''
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/paths/~1items/put/requestBody/content/application~1json/schema'
      responses:
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorModel'
`)

	loader := NewLoader()
	doc, err := loader.LoadFromData(spec)
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
}

func TestServersVariables(t *testing.T) {
	const spec = `
openapi: 3.0.1
info:
  title: My API
  version: 1.0.0
paths: {}
servers:
- @@@
`
	for value, expected := range map[string]string{
		`{url: /}`:                            "",
		`{url: "http://{x}.{y}.example.com"}`: "invalid servers: server has undeclared variables",
		`{url: "http://{x}.y}.example.com"}`:  "invalid servers: server URL has mismatched { and }",
		`{url: "http://{x.example.com"}`:      "invalid servers: server URL has mismatched { and }",
		`{url: "http://{x}.example.com", variables: {x: {default: "www"}}}`:                "",
		`{url: "http://{x}.example.com", variables: {x: {default: "www", enum: ["www"]}}}`: "",
		`{url: "http://{x}.example.com", variables: {x: {enum: ["www"]}}}`:                 `invalid servers: field default is required in {"enum":["www"]}`,
		`{url: "http://www.example.com", variables: {x: {enum: ["www"]}}}`:                 "invalid servers: server has undeclared variables",
		`{url: "http://{y}.example.com", variables: {x: {enum: ["www"]}}}`:                 "invalid servers: server has undeclared variables",
	} {
		t.Run(value, func(t *testing.T) {
			loader := NewLoader()
			doc, err := loader.LoadFromData([]byte(strings.Replace(spec, "@@@", value, 1)))
			require.NoError(t, err)
			err = doc.Validate(loader.Context)
			if expected == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, expected)
			}
		})
	}
}
