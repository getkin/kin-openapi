package openapi3filter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	legacyrouter "github.com/getkin/kin-openapi/routers/legacy"
)

var (
	explode   = openapi3.BoolPtr(true)
	noExplode = openapi3.BoolPtr(false)
	arrayOf   = func(items *openapi3.SchemaRef) *openapi3.SchemaRef {
		return &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"array"}, Items: items}}
	}
	objectOf = func(args ...interface{}) *openapi3.SchemaRef {
		s := &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"object"}, Properties: make(map[string]*openapi3.SchemaRef)}}
		if len(args)%2 != 0 {
			panic("invalid arguments. must be an even number of arguments")
		}
		for i := 0; i < len(args)/2; i++ {
			propName := args[i*2].(string)
			propSchema := args[i*2+1].(*openapi3.SchemaRef)
			s.Value.Properties[propName] = propSchema
		}
		return s
	}

	additionalPropertiesObjectOf = func(schema *openapi3.SchemaRef) *openapi3.SchemaRef {
		return &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"object"}, AdditionalProperties: openapi3.AdditionalProperties{Schema: schema}}}
	}

	integerSchema                          = &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}}}
	numberSchema                           = &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"number"}}}
	booleanSchema                          = &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"boolean"}}}
	stringSchema                           = &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}
	additionalPropertiesObjectStringSchema = &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"object"}, AdditionalProperties: openapi3.AdditionalProperties{Schema: stringSchema}}}
	additionalPropertiesObjectBoolSchema   = &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"object"}, AdditionalProperties: openapi3.AdditionalProperties{Schema: booleanSchema}}}
	allofSchema                            = &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			AllOf: []*openapi3.SchemaRef{
				integerSchema,
				numberSchema,
			},
		},
	}
	anyofSchema = &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			AnyOf: []*openapi3.SchemaRef{
				integerSchema,
				stringSchema,
			},
		},
	}
	oneofSchema = &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			OneOf: []*openapi3.SchemaRef{
				booleanSchema,
				integerSchema,
			},
		},
	}
	stringArraySchema  = arrayOf(stringSchema)
	integerArraySchema = arrayOf(integerSchema)
	objectSchema       = objectOf("id", stringSchema, "name", stringSchema)
)

func TestDeepGet(t *testing.T) {
	iarray := map[string]interface{}{
		"0": map[string]interface{}{
			"foo": 111,
		},
		"1": map[string]interface{}{
			"bar": 222,
		},
	}

	tests := []struct {
		name       string
		m          map[string]interface{}
		keys       []string
		expected   interface{}
		shouldFind bool
	}{
		{
			name: "Simple map - key exists",
			m: map[string]interface{}{
				"foo": "bar",
			},
			keys:       []string{"foo"},
			expected:   "bar",
			shouldFind: true,
		},
		{
			name: "Nested map - key exists",
			m: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": "baz",
				},
			},
			keys:       []string{"foo", "bar"},
			expected:   "baz",
			shouldFind: true,
		},
		{
			name: "Nested map - key does not exist",
			m: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": "baz",
				},
			},
			keys:       []string{"foo", "baz"},
			expected:   nil,
			shouldFind: false,
		},
		{
			name: "Array - element exists",
			m: map[string]interface{}{
				"array": map[string]interface{}{"0": "a", "1": "b", "2": "c"},
			},
			keys:       []string{"array", "1"},
			expected:   "b",
			shouldFind: true,
		},
		{
			name: "Array - element does not exist - invalid index",
			m: map[string]interface{}{
				"array": map[string]interface{}{"0": "a", "1": "b", "2": "c"},
			},
			keys:       []string{"array", "3"},
			expected:   nil,
			shouldFind: false,
		},
		{
			name: "Array - element does not exist - invalid keys",
			m: map[string]interface{}{
				"array": map[string]interface{}{"0": "a", "1": "b", "2": "c"},
			},
			keys:       []string{"array", "a", "999"},
			expected:   nil,
			shouldFind: false,
		},
		{
			name: "Array of objects - element exists 1",
			m: map[string]interface{}{
				"array": iarray,
			},
			keys:       []string{"array", "1", "bar"},
			expected:   222,
			shouldFind: true,
		},
		{
			name: "Array of objects - element exists 2",
			m: map[string]interface{}{
				"array": iarray,
			},
			keys: []string{"array", "0"},
			expected: map[string]interface{}{
				"foo": 111,
			},
			shouldFind: true,
		},
		{
			name: "Array of objects - element exists 3",
			m: map[string]interface{}{
				"array": iarray,
			},
			keys:       []string{"array"},
			expected:   iarray,
			shouldFind: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc

			result, found := deepGet(tc.m, tc.keys...)
			require.Equal(t, tc.shouldFind, found, "shouldFind mismatch")
			require.Equal(t, tc.expected, result, "result mismatch")
		})
	}
}

func TestDeepSet(t *testing.T) {
	tests := []struct {
		name     string
		inputMap map[string]interface{}
		keys     []string
		value    interface{}
		expected map[string]interface{}
	}{
		{
			name:     "simple set",
			inputMap: map[string]interface{}{},
			keys:     []string{"key"},
			value:    "value",
			expected: map[string]interface{}{"key": "value"},
		},
		{
			name:     "intermediate array of objects",
			inputMap: map[string]interface{}{},
			keys:     []string{"nested", "0", "key"},
			value:    true,
			expected: map[string]interface{}{
				"nested": map[string]interface{}{
					"0": map[string]interface{}{
						"key": true,
					},
				},
			},
		},
		{
			name:     "existing nested array of objects",
			inputMap: map[string]interface{}{"nested": map[string]interface{}{"0": map[string]interface{}{"existingKey": "existingValue"}}},
			keys:     []string{"nested", "0", "newKey"},
			value:    "newValue",
			expected: map[string]interface{}{
				"nested": map[string]interface{}{
					"0": map[string]interface{}{
						"existingKey": "existingValue",
						"newKey":      "newValue",
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			deepSet(tc.inputMap, tc.keys, tc.value)
			require.EqualValues(t, tc.expected, tc.inputMap)
		})
	}
}

func TestDecodeParameter(t *testing.T) {
	type testCase struct {
		name   string
		param  *openapi3.Parameter
		path   string
		query  string
		header string
		cookie string
		want   interface{}
		found  bool
		err    error
	}

	testGroups := []struct {
		name      string
		testCases []testCase
	}{
		{
			name: "path primitive",
			testCases: []testCase{
				{
					name:  "simple",
					param: &openapi3.Parameter{Name: "param", In: "path", Style: "simple", Explode: noExplode, Schema: stringSchema},
					path:  "/foo",
					want:  "foo",
					found: true,
				},
				{
					name:  "simple explode",
					param: &openapi3.Parameter{Name: "param", In: "path", Style: "simple", Explode: explode, Schema: stringSchema},
					path:  "/foo",
					want:  "foo",
					found: true,
				},
				{
					name:  "label",
					param: &openapi3.Parameter{Name: "param", In: "path", Style: "label", Explode: noExplode, Schema: stringSchema},
					path:  "/.foo",
					want:  "foo",
					found: true,
				},
				{
					name:  "label invalid",
					param: &openapi3.Parameter{Name: "param", In: "path", Style: "label", Explode: noExplode, Schema: stringSchema},
					path:  "/foo",
					found: true,
					err:   &ParseError{Kind: KindInvalidFormat, Value: "foo"},
				},
				{
					name:  "label explode",
					param: &openapi3.Parameter{Name: "param", In: "path", Style: "label", Explode: explode, Schema: stringSchema},
					path:  "/.foo",
					want:  "foo",
					found: true,
				},
				{
					name:  "label explode invalid",
					param: &openapi3.Parameter{Name: "param", In: "path", Style: "label", Explode: explode, Schema: stringSchema},
					path:  "/foo",
					found: true,
					err:   &ParseError{Kind: KindInvalidFormat, Value: "foo"},
				},
				{
					name:  "matrix",
					param: &openapi3.Parameter{Name: "param", In: "path", Style: "matrix", Explode: noExplode, Schema: stringSchema},
					path:  "/;param=foo",
					want:  "foo",
					found: true,
				},
				{
					name:  "matrix invalid",
					param: &openapi3.Parameter{Name: "param", In: "path", Style: "matrix", Explode: noExplode, Schema: stringSchema},
					path:  "/foo",
					found: true,
					err:   &ParseError{Kind: KindInvalidFormat, Value: "foo"},
				},
				{
					name:  "matrix explode",
					param: &openapi3.Parameter{Name: "param", In: "path", Style: "matrix", Explode: explode, Schema: stringSchema},
					path:  "/;param=foo",
					want:  "foo",
					found: true,
				},
				{
					name:  "matrix explode invalid",
					param: &openapi3.Parameter{Name: "param", In: "path", Style: "matrix", Explode: explode, Schema: stringSchema},
					path:  "/foo",
					found: true,
					err:   &ParseError{Kind: KindInvalidFormat, Value: "foo"},
				},
				{
					name:  "default",
					param: &openapi3.Parameter{Name: "param", In: "path", Schema: stringSchema},
					path:  "/foo",
					want:  "foo",
					found: true,
				},
				{
					name:  "string",
					param: &openapi3.Parameter{Name: "param", In: "path", Schema: stringSchema},
					path:  "/foo",
					want:  "foo",
					found: true,
				},
				{
					name:  "integer",
					param: &openapi3.Parameter{Name: "param", In: "path", Schema: integerSchema},
					path:  "/1",
					want:  int64(1),
					found: true,
				},
				{
					name:  "integer invalid",
					param: &openapi3.Parameter{Name: "param", In: "path", Schema: integerSchema},
					path:  "/foo",
					found: true,
					err:   &ParseError{Kind: KindInvalidFormat, Value: "foo"},
				},
				{
					name:  "number",
					param: &openapi3.Parameter{Name: "param", In: "path", Schema: numberSchema},
					path:  "/1.1",
					want:  1.1,
					found: true,
				},
				{
					name:  "number invalid",
					param: &openapi3.Parameter{Name: "param", In: "path", Schema: numberSchema},
					path:  "/foo",
					found: true,
					err:   &ParseError{Kind: KindInvalidFormat, Value: "foo"},
				},
				{
					name:  "boolean",
					param: &openapi3.Parameter{Name: "param", In: "path", Schema: booleanSchema},
					path:  "/true",
					want:  true,
					found: true,
				},
				{
					name:  "boolean invalid",
					param: &openapi3.Parameter{Name: "param", In: "path", Schema: booleanSchema},
					path:  "/foo",
					found: true,
					err:   &ParseError{Kind: KindInvalidFormat, Value: "foo"},
				},
			},
		},
		{
			name: "path array",
			testCases: []testCase{
				{
					name:  "simple",
					param: &openapi3.Parameter{Name: "param", In: "path", Style: "simple", Explode: noExplode, Schema: stringArraySchema},
					path:  "/foo,bar",
					want:  []interface{}{"foo", "bar"},
					found: true,
				},
				{
					name:  "simple explode",
					param: &openapi3.Parameter{Name: "param", In: "path", Style: "simple", Explode: explode, Schema: stringArraySchema},
					path:  "/foo,bar",
					want:  []interface{}{"foo", "bar"},
					found: true,
				},
				{
					name:  "label",
					param: &openapi3.Parameter{Name: "param", In: "path", Style: "label", Explode: noExplode, Schema: stringArraySchema},
					path:  "/.foo,bar",
					want:  []interface{}{"foo", "bar"},
					found: true,
				},
				{
					name:  "label invalid",
					param: &openapi3.Parameter{Name: "param", In: "path", Style: "label", Explode: noExplode, Schema: stringArraySchema},
					path:  "/foo,bar",
					found: true,
					err:   &ParseError{Kind: KindInvalidFormat, Value: "foo,bar"},
				},
				{
					name:  "label explode",
					param: &openapi3.Parameter{Name: "param", In: "path", Style: "label", Explode: explode, Schema: stringArraySchema},
					path:  "/.foo.bar",
					want:  []interface{}{"foo", "bar"},
					found: true,
				},
				{
					name:  "label explode invalid",
					param: &openapi3.Parameter{Name: "param", In: "path", Style: "label", Explode: explode, Schema: stringArraySchema},
					path:  "/foo.bar",
					found: true,
					err:   &ParseError{Kind: KindInvalidFormat, Value: "foo.bar"},
				},
				{
					name:  "matrix",
					param: &openapi3.Parameter{Name: "param", In: "path", Style: "matrix", Explode: noExplode, Schema: stringArraySchema},
					path:  "/;param=foo,bar",
					want:  []interface{}{"foo", "bar"},
					found: true,
				},
				{
					name:  "matrix invalid",
					param: &openapi3.Parameter{Name: "param", In: "path", Style: "matrix", Explode: noExplode, Schema: stringArraySchema},
					path:  "/foo,bar",
					found: true,
					err:   &ParseError{Kind: KindInvalidFormat, Value: "foo,bar"},
				},
				{
					name:  "matrix explode",
					param: &openapi3.Parameter{Name: "param", In: "path", Style: "matrix", Explode: explode, Schema: stringArraySchema},
					path:  "/;param=foo;param=bar",
					want:  []interface{}{"foo", "bar"},
					found: true,
				},
				{
					name:  "matrix explode invalid",
					param: &openapi3.Parameter{Name: "param", In: "path", Style: "matrix", Explode: explode, Schema: stringArraySchema},
					path:  "/foo,bar",
					found: true,
					err:   &ParseError{Kind: KindInvalidFormat, Value: "foo,bar"},
				},
				{
					name:  "default",
					param: &openapi3.Parameter{Name: "param", In: "path", Schema: stringArraySchema},
					path:  "/foo,bar",
					want:  []interface{}{"foo", "bar"},
					found: true,
				},
				{
					name:  "invalid integer items",
					param: &openapi3.Parameter{Name: "param", In: "path", Schema: arrayOf(integerSchema)},
					path:  "/1,foo",
					found: true,
					err:   &ParseError{path: []interface{}{1}, Cause: &ParseError{Kind: KindInvalidFormat, Value: "foo"}},
				},
				{
					name:  "invalid number items",
					param: &openapi3.Parameter{Name: "param", In: "path", Schema: arrayOf(numberSchema)},
					path:  "/1.1,foo",
					found: true,
					err:   &ParseError{path: []interface{}{1}, Cause: &ParseError{Kind: KindInvalidFormat, Value: "foo"}},
				},
				{
					name:  "invalid boolean items",
					param: &openapi3.Parameter{Name: "param", In: "path", Schema: arrayOf(booleanSchema)},
					path:  "/true,foo",
					found: true,
					err:   &ParseError{path: []interface{}{1}, Cause: &ParseError{Kind: KindInvalidFormat, Value: "foo"}},
				},
			},
		},
		{
			name: "path object",
			testCases: []testCase{
				{
					name:  "simple",
					param: &openapi3.Parameter{Name: "param", In: "path", Style: "simple", Explode: noExplode, Schema: objectSchema},
					path:  "/id,foo,name,bar",
					want:  map[string]interface{}{"id": "foo", "name": "bar"},
					found: true,
				},
				{
					name:  "simple explode",
					param: &openapi3.Parameter{Name: "param", In: "path", Style: "simple", Explode: explode, Schema: objectSchema},
					path:  "/id=foo,name=bar",
					want:  map[string]interface{}{"id": "foo", "name": "bar"},
					found: true,
				},
				{
					name:  "label",
					param: &openapi3.Parameter{Name: "param", In: "path", Style: "label", Explode: noExplode, Schema: objectSchema},
					path:  "/.id,foo,name,bar",
					want:  map[string]interface{}{"id": "foo", "name": "bar"},
					found: true,
				},
				{
					name:  "label invalid",
					param: &openapi3.Parameter{Name: "param", In: "path", Style: "label", Explode: noExplode, Schema: objectSchema},
					path:  "/id,foo,name,bar",
					found: true,
					err:   &ParseError{Kind: KindInvalidFormat, Value: "id,foo,name,bar"},
				},
				{
					name:  "label explode",
					param: &openapi3.Parameter{Name: "param", In: "path", Style: "label", Explode: explode, Schema: objectSchema},
					path:  "/.id=foo.name=bar",
					want:  map[string]interface{}{"id": "foo", "name": "bar"},
					found: true,
				},
				{
					name:  "label explode invalid",
					param: &openapi3.Parameter{Name: "param", In: "path", Style: "label", Explode: explode, Schema: objectSchema},
					path:  "/id=foo.name=bar",
					found: true,
					err:   &ParseError{Kind: KindInvalidFormat, Value: "id=foo.name=bar"},
				},
				{
					name:  "matrix",
					param: &openapi3.Parameter{Name: "param", In: "path", Style: "matrix", Explode: noExplode, Schema: objectSchema},
					path:  "/;param=id,foo,name,bar",
					want:  map[string]interface{}{"id": "foo", "name": "bar"},
					found: true,
				},
				{
					name:  "matrix invalid",
					param: &openapi3.Parameter{Name: "param", In: "path", Style: "matrix", Explode: noExplode, Schema: objectSchema},
					path:  "/id,foo,name,bar",
					found: true,
					err:   &ParseError{Kind: KindInvalidFormat, Value: "id,foo,name,bar"},
				},
				{
					name:  "matrix explode",
					param: &openapi3.Parameter{Name: "param", In: "path", Style: "matrix", Explode: explode, Schema: objectSchema},
					path:  "/;id=foo;name=bar",
					want:  map[string]interface{}{"id": "foo", "name": "bar"},
					found: true,
				},
				{
					name:  "matrix explode invalid",
					param: &openapi3.Parameter{Name: "param", In: "path", Style: "matrix", Explode: explode, Schema: objectSchema},
					path:  "/id=foo;name=bar",
					found: true,
					err:   &ParseError{Kind: KindInvalidFormat, Value: "id=foo;name=bar"},
				},
				{
					name:  "default",
					param: &openapi3.Parameter{Name: "param", In: "path", Schema: objectSchema},
					path:  "/id,foo,name,bar",
					want:  map[string]interface{}{"id": "foo", "name": "bar"},
					found: true,
				},
				{
					name:  "invalid integer prop",
					param: &openapi3.Parameter{Name: "param", In: "path", Schema: objectOf("foo", integerSchema)},
					path:  "/foo,bar",
					found: true,
					err:   &ParseError{path: []interface{}{"foo"}, Cause: &ParseError{Kind: KindInvalidFormat, Value: "bar"}},
				},
				{
					name:  "invalid number prop",
					param: &openapi3.Parameter{Name: "param", In: "path", Schema: objectOf("foo", numberSchema)},
					path:  "/foo,bar",
					found: true,
					err:   &ParseError{path: []interface{}{"foo"}, Cause: &ParseError{Kind: KindInvalidFormat, Value: "bar"}},
				},
				{
					name:  "invalid boolean prop",
					param: &openapi3.Parameter{Name: "param", In: "path", Schema: objectOf("foo", booleanSchema)},
					path:  "/foo,bar",
					found: true,
					err:   &ParseError{path: []interface{}{"foo"}, Cause: &ParseError{Kind: KindInvalidFormat, Value: "bar"}},
				},
			},
		},
		{
			name: "query primitive",
			testCases: []testCase{
				{
					name:  "form",
					param: &openapi3.Parameter{Name: "param", In: "query", Style: "form", Explode: noExplode, Schema: stringSchema},
					query: "param=foo",
					want:  "foo",
					found: true,
				},
				{
					name:  "form explode",
					param: &openapi3.Parameter{Name: "param", In: "query", Style: "form", Explode: explode, Schema: stringSchema},
					query: "param=foo",
					want:  "foo",
					found: true,
				},
				{
					name:  "default",
					param: &openapi3.Parameter{Name: "param", In: "query", Schema: stringSchema},
					query: "param=foo",
					want:  "foo",
					found: true,
				},
				{
					name:  "string",
					param: &openapi3.Parameter{Name: "param", In: "query", Schema: stringSchema},
					query: "param=foo",
					want:  "foo",
					found: true,
				},
				{
					name:  "integer",
					param: &openapi3.Parameter{Name: "param", In: "query", Schema: integerSchema},
					query: "param=1",
					want:  int64(1),
					found: true,
				},
				{
					name:  "integer invalid",
					param: &openapi3.Parameter{Name: "param", In: "query", Schema: integerSchema},
					query: "param=foo",
					found: true,
					err:   &ParseError{Kind: KindInvalidFormat, Value: "foo"},
				},
				{
					name:  "number",
					param: &openapi3.Parameter{Name: "param", In: "query", Schema: numberSchema},
					query: "param=1.1",
					want:  1.1,
					found: true,
				},
				{
					name:  "number invalid",
					param: &openapi3.Parameter{Name: "param", In: "query", Schema: numberSchema},
					query: "param=foo",
					found: true,
					err:   &ParseError{Kind: KindInvalidFormat, Value: "foo"},
				},
				{
					name:  "boolean",
					param: &openapi3.Parameter{Name: "param", In: "query", Schema: booleanSchema},
					query: "param=true",
					want:  true,
					found: true,
				},
				{
					name:  "boolean invalid",
					param: &openapi3.Parameter{Name: "param", In: "query", Schema: booleanSchema},
					query: "param=foo",
					found: true,
					err:   &ParseError{Kind: KindInvalidFormat, Value: "foo"},
				},
			},
		},
		{
			name: "query Allof",
			testCases: []testCase{
				{
					name:  "allofSchema integer and number",
					param: &openapi3.Parameter{Name: "param", In: "query", Schema: allofSchema},
					query: "param=1",
					want:  float64(1),
					found: true,
				},
				{
					name:  "allofSchema string",
					param: &openapi3.Parameter{Name: "param", In: "query", Schema: allofSchema},
					query: "param=abdf",
					found: true,
					err:   &ParseError{Kind: KindInvalidFormat, Value: "abdf"},
				},
			},
		},
		{
			name: "query Anyof",
			testCases: []testCase{
				{
					name:  "anyofSchema integer",
					param: &openapi3.Parameter{Name: "param", In: "query", Schema: anyofSchema},
					query: "param=1",
					want:  int64(1),
					found: true,
				},
				{
					name:  "anyofSchema string",
					param: &openapi3.Parameter{Name: "param", In: "query", Schema: anyofSchema},
					query: "param=abdf",
					want:  "abdf",
					found: true,
				},
			},
		},
		{
			name: "query Oneof",
			testCases: []testCase{
				{
					name:  "oneofSchema boolean",
					param: &openapi3.Parameter{Name: "param", In: "query", Schema: oneofSchema},
					query: "param=true",
					want:  true,
					found: true,
				},
				{
					name:  "oneofSchema int",
					param: &openapi3.Parameter{Name: "param", In: "query", Schema: oneofSchema},
					query: "param=1122",
					want:  int64(1122),
					found: true,
				},
				{
					name:  "oneofSchema string",
					param: &openapi3.Parameter{Name: "param", In: "query", Schema: oneofSchema},
					query: "param=abcd",
					want:  nil,
					found: true,
				},
			},
		},
		{
			name: "query array",
			testCases: []testCase{
				{
					name:  "form",
					param: &openapi3.Parameter{Name: "param", In: "query", Style: "form", Explode: noExplode, Schema: stringArraySchema},
					query: "param=foo,bar",
					want:  []interface{}{"foo", "bar"},
					found: true,
				},
				{
					name:  "form explode",
					param: &openapi3.Parameter{Name: "param", In: "query", Style: "form", Explode: explode, Schema: stringArraySchema},
					query: "param=foo&param=bar",
					want:  []interface{}{"foo", "bar"},
					found: true,
				},
				{
					name:  "spaceDelimited",
					param: &openapi3.Parameter{Name: "param", In: "query", Style: "spaceDelimited", Explode: noExplode, Schema: stringArraySchema},
					query: "param=foo bar",
					want:  []interface{}{"foo", "bar"},
					found: true,
				},
				{
					name:  "spaceDelimited explode",
					param: &openapi3.Parameter{Name: "param", In: "query", Style: "spaceDelimited", Explode: explode, Schema: stringArraySchema},
					query: "param=foo&param=bar",
					want:  []interface{}{"foo", "bar"},
					found: true,
				},
				{
					name:  "pipeDelimited",
					param: &openapi3.Parameter{Name: "param", In: "query", Style: "pipeDelimited", Explode: noExplode, Schema: stringArraySchema},
					query: "param=foo|bar",
					want:  []interface{}{"foo", "bar"},
					found: true,
				},
				{
					name:  "pipeDelimited explode",
					param: &openapi3.Parameter{Name: "param", In: "query", Style: "pipeDelimited", Explode: explode, Schema: stringArraySchema},
					query: "param=foo&param=bar",
					want:  []interface{}{"foo", "bar"},
					found: true,
				},
				{
					name:  "default",
					param: &openapi3.Parameter{Name: "param", In: "query", Schema: stringArraySchema},
					query: "param=foo&param=bar",
					want:  []interface{}{"foo", "bar"},
					found: true,
				},
				{
					name:  "invalid integer items",
					param: &openapi3.Parameter{Name: "param", In: "query", Schema: arrayOf(integerSchema)},
					query: "param=1&param=foo",
					found: true,
					err:   &ParseError{path: []interface{}{1}, Cause: &ParseError{Kind: KindInvalidFormat, Value: "foo"}},
				},
				{
					name:  "invalid number items",
					param: &openapi3.Parameter{Name: "param", In: "query", Schema: arrayOf(numberSchema)},
					query: "param=1.1&param=foo",
					found: true,
					err:   &ParseError{path: []interface{}{1}, Cause: &ParseError{Kind: KindInvalidFormat, Value: "foo"}},
				},
				{
					name:  "invalid boolean items",
					param: &openapi3.Parameter{Name: "param", In: "query", Schema: arrayOf(booleanSchema)},
					query: "param=true&param=foo",
					found: true,
					err:   &ParseError{path: []interface{}{1}, Cause: &ParseError{Kind: KindInvalidFormat, Value: "foo"}},
				},
			},
		},
		{
			name: "query object",
			testCases: []testCase{
				{
					name:  "form",
					param: &openapi3.Parameter{Name: "param", In: "query", Style: "form", Explode: noExplode, Schema: objectSchema},
					query: "param=id,foo,name,bar",
					want:  map[string]interface{}{"id": "foo", "name": "bar"},
					found: true,
				},
				{
					name:  "form explode",
					param: &openapi3.Parameter{Name: "param", In: "query", Style: "form", Explode: explode, Schema: objectSchema},
					query: "id=foo&name=bar",
					want:  map[string]interface{}{"id": "foo", "name": "bar"},
					found: true,
				},
				{
					name:  "deepObject explode",
					param: &openapi3.Parameter{Name: "param", In: "query", Style: "deepObject", Explode: explode, Schema: objectSchema},
					query: "param[id]=foo&param[name]=bar",
					want:  map[string]interface{}{"id": "foo", "name": "bar"},
					found: true,
				},
				// NOTE: does not error out when only one array element is present (no delimiter found),
				// so it is only catched as request error with a generic schema error
				// {
				// 	name: "deepObject explode additionalProperties with object properties - missing index on nested array",
				// 	param: &openapi3.Parameter{
				// 		Name: "param", In: "query", Style: "deepObject", Explode: explode,
				// 		Schema: objectOf(
				// 			"obj", additionalPropertiesObjectOf(objectOf("item1", integerSchema, "item2", stringArraySchema)),
				// 			"objIgnored", objectOf("items", stringArraySchema),
				// 		),
				// 	},
				// 	query: "param[obj][prop2][item2]=def",
				// 	err:   &ParseError{path: []interface{}{"obj", "prop2", "item2"}, Kind: KindInvalidFormat, Reason: "array items must be set with indexes"},
				// },
				{
					name:  "deepObject explode array - missing indexes",
					param: &openapi3.Parameter{Name: "param", In: "query", Style: "deepObject", Explode: explode, Schema: objectOf("items", stringArraySchema)},
					query: "param[items]=f%26oo&param[items]=bar",
					found: true,
					err:   &ParseError{path: []interface{}{"items"}, Kind: KindInvalidFormat, Reason: "array items must be set with indexes"},
				},
				{
					name:  "deepObject explode array",
					param: &openapi3.Parameter{Name: "param", In: "query", Style: "deepObject", Explode: explode, Schema: objectOf("items", integerArraySchema)},
					query: "param[items][1]=456&param[items][0]=123",
					want:  map[string]interface{}{"items": []interface{}{int64(123), int64(456)}},
					found: true,
				},
				{
					name: "deepObject explode nested object additionalProperties",
					param: &openapi3.Parameter{
						Name: "param", In: "query", Style: "deepObject", Explode: explode,
						Schema: objectOf(
							"obj", additionalPropertiesObjectStringSchema,
							"objTwo", stringSchema,
							"objIgnored", objectOf("items", stringArraySchema),
						),
					},
					query: "param[obj][prop1]=bar&param[obj][prop2]=foo&param[objTwo]=string",
					want: map[string]interface{}{
						"obj":    map[string]interface{}{"prop1": "bar", "prop2": "foo"},
						"objTwo": "string",
					},
					found: true,
				},
				{
					name: "deepObject explode additionalProperties with object properties - sharing property",
					param: &openapi3.Parameter{
						Name: "param", In: "query", Style: "deepObject", Explode: explode,
						Schema: objectOf(
							"obj", additionalPropertiesObjectOf(objectOf("item1", integerSchema, "item2", stringSchema)),
							"objIgnored", objectOf("items", stringArraySchema),
						),
					},
					query: "param[obj][prop1][item1]=1&param[obj][prop1][item2]=abc",
					want: map[string]interface{}{
						"obj": map[string]interface{}{"prop1": map[string]interface{}{
							"item1": int64(1),
							"item2": "abc",
						}},
					},
					found: true,
				},
				{
					name: "deepObject explode nested object additionalProperties - bad value",
					param: &openapi3.Parameter{
						Name: "param", In: "query", Style: "deepObject", Explode: explode,
						Schema: objectOf(
							"obj", additionalPropertiesObjectBoolSchema,
							"objTwo", stringSchema,
							"objIgnored", objectOf("items", stringArraySchema),
						),
					},
					query: "param[obj][prop1]=notbool&param[objTwo]=string",
					err:   &ParseError{path: []interface{}{"obj", "prop1"}, Cause: &ParseError{Kind: KindInvalidFormat, Value: "notbool"}},
				},
				{
					name: "deepObject explode nested object additionalProperties - bad index inside additionalProperties",
					param: &openapi3.Parameter{
						Name: "param", In: "query", Style: "deepObject", Explode: explode,
						Schema: objectOf(
							"obj", additionalPropertiesObjectStringSchema,
							"objTwo", stringSchema,
							"objIgnored", objectOf("items", stringArraySchema),
						),
					},
					query: "param[obj][prop1]=bar&param[obj][prop2][badindex]=bad&param[objTwo]=string",
					err: &ParseError{
						path:   []interface{}{"obj", "prop2"},
						Reason: `path is not convertible to primitive`,
						Kind:   KindInvalidFormat,
						Value:  map[string]interface{}{"badindex": "bad"},
					},
				},
				{
					name: "deepObject explode nested object",
					param: &openapi3.Parameter{
						Name: "param", In: "query", Style: "deepObject", Explode: explode,
						Schema: objectOf(
							"obj", objectOf("nestedObjOne", stringSchema, "nestedObjTwo", stringSchema),
							"objTwo", stringSchema,
							"objIgnored", objectOf("items", stringArraySchema),
						),
					},
					query: "param[obj][nestedObjOne]=bar&param[obj][nestedObjTwo]=foo&param[objTwo]=string",
					want: map[string]interface{}{
						"obj":    map[string]interface{}{"nestedObjOne": "bar", "nestedObjTwo": "foo"},
						"objTwo": "string",
					},
					found: true,
				},

				{
					name: "deepObject explode nested object - extraneous param ignored",
					param: &openapi3.Parameter{
						Name: "param", In: "query", Style: "deepObject", Explode: explode,
						Schema: objectOf(
							"obj", objectOf("nestedObjOne", stringSchema, "nestedObjTwo", stringSchema),
						),
					},
					query: "anotherparam=bar",
					want:  map[string]interface{}(nil),
				},
				{
					name: "deepObject explode nested object - bad array item type",
					param: &openapi3.Parameter{
						Name: "param", In: "query", Style: "deepObject", Explode: explode,
						Schema: objectOf(
							"objTwo", integerArraySchema,
						),
					},
					query: "param[objTwo]=badint",
					err:   &ParseError{path: []interface{}{"objTwo"}, Cause: &ParseError{Kind: KindInvalidFormat, Value: "badint"}},
				},
				{
					name: "deepObject explode deeply nested object - bad array item type",
					param: &openapi3.Parameter{
						Name: "param", In: "query", Style: "deepObject", Explode: explode,
						Schema: objectOf(
							"obj", objectOf("nestedObjOne", integerArraySchema),
						),
					},
					query: "param[obj][nestedObjOne]=badint",
					err:   &ParseError{path: []interface{}{"obj", "nestedObjOne"}, Cause: &ParseError{Kind: KindInvalidFormat, Value: "badint"}},
				},
				{
					name: "deepObject explode nested object with array",
					param: &openapi3.Parameter{
						Name: "param", In: "query", Style: "deepObject", Explode: explode,
						Schema: objectOf(
							"obj", objectOf("nestedObjOne", stringSchema, "nestedObjTwo", stringSchema),
							"objTwo", stringArraySchema,
							"objIgnored", objectOf("items", stringArraySchema),
						),
					},
					query: "param[obj][nestedObjOne]=bar&param[obj][nestedObjTwo]=foo&param[objTwo][0]=f%26oo&param[objTwo][1]=bar",
					want: map[string]interface{}{
						"obj":    map[string]interface{}{"nestedObjOne": "bar", "nestedObjTwo": "foo"},
						"objTwo": []interface{}{"f%26oo", "bar"},
					},
					found: true,
				},
				{
					name: "deepObject explode nested object with array - bad value",
					param: &openapi3.Parameter{
						Name: "param", In: "query", Style: "deepObject", Explode: explode,
						Schema: objectOf(
							"obj", objectOf("nestedObjOne", stringSchema, "nestedObjTwo", booleanSchema),
							"objTwo", stringArraySchema,
							"objIgnored", objectOf("items", stringArraySchema),
						),
					},
					query: "param[obj][nestedObjOne]=bar&param[obj][nestedObjTwo]=bad&param[objTwo][0]=f%26oo&param[objTwo][1]=bar",
					err:   &ParseError{path: []interface{}{"obj", "nestedObjTwo"}, Cause: &ParseError{Kind: KindInvalidFormat, Value: "bad"}},
				},
				{
					name: "deepObject explode nested object with nested array",
					param: &openapi3.Parameter{
						Name: "param", In: "query", Style: "deepObject", Explode: explode,
						Schema: objectOf(
							"obj", objectOf("nestedObjOne", stringSchema, "nestedObjTwo", stringSchema),
							"objTwo", objectOf("items", stringArraySchema),
							"objIgnored", objectOf("items", stringArraySchema),
						),
					},
					query: "param[obj][nestedObjOne]=bar&param[obj][nestedObjTwo]=foo&param[objTwo][items][0]=f%26oo&param[objTwo][items][1]=bar",
					want: map[string]interface{}{
						"obj":    map[string]interface{}{"nestedObjOne": "bar", "nestedObjTwo": "foo"},
						"objTwo": map[string]interface{}{"items": []interface{}{"f%26oo", "bar"}},
					},
					found: true,
				},
				{
					name: "deepObject explode nested object with nested array on different levels",
					param: &openapi3.Parameter{
						Name: "param", In: "query", Style: "deepObject", Explode: explode,
						Schema: objectOf(
							"obj", objectOf("nestedObjOne", objectOf("items", stringArraySchema)),
							"objTwo", objectOf("items", stringArraySchema),
						),
					},
					query: "param[obj][nestedObjOne][items][0]=baz&param[objTwo][items][0]=foo&param[objTwo][items][1]=bar",
					want: map[string]interface{}{
						"obj":    map[string]interface{}{"nestedObjOne": map[string]interface{}{"items": []interface{}{"baz"}}},
						"objTwo": map[string]interface{}{"items": []interface{}{"foo", "bar"}},
					},
					found: true,
				},
				{
					name: "deepObject explode array of arrays",
					param: &openapi3.Parameter{
						Name: "param", In: "query", Style: "deepObject", Explode: explode,
						Schema: objectOf(
							"arr", arrayOf(arrayOf(integerSchema)),
						),
					},
					query: "param[arr][1][1]=123&param[arr][1][2]=456",
					want: map[string]interface{}{
						"arr": []interface{}{
							nil,
							[]interface{}{nil, int64(123), int64(456)},
						},
					},
					found: true,
				},
				{
					name: "deepObject explode nested array of objects - missing intermediate array index",
					param: &openapi3.Parameter{
						Name: "param", In: "query", Style: "deepObject", Explode: explode,
						Schema: objectOf(
							"arr", arrayOf(objectOf("key", booleanSchema)),
						),
					},
					query: "param[arr][3][key]=true&param[arr][0][key]=false",
					want: map[string]interface{}{
						"arr": []interface{}{
							map[string]interface{}{"key": false},
							nil,
							nil,
							map[string]interface{}{"key": true},
						},
					},
					found: true,
				},
				{
					name: "deepObject explode nested array of objects",
					param: &openapi3.Parameter{
						Name: "param", In: "query", Style: "deepObject", Explode: explode,
						Schema: objectOf(
							"arr", arrayOf(objectOf("key", booleanSchema)),
						),
					},
					query: "param[arr][0][key]=true&param[arr][1][key]=false",
					found: true,
					want: map[string]interface{}{
						"arr": []interface{}{
							map[string]interface{}{"key": true},
							map[string]interface{}{"key": false},
						},
					},
				},
				{
					name:  "default",
					param: &openapi3.Parameter{Name: "param", In: "query", Schema: objectSchema},
					query: "id=foo&name=bar",
					want:  map[string]interface{}{"id": "foo", "name": "bar"},
					found: true,
				},
				{
					name:  "invalid integer prop",
					param: &openapi3.Parameter{Name: "param", In: "query", Schema: objectOf("foo", integerSchema)},
					query: "foo=bar",
					found: true,
					err:   &ParseError{path: []interface{}{"foo"}, Cause: &ParseError{Kind: KindInvalidFormat, Value: "bar"}},
				},
				{
					name:  "invalid number prop",
					param: &openapi3.Parameter{Name: "param", In: "query", Schema: objectOf("foo", numberSchema)},
					query: "foo=bar",
					found: true,
					err:   &ParseError{path: []interface{}{"foo"}, Cause: &ParseError{Kind: KindInvalidFormat, Value: "bar"}},
				},
				{
					name:  "invalid boolean prop",
					param: &openapi3.Parameter{Name: "param", In: "query", Schema: objectOf("foo", booleanSchema)},
					query: "foo=bar",
					found: true,
					err:   &ParseError{path: []interface{}{"foo"}, Cause: &ParseError{Kind: KindInvalidFormat, Value: "bar"}},
				},
			},
		},
		{
			name: "header primitive",
			testCases: []testCase{
				{
					name:   "simple",
					param:  &openapi3.Parameter{Name: "X-Param", In: "header", Style: "simple", Explode: noExplode, Schema: stringSchema},
					header: "X-Param:foo",
					want:   "foo",
					found:  true,
				},
				{
					name:   "simple explode",
					param:  &openapi3.Parameter{Name: "X-Param", In: "header", Style: "simple", Explode: explode, Schema: stringSchema},
					header: "X-Param:foo",
					want:   "foo",
					found:  true,
				},
				{
					name:   "default",
					param:  &openapi3.Parameter{Name: "X-Param", In: "header", Schema: stringSchema},
					header: "X-Param:foo",
					want:   "foo",
					found:  true,
				},
				{
					name:   "string",
					param:  &openapi3.Parameter{Name: "X-Param", In: "header", Schema: stringSchema},
					header: "X-Param:foo",
					want:   "foo",
					found:  true,
				},
				{
					name:   "integer",
					param:  &openapi3.Parameter{Name: "X-Param", In: "header", Schema: integerSchema},
					header: "X-Param:1",
					want:   int64(1),
					found:  true,
				},
				{
					name:   "integer invalid",
					param:  &openapi3.Parameter{Name: "X-Param", In: "header", Schema: integerSchema},
					header: "X-Param:foo",
					found:  true,
					err:    &ParseError{Kind: KindInvalidFormat, Value: "foo"},
				},
				{
					name:   "number",
					param:  &openapi3.Parameter{Name: "X-Param", In: "header", Schema: numberSchema},
					header: "X-Param:1.1",
					want:   1.1,
					found:  true,
				},
				{
					name:   "number invalid",
					param:  &openapi3.Parameter{Name: "X-Param", In: "header", Schema: numberSchema},
					header: "X-Param:foo",
					found:  true,
					err:    &ParseError{Kind: KindInvalidFormat, Value: "foo"},
				},
				{
					name:   "boolean",
					param:  &openapi3.Parameter{Name: "X-Param", In: "header", Schema: booleanSchema},
					header: "X-Param:true",
					want:   true,
					found:  true,
				},
				{
					name:   "boolean invalid",
					param:  &openapi3.Parameter{Name: "X-Param", In: "header", Schema: booleanSchema},
					header: "X-Param:foo",
					found:  true,
					err:    &ParseError{Kind: KindInvalidFormat, Value: "foo"},
				},
			},
		},
		{
			name: "header array",
			testCases: []testCase{
				{
					name:   "simple",
					param:  &openapi3.Parameter{Name: "X-Param", In: "header", Style: "simple", Explode: noExplode, Schema: stringArraySchema},
					header: "X-Param:foo,bar",
					want:   []interface{}{"foo", "bar"},
					found:  true,
				},
				{
					name:   "simple explode",
					param:  &openapi3.Parameter{Name: "X-Param", In: "header", Style: "simple", Explode: explode, Schema: stringArraySchema},
					header: "X-Param:foo,bar",
					want:   []interface{}{"foo", "bar"},
					found:  true,
				},
				{
					name:   "default",
					param:  &openapi3.Parameter{Name: "X-Param", In: "header", Schema: stringArraySchema},
					header: "X-Param:foo,bar",
					want:   []interface{}{"foo", "bar"},
					found:  true,
				},
				{
					name:   "invalid integer items",
					param:  &openapi3.Parameter{Name: "X-Param", In: "header", Schema: arrayOf(integerSchema)},
					header: "X-Param:1,foo",
					found:  true,
					err:    &ParseError{path: []interface{}{1}, Cause: &ParseError{Kind: KindInvalidFormat, Value: "foo"}},
				},
				{
					name:   "invalid number items",
					param:  &openapi3.Parameter{Name: "X-Param", In: "header", Schema: arrayOf(numberSchema)},
					header: "X-Param:1.1,foo",
					found:  true,
					err:    &ParseError{path: []interface{}{1}, Cause: &ParseError{Kind: KindInvalidFormat, Value: "foo"}},
				},
				{
					name:   "invalid boolean items",
					param:  &openapi3.Parameter{Name: "X-Param", In: "header", Schema: arrayOf(booleanSchema)},
					header: "X-Param:true,foo",
					found:  true,
					err:    &ParseError{path: []interface{}{1}, Cause: &ParseError{Kind: KindInvalidFormat, Value: "foo"}},
				},
			},
		},
		{
			name: "header object",
			testCases: []testCase{
				{
					name:   "simple",
					param:  &openapi3.Parameter{Name: "X-Param", In: "header", Style: "simple", Explode: noExplode, Schema: objectSchema},
					header: "X-Param:id,foo,name,bar",
					want:   map[string]interface{}{"id": "foo", "name": "bar"},
					found:  true,
				},
				{
					name:   "simple explode",
					param:  &openapi3.Parameter{Name: "X-Param", In: "header", Style: "simple", Explode: explode, Schema: objectSchema},
					header: "X-Param:id=foo,name=bar",
					want:   map[string]interface{}{"id": "foo", "name": "bar"},
					found:  true,
				},
				{
					name:   "default",
					param:  &openapi3.Parameter{Name: "X-Param", In: "header", Schema: objectSchema},
					header: "X-Param:id,foo,name,bar",
					want:   map[string]interface{}{"id": "foo", "name": "bar"},
					found:  true,
				},
				{
					name:   "valid integer prop",
					param:  &openapi3.Parameter{Name: "X-Param", In: "header", Schema: integerSchema},
					header: "X-Param:88",
					found:  true,
					want:   int64(88),
				},
				{
					name:   "invalid integer prop",
					param:  &openapi3.Parameter{Name: "X-Param", In: "header", Schema: objectOf("foo", integerSchema)},
					header: "X-Param:foo,bar",
					found:  true,
					err:    &ParseError{path: []interface{}{"foo"}, Cause: &ParseError{Kind: KindInvalidFormat, Value: "bar"}},
				},
				{
					name:   "invalid number prop",
					param:  &openapi3.Parameter{Name: "X-Param", In: "header", Schema: objectOf("foo", numberSchema)},
					header: "X-Param:foo,bar",
					found:  true,
					err:    &ParseError{path: []interface{}{"foo"}, Cause: &ParseError{Kind: KindInvalidFormat, Value: "bar"}},
				},
				{
					name:   "invalid boolean prop",
					param:  &openapi3.Parameter{Name: "X-Param", In: "header", Schema: objectOf("foo", booleanSchema)},
					header: "X-Param:foo,bar",
					found:  true,
					err:    &ParseError{path: []interface{}{"foo"}, Cause: &ParseError{Kind: KindInvalidFormat, Value: "bar"}},
				},
			},
		},
		{
			name: "cookie primitive",
			testCases: []testCase{
				{
					name:   "form",
					param:  &openapi3.Parameter{Name: "X-Param", In: "cookie", Style: "form", Explode: noExplode, Schema: stringSchema},
					cookie: "X-Param:foo",
					want:   "foo",
					found:  true,
				},
				{
					name:   "form explode",
					param:  &openapi3.Parameter{Name: "X-Param", In: "cookie", Style: "form", Explode: explode, Schema: stringSchema},
					cookie: "X-Param:foo",
					want:   "foo",
					found:  true,
				},
				{
					name:   "default",
					param:  &openapi3.Parameter{Name: "X-Param", In: "cookie", Schema: stringSchema},
					cookie: "X-Param:foo",
					want:   "foo",
					found:  true,
				},
				{
					name:   "string",
					param:  &openapi3.Parameter{Name: "X-Param", In: "cookie", Schema: stringSchema},
					cookie: "X-Param:foo",
					want:   "foo",
					found:  true,
				},
				{
					name:   "integer",
					param:  &openapi3.Parameter{Name: "X-Param", In: "cookie", Schema: integerSchema},
					cookie: "X-Param:1",
					want:   int64(1),
					found:  true,
				},
				{
					name:   "integer invalid",
					param:  &openapi3.Parameter{Name: "X-Param", In: "cookie", Schema: integerSchema},
					cookie: "X-Param:foo",
					found:  true,
					err:    &ParseError{Kind: KindInvalidFormat, Value: "foo"},
				},
				{
					name:   "number",
					param:  &openapi3.Parameter{Name: "X-Param", In: "cookie", Schema: numberSchema},
					cookie: "X-Param:1.1",
					want:   1.1,
					found:  true,
				},
				{
					name:   "number invalid",
					param:  &openapi3.Parameter{Name: "X-Param", In: "cookie", Schema: numberSchema},
					cookie: "X-Param:foo",
					found:  true,
					err:    &ParseError{Kind: KindInvalidFormat, Value: "foo"},
				},
				{
					name:   "boolean",
					param:  &openapi3.Parameter{Name: "X-Param", In: "cookie", Schema: booleanSchema},
					cookie: "X-Param:true",
					want:   true,
					found:  true,
				},
				{
					name:   "boolean invalid",
					param:  &openapi3.Parameter{Name: "X-Param", In: "cookie", Schema: booleanSchema},
					cookie: "X-Param:foo",
					found:  true,
					err:    &ParseError{Kind: KindInvalidFormat, Value: "foo"},
				},
			},
		},
		{
			name: "cookie array",
			testCases: []testCase{
				{
					name:   "form",
					param:  &openapi3.Parameter{Name: "X-Param", In: "cookie", Style: "form", Explode: noExplode, Schema: stringArraySchema},
					cookie: "X-Param:foo,bar",
					want:   []interface{}{"foo", "bar"},
					found:  true,
				},
				{
					name:   "invalid integer items",
					param:  &openapi3.Parameter{Name: "X-Param", In: "cookie", Style: "form", Explode: noExplode, Schema: arrayOf(integerSchema)},
					cookie: "X-Param:1,foo",
					found:  true,
					err:    &ParseError{path: []interface{}{1}, Cause: &ParseError{Kind: KindInvalidFormat, Value: "foo"}},
				},
				{
					name:   "invalid number items",
					param:  &openapi3.Parameter{Name: "X-Param", In: "cookie", Style: "form", Explode: noExplode, Schema: arrayOf(numberSchema)},
					cookie: "X-Param:1.1,foo",
					found:  true,
					err:    &ParseError{path: []interface{}{1}, Cause: &ParseError{Kind: KindInvalidFormat, Value: "foo"}},
				},
				{
					name:   "invalid boolean items",
					param:  &openapi3.Parameter{Name: "X-Param", In: "cookie", Style: "form", Explode: noExplode, Schema: arrayOf(booleanSchema)},
					cookie: "X-Param:true,foo",
					found:  true,
					err:    &ParseError{path: []interface{}{1}, Cause: &ParseError{Kind: KindInvalidFormat, Value: "foo"}},
				},
			},
		},
		{
			name: "cookie object",
			testCases: []testCase{
				{
					name:   "form",
					param:  &openapi3.Parameter{Name: "X-Param", In: "cookie", Style: "form", Explode: noExplode, Schema: objectSchema},
					cookie: "X-Param:id,foo,name,bar",
					want:   map[string]interface{}{"id": "foo", "name": "bar"},
					found:  true,
				},
				{
					name:   "invalid integer prop",
					param:  &openapi3.Parameter{Name: "X-Param", In: "cookie", Style: "form", Explode: noExplode, Schema: objectOf("foo", integerSchema)},
					cookie: "X-Param:foo,bar",
					found:  true,
					err:    &ParseError{path: []interface{}{"foo"}, Cause: &ParseError{Kind: KindInvalidFormat, Value: "bar"}},
				},
				{
					name:   "invalid number prop",
					param:  &openapi3.Parameter{Name: "X-Param", In: "cookie", Style: "form", Explode: noExplode, Schema: objectOf("foo", numberSchema)},
					cookie: "X-Param:foo,bar",
					found:  true,
					err:    &ParseError{path: []interface{}{"foo"}, Cause: &ParseError{Kind: KindInvalidFormat, Value: "bar"}},
				},
				{
					name:   "invalid boolean prop",
					param:  &openapi3.Parameter{Name: "X-Param", In: "cookie", Style: "form", Explode: noExplode, Schema: objectOf("foo", booleanSchema)},
					cookie: "X-Param:foo,bar",
					found:  true,
					err:    &ParseError{path: []interface{}{"foo"}, Cause: &ParseError{Kind: KindInvalidFormat, Value: "bar"}},
				},
			},
		},
	}

	for _, tg := range testGroups {
		t.Run(tg.name, func(t *testing.T) {
			for _, tc := range tg.testCases {
				t.Run(tc.name, func(t *testing.T) {
					req, err := http.NewRequest(http.MethodGet, "http://test.org/test"+tc.path, nil)
					require.NoError(t, err, "failed to create a test request")

					if tc.query != "" {
						query := req.URL.Query()
						for _, param := range strings.Split(tc.query, "&") {
							v := strings.Split(param, "=")
							query.Add(v[0], v[1])
						}
						req.URL.RawQuery = query.Encode()
					}

					if tc.header != "" {
						v := strings.Split(tc.header, ":")
						req.Header.Add(v[0], v[1])
					}

					if tc.cookie != "" {
						v := strings.Split(tc.cookie, ":")
						req.AddCookie(&http.Cookie{Name: v[0], Value: v[1]})
					}

					path := "/test"
					if tc.path != "" {
						path += "/{" + tc.param.Name + "}"
						tc.param.Required = true
					}

					info := &openapi3.Info{
						Title:   "MyAPI",
						Version: "0.1",
					}
					doc := &openapi3.T{OpenAPI: "3.0.0", Info: info, Paths: openapi3.NewPaths()}
					op := &openapi3.Operation{
						OperationID: "test",
						Parameters:  []*openapi3.ParameterRef{{Value: tc.param}},
						Responses:   openapi3.NewResponses(),
					}
					doc.AddOperation(path, http.MethodGet, op)
					err = doc.Validate(context.Background())
					require.NoError(t, err)
					router, err := legacyrouter.NewRouter(doc)
					require.NoError(t, err)

					route, pathParams, err := router.FindRoute(req)
					require.NoError(t, err)

					input := &RequestValidationInput{Request: req, PathParams: pathParams, Route: route}
					got, found, err := decodeStyledParameter(tc.param, input)

					if tc.err != nil {
						require.Error(t, err)
						matchParseError(t, err, tc.err)

						return
					}

					require.NoError(t, err)
					require.EqualValues(t, tc.want, got)

					require.Truef(t, found == tc.found, "got found: %t, want found: %t", found, tc.found)
				})
			}
		})
	}
}

func TestDecodeBody(t *testing.T) {
	urlencodedForm := make(url.Values)
	urlencodedForm.Set("a", "a1")
	urlencodedForm.Set("b", "10")
	urlencodedForm.Add("c", "c1")
	urlencodedForm.Add("c", "c2")

	urlencodedSpaceDelim := make(url.Values)
	urlencodedSpaceDelim.Set("a", "a1")
	urlencodedSpaceDelim.Set("b", "10")
	urlencodedSpaceDelim.Add("c", "c1 c2")

	urlencodedPipeDelim := make(url.Values)
	urlencodedPipeDelim.Set("a", "a1")
	urlencodedPipeDelim.Set("b", "10")
	urlencodedPipeDelim.Add("c", "c1|c2")

	d, err := json.Marshal(map[string]interface{}{"d1": "d1"})
	require.NoError(t, err)
	multipartForm, multipartFormMime, err := newTestMultipartForm([]*testFormPart{
		{name: "a", contentType: "text/plain", data: strings.NewReader("a1")},
		{name: "b", contentType: "application/json", data: strings.NewReader("10")},
		{name: "c", contentType: "text/plain", data: strings.NewReader("c1")},
		{name: "c", contentType: "text/plain", data: strings.NewReader("c2")},
		{name: "d", contentType: "application/json", data: bytes.NewReader(d)},
		{name: "f", contentType: "application/octet-stream", data: strings.NewReader("foo"), filename: "f1"},
		{name: "g", data: strings.NewReader("g1")},
	})
	require.NoError(t, err)

	multipartFormExtraPart, multipartFormMimeExtraPart, err := newTestMultipartForm([]*testFormPart{
		{name: "a", contentType: "text/plain", data: strings.NewReader("a1")},
		{name: "x", contentType: "text/plain", data: strings.NewReader("x1")},
	})
	require.NoError(t, err)

	multipartAnyAdditionalProps, multipartMimeAnyAdditionalProps, err := newTestMultipartForm([]*testFormPart{
		{name: "a", contentType: "text/plain", data: strings.NewReader("a1")},
		{name: "x", contentType: "text/plain", data: strings.NewReader("x1")},
	})
	require.NoError(t, err)

	multipartAdditionalProps, multipartMimeAdditionalProps, err := newTestMultipartForm([]*testFormPart{
		{name: "a", contentType: "text/plain", data: strings.NewReader("a1")},
		{name: "x", contentType: "text/plain", data: strings.NewReader("x1")},
	})
	require.NoError(t, err)

	multipartAdditionalPropsErr, multipartMimeAdditionalPropsErr, err := newTestMultipartForm([]*testFormPart{
		{name: "a", contentType: "text/plain", data: strings.NewReader("a1")},
		{name: "x", contentType: "text/plain", data: strings.NewReader("x1")},
		{name: "y", contentType: "text/plain", data: strings.NewReader("y1")},
	})
	require.NoError(t, err)

	testCases := []struct {
		name     string
		mime     string
		body     io.Reader
		schema   *openapi3.Schema
		encoding map[string]*openapi3.Encoding
		want     interface{}
		wantErr  error
	}{
		{
			name:    prefixUnsupportedCT,
			mime:    "application/xml",
			wantErr: &ParseError{Kind: KindUnsupportedFormat},
		},
		{
			name:    "invalid body data",
			mime:    "application/json",
			body:    strings.NewReader("invalid"),
			wantErr: &ParseError{Kind: KindInvalidFormat},
		},
		{
			name: "plain text",
			mime: "text/plain",
			body: strings.NewReader("text"),
			want: "text",
		},
		{
			name: "json",
			mime: "application/json",
			body: strings.NewReader("\"foo\""),
			want: "foo",
		},
		{
			name: "x-yaml",
			mime: "application/x-yaml",
			body: strings.NewReader("foo"),
			want: "foo",
		},
		{
			name: "yaml",
			mime: "application/yaml",
			body: strings.NewReader("foo"),
			want: "foo",
		},
		{
			name: "urlencoded form",
			mime: "application/x-www-form-urlencoded",
			body: strings.NewReader(urlencodedForm.Encode()),
			schema: openapi3.NewObjectSchema().
				WithProperty("a", openapi3.NewStringSchema()).
				WithProperty("b", openapi3.NewIntegerSchema()).
				WithProperty("c", openapi3.NewArraySchema().WithItems(openapi3.NewStringSchema())),
			want: map[string]interface{}{"a": "a1", "b": int64(10), "c": []interface{}{"c1", "c2"}},
		},
		{
			name: "urlencoded space delimited",
			mime: "application/x-www-form-urlencoded",
			body: strings.NewReader(urlencodedSpaceDelim.Encode()),
			schema: openapi3.NewObjectSchema().
				WithProperty("a", openapi3.NewStringSchema()).
				WithProperty("b", openapi3.NewIntegerSchema()).
				WithProperty("c", openapi3.NewArraySchema().WithItems(openapi3.NewStringSchema())),
			encoding: map[string]*openapi3.Encoding{
				"c": {Style: openapi3.SerializationSpaceDelimited, Explode: openapi3.BoolPtr(false)},
			},
			want: map[string]interface{}{"a": "a1", "b": int64(10), "c": []interface{}{"c1", "c2"}},
		},
		{
			name: "urlencoded pipe delimited",
			mime: "application/x-www-form-urlencoded",
			body: strings.NewReader(urlencodedPipeDelim.Encode()),
			schema: openapi3.NewObjectSchema().
				WithProperty("a", openapi3.NewStringSchema()).
				WithProperty("b", openapi3.NewIntegerSchema()).
				WithProperty("c", openapi3.NewArraySchema().WithItems(openapi3.NewStringSchema())),
			encoding: map[string]*openapi3.Encoding{
				"c": {Style: openapi3.SerializationPipeDelimited, Explode: openapi3.BoolPtr(false)},
			},
			want: map[string]interface{}{"a": "a1", "b": int64(10), "c": []interface{}{"c1", "c2"}},
		},
		{
			name: "multipart",
			mime: multipartFormMime,
			body: multipartForm,
			schema: openapi3.NewObjectSchema().
				WithProperty("a", openapi3.NewStringSchema()).
				WithProperty("b", openapi3.NewIntegerSchema()).
				WithProperty("c", openapi3.NewArraySchema().WithItems(openapi3.NewStringSchema())).
				WithProperty("d", openapi3.NewObjectSchema().WithProperty("d1", openapi3.NewStringSchema())).
				WithProperty("f", openapi3.NewStringSchema().WithFormat("binary")).
				WithProperty("g", openapi3.NewStringSchema()),
			want: map[string]interface{}{"a": "a1", "b": json.Number("10"), "c": []interface{}{"c1", "c2"}, "d": map[string]interface{}{"d1": "d1"}, "f": "foo", "g": "g1"},
		},
		{
			name: "multipartExtraPart",
			mime: multipartFormMimeExtraPart,
			body: multipartFormExtraPart,
			schema: openapi3.NewObjectSchema().
				WithProperty("a", openapi3.NewStringSchema()),
			want:    map[string]interface{}{"a": "a1"},
			wantErr: &ParseError{Kind: KindOther},
		},
		{
			name: "multipartAnyAdditionalProperties",
			mime: multipartMimeAnyAdditionalProps,
			body: multipartAnyAdditionalProps,
			schema: openapi3.NewObjectSchema().
				WithAnyAdditionalProperties().
				WithProperty("a", openapi3.NewStringSchema()),
			want: map[string]interface{}{"a": "a1"},
		},
		{
			name: "multipartWithAdditionalProperties",
			mime: multipartMimeAdditionalProps,
			body: multipartAdditionalProps,
			schema: openapi3.NewObjectSchema().
				WithAdditionalProperties(openapi3.NewObjectSchema().
					WithProperty("x", openapi3.NewStringSchema())).
				WithProperty("a", openapi3.NewStringSchema()),
			want: map[string]interface{}{"a": "a1", "x": "x1"},
		},
		{
			name: "multipartWithAdditionalPropertiesError",
			mime: multipartMimeAdditionalPropsErr,
			body: multipartAdditionalPropsErr,
			schema: openapi3.NewObjectSchema().
				WithAdditionalProperties(openapi3.NewObjectSchema().
					WithProperty("x", openapi3.NewStringSchema())).
				WithProperty("a", openapi3.NewStringSchema()),
			want:    map[string]interface{}{"a": "a1", "x": "x1"},
			wantErr: &ParseError{Kind: KindOther},
		},
		{
			name: "file",
			mime: "application/octet-stream",
			body: strings.NewReader("foo"),
			want: "foo",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := make(http.Header)
			h.Set(headerCT, tc.mime)
			var schemaRef *openapi3.SchemaRef
			if tc.schema != nil {
				schemaRef = tc.schema.NewRef()
			}
			encFn := func(name string) *openapi3.Encoding {
				if tc.encoding == nil {
					return nil
				}
				return tc.encoding[name]
			}
			_, got, err := decodeBody(tc.body, h, schemaRef, encFn)

			if tc.wantErr != nil {
				require.Error(t, err)
				matchParseError(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
			require.Truef(t, reflect.DeepEqual(got, tc.want), "got %v, want %v", got, tc.want)
		})
	}
}

type testFormPart struct {
	name        string
	contentType string
	data        io.Reader
	filename    string
}

func newTestMultipartForm(parts []*testFormPart) (io.Reader, string, error) {
	form := &bytes.Buffer{}
	w := multipart.NewWriter(form)
	defer w.Close()

	for _, p := range parts {
		var disp string
		if p.filename == "" {
			disp = fmt.Sprintf("form-data; name=%q", p.name)
		} else {
			disp = fmt.Sprintf("form-data; name=%q; filename=%q", p.name, p.filename)
		}

		h := make(textproto.MIMEHeader)
		h.Set(headerCT, p.contentType)
		h.Set("Content-Disposition", disp)
		pw, err := w.CreatePart(h)
		if err != nil {
			return nil, "", err
		}
		if _, err = io.Copy(pw, p.data); err != nil {
			return nil, "", err
		}
	}
	return form, w.FormDataContentType(), nil
}

func TestRegisterAndUnregisterBodyDecoder(t *testing.T) {
	var decoder BodyDecoder
	decoder = func(body io.Reader, h http.Header, schema *openapi3.SchemaRef, encFn EncodingFn) (decoded interface{}, err error) {
		var data []byte
		if data, err = io.ReadAll(body); err != nil {
			return
		}
		return strings.Split(string(data), ","), nil
	}
	contentType := "application/csv"
	h := make(http.Header)
	h.Set(headerCT, contentType)

	originalDecoder := RegisteredBodyDecoder(contentType)
	require.Nil(t, originalDecoder)

	RegisterBodyDecoder(contentType, decoder)
	require.Equal(t, fmt.Sprintf("%v", decoder), fmt.Sprintf("%v", RegisteredBodyDecoder(contentType)))

	body := strings.NewReader("foo,bar")
	schema := openapi3.NewArraySchema().WithItems(openapi3.NewStringSchema()).NewRef()
	encFn := func(string) *openapi3.Encoding { return nil }
	_, got, err := decodeBody(body, h, schema, encFn)

	require.NoError(t, err)
	require.Equal(t, []string{"foo", "bar"}, got)

	UnregisterBodyDecoder(contentType)

	originalDecoder = RegisteredBodyDecoder(contentType)
	require.Nil(t, originalDecoder)

	_, _, err = decodeBody(body, h, schema, encFn)
	require.Equal(t, &ParseError{
		Kind:   KindUnsupportedFormat,
		Reason: prefixUnsupportedCT + ` "application/csv"`,
	}, err)
}

func matchParseError(t *testing.T, got, want error) {
	t.Helper()

	wErr, ok := want.(*ParseError)
	if !ok {
		t.Errorf("want error is not a ParseError")
		return
	}
	gErr, ok := got.(*ParseError)
	if !ok {
		t.Errorf("got error is not a ParseError")
		return
	}
	assert.Equalf(t, wErr.Kind, gErr.Kind, "ParseError Kind differs")
	assert.Equalf(t, wErr.Value, gErr.Value, "ParseError Value differs")
	assert.Equalf(t, wErr.Path(), gErr.Path(), "ParseError Path differs")

	if wErr.Reason != "" {
		assert.Equalf(t, wErr.Reason, gErr.Reason, "ParseError Reason differs")
	}
	if wErr.Cause != nil {
		matchParseError(t, gErr.Cause, wErr.Cause)
	}
}
