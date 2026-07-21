package openapi3filter_test

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

// These tests cover the fix for a nil-pointer panic: an OpenAPI 3.1 document
// may legally declare `type: array` without `items`, which passes
// doc.Validate() with Schema.Items == nil. Several openapi3filter decoders
// used to dereference Items with no nil check, panicking on any request that
// reached the affected operation. The decoders must instead return a clean
// error.

const urlencodedArrayNoItemsSpec31 = `
openapi: 3.1.0
info: {title: t, version: 1.0.0}
paths:
  /f:
    post:
      requestBody:
        required: true
        content:
          application/x-www-form-urlencoded:
            schema:
              type: object
              properties:
                tags: {type: array}
      responses:
        '200': {description: ok}
`

const multipartArrayNoItemsSpec31 = `
openapi: 3.1.0
info: {title: t, version: 1.0.0}
paths:
  /upload:
    post:
      requestBody:
        required: true
        content:
          multipart/form-data:
            schema:
              type: object
              properties:
                tags: {type: array}
      responses:
        '200': {description: ok}
`

const responseHeaderArrayNoItemsSpec31 = `
openapi: 3.1.0
info: {title: t, version: 1.0.0}
paths:
  /hdr:
    get:
      responses:
        '200':
          description: ok
          headers:
            X-Tags:
              schema: {type: array}
`

func catchPanicValue(fn func()) (panicValue any) {
	defer func() { panicValue = recover() }()
	fn()
	return nil
}

// The 3.1 schema without items must still be accepted by Validate (spec-legal
// under JSON Schema 2020-12), leaving Items == nil.
func TestArrayItemsNil_OpenAPI31Accepts(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(urlencodedArrayNoItemsSpec31))
	require.NoError(t, err)
	require.NoError(t, doc.Validate(loader.Context))
	schema := doc.Paths.Value("/f").Post.RequestBody.Value.Content.
		Get("application/x-www-form-urlencoded").Schema.Value.Properties["tags"].Value
	require.True(t, schema.Type.Is("array"))
	require.Nil(t, schema.Items)
}

func TestArrayItemsNil_UrlencodedNoPanic(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(urlencodedArrayNoItemsSpec31))
	require.NoError(t, err)
	require.NoError(t, doc.Validate(loader.Context))
	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/f", strings.NewReader("tags=a"))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	route, pathParams, err := router.FindRoute(req)
	require.NoError(t, err)

	var reqErr error
	panicVal := catchPanicValue(func() {
		reqErr = openapi3filter.ValidateRequest(context.Background(), &openapi3filter.RequestValidationInput{
			Request: req, PathParams: pathParams, Route: route,
			Options: &openapi3filter.Options{AuthenticationFunc: openapi3filter.NoopAuthenticationFunc},
		})
	})
	require.Nil(t, panicVal, "must not panic")
	require.Error(t, reqErr, "must return a clean error")
	require.Contains(t, reqErr.Error(), "array items required")
}

func TestArrayItemsNil_MultipartNoPanic(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(multipartArrayNoItemsSpec31))
	require.NoError(t, err)
	require.NoError(t, doc.Validate(loader.Context))
	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	require.NoError(t, w.WriteField("tags", "a"))
	require.NoError(t, w.Close())

	req, err := http.NewRequest(http.MethodPost, "/upload", &buf)
	require.NoError(t, err)
	req.Header.Set("Content-Type", w.FormDataContentType())
	route, pathParams, err := router.FindRoute(req)
	require.NoError(t, err)

	var reqErr error
	panicVal := catchPanicValue(func() {
		reqErr = openapi3filter.ValidateRequest(context.Background(), &openapi3filter.RequestValidationInput{
			Request: req, PathParams: pathParams, Route: route,
			Options: &openapi3filter.Options{AuthenticationFunc: openapi3filter.NoopAuthenticationFunc},
		})
	})
	require.Nil(t, panicVal, "must not panic")
	require.Error(t, reqErr, "must return a clean error")
}

func TestArrayItemsNil_ResponseHeaderNoPanic(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(responseHeaderArrayNoItemsSpec31))
	require.NoError(t, err)
	require.NoError(t, doc.Validate(loader.Context))
	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "/hdr", nil)
	require.NoError(t, err)
	route, pathParams, err := router.FindRoute(req)
	require.NoError(t, err)

	var respErr error
	panicVal := catchPanicValue(func() {
		respErr = openapi3filter.ValidateResponse(context.Background(), &openapi3filter.ResponseValidationInput{
			RequestValidationInput: &openapi3filter.RequestValidationInput{
				Request: req, PathParams: pathParams, Route: route,
				Options: &openapi3filter.Options{AuthenticationFunc: openapi3filter.NoopAuthenticationFunc},
			},
			Status: http.StatusOK,
			Header: http.Header{"X-Tags": []string{"a,b"}},
			Body:   io.NopCloser(strings.NewReader("")),
		})
	})
	require.Nil(t, panicVal, "must not panic")
	require.Error(t, respErr, "must return a clean error")
}
