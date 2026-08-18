package openapi3filter_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3filter"
)

func TestEncodedPathAuthorizationBypass(t *testing.T) {
	var authCalls atomic.Int64
	var downstreamRoute atomic.Value
	var protectedID atomic.Value

	handler := &openapi3filter.ValidationHandler{
		File: "testdata/encoded_path_authorization_bypass.yml",
		AuthenticationFunc: func(_ context.Context, input *openapi3filter.AuthenticationInput) error {
			authCalls.Add(1)
			if input.RequestValidationInput.Request.Header.Get("X-API-Key") != "good" {
				return input.NewError(errors.New("missing or invalid X-API-Key"))
			}
			return nil
		},
		ErrorEncoder: func(_ context.Context, _ error, w http.ResponseWriter) {
			w.WriteHeader(http.StatusUnauthorized)
		},
	}
	err := handler.Load()
	require.NoError(t, err)

	downstream := mux.NewRouter().UseEncodedPath()
	downstream.HandleFunc("/public/status", func(w http.ResponseWriter, _ *http.Request) {
		downstreamRoute.Store("/public/status")
		w.WriteHeader(http.StatusNoContent)
	}).Methods(http.MethodGet)
	downstream.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		downstreamRoute.Store("/{id}")
		protectedID.Store(mux.Vars(r)["id"])
		w.WriteHeader(http.StatusAccepted)
	}).Methods(http.MethodGet)

	server := httptest.NewServer(handler.Middleware(downstream))
	defer server.Close()

	request := func(path string) int {
		t.Helper()
		response, err := http.Get(server.URL + path)
		require.NoError(t, err)
		defer response.Body.Close()
		return response.StatusCode
	}

	t.Run("canonical protected route is rejected", func(t *testing.T) {
		before := authCalls.Load()
		status := request("/ordinary")
		require.Equal(t, http.StatusUnauthorized, status)
		require.Equal(t, before+1, authCalls.Load())
	})

	t.Run("canonical public route reaches public handler", func(t *testing.T) {
		before := authCalls.Load()
		status := request("/public/status")
		require.Equal(t, http.StatusNoContent, status)
		route, _ := downstreamRoute.Load().(string)
		require.Equal(t, "/public/status", route)
		require.Equal(t, before, authCalls.Load())
	})

	t.Run("encoded slash skips authentication and reaches protected handler", func(t *testing.T) {
		before := authCalls.Load()
		status := request("/public%2Fstatus")
		require.Equal(t, http.StatusUnauthorized, status)
		route, _ := downstreamRoute.Load().(string)
		require.Equal(t, "/public/status", route)
		id, _ := protectedID.Load().(string)
		require.Equal(t, "", id)
		require.Equal(t, before, authCalls.Load())
	})

	t.Run("documented Router.Use wiring has the same bypass", func(t *testing.T) {
		var useAuthCalls atomic.Int64
		var useRoute atomic.Value
		var useProtectedID atomic.Value
		useHandler := &openapi3filter.ValidationHandler{
			File: "testdata/encoded_path_authorization_bypass.yml",
			AuthenticationFunc: func(_ context.Context, input *openapi3filter.AuthenticationInput) error {
				useAuthCalls.Add(1)
				return input.NewError(errors.New("missing or invalid X-API-Key"))
			},
			ErrorEncoder: func(_ context.Context, _ error, w http.ResponseWriter) {
				w.WriteHeader(http.StatusUnauthorized)
			},
		}
		err := useHandler.Load()
		require.NoError(t, err)

		useRouter := mux.NewRouter().UseEncodedPath()
		useRouter.HandleFunc("/public/status", func(w http.ResponseWriter, _ *http.Request) {
			useRoute.Store("/public/status")
			w.WriteHeader(http.StatusNoContent)
		}).Methods(http.MethodGet)
		useRouter.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
			useRoute.Store("/{id}")
			useProtectedID.Store(mux.Vars(r)["id"])
			w.WriteHeader(http.StatusAccepted)
		}).Methods(http.MethodGet)
		useRouter.Use(useHandler.Middleware)

		useServer := httptest.NewServer(useRouter)
		defer useServer.Close()
		response, err := http.Get(useServer.URL + "/public%2Fstatus")
		require.NoError(t, err)
		defer response.Body.Close()

		require.Equal(t, http.StatusUnauthorized, response.StatusCode)
		route, _ := useRoute.Load().(string)
		require.Equal(t, "", route)
		id, _ := useProtectedID.Load().(string)
		require.Equal(t, "", id)
		require.Equal(t, int64(0), useAuthCalls.Load())
	})
}
