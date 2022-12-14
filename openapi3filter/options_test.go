package openapi3filter_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func ExampleOptions_WithCustomSchemaErrorFunc() {
	const spec = `
openapi: 3.0.0
info:
  title: 'Validator'
  version: 0.0.1
paths:
  /some:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                field:
                  title: Some field
                  type: integer
      responses:
        '200':
          description: Created
`

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	if err != nil {
		panic(err)
	}

	if err = doc.Validate(loader.Context); err != nil {
		panic(err)
	}

	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		panic(err)
	}

	opts := &openapi3filter.Options{}

	opts.WithCustomSchemaErrorFunc(func(err *openapi3.SchemaError) string {
		return fmt.Sprintf(`field "%s" must be an integer`, err.Schema.Title)
	})

	req, err := http.NewRequest(http.MethodPost, "/some", strings.NewReader(`{"field":"not integer"}`))
	if err != nil {
		panic(err)
	}

	req.Header.Add("Content-Type", "application/json")

	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		panic(err)
	}

	validationInput := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
		Options:    opts,
	}
	err = openapi3filter.ValidateRequest(context.Background(), validationInput)

	fmt.Println(err.Error())

	// Output: request body has an error: doesn't match schema: field "Some field" must be an integer
}
