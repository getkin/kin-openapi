package openapi3_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestInternalizeRefsDottedFilenames(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, suffix string
	}{
		{"yaml", ".yaml"},
		{"yml", ".yml"},
		{"json", ".json"},
		{"uppercase", ".YAML"},
		{"no document extension", ""},
		{"other extension", ".schema"},
		{"repeated document extension", ".json.yaml"},
		{"repeated document extension reversed", ".yaml.json"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			const root = `{
  "openapi": "3.1.0", "info": {"title": "t", "version": "1"},
  "paths": {"/p": {"get": {
    "parameters": [
      {"$ref": "./components.v1/Lib.AccountId%s"},
      {"$ref": "./components.v1/Lib.PageSize%s"}
    ],
    "responses": {"200": {"description": "OK", "content": {
      "application/json": {"schema": {"$ref": "./components.v1/Models.User%s"}}
    }}}
  }}}
}`
			require.NoError(t, os.Mkdir(filepath.Join(dir, "components.v1"), 0o700))
			for name, content := range map[string]string{
				"root.yaml": fmt.Sprintf(root, test.suffix, test.suffix, test.suffix),
				"components.v1/Lib.AccountId" + test.suffix: `{"name":"account_id","in":"query","schema":{"type":"string"}}`,
				"components.v1/Lib.PageSize" + test.suffix:  `{"name":"page_size","in":"query","schema":{"type":"integer"}}`,
				"components.v1/Models.User" + test.suffix:   fmt.Sprintf(`{"type":"object","properties":{"group":{"$ref":"./Models.Group%s"}}}`, test.suffix),
				"components.v1/Models.Group" + test.suffix:  `{"type":"integer"}`,
			} {
				require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
			}

			loader := openapi3.NewLoader()
			loader.IsExternalRefsAllowed = true
			doc, err := loader.LoadFromFile(filepath.Join(dir, "root.yaml"))
			require.NoError(t, err)
			require.NoError(t, doc.Validate(t.Context()))
			doc.InternalizeRefs(t.Context(), nil)
			require.Len(t, doc.Components.Parameters, 2)
			require.Len(t, doc.Components.Schemas, 2)
			require.NoError(t, doc.Validate(t.Context()))

			data, err := doc.MarshalJSON()
			require.NoError(t, err)
			// A fresh loader cannot read external files: the bundle must stand alone.
			reloadLoader := openapi3.NewLoader()
			reloaded, err := reloadLoader.LoadFromData(data)
			require.NoError(t, err)
			require.NoError(t, reloaded.Validate(t.Context()))
			operation := reloaded.Paths.Value("/p").Get
			require.Equal(t, "account_id", operation.Parameters[0].Value.Name)
			require.True(t, operation.Parameters[0].Value.Schema.Value.Type.Is("string"))
			require.Equal(t, "page_size", operation.Parameters[1].Value.Name)
			require.True(t, operation.Parameters[1].Value.Schema.Value.Type.Is("integer"))
			user := operation.Responses.Status(200).Value.Content.Get("application/json").Schema
			require.True(t, user.Value.Type.Is("object"))
			require.True(t, user.Value.Properties["group"].Value.Type.Is("integer"))
			require.NotEqual(t, user.Ref, user.Value.Properties["group"].Ref)

			doc.InternalizeRefs(t.Context(), nil)
			again, err := doc.MarshalJSON()
			require.NoError(t, err)
			require.JSONEq(t, string(data), string(again))
		})
	}
}
