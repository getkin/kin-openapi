package openapi3filter

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	legacyrouter "github.com/getkin/kin-openapi/routers/legacy"
)

func TestValidationWithIntegerEnum(t *testing.T) {
	t.Run("PUT Request", func(t *testing.T) {
		const spec = `
openapi: 3.0.0
info:
  title: Example integer enum
  version: '0.1'
paths:
  /sample:
    put:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                exenum:
                  type: integer
                  enum:
                    - 0
                    - 1
                    - 2
                    - 3
                  example: 0
                  nullable: true
      responses:
        '200':
          description: Ok
`

		loader := openapi3.NewLoader()
		doc, err := loader.LoadFromData([]byte(spec))
		require.NoError(t, err)

		router, err := legacyrouter.NewRouter(doc)
		require.NoError(t, err)

		tests := []struct {
			data    []byte
			wantErr bool
		}{
			{
				[]byte(`{"exenum": 1}`),
				false,
			},
			{
				[]byte(`{"exenum": "1"}`),
				true,
			},
			{
				[]byte(`{"exenum": null}`),
				false,
			},
			{
				[]byte(`{}`),
				false,
			},
		}

		for _, tt := range tests {
			body := bytes.NewReader(tt.data)
			req, err := http.NewRequest("PUT", "/sample", body)
			require.NoError(t, err)
			req.Header.Add(headerCT, "application/json")

			route, pathParams, err := router.FindRoute(req)
			require.NoError(t, err)

			requestValidationInput := &RequestValidationInput{
				Request:    req,
				PathParams: pathParams,
				Route:      route,
			}
			err = ValidateRequest(loader.Context, requestValidationInput)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		}
	})

	t.Run("GET Request", func(t *testing.T) {
		const spec = `
openapi: 3.0.0
info:
  title: Example integer enum
  version: '0.1'
paths:
  /sample:
    get:
      parameters:
        - in: query
          name: exenum
          schema:
            type: integer
            enum:
              - 0
              - 1
              - 2
              - 3
      responses:
        '200':
          description: Ok
`

		loader := openapi3.NewLoader()
		doc, err := loader.LoadFromData([]byte(spec))
		require.NoError(t, err)

		router, err := legacyrouter.NewRouter(doc)
		require.NoError(t, err)

		tests := []struct {
			exenum  string
			wantErr bool
		}{
			{
				"1",
				false,
			},
			{
				"4",
				true,
			},
		}

		for _, tt := range tests {
			req, err := http.NewRequest("GET", "/sample?exenum="+tt.exenum, nil)
			require.NoError(t, err)
			route, pathParams, err := router.FindRoute(req)
			require.NoError(t, err)

			requestValidationInput := &RequestValidationInput{
				Request:    req,
				PathParams: pathParams,
				Route:      route,
			}
			err = ValidateRequest(loader.Context, requestValidationInput)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		}
	})
}

func TestValidationWithStringEnum(t *testing.T) {
	const spec = `
openapi: 3.0.0
info:
  title: Example string enum
  version: '0.1'
paths:
  /sample:
    put:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                exenum:
                  type: string
                  enum:
                    - "0"
                    - "1"
                    - "2"
                    - "3"
                  example: "0"
      responses:
        '200':
          description: Ok
`

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	router, err := legacyrouter.NewRouter(doc)
	require.NoError(t, err)

	tests := []struct {
		data    []byte
		wantErr bool
	}{
		{
			[]byte(`{"exenum": "1"}`),
			false,
		},
		{
			[]byte(`{"exenum": 1}`),
			true,
		},
		{
			[]byte(`{"exenum": null}`),
			true,
		},
		{
			[]byte(`{}`),
			false,
		},
	}

	for _, tt := range tests {
		body := bytes.NewReader(tt.data)
		req, err := http.NewRequest("PUT", "/sample", body)
		require.NoError(t, err)
		req.Header.Add(headerCT, "application/json")

		route, pathParams, err := router.FindRoute(req)
		require.NoError(t, err)

		requestValidationInput := &RequestValidationInput{
			Request:    req,
			PathParams: pathParams,
			Route:      route,
		}
		err = ValidateRequest(loader.Context, requestValidationInput)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}
