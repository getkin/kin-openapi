package jsoninfo

import (
	"encoding/json"
)

type StrictStruct interface {
	MarshalJSONUnsupportedFields(result map[string]json.RawMessage) error
	UnmarshalJSONUnsupportedFields(data []byte, unsupportedFields map[string]json.RawMessage) error
}
