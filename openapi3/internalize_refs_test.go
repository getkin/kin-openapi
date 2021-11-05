package openapi3

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func baseResolver(ref string) string {
	split := strings.Split(ref, "#")
	if len(split) == 2 {
		return filepath.Base(split[1])
	}
	ref = split[0]
	for ext := filepath.Ext(ref); len(ext) > 0; ext = filepath.Ext(ref) {
		ref = strings.TrimSuffix(ref, ext)
	}
	return filepath.Base(ref)
}

func TestInternalizeRefs(t *testing.T) {
	var regexpRef = regexp.MustCompile(`"\$ref":`)
	var regexpRefInternal = regexp.MustCompile(`"\$ref":"\#`)

	tests := []struct {
		filename string
	}{
		{"testdata/testref.openapi.yml"},
		{"testdata/recursiveRef/openapi.yml"},
		{"testdata/spec.yaml"},
		{"testdata/callbacks.yml"},
	}

	for _, test := range tests {
		t.Run(test.filename, func(t *testing.T) {
			// Load in the reference spec from the testdata
			sl := NewLoader()
			sl.IsExternalRefsAllowed = true
			doc, err := sl.LoadFromFile(test.filename)
			require.NoError(t, err, "loading test file")

			// Internalize the references
			doc = doc.InternalizeRefs(context.Background(), baseResolver)

			// Validate the internalized spec
			require.Nil(t, doc.Validate(context.Background()), "validating internalized spec")

			// write it out to the current directory, a different context to where
			// we read it from which should fail if there are external references
			data, err := doc.MarshalJSON()
			require.NoError(t, err, "marshalling interalized spec")

			// run a static check over the file, making sure each occurence of a
			// reference is followed by a #
			s, _ := json.MarshalIndent(doc, "", "  ")
			fmt.Println(string(s))
			require.Equal(t, len(regexpRef.FindAll(data, -1)), len(regexpRefInternal.FindAll(data, -1)), "checking all references are internal")

			require.NoError(t, os.WriteFile("__internalized.openapi.json", data, 0644), "writing internalized spec to new context")

			doc2, err := sl.LoadFromFile(test.filename)
			require.NoError(t, err, "reloading spec")
			require.Nil(t, doc2.Validate(context.Background()), "validating reloaded spec")

			// try delete our artifact
			require.NoError(t, os.Remove("__internalized.openapi.json"), "removing test artifact")
		})
	}
}
