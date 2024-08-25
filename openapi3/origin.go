package openapi3

import "encoding/json"

const originKey = "origin"

type Origin struct {
	Key    *Location           `json:"key,omitempty" yaml:"key,omitempty"`
	Fields map[string]Location `json:"fields,omitempty" yaml:"fields,omitempty"`
}

type Location struct {
	Line   int `json:"line,omitempty" yaml:"line,omitempty"`
	Column int `json:"column,omitempty" yaml:"column,omitempty"`
}

// unmarshalStringMapP unmarshals given json into a map[string]*V
func unmarshalStringMapP[V any](data []byte) (map[string]*V, error) {
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	// TODO: add origin to the result
	delete(m, originKey)

	result := make(map[string]*V, len(m))
	for k, v := range m {
		if data, err := json.Marshal(v); err != nil {
			return nil, err
		} else {
			var v V
			if err = json.Unmarshal(data, &v); err != nil {
				return nil, err
			}
			result[k] = &v
		}
	}

	return result, nil
}

// unmarshalStringMap unmarshals given json into a map[string]V
func unmarshalStringMap[V any](data []byte) (map[string]V, error) {
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	// TODO: add origin to the result
	delete(m, originKey)

	result := make(map[string]V, len(m))
	for k, v := range m {
		if data, err := json.Marshal(v); err != nil {
			return nil, err
		} else {
			var v V
			if err = json.Unmarshal(data, &v); err != nil {
				return nil, err
			}
			result[k] = v
		}
	}

	return result, nil
}
