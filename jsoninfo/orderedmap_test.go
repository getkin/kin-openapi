package jsoninfo

import (
	"reflect"
	"testing"
)

func TestExtractObjectKeys(t *testing.T) {
	const j = `{
		"foo": {"bar": 1},
		"baz": "qux",
		"quux": "quuz"
	}`

	keys, err := ExtractObjectKeys([]byte(j))
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(keys, []string{"foo", "baz", "quux"}) {
		t.Fatalf("expected %v, got %v", []string{"foo", "baz", "quux"}, keys)
	}
}
