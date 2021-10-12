package openapi3filter_test

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func Example_validateMultipartFormData() {
	const spec = `
openapi: 3.0.0
info:
  title: 'Validator'
  version: 0.0.1
paths:
  /test:
    post:
      requestBody:
        required: true
        content:
          multipart/form-data:
            schema:
              type: object
              required:
                - file
              properties:
                file:
                  type: string
                  format: binary
                categories:
                  type: array
                  items:
                    $ref: "#/components/schemas/Category"
      responses:
        '200':
          description: Created

components:
  schemas:
    Category:
      type: object
      properties:
        name:
          type: string
      required:
        - name
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

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	{ // Add a single "categories" item as part data
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="categories"`)
		h.Set("Content-Type", "application/json")
		fw, err := writer.CreatePart(h)
		if err != nil {
			panic(err)
		}
		if _, err = io.Copy(fw, strings.NewReader(`{"name": "foo"}`)); err != nil {
			panic(err)
		}
	}

	{ // Add a single "categories" item as part data, again
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="categories"`)
		h.Set("Content-Type", "application/json")
		fw, err := writer.CreatePart(h)
		if err != nil {
			panic(err)
		}
		if _, err = io.Copy(fw, strings.NewReader(`{"name": "bar"}`)); err != nil {
			panic(err)
		}
	}

	{ // Add file data
		fw, err := writer.CreateFormFile("file", "hello.txt")
		if err != nil {
			panic(err)
		}
		if _, err = io.Copy(fw, strings.NewReader("hello")); err != nil {
			panic(err)
		}
	}

	writer.Close()

	req, err := http.NewRequest(http.MethodPost, "/test", bytes.NewReader(body.Bytes()))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		panic(err)
	}

	if err = openapi3filter.ValidateRequestBody(
		context.Background(),
		&openapi3filter.RequestValidationInput{
			Request:    req,
			PathParams: pathParams,
			Route:      route,
		},
		route.Operation.RequestBody.Value,
	); err != nil {
		panic(err)
	}
	// Output:
}
