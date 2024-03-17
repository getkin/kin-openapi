package openapi3_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestIssue856(t *testing.T) {
	const pref = "testdata/issue856/"

	err := os.MkdirAll(pref, 0750)
	require.NoError(t, err)
	defer os.RemoveAll(pref)

	file := func(t *testing.T, name, contents string) string {
		path := pref + name
		err := os.WriteFile(path, []byte(contents), 0644)
		require.NoError(t, err)
		return path
	}

	type Case struct {
		suff          string
		first, second string
	}

	cases := make(map[string]Case)

	cases["frag #/hm"] = Case{suff: "", second: `
hm:  # <--
    type: object
    properties:
        children:
            type: array
            items:
                $ref: '#/hm'  # <--
    description: test
`[1:], first: `
openapi: 3.0.0
info:
    title: Circular Reference Example
    version: 1.0.0
paths:
    /sample:
        put:
            requestBody:
                required: true
                content:
                    application/json:
                        schema:
                            $ref: './AWSEnvironmentSettings.yaml#/hm'  # <--
            responses:
                '200':
                    description: Ok
`[1:]}

	cases["frag #/"] = Case{suff: "_bis", second: `
type: object  # <--
properties:
    children:
        type: array
        items:
            $ref: '#/'  # <--
description: test
`[1:], first: `
openapi: 3.0.0
info:
    title: Circular Reference Example
    version: 1.0.0
paths:
    /sample:
        put:
            requestBody:
                required: true
                content:
                    application/json:
                        schema:
                            $ref: './AWSEnvironmentSettings_bis.yaml#/' # <--
            responses:
                '200':
                    description: Ok
`[1:]}

	cases["rel #/"] = Case{suff: "_rel", second: `
type: object  # <--
properties:
    children:
        type: array
        items:
            $ref: './AWSEnvironmentSettings_rel.yaml#/'  # <--
description: test
`[1:], first: `
openapi: 3.0.0
info:
    title: Circular Reference Example
    version: 1.0.0
paths:
    /sample:
        put:
            requestBody:
                required: true
                content:
                    application/json:
                        schema:
                            $ref: './AWSEnvironmentSettings_rel.yaml#/' # <--
            responses:
                '200':
                    description: Ok
`[1:]}

	cases["no #/"] = Case{suff: "_no", second: `
type: object  # <--
properties:
    children:
        type: array
        items:
            $ref: './AWSEnvironmentSettings_no.yaml'  # <--
description: test
`[1:], first: `
openapi: 3.0.0
info:
    title: Circular Reference Example
    version: 1.0.0
paths:
    /sample:
        put:
            requestBody:
                required: true
                content:
                    application/json:
                        schema:
                            $ref: './AWSEnvironmentSettings_no.yaml' # <--
            responses:
                '200':
                    description: Ok
`[1:]}

	for title, kase := range cases {
		title, kase := title, kase
		t.Run(title, func(t *testing.T) {
			f := file(t, fmt.Sprintf("AWSEnvironmentSettings%s.yaml", kase.suff), kase.second)
			defer os.Remove(f)

			docpath := fmt.Sprintf("circular2%s.yaml", kase.suff)
			g := file(t, docpath, kase.first)
			defer os.Remove(g)

			loader := openapi3.NewLoader()
			loader.IsExternalRefsAllowed = true
			_, err := loader.LoadFromFile(pref + docpath)
			require.ErrorContains(t, err, openapi3.CircularReferenceError)
		})
	}
}
