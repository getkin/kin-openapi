package openapi3

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestIssue972(t *testing.T) {
	type testcase struct {
		spec                    string
		validationErrorContains string
	}

	base := `
openapi: 3.0.2
paths: {}
components: {}
`

	for _, tc := range []testcase{{
		spec:                    base,
		validationErrorContains: "invalid info: must be an object",
	}, {
		spec: base + `
info:
  title: "Hello World REST APIs"
  version: "1.0"
`,
	}, {
		spec: base + `
info: null
`,
		validationErrorContains: "invalid info: must be an object",
	}, {
		spec: base + `
info: {}
`,
		validationErrorContains: "invalid info: value of version must be a non-empty string",
	}, {
		spec: base + `
info:
  title: "Hello World REST APIs"
`,
		validationErrorContains: "invalid info: value of version must be a non-empty string",
	}} {
		t.Logf("spec: %s", tc.spec)

		loader := &Loader{}
		doc, err := loader.LoadFromData([]byte(tc.spec))
		assert.NoError(t, err)
		assert.NotNil(t, doc)

		err = doc.Validate(loader.Context)
		if e := tc.validationErrorContains; e != "" {
			assert.ErrorContains(t, err, e)
		} else {
			assert.NoError(t, err)
		}

		txt, err := yaml.Marshal(doc)
		assert.NoError(t, err)
		assert.NotEmpty(t, txt)
	}
}
