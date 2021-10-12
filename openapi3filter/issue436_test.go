package openapi3filter

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/stretchr/testify/require"
)

func TestIssue436(t *testing.T) {
	const testSchema = `
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
          #application/json:
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

	doc, err := openapi3.NewLoader().LoadFromData([]byte(testSchema))
	require.NoError(t, err)
	err = doc.Validate(context.Background())
	require.NoError(t, err)

	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"`, "categories"))
	h.Set("Content-Type", "application/json")
	fw, err := writer.CreatePart(h)
	require.NoError(t, err)
	_, err = io.Copy(fw, strings.NewReader(`[{"name": "foo"}]`))
	require.NoError(t, err)

	fw, err = writer.CreateFormFile("file", "hello.txt")
	require.NoError(t, err)
	_, err = io.Copy(fw, strings.NewReader("hello"))
	require.NoError(t, err)

	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/test", bytes.NewReader(body.Bytes()))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	route, pathParams, _ := router.FindRoute(req)

	reqBody := route.Operation.RequestBody.Value

	requestValidationInput := &RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	}

	err = ValidateRequestBody(context.TODO(), requestValidationInput, reqBody)
	if err == nil {
		fmt.Println("Valid")
	} else {
		fmt.Printf("NOT valid. %s\n", err)
	}
}
