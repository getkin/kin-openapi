package openapi3gen_test

import (
	"github.com/jban332/kinapi/openapi3gen"
	"github.com/jban332/kincore/jsontest"
	"testing"
)

type CyclicType0 struct {
	CyclicField *CyclicType1 `json:"a"`
}

type CyclicType1 struct {
	CyclicField *CyclicType0 `json:"b"`
}

func Test_Schema_Generation(t *testing.T) {
	schema, err := openapi3gen.SchemaFromInstance(&CyclicType0{})
	jsontest.ExpectWithErr(t, schema, err).ErrOfType(&openapi3gen.CycleError{})
}
