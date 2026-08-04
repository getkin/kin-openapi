package gorillamux

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers"
)

// customMethodsSpec exercises the OpenAPI 3.2 `query` field and custom methods
// held in `additionalOperations`.
const customMethodsSpec = `
openapi: 3.2.0
info:
  title: MyAPI
  version: "0.1"
paths:
  /pets:
    query:
      responses:
        "200":
          description: OK
    additionalOperations:
      COPY:
        responses:
          "200":
            description: OK
  /pets/{id}:
    parameters:
      - name: id
        in: path
        required: true
        schema:
          type: string
    get:
      responses:
        "200":
          description: OK
    query:
      responses:
        "200":
          description: OK
    additionalOperations:
      COPY:
        responses:
          "200":
            description: OK
`

func TestRouterCustomMethods(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(customMethodsSpec))
	require.NoError(t, err)
	require.NoError(t, doc.Validate(context.Background()))

	router, err := NewRouter(doc)
	require.NoError(t, err)

	pets := doc.Paths.Value("/pets")
	petsID := doc.Paths.Value("/pets/{id}")

	for _, tc := range []struct {
		method    string
		uri       string
		operation *openapi3.Operation
		params    map[string]string
	}{
		{openapi3.MethodQuery, "/pets", pets.Query, nil},
		{"COPY", "/pets", pets.AdditionalOperations["COPY"], nil},
		{openapi3.MethodQuery, "/pets/fido", petsID.Query, map[string]string{"id": "fido"}},
		{"COPY", "/pets/fido", petsID.AdditionalOperations["COPY"], map[string]string{"id": "fido"}},
		{http.MethodGet, "/pets/fido", petsID.Get, map[string]string{"id": "fido"}},
	} {
		t.Run(tc.method+" "+tc.uri, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, tc.uri, nil)
			require.NoError(t, err)

			route, pathParams, err := router.FindRoute(req)
			require.NoError(t, err)
			require.Same(t, tc.operation, route.Operation)
			require.Equal(t, tc.method, route.Method)
			for name, want := range tc.params {
				require.Equal(t, want, pathParams[name])
			}
		})
	}

	t.Run("unregistered method", func(t *testing.T) {
		for _, uri := range []string{"/pets", "/pets/fido"} {
			req, err := http.NewRequest("PURGE", uri, nil)
			require.NoError(t, err)

			route, pathParams, err := router.FindRoute(req)
			require.EqualError(t, err, routers.ErrMethodNotAllowed.Error())
			require.Nil(t, route)
			require.Nil(t, pathParams)
		}
	})
}
