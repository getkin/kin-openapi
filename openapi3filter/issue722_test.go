package openapi3filter_test

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func TestValidateMultipartFormDataContainingAllOf(t *testing.T) {
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
              allOf:
              - $ref: '#/components/schemas/Category'
              - properties:
                  file:
                    type: string
                    format: binary
                  description:
                    type: string
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
		t.Fatal(err)
	}
	if err = doc.Validate(loader.Context); err != nil {
		t.Fatal(err)
	}

	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		t.Fatal(err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	{ // Add file data
		fw, err := writer.CreateFormFile("file", "hello.txt")
		if err != nil {
			t.Fatal(err)
		}
		if _, err = io.Copy(fw, strings.NewReader("hello")); err != nil {
			t.Fatal(err)
		}
	}

	{ // Add a single "name" item as part data
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="name"`)
		fw, err := writer.CreatePart(h)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = io.Copy(fw, strings.NewReader(`foo`)); err != nil {
			t.Fatal(err)
		}
	}

	{ // Add a single "description" item as part data
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="description"`)
		fw, err := writer.CreatePart(h)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = io.Copy(fw, strings.NewReader(`description note`)); err != nil {
			t.Fatal(err)
		}
	}

	writer.Close()

	req, err := http.NewRequest(http.MethodPost, "/test", bytes.NewReader(body.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		t.Fatal(err)
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
		t.Error(err)
	}
}
