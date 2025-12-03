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

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func TestValidateCsvFileUpload(t *testing.T) {
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
                  format: string
      responses:
        '200':
          description: Created
`

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	tests := []struct {
		csvData string
		wantErr bool
	}{
		{
			`foo,bar`,
			false,
		},
		{
			`"foo","bar"`,
			false,
		},
		{
			`foo,bar
baz,qux`,
			false,
		},
		{
			`foo,bar
baz,qux,quux`,
			true,
		},
		{
			`"""`,
			true,
		},
	}
	for _, tt := range tests {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		{ // Add file data
			h := make(textproto.MIMEHeader)
			h.Set("Content-Disposition", `form-data; name="file"; filename="hello.csv"`)
			h.Set("Content-Type", "text/csv")

			fw, err := writer.CreatePart(h)
			require.NoError(t, err)
			_, err = io.Copy(fw, strings.NewReader(tt.csvData))

			require.NoError(t, err)
		}

		writer.Close()

		req, err := http.NewRequest(http.MethodPost, "/test", bytes.NewReader(body.Bytes()))
		require.NoError(t, err)

		req.Header.Set("Content-Type", writer.FormDataContentType())

		route, pathParams, err := router.FindRoute(req)
		require.NoError(t, err)

		if err = openapi3filter.ValidateRequestBody(
			context.Background(),
			&openapi3filter.RequestValidationInput{
				Request:    req,
				PathParams: pathParams,
				Route:      route,
			},
			route.Operation.RequestBody.Value,
		); err != nil {
			if !tt.wantErr {
				t.Errorf("got %v", err)
			}
			continue
		}
		if tt.wantErr {
			t.Errorf("want err")
		}
	}
}
