package openapi3gen_test

import (
	"encoding/json"
	"github.com/ronniedada/kin-openapi/openapi3gen"
	"github.com/ronniedada/kin-test/jsontest"
	"testing"
	"time"
)

type CyclicType0 struct {
	CyclicField *CyclicType1 `json:"a"`
}

type CyclicType1 struct {
	CyclicField *CyclicType0 `json:"b"`
}

func TestCyclic(t *testing.T) {
	schema, _, err := openapi3gen.NewSchemaRefForValue(&CyclicType0{})
	jsontest.ExpectWithErr(t, schema, err).ErrOfType(&openapi3gen.CycleError{})
}

type Example struct {
	Bool    bool                     `json:"bool"`
	Int     int                      `json:"int"`
	Int64   int64                    `json:"int64"`
	Float64 float64                  `json:"float64"`
	String  string                   `json:"string"`
	Bytes   []byte                   `json:"bytes"`
	JSON    json.RawMessage          `json:"json"`
	Time    time.Time                `json:"time"`
	Slice   []*ExampleChild          `json:"slice"`
	Map     map[string]*ExampleChild `json:"map"`
	Struct  struct {
		X string `json:"x"`
	} `json:"struct"`
	EmptyStruct struct {
		X string
	} `json:"structWithoutFields"`
	Ptr *ExampleChild `json:"ptr"`
}

type ExampleChild string

type Array []interface{}
type Object map[string]interface{}

func TestSimple(t *testing.T) {
	schema, _, err := openapi3gen.NewSchemaRefForValue(&Example{})
	jsontest.ExpectWithErr(t, schema, err).Value(Object{
		"type": "object",
		"properties": Object{
			"bool": Object{
				"type": "boolean",
			},
			"int": Object{
				"type":   "number",
				"format": "int64",
			},
			"int64": Object{
				"type":   "number",
				"format": "int64",
			},
			"float64": Object{
				"type": "number",
			},
			"time": Object{
				"type":   "string",
				"format": "date-time",
			},
			"string": Object{
				"type": "string",
			},
			"bytes": Object{
				"type":   "string",
				"format": "byte",
			},
			"json": Object{},
			"slice": Object{
				"type": "array",
				"items": Object{
					"type": "string",
				},
			},
			"map": Object{
				"type": "object",
				"additionalProperties": Object{
					"type": "string",
				},
			},
			"struct": Object{
				"type": "object",
				"properties": Object{
					"x": Object{
						"type": "string",
					},
				},
			},
			"structWithoutFields": Object{},
			"ptr": Object{
				"type": "string",
			},
		},
	})
}
