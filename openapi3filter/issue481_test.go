package openapi3filter_test

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func TestIssue481(t *testing.T) {
	spec := `
openapi: 3.0.0
info:
  title: 'Validator'
  version: 0.0.1
paths:
  /upload:
    post:
      responses:
        '200':
          description: OK
      requestBody:
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
            encoding:
              file:
                contentType: '*/*'
`

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	var calls int
	openapi3filter.RegisterBodyDecoder("*/*", func(body io.Reader, _ http.Header, _ *openapi3.SchemaRef, _ openapi3filter.EncodingFn) (any, error) {
		calls++
		data, err := io.ReadAll(body)
		if err != nil {
			return nil, err
		}
		return string(data), nil
	})
	defer openapi3filter.UnregisterBodyDecoder("*/*")

	// Only the "*/*" catch-all decoder is registered: it decodes every one of
	// these media types.
	for _, contentType := range []string{
		"application/pdf",
		"image/png",
		"application/x-made-up",
	} {
		t.Run(contentType, func(t *testing.T) {
			previous := openapi3filter.RegisteredBodyDecoder(contentType)
			require.Nil(t, previous)
			calls = 0

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			h := make(textproto.MIMEHeader)
			h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file", "some.file"))
			h.Set("Content-Type", contentType)
			fw, err := writer.CreatePart(h)
			require.NoError(t, err)
			_, err = fw.Write([]byte("%PDF-1.4 whatever"))
			require.NoError(t, err)
			require.NoError(t, writer.Close())

			req, err := http.NewRequest(http.MethodPost, "/upload", body)
			require.NoError(t, err)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			route, pathParams, err := router.FindRoute(req)
			require.NoError(t, err)

			err = openapi3filter.ValidateRequest(t.Context(), &openapi3filter.RequestValidationInput{
				Request:    req,
				PathParams: pathParams,
				Route:      route,
			})
			require.NoError(t, err)
			require.Equal(t, 1, calls)
		})
	}
}
