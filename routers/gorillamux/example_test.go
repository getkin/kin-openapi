package gorillamux_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func Example() {
	ctx := context.Background()
	loader := &openapi3.Loader{Context: ctx, IsExternalRefsAllowed: true}
	doc, err := loader.LoadFromFile("../../openapi3/testdata/pathref.openapi.yml")
	if err != nil {
		panic(err)
	}
	if err = doc.Validate(ctx); err != nil {
		panic(err)
	}
	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		panic(err)
	}
	httpReq, err := http.NewRequest(http.MethodGet, "/test", nil)
	if err != nil {
		panic(err)
	}

	route, pathParams, err := router.FindRoute(httpReq)
	if err != nil {
		panic(err)
	}

	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    httpReq,
		PathParams: pathParams,
		Route:      route,
	}
	if err := openapi3filter.ValidateRequest(ctx, requestValidationInput); err != nil {
		panic(err)
	}

	responseValidationInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: requestValidationInput,
		Status:                 200,
		Header:                 http.Header{"Content-Type": []string{"application/json"}},
	}
	responseValidationInput.SetBodyBytes([]byte(`{}`))

	err = openapi3filter.ValidateResponse(ctx, responseValidationInput)
	fmt.Println(err)
	// Output:
	// response body doesn't match schema pathref.openapi.yml#/components/schemas/TestSchema: value must be a string
	// Schema:
	//   {
	//     "type": "string"
	//   }
	//
	// Value:
	//   {}
}
