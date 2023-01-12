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
		_, err := w.Write([]byte(body))
		if err != nil {
			panic(err)
		}
	}))
	defer ts.Close()

	rootSpec := []byte(fmt.Sprintf(
		`{"openapi":"3.0.0","info":{"title":"MyAPI","version":"0.1","description":"An API"},"paths":{},"components":{"schemas":{"Bar1":{"$ref":"%s#/components/schemas/Foo"}}}}`,
		ts.URL,
	))

	wg := &sync.WaitGroup{}
	n := 10
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			loader := NewLoader()
			loader.IsExternalRefsAllowed = true
			doc, err := loader.LoadFromData(rootSpec)
			require.NoError(t, err)
			require.NotNil(t, doc)
		}()
	}
	wg.Wait()
}
