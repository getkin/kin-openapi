package openapi3

import (
	"context"
	"io/ioutil"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInternalizeRefs(t *testing.T) {
	ctx := context.Background()

	regexpRef := regexp.MustCompile(`"\$ref":`)
	regexpRefInternal := regexp.MustCompile(`"\$ref":"#`)

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
			err = doc.Validate(ctx)
			require.NoError(t, err, "validating spec")

			// Internalize the references
			doc.InternalizeRefs(ctx, nil)

			// Validate the internalized spec
			err = doc.Validate(ctx)
			require.NoError(t, err, "validating internalized spec")

			data, err := doc.MarshalJSON()
			require.NoError(t, err, "marshalling internalized spec")

			// run a static check over the file, making sure each occurence of a
			// reference is followed by a #
			numRefs := len(regexpRef.FindAll(data, -1))
			numInternalRefs := len(regexpRefInternal.FindAll(data, -1))
			require.Equal(t, numRefs, numInternalRefs, "checking all references are internal")

			// load from data, but with the path set to the current directory
			doc2, err := sl.LoadFromData(data)
			require.NoError(t, err, "reloading spec")
			err = doc2.Validate(ctx)
			require.NoError(t, err, "validating reloaded spec")

			// compare with expected
			data0, err := ioutil.ReadFile(test.filename + ".internalized.yml")
			require.NoError(t, err)
			require.JSONEq(t, string(data), string(data0))
		})
	}
}
