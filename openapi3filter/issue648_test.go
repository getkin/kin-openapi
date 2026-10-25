package openapi3filter_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func TestIssue648(t *testing.T) {
	t.Parallel()

	spec := `
openapi: 3.0.0
info:
  version: 1.0.0
  title: Sample API
paths:
  /items:
    get:
      parameters:
      - name: spaces
        in: query
        style: spaceDelimited
        schema:
          type: array
          items:
            type: integer
      - name: pipes
        in: query
        style: pipeDelimited
        schema:
          type: array
          items:
            type: integer
      - name: form
        in: query
        schema:
          type: array
          items:
            type: integer
      - name: filter
        in: query
        style: deepObject
        schema:
          type: object
          properties:
            a:
              type: integer
      responses:
        '200':
          description: Successful response
`[1:]

	loader := openapi3.NewLoader()
	ctx := loader.Context

	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	err = doc.Validate(ctx)
	require.NoError(t, err)

	params := doc.Paths.Value("/items").Get.Parameters
	for name, want := range map[string]*openapi3.SerializationMethod{
		"spaces": {Style: openapi3.SerializationSpaceDelimited, Explode: false},
		"pipes":  {Style: openapi3.SerializationPipeDelimited, Explode: false},
		"form":   {Style: openapi3.SerializationForm, Explode: true},
		"filter": {Style: openapi3.SerializationDeepObject, Explode: true},
	} {
		sm, err := params.GetByInAndName(openapi3.ParameterInQuery, name).SerializationMethod()
		require.NoError(t, err)
		require.Equal(t, want, sm, name)
	}

	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	for _, query := range []string{
		"spaces=1%202%203",
		"pipes=1|2|3",
		"form=1&form=2&form=3",
		"filter[a]=1",
	} {
		t.Run(query, func(t *testing.T) {
			httpReq, err := http.NewRequest(http.MethodGet, "/items?"+query, nil)
			require.NoError(t, err)

			route, pathParams, err := router.FindRoute(httpReq)
			require.NoError(t, err)

			err = openapi3filter.ValidateRequest(ctx, &openapi3filter.RequestValidationInput{
				Request:    httpReq,
				PathParams: pathParams,
				Route:      route,
			})
			require.NoError(t, err)
		})
	}
}
