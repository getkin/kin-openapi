package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestIssue767(t *testing.T) {
	t.Parallel()

	tests := [...]struct {
		name     string
		schema   *openapi3.Schema
		value    map[string]interface{}
		opts     []openapi3.SchemaValidationOption
		checkErr require.ErrorAssertionFunc
	}{
		{
			name: "default values disabled should fail with minProps 1",
			schema: openapi3.NewSchema().WithProperties(map[string]*openapi3.Schema{
				"foo": {Type: "boolean", Default: true}}).WithMinProperties(1),
			value: map[string]interface{}{},
			opts: []openapi3.SchemaValidationOption{
				openapi3.VisitAsRequest(),
			},
			checkErr: require.Error,
		},
		{
			name: "default values enabled should pass with minProps 1",
			schema: openapi3.NewSchema().WithProperties(map[string]*openapi3.Schema{
				"foo": {Type: "boolean", Default: true}}).WithMinProperties(1),
			value: map[string]interface{}{},
			opts: []openapi3.SchemaValidationOption{
				openapi3.VisitAsRequest(),
				openapi3.DefaultsSet(func() {}),
			},
			checkErr: require.NoError,
		},
		{
			name: "default values enabled should pass with minProps 2",
			schema: openapi3.NewSchema().WithProperties(map[string]*openapi3.Schema{
				"foo": {Type: "boolean", Default: true},
				"bar": {Type: "boolean"},
			}).WithMinProperties(2),
			value: map[string]interface{}{"bar": false},
			opts: []openapi3.SchemaValidationOption{
				openapi3.VisitAsRequest(),
				openapi3.DefaultsSet(func() {}),
			},
			checkErr: require.NoError,
		},
		{
			name: "default values enabled should fail with maxProps 1",
			schema: openapi3.NewSchema().WithProperties(map[string]*openapi3.Schema{
				"foo": {Type: "boolean", Default: true},
				"bar": {Type: "boolean"},
			}).WithMaxProperties(1),
			value: map[string]interface{}{"bar": false},
			opts: []openapi3.SchemaValidationOption{
				openapi3.VisitAsRequest(),
				openapi3.DefaultsSet(func() {}),
			},
			checkErr: require.Error,
		},
		{
			name: "default values disabled should pass with maxProps 1",
			schema: openapi3.NewSchema().WithProperties(map[string]*openapi3.Schema{
				"foo": {Type: "boolean", Default: true},
				"bar": {Type: "boolean"},
			}).WithMaxProperties(1),
			value: map[string]interface{}{"bar": false},
			opts: []openapi3.SchemaValidationOption{
				openapi3.VisitAsRequest(),
			},
			checkErr: require.NoError,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := test.schema.VisitJSON(test.value, test.opts...)
			test.checkErr(t, err)
		})
	}
}
