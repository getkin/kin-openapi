package jsoninfo_test

import (
	"encoding/json"
	"github.com/jban332/kin-openapi/jsoninfo"
	"testing"
	"time"
)

type SimpleType struct {
	Bool    bool      `json:"bool,omitempty"`
	Int     int       `json:"int,omitempty"`
	Int64   int64     `json:"int64,omitempty"`
	Float64 float64   `json:"float64,omitempty"`
	Time    time.Time `json:"time,omitempty"`
	String  string    `json:"string,omitempty"`
	Bytes   []byte    `json:"bytes,omitempty"`
}

type SimplePtrType struct {
	Bool    *bool      `json:"bool,omitempty"`
	Int     *int       `json:"int,omitempty"`
	Int64   *int64     `json:"int64,omitempty"`
	Float64 *float64   `json:"float64,omitempty"`
	Time    *time.Time `json:"time,omitempty"`
	String  *string    `json:"string,omitempty"`
	Bytes   *[]byte    `json:"bytes,omitempty"`
}

type EmptyType struct {
	Bool    bool      `json:"bool"`
	Int     int       `json:"int"`
	Int64   int64     `json:"int64"`
	Float64 float64   `json:"float64"`
	Time    time.Time `json:"time"`
	String  string    `json:"string"`
	Bytes   []byte    `json:"bytes"`
}

type OriginalNameType struct {
	Field string `json:",omitempty"`
}

type RootType struct {
	jsoninfo.ExtensionProps
	EmbeddedType0
	EmbeddedType1
}

type EmbeddedType0 struct {
	Field0 string `json:"embedded0,omitempty"`
}

type EmbeddedType1 struct {
	Field1 string `json:"embedded1,omitempty"`
}

type Example struct {
	NoMarshal   bool
	NoUnmarshal bool
	Value       interface{}
	JSON        interface{}
}

var Examples = []Example{
	// Primitives
	{
		Value: &SimpleType{},
		JSON: Object{
			"time": time.Unix(0, 0),
		},
	},
	{
		Value: &SimpleType{},
		JSON: Object{
			"bool":    true,
			"int":     42,
			"int64":   42,
			"float64": 3.14,
			"string":  "abc",
			"bytes":   []byte{1, 2, 3},
			"time":    time.Unix(1, 0),
		},
	},

	// Pointers
	{
		Value: &SimplePtrType{},
		JSON:  Object{},
	},
	{
		Value: &SimplePtrType{},
		JSON: Object{
			"bool":    true,
			"int":     42,
			"int64":   42,
			"float64": 3.14,
			"string":  "abc",
			"bytes":   []byte{1, 2, 3},
			"time":    time.Unix(1, 0),
		},
	},

	// JSON tag "fieldName"
	{
		Value: &EmptyType{},
		JSON: Object{
			"bool":    false,
			"int":     0,
			"int64":   0,
			"float64": 0,
			"string":  "",
			"bytes":   []byte{},
			"time":    time.Unix(0, 0),
		},
	},

	// JSON tag ",omitempty"
	{
		Value: &OriginalNameType{},
		JSON: Object{
			"Field": "abc",
		},
	},

	// Embedding
	{
		Value: &RootType{},
		JSON:  Object{},
	},
	{
		Value: &RootType{},
		JSON: Object{
			"embedded0": "0",
			"embedded1": "1",
			"x-other":   nil,
		},
	},
}

type Object map[string]interface{}
type Array []interface{}

func TestExtensions(t *testing.T) {
	for _, example := range Examples {
		// Define JSON that will be unmarshalled
		expectedData, err := json.Marshal(example.JSON)
		if err != nil {
			panic(err)
		}
		expected := string(expectedData)

		// Define value that will marshalled
		x := example.Value

		// Unmarshal
		if !example.NoUnmarshal {
			t.Logf("Unmarshalling %T", x)
			err = jsoninfo.UnmarshalStructFields(expectedData, x)
			if err != nil {
				t.Fatalf("Error unmarshalling %T: %v", x, err)
			}
			t.Logf("Marshalling %T", x)
		}

		// Marshal
		if !example.NoMarshal {
			data, err := jsoninfo.MarshalStructFields(x)
			if err != nil {
				t.Fatalf("Error marshalling: %v", err)
			}
			actually := string(data)

			if actually != expected {
				t.Fatalf("Error!\nExpected: %s\nActually: %s", expected, actually)
			}
		}
	}
}
