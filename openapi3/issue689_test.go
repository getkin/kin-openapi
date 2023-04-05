package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestIssue689(t *testing.T) {
	t.Parallel()

	tests := [...]struct {
		name     string
		schema   *openapi3.Schema
		value    map[string]interface{}
		opts     []openapi3.SchemaValidationOption
		checkErr require.ErrorAssertionFunc
	}{
		// read-only
		{
			name: "read-only property succeeds when read-only validation is disabled",
			schema: openapi3.NewSchema().WithProperties(map[string]*openapi3.Schema{
				"foo": {Type: "boolean", ReadOnly: true}}),
			value: map[string]interface{}{"foo": true},
			opts: []openapi3.SchemaValidationOption{
				openapi3.VisitAsRequest(),
				openapi3.DisableReadOnlyValidation()},
			checkErr: require.NoError,
		},
		{
			name: "non read-only property succeeds when read-only validation is disabled",
			schema: openapi3.NewSchema().WithProperties(map[string]*openapi3.Schema{
				"foo": {Type: "boolean", ReadOnly: false}}),
			opts: []openapi3.SchemaValidationOption{
				openapi3.VisitAsRequest()},
			value:    map[string]interface{}{"foo": true},
			checkErr: require.NoError,
		},
		{
			name: "read-only property fails when read-only validation is enabled",
			schema: openapi3.NewSchema().WithProperties(map[string]*openapi3.Schema{
				"foo": {Type: "boolean", ReadOnly: true}}),
			opts: []openapi3.SchemaValidationOption{
				openapi3.VisitAsRequest()},
			value:    map[string]interface{}{"foo": true},
			checkErr: require.Error,
		},
		{
			name: "non read-only property succeeds when read-only validation is enabled",
			schema: openapi3.NewSchema().WithProperties(map[string]*openapi3.Schema{
				"foo": {Type: "boolean", ReadOnly: false}}),
			opts: []openapi3.SchemaValidationOption{
				openapi3.VisitAsRequest()},
			value:    map[string]interface{}{"foo": true},
			checkErr: require.NoError,
		},
		// write-only
		{
			name: "write-only property succeeds when write-only validation is disabled",
			schema: openapi3.NewSchema().WithProperties(map[string]*openapi3.Schema{
				"foo": {Type: "boolean", WriteOnly: true}}),
			value: map[string]interface{}{"foo": true},
			opts: []openapi3.SchemaValidationOption{
				openapi3.VisitAsResponse(),
				openapi3.DisableWriteOnlyValidation()},
			checkErr: require.NoError,
		},
		{
			name: "non write-only property succeeds when write-only validation is disabled",
			schema: openapi3.NewSchema().WithProperties(map[string]*openapi3.Schema{
				"foo": {Type: "boolean", WriteOnly: false}}),
			opts: []openapi3.SchemaValidationOption{
				openapi3.VisitAsResponse()},
			value:    map[string]interface{}{"foo": true},
			checkErr: require.NoError,
		},
		{
			name: "write-only property fails when write-only validation is enabled",
			schema: openapi3.NewSchema().WithProperties(map[string]*openapi3.Schema{
				"foo": {Type: "boolean", WriteOnly: true}}),
			opts: []openapi3.SchemaValidationOption{
				openapi3.VisitAsResponse()},
			value:    map[string]interface{}{"foo": true},
			checkErr: require.Error,
		},
		{
			name: "non write-only property succeeds when write-only validation is enabled",
			schema: openapi3.NewSchema().WithProperties(map[string]*openapi3.Schema{
				"foo": {Type: "boolean", WriteOnly: false}}),
			opts: []openapi3.SchemaValidationOption{
				openapi3.VisitAsResponse()},
			value:    map[string]interface{}{"foo": true},
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
