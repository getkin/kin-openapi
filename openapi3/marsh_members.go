package openapi3

import (
	"bytes"
	"encoding/json"
	"errors"
	"slices"
	"unicode/utf8"
)

var (
	errNotJSONObject = errors.New("not a JSON object")
	errFound         = errors.New("found")
)

// unmarshalExtensions decodes the members of data whose names are not in known.
func unmarshalExtensions(data []byte, known ...string) map[string]any {
	var extensions map[string]any
	_ = jsonMembers(data, func(name string, value []byte) error {
		if slices.Contains(known, name) {
			return nil
		}
		var v any
		if err := unmarshalJSON(value, &v); err != nil {
			return err
		}
		if extensions == nil {
			extensions = make(map[string]any)
		}
		extensions[name] = v
		return nil
	})
	return extensions
}

// unmarshalRef decodes data into ref only if data has a "$ref" member.
func unmarshalRef(data []byte, ref *Ref) error {
	found := false
	_ = jsonMembers(data, func(name string, _ []byte) error {
		if name == "$ref" {
			found = true
			return errFound
		}
		return nil
	})
	if !found {
		return nil
	}
	return unmarshalJSON(data, ref)
}

// jsonMemberIsNull reports whether the JSON object data has a member named
// name whose value is null.
func jsonMemberIsNull(data []byte, name string) bool {
	null := false
	_ = jsonMembers(data, func(n string, value []byte) error {
		if n == name {
			null = string(value) == "null"
		}
		return nil
	})
	return null
}

// jsonMembers calls f with the name and value of each member of the JSON object data.
func jsonMembers(data []byte, f func(name string, value []byte) error) error {
	i := skipJSONSpace(data, 0)
	if bytes.HasPrefix(data[i:], []byte("null")) {
		return nil
	}
	if i >= len(data) || data[i] != '{' {
		return errNotJSONObject
	}
	i = skipJSONSpace(data, i+1)
	if i < len(data) && data[i] == '}' {
		return nil
	}
	for i < len(data) {
		end := skipJSONValue(data, i)
		if end < 0 || data[i] != '"' {
			return errNotJSONObject
		}
		name, err := jsonMemberName(data[i:end])
		if err != nil {
			return err
		}
		if i = skipJSONSpace(data, end); i >= len(data) || data[i] != ':' {
			return errNotJSONObject
		}
		i = skipJSONSpace(data, i+1)
		if end = skipJSONValue(data, i); end < 0 {
			return errNotJSONObject
		}
		if err := f(name, data[i:end]); err != nil {
			return err
		}
		if i = skipJSONSpace(data, end); i >= len(data) {
			break
		}
		switch data[i] {
		case '}':
			return nil
		case ',':
			i = skipJSONSpace(data, i+1)
		default:
			return errNotJSONObject
		}
	}
	return errNotJSONObject
}

// jsonMemberName decodes the JSON string quoted.
func jsonMemberName(quoted []byte) (string, error) {
	if s := quoted[1 : len(quoted)-1]; bytes.IndexByte(s, '\\') < 0 && utf8.Valid(s) {
		return string(s), nil
	}
	var name string
	err := json.Unmarshal(quoted, &name)
	return name, err
}

func skipJSONSpace(data []byte, i int) int {
	for i < len(data) && (data[i] == ' ' || data[i] == '\t' || data[i] == '\n' || data[i] == '\r') {
		i++
	}
	return i
}

// skipJSONValue returns the end of the JSON value that starts at data[i], or
// -1 if it does not end.
func skipJSONValue(data []byte, i int) int {
	if i >= len(data) {
		return -1
	}
	switch data[i] {
	case '"':
		for i++; i < len(data); i++ {
			switch data[i] {
			case '\\':
				i++
			case '"':
				return i + 1
			}
		}
		return -1
	case '{', '[':
		depth := 0
		for ; i < len(data); i++ {
			switch data[i] {
			case '"':
				end := skipJSONValue(data, i)
				if end < 0 {
					return -1
				}
				i = end - 1
			case '{', '[':
				depth++
			case '}', ']':
				if depth--; depth == 0 {
					return i + 1
				}
			}
		}
		return -1
	default:
		for ; i < len(data); i++ {
			switch data[i] {
			case ',', '}', ']', ' ', '\t', '\n', '\r':
				return i
			}
		}
		return i
	}
}
