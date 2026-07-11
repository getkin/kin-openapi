package openapi3

import "encoding/json"

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
