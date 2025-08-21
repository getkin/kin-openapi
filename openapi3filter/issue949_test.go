package openapi3filter_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/stretchr/testify/require"
)

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
            schema:
              type: object
              properties:
                file:
                  type: string
                  format: binary
                counts:
                  type: object
                  properties:
                    name:
                      type: string
                    count:
                      type: integer
                primitive:
                  type: integer
      responses:
        '200':
          description: OK
`

type count struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func TestIssue949(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(testSchema))
	require.NoError(t, err)

	err = doc.Validate(context.Background())
	require.NoError(t, err)

	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the counts object to the request body
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"`, "counts"))
	h.Set("Content-Type", "application/json")
	fw, err := writer.CreatePart(h)
	require.NoError(t, err)

	countStruct := count{Name: "foo", Count: 7}
	countBody, err := json.Marshal(countStruct)
	require.NoError(t, err)
	_, err = fw.Write(countBody)
	require.NoError(t, err)

	// Add the file to the request body
	fw, err = writer.CreateFormFile("file", "hello.txt")
	require.NoError(t, err)

	_, err = io.Copy(fw, strings.NewReader("hello"))
	require.NoError(t, err)

	// Add the primitive integer to the request body
	fw, err = writer.CreateFormField("primitive")
	require.NoError(t, err)
	_, err = fw.Write([]byte("1"))
	require.NoError(t, err)

	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/test", bytes.NewReader(body.Bytes()))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	route, pathParams, err := router.FindRoute(req)
	require.NoError(t, err)

	reqBody := route.Operation.RequestBody.Value

	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	}

	err = openapi3filter.ValidateRequestBody(context.TODO(), requestValidationInput, reqBody)
	require.NoError(t, err)
}
