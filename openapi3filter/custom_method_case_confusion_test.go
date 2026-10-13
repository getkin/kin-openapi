package openapi3filter_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3filter"
)

const customMethodCaseSpec = `openapi: 3.2.0
info:
  title: custom method identity fixture
  version: 1.0.0
security:
  - ApiKeyAuth: []
paths:
  /objects:
    post:
      operationId: protected-post-control
      responses:
        "204":
          description: protected control
    additionalOperations:
      copy:
        operationId: public-lowercase-copy
        security: []
        responses:
          "204":
            description: intentionally public lowercase custom method
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: X-API-Key
`

func TestCustomMethodCaseAuthenticationBypass(t *testing.T) {
	t.Parallel()

	var authCalls atomic.Int64
	var downstreamMethods atomic.Value

	specPath := filepath.Join(t.TempDir(), "openapi.yaml")
	err := os.WriteFile(specPath, []byte(customMethodCaseSpec), 0o600)
	require.NoError(t, err)

	handler := &openapi3filter.ValidationHandler{
		File: specPath,
		AuthenticationFunc: func(_ context.Context, input *openapi3filter.AuthenticationInput) error {
			authCalls.Add(1)
			return input.NewError(errors.New("missing API key"))
		},
		ErrorEncoder: func(_ context.Context, _ error, w http.ResponseWriter) {
			w.WriteHeader(http.StatusUnauthorized)
		},
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			downstreamMethods.Store(r.Method)
			w.WriteHeader(http.StatusNoContent)
		}),
	}
	err = handler.Load()
	require.NoError(t, err)

	serve := func(method string) int {
		t.Helper()
		downstreamMethods.Store("")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(method, "/objects", nil))
		return recorder.Code
	}

	t.Run("protected standard method invokes authentication", func(t *testing.T) {
		before := authCalls.Load()
		status := serve(http.MethodPost)
		require.Equal(t, http.StatusUnauthorized, status)
		require.Equal(t, before+1, authCalls.Load())
		require.Equal(t, "", downstreamMethods.Load())
	})

	t.Run("declared lowercase custom method matches its public operation", func(t *testing.T) {
		before := authCalls.Load()
		status := serve("copy")
		require.Equal(t, http.StatusNoContent, status)
		require.Equal(t, before, authCalls.Load())
		require.Equal(t, "copy", downstreamMethods.Load())
	})

	t.Run("uppercase custom method does not inherit lowercase public operation", func(t *testing.T) {
		before := authCalls.Load()
		status := serve("COPY")
		require.Equal(t, http.StatusUnauthorized, status)
		require.Equal(t, before, authCalls.Load())
		require.Equal(t, "", downstreamMethods.Load())
	})
}
