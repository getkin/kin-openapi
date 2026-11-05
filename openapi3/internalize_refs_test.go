package openapi3_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestInternalizeRefs(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	regexpRef := regexp.MustCompile(`"\$ref":`)
	regexpRefInternal := regexp.MustCompile(`"\$ref":"#`)

	tests := []struct {
		filename string
	}{
		{"testdata/testref.openapi.yml"},
		{"testdata/recursiveRef/openapi.yml"},
		{"testdata/spec.yaml"},
		{"testdata/callbacks.yml"},
		{"testdata/issue831/testref.internalizepath.openapi.yml"},
		{"testdata/issue959/openapi.yml"},
		{"testdata/interalizationNameCollision/api.yml"},
		{"testdata/discriminator.yml"},
		{"testdata/discriminatorLocalMapping.yml"},
		{"testdata/issue1205/openapi.yml"},
		{"testdata/issue1205/mutual.yml"},
		{"testdata/issue1395/openapi.yaml"},
	}

	for _, test := range tests {
		t.Run(test.filename, func(t *testing.T) {
			// Load in the reference spec from the testdata
			sl := openapi3.NewLoader()
			sl.IsExternalRefsAllowed = true
			doc, err := sl.LoadFromFile(test.filename)
			require.NoError(t, err, "loading test file")
			err = doc.Validate(ctx)
			require.NoError(t, err, "validating spec")

			// Internalize the references
			doc.InternalizeRefs(ctx, nil)

			// Validate the internalized spec
			err = doc.Validate(ctx)
			require.NoError(t, err, "validating internalized spec")

			actual, err := doc.MarshalJSON()
			require.NoError(t, err, "marshaling internalized spec")

			// run a static check over the file, making sure each occurrence of a
			// reference is followed by a #
			numRefs := len(regexpRef.FindAll(actual, -1))
			numInternalRefs := len(regexpRefInternal.FindAll(actual, -1))
			require.Equal(t, numRefs, numInternalRefs, "checking all references are internal")

			// load from actual, but with the path set to the current directory
			doc2, err := sl.LoadFromData(actual)
			require.NoError(t, err, "reloading spec")
			err = doc2.Validate(ctx)
			require.NoError(t, err, "validating reloaded spec")

			// compare with expected
			expected, err := os.ReadFile(test.filename + ".internalized.yml")
			require.NoError(t, err)
			require.JSONEq(t, string(expected), string(actual))
		})
	}
}

func TestInternalizeRefsReusesRootSchema(t *testing.T) {
	t.Parallel()

	for _, version := range []string{"3.0.3", "3.1.0"} {
		for _, fragment := range []bool{false, true} {
			for _, wrapper := range []string{"Box", "Wrapper"} {
				t.Run(fmt.Sprintf("%s/fragment=%t/%s", version, fragment, wrapper), func(t *testing.T) {
					t.Parallel()

					dir := t.TempDir()
					itemRef, wrapperRef := "./schemas/Item.yaml", "./schemas/Wrapper.yaml"
					files := map[string]string{
						"schemas/Item.yaml": "type: string\n",
						"schemas/Wrapper.yaml": `type: object
properties:
  item:
    $ref: ./Item.yaml
`,
					}
					if fragment {
						itemRef = "./common.yaml#/components/schemas/Item"
						wrapperRef = "./common.yaml#/components/schemas/Wrapper"
						files = map[string]string{
							"common.yaml": `components:
  schemas:
    Item:
      type: string
    Wrapper:
      type: object
      properties:
        item:
          $ref: '#/components/schemas/Item'
`,
						}
					}
					files["openapi.yaml"] = fmt.Sprintf(`openapi: %s
info:
  title: Root component references
  version: "1"
paths:
  /item:
    get:
      responses:
        "200":
          description: An item
          content:
            application/json:
              schema:
                $ref: %s
        "201":
          description: A wrapper
          content:
            application/json:
              schema:
                $ref: %s
components:
  schemas:
    Item:
      $ref: %s
    %s:
      $ref: %s
`, version, itemRef, wrapperRef, itemRef, wrapper, wrapperRef)
					for name, data := range files {
						filename := filepath.Join(dir, name)
						require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0o700))
						require.NoError(t, os.WriteFile(filename, []byte(data), 0o600))
					}

					loader := openapi3.NewLoader()
					loader.IsExternalRefsAllowed = true
					doc, err := loader.LoadFromFile(filepath.Join(dir, "openapi.yaml"))
					require.NoError(t, err)
					require.NoError(t, doc.Validate(t.Context()))

					doc.InternalizeRefs(t.Context(), nil)

					require.Len(t, doc.Components.Schemas, 2)
					require.Empty(t, doc.Components.Schemas["Item"].Ref)
					require.Empty(t, doc.Components.Schemas[wrapper].Ref)
					require.Equal(t, "#/components/schemas/Item", doc.Components.Schemas[wrapper].Value.Properties["item"].Ref)
					responseSchema := doc.Paths.Value("/item").Get.Responses.Status(200).Value.Content.Get("application/json").Schema
					require.Equal(t, "#/components/schemas/Item", responseSchema.Ref)
					wrapperSchema := doc.Paths.Value("/item").Get.Responses.Status(201).Value.Content.Get("application/json").Schema
					require.Equal(t, "#/components/schemas/"+wrapper, wrapperSchema.Ref)
					require.NoError(t, doc.Validate(t.Context()))

					data, err := doc.MarshalJSON()
					require.NoError(t, err)
					internalLoader := openapi3.NewLoader()
					reloaded, err := internalLoader.LoadFromData(data)
					require.NoError(t, err)
					require.NoError(t, reloaded.Validate(t.Context()))

					doc.InternalizeRefs(t.Context(), nil)
					again, err := doc.MarshalJSON()
					require.NoError(t, err)
					require.JSONEq(t, string(data), string(again))
				})
			}
		}
	}
}

func TestInternalizeRefsAbsolutePath(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	filename, err := filepath.Abs("testdata/recursiveRef/openapi.yml")
	require.NoError(t, err)

	sl := openapi3.NewLoader()
	sl.IsExternalRefsAllowed = true
	doc, err := sl.LoadFromFile(filename)
	require.NoError(t, err, "loading test file")

	done := make(chan struct{})
	go func() {
		defer close(done)
		doc.InternalizeRefs(ctx, nil)
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("InternalizeRefs did not return for a document loaded from an absolute path")
	}

	err = doc.Validate(ctx)
	require.NoError(t, err, "validating internalized spec")

	actual, err := doc.MarshalJSON()
	require.NoError(t, err, "marshaling internalized spec")

	expected, err := os.ReadFile("testdata/recursiveRef/openapi.yml.internalized.yml")
	require.NoError(t, err)
	require.JSONEq(t, string(expected), string(actual))
}
