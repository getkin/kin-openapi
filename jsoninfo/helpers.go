package jsoninfo

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// DebugString returns a debugging-friendly JSON representation of the value.
//
// Uses indentation and disables HTML escaping.
// Panics if marshaling fails.
func DebugString(value interface{}) string {
	w := bytes.NewBuffer(make([]byte, 0, 255))
	e := json.NewEncoder(w)
	e.SetEscapeHTML(false)
	e.SetIndent("", "  ")
	err := e.Encode(value)
	if err != nil {
		panic(fmt.Errorf("Marshalling instance of type '%T' failed: %v", value, err))
	}
	return w.String()
}
