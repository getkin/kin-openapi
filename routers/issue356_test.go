package routers_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/getkin/kin-openapi/routers/legacy"
	"github.com/stretchr/testify/require"
)

func TestIssue356(t *testing.T) {
	spec := func(servers string) []byte {
		return []byte(`
openapi: 3.0.0
info:
  title: Example
  version: '1.0'
  description: test
servers:
` + servers + `
paths:
  /test:
    post:
      responses:
        '201':
          description: Created
          content:
            application/json:
              schema: {type: object}
      requestBody:
        content:
          application/json:
            schema: {type: object}
        description: ''
      description: Create a test object
`)
	}

	for servers, expectError := range map[string]bool{
		`
- url: http://localhost:3000/base
- url: /base
`: false,

		`
- url: /base
- url: http://localhost:3000/base
`: false,

		`- url: /base`: false,

		`- url: http://localhost:3000/base`: true,

		``: true,
	} {
		loader := &openapi3.Loader{Context: context.Background()}
		t.Logf("using servers: %q (%v)", servers, expectError)
		doc, err := loader.LoadFromData(spec(servers))
		require.NoError(t, err)
		err = doc.Validate(context.Background())
		require.NoError(t, err)

		for i, newRouter := range []func(*openapi3.T) (routers.Router, error){gorillamux.NewRouter, legacy.NewRouter} {
			t.Logf("using NewRouter from %s", map[int]string{0: "gorillamux", 1: "legacy"}[i])
			router, err := newRouter(doc)
			require.NoError(t, err)

			if true {
				t.Logf("using naked newRouter")
				httpReq, err := http.NewRequest(http.MethodPost, "/base/test", strings.NewReader(`{}`))
				require.NoError(t, err)
				httpReq.Header.Set("Content-Type", "application/json")

				route, pathParams, err := router.FindRoute(httpReq)
				if expectError {
					require.Error(t, err, routers.ErrPathNotFound)
					return
				}
				require.NoError(t, err)

				requestValidationInput := &openapi3filter.RequestValidationInput{
					Request:    httpReq,
					PathParams: pathParams,
					Route:      route,
				}
				err = openapi3filter.ValidateRequest(context.Background(), requestValidationInput)
				require.NoError(t, err)
			}

			if true {
				t.Logf("using httptest.NewServer")
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					route, pathParams, err := router.FindRoute(r)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(err.Error()))
						return
					}

					requestValidationInput := &openapi3filter.RequestValidationInput{
						Request:    r,
						PathParams: pathParams,
						Route:      route,
					}
					err = openapi3filter.ValidateRequest(r.Context(), requestValidationInput)
					require.NoError(t, err)

					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte("{}"))
				}))
				defer ts.Close()

				req, err := http.NewRequest(http.MethodPost, ts.URL+"/base/test", strings.NewReader(`{}`))
				require.NoError(t, err)
				req.Header.Set("Content-Type", "application/json")

				rep, err := http.DefaultClient.Do(req)
				require.NoError(t, err)
				defer rep.Body.Close()
				body, err := ioutil.ReadAll(rep.Body)
				require.NoError(t, err)

				if expectError {
					require.Equal(t, 500, rep.StatusCode)
					require.Equal(t, routers.ErrPathNotFound.Error(), string(body))
					return
				}
				require.Equal(t, 200, rep.StatusCode)
				require.Equal(t, "{}", string(body))
			}
		}
	}
}
