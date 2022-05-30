package openapi3gen_test

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/getkin/kin-openapi/openapi3gen"
)

type (
	SomeStruct struct {
		Time        time.Time                 `json:"time"`
		Ptr         *SomeOtherType            `json:"ptr"`
		Map         map[string]*SomeOtherType `json:"map"`
		EmptyStruct struct{ Y string }        `json:"structWithoutFields"`
		Struct      struct {
			X string `json:"x"`
		} `json:"struct"`
		String  string          `json:"string"`
		JSON    json.RawMessage `json:"json"`
		Bytes   []byte          `json:"bytes"`
		Slice   []SomeOtherType `json:"slice"`
		Float64 float64         `json:"float64"`
		Int64   int64           `json:"int64"`
		Int     int             `json:"int"`
		Bool    bool            `json:"bool"`
	}

	SomeOtherType string
)

func Example() {
	schemaRef, err := openapi3gen.NewSchemaRefForValue(&SomeStruct{}, nil)
	if err != nil {
		panic(err)
	}

	data, err := json.MarshalIndent(schemaRef, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", data)
	// Output:
	// {
	//   "properties": {
	//     "bool": {
	//       "type": "boolean"
	//     },
	//     "bytes": {
	//       "format": "byte",
	//       "type": "string"
	//     },
	//     "float64": {
	//       "format": "double",
	//       "type": "number"
	//     },
	//     "int": {
	//       "type": "integer"
	//     },
	//     "int64": {
	//       "format": "int64",
	//       "type": "integer"
	//     },
	//     "json": {},
	//     "map": {
	//       "additionalProperties": {
	//         "type": "string"
	//       },
	//       "type": "object"
	//     },
	//     "ptr": {
	//       "type": "string"
	//     },
	//     "slice": {
	//       "items": {
	//         "type": "string"
	//       },
	//       "type": "array"
	//     },
	//     "string": {
	//       "type": "string"
	//     },
	//     "struct": {
	//       "properties": {
	//         "x": {
	//           "type": "string"
	//         }
	//       },
	//       "type": "object"
	//     },
	//     "structWithoutFields": {},
	//     "time": {
	//       "format": "date-time",
	//       "type": "string"
	//     }
	//   },
	//   "type": "object"
	// }
}
