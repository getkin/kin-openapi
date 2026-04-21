package openapi3

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue741(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		body := `{"openapi":"3.0.0","info":{"title":"MyAPI","version":"0.1","description":"An API"},"paths":{},"components":{"schemas":{"Foo":{"type":"string"}}}}`
		if _, err := w.Write([]byte(body)); err != nil {
			panic(err)
		}
	}))
	defer ts.Close()

	rootSpec := fmt.Appendf(nil,
		`{"openapi":"3.0.0","info":{"title":"MyAPI","version":"0.1","description":"An API"},"paths":{},"components":{"schemas":{"Bar1":{"$ref":"%s#/components/schemas/Foo"}}}}`,
		ts.URL,
	)

	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			loader := NewLoader()
			loader.IsExternalRefsAllowed = true
			doc, err := loader.LoadFromData(rootSpec)
			require.NoError(t, err)
			require.NotNil(t, doc)
		})
	}
	wg.Wait()
}
