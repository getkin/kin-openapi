package openapi3

import "encoding/json"

// StringMap is a generic map[string]V with a custom UnmarshalJSON.
type StringMap[V any] map[string]V

// UnmarshalJSON sets StringMap to a copy of data.
func (stringMap *StringMap[V]) UnmarshalJSON(data []byte) (err error) {
	*stringMap, err = unmarshalStringMap[V](data)
	return
}

// unmarshalStringMapP unmarshals given json into a map[string]*V.
func unmarshalStringMapP[V any](data []byte) (map[string]*V, error) {
	var m map[string]*V
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// unmarshalStringMap unmarshals given json into a map[string]V.
func unmarshalStringMap[V any](data []byte) (map[string]V, error) {
	var m map[string]V
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}
