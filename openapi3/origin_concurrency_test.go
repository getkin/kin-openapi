package openapi3

import (
	"sync"
	"testing"

	goyaml "go.yaml.in/yaml/v3"
)

const concurrentSpec = `
openapi: 3.0.0
info: {title: t, version: "1"}
paths:
  /a:
    get:
      responses:
        "200": {description: ok}
`

// A decode reaches UnmarshalYAML through package state, so a caller handing
// these types to the yaml package itself decodes while another goroutine holds
// that state. Such a decode indexed no tree, so it gets no origins, and reads
// the state without racing the decode that set it.
//
// Run with -race, which the workflow does by name. Skipped in -short mode: the
// concurrency is the point of the test, and the Windows runners already retry
// the short suite around flaky memory corruption in Go 1.26 and 1.27.
func TestConcurrentDecodeDoesNotLeakOrigins(t *testing.T) {
	if testing.Short() {
		t.Skip("decodes concurrently; runs in the dedicated -race step")
	}

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		for range 50 {
			loader := NewLoader()
			loader.IncludeOrigin = true
			if _, err := loader.LoadFromData([]byte(concurrentSpec)); err != nil {
				t.Error(err)
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		for range 50 {
			var doc T
			if err := goyaml.Unmarshal([]byte(concurrentSpec), &doc); err != nil {
				t.Error(err)
				return
			}
			if doc.Info != nil && doc.Info.Origin != nil {
				t.Error("a decode through the yaml package took the origins of a decode running beside it")
				return
			}
		}
	}()

	// A Loader that asked for no origins holds no lock, so it runs beside the
	// one above rather than waiting for it, and must still come back bare.
	go func() {
		defer wg.Done()
		for range 50 {
			loader := NewLoader()
			doc, err := loader.LoadFromData([]byte(concurrentSpec))
			if err != nil {
				t.Error(err)
				return
			}
			if doc.Info.Origin != nil {
				t.Error("a load that asked for no origins took the origins of a decode running beside it")
				return
			}
		}
	}()

	wg.Wait()
}
