package jsoninfo

import (
	"encoding/json"
)

type Extension interface {
}

type HasRawJSON interface {
	GetRawJSON() []byte
	SetRawJSON(value []byte)
}

type ExtensionInfo struct {
	MarshalFunc   []func(value interface{}, dest map[string]json.RawMessage) error
	UnmarshalFunc []func(value interface{}, raw []byte, unsupportedFields map[string]json.RawMessage) error
}
