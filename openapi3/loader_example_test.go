package openapi3_test

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// ExampleLoader_JoinFunc demonstrates how to use JoinFunc to load a multi-file
// OpenAPI spec from a virtual path scheme (e.g. git refs like "rev:file.yaml").
//
// When loading specs from non-filesystem sources via LoadFromDataWithPath and
// ReadFromURIFunc, the base path may use a custom prefix convention. The default
// path resolution uses path.Dir which does not understand such prefixes. JoinFunc
// lets the caller override path resolution to preserve the prefix.
func ExampleLoader_JoinFunc() {
	// Set up test files in a temp directory.
	dir, _ := os.MkdirTemp("", "joinfunc-example")
	defer os.RemoveAll(dir)

	root := `openapi: "3.0.0"
info:
  title: Pet API
  version: "1.0"
paths: {}
components:
  schemas:
    Pet:
      $ref: "./schemas/pet.yaml"
`
	pet := `type: object
properties:
  name:
    type: string
`
	os.MkdirAll(filepath.Join(dir, "schemas"), 0o755)
	os.WriteFile(filepath.Join(dir, "root.yaml"), []byte(root), 0o644)
	os.WriteFile(filepath.Join(dir, "schemas", "pet.yaml"), []byte(pet), 0o644)

	// Use a "rev:" prefix to simulate a virtual path scheme.
	const prefix = "rev:"

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	// ReadFromURIFunc strips the prefix and reads from the real filesystem.
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, location *url.URL) ([]byte, error) {
		p := location.Path
		if strings.HasPrefix(p, prefix) {
			p = p[len(prefix):]
		}
		return os.ReadFile(filepath.Join(dir, filepath.FromSlash(p)))
	}

	// JoinFunc preserves the prefix when resolving relative $ref paths.
	// Without this, path.Dir("rev:root.yaml") returns "." and $ref resolution breaks.
	loader.JoinFunc = func(basePath *url.URL, relativePath *url.URL) *url.URL {
		if basePath == nil {
			return relativePath
		}
		result := *basePath
		base := basePath.Path
		if i := strings.IndexByte(base, ':'); i >= 0 {
			pfx := base[:i+1]
			filePart := base[i+1:]
			result.Path = pfx + path.Join(path.Dir(filePart), relativePath.Path)
		} else {
			result.Path = path.Join(path.Dir(base), relativePath.Path)
		}
		return &result
	}

	rootContent, _ := os.ReadFile(filepath.Join(dir, "root.yaml"))
	doc, err := loader.LoadFromDataWithPath(rootContent, &url.URL{Path: prefix + "root.yaml"})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	petSchema := doc.Components.Schemas["Pet"]
	nameType := petSchema.Value.Properties["name"].Value.Type.Slice()[0]
	fmt.Println("pet.name type:", nameType)
	// Output: pet.name type: string
}
