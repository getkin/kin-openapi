// +build race

package openapi3_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

var specYAML = []byte(`
openapi: '3.0'
info:
  title: MyAPI
  version: '0.1'
paths:
  /test:
    get:
      parameters:
      - in: query
        name: example
        required: true
        schema:
          description: Some schemas
          type: string
          pattern: ^([-._a-zA-Z0-9])+$
      responses:
        200:
          "$ref": "#/components/responses/someResponse"
components:
  responses:
    someResponse:
      description: Some response
`)

type MiddlewareOpts struct {
	Router routers.Router
	t      *testing.T
}
type middleware = func(next http.Handler) http.Handler

func validator(opts MiddlewareOpts) middleware {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			openapi3.SchemaErrorDetailsDisabled = true
			route, pathParams, _ := opts.Router.FindRoute(r)

			requestValidationInput := &openapi3filter.RequestValidationInput{
				Request:    r,
				PathParams: pathParams,
				Route:      route,
				Options: &openapi3filter.Options{
					ExcludeResponseBody: true,
				},
			}
			err := openapi3filter.ValidateRequest(context.TODO(), requestValidationInput)
			require.NoError(opts.t, err)
			handler.ServeHTTP(w, r)
		})
	}
}

func testFunc(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Response")

}

func TestVisitJSONForDataRace(t *testing.T) {
	doc, err := openapi3.NewLoader().LoadFromData(specYAML)
	require.NoError(t, err)
	router, _ := gorillamux.NewRouter(doc)

	// setup middleware
	middle := validator(MiddlewareOpts{
		Router: router,
		t:      t,
	})
	h := middle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { testFunc(w, r) }))

	// spin up server with middleware
	srv := httptest.NewServer(h)
	defer srv.Close()

	r, err := http.NewRequest(http.MethodGet, srv.URL+"/test?example=alonso", nil)
	if err != nil {
		t.Fatal(err)
	}
	// execute simultaneously request
	execute(t, 2, srv, r)
}

func execute(t *testing.T, noReqs int, srv *httptest.Server, r *http.Request) {

	wg := &sync.WaitGroup{}
	wg.Add(noReqs)
	starter := make(chan struct{})
	for i := 0; i < noReqs; i++ {
		go func() {
			<-starter
			resp, err := srv.Client().Do(r)
			require.NoError(t, err)
			body, err := io.ReadAll(resp.Body)
			require.EqualValues(t, "Response", string(body))
			require.NoError(t, err)
			wg.Done()
		}()
	}
	close(starter)
	wg.Wait()
}
