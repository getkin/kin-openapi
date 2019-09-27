package openapi3

import (
	"context"
	"fmt"

	"github.com/getkin/kin-openapi/jsoninfo"
)

type CallbackRef struct {
	Ref   string
	Value *Callback
}

func (value CallbackRef) String() string {
	return fmt.Sprintf("%s:CallbackRef", value.Ref)
}

func (value *CallbackRef) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalRef(value.Ref, value.Value)
}

func (value *CallbackRef) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalRef(data, &value.Ref, &value.Value)
}

func (value *CallbackRef) Validate(c context.Context) error {
	v := value.Value
	if v == nil {
		return foundUnresolvedRef(value.Ref)
	}
	return v.Validate(c)
}

type ExampleRef struct {
	Ref   string
	Value *Example
}

func (value ExampleRef) String() string {
	return fmt.Sprintf("%s:ExampleRef", value.Ref)
}

func (value ExampleRef) Empty() bool {
	return value.Ref == ""
}

func (value *ExampleRef) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalRef(value.Ref, value.Value)
}

func (value *ExampleRef) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalRef(data, &value.Ref, &value.Value)
}

func (value *ExampleRef) Validate(c context.Context) error {
	return nil
}

type HeaderRef struct {
	Ref   string
	Value *Header
}

func (value HeaderRef) String() string {
	return fmt.Sprintf("%s:HeaderRef", value.Ref)
}

func (value HeaderRef) Empty() bool {
	return value.Ref == ""
}

func (value *HeaderRef) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalRef(value.Ref, value.Value)
}

func (value *HeaderRef) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalRef(data, &value.Ref, &value.Value)
}

func (value *HeaderRef) Validate(c context.Context) error {
	v := value.Value
	if v == nil {
		return foundUnresolvedRef(value.Ref)
	}
	return v.Validate(c)
}

type LinkRef struct {
	Ref   string
	Value *Link
}

func (value LinkRef) String() string {
	return fmt.Sprintf("%s:LinkRef", value.Ref)
}

func (value LinkRef) Empty() bool {
	return value.Ref == ""
}

func (value *LinkRef) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalRef(value.Ref, value.Value)
}

func (value *LinkRef) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalRef(data, &value.Ref, &value.Value)
}

func (value *LinkRef) Validate(c context.Context) error {
	v := value.Value
	if v == nil {
		return foundUnresolvedRef(value.Ref)
	}
	return v.Validate(c)
}

type ParameterRef struct {
	Ref   string
	Value *Parameter
}

func (value ParameterRef) String() string {
	return fmt.Sprintf("%s:ParameterRef", value.Ref)
}

func (value ParameterRef) Empty() bool {
	return value.Ref == ""
}

func (value *ParameterRef) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalRef(value.Ref, value.Value)
}

func (value *ParameterRef) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalRef(data, &value.Ref, &value.Value)
}

func (value *ParameterRef) Validate(c context.Context) error {
	v := value.Value
	if v == nil {
		return foundUnresolvedRef(value.Ref)
	}
	return v.Validate(c)
}

type ResponseRef struct {
	Ref   string
	Value *Response
}

func (value ResponseRef) String() string {
	return fmt.Sprintf("%s:ResponseRef", value.Ref)
}

func (value ResponseRef) Empty() bool {
	return value.Ref == ""
}

func (value *ResponseRef) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalRef(value.Ref, value.Value)
}

func (value *ResponseRef) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalRef(data, &value.Ref, &value.Value)
}

func (value *ResponseRef) Validate(c context.Context) error {
	v := value.Value
	if v == nil {
		return foundUnresolvedRef(value.Ref)
	}
	return v.Validate(c)
}

type RequestBodyRef struct {
	Ref   string
	Value *RequestBody
}

func (value RequestBodyRef) String() string {
	return fmt.Sprintf("%s:RequestBodyRef", value.Ref)
}

func (value RequestBodyRef) Empty() bool {
	return value.Ref == ""
}

func (value *RequestBodyRef) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalRef(value.Ref, value.Value)
}

func (value *RequestBodyRef) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalRef(data, &value.Ref, &value.Value)
}

func (value *RequestBodyRef) Validate(c context.Context) error {
	v := value.Value
	if v == nil {
		return foundUnresolvedRef(value.Ref)
	}
	return v.Validate(c)
}

type SchemaRef struct {
	Ref   string
	Value *Schema
}

func (value SchemaRef) String() string {
	return fmt.Sprintf("%s:SchemaRef", value.Ref)
}

func (value SchemaRef) Empty() bool {
	return value.Ref == ""
}

func NewSchemaRef(ref string, value *Schema) *SchemaRef {
	return &SchemaRef{
		Ref:   ref,
		Value: value,
	}
}

func (value *SchemaRef) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalRef(value.Ref, value.Value)
}

func (value *SchemaRef) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalRef(data, &value.Ref, &value.Value)
}

func (value *SchemaRef) Validate(c context.Context) error {
	v := value.Value
	if v == nil {
		return foundUnresolvedRef(value.Ref)
	}
	return v.Validate(c)
}

type SecuritySchemeRef struct {
	Ref   string
	Value *SecurityScheme
}

func (value SecuritySchemeRef) String() string {
	return fmt.Sprintf("%s:SecuritySchemeRef", value.Ref)
}

func (value SecuritySchemeRef) Empty() bool {
	return value.Ref == ""
}

func (value *SecuritySchemeRef) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalRef(value.Ref, value.Value)
}

func (value *SecuritySchemeRef) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalRef(data, &value.Ref, &value.Value)
}

func (value *SecuritySchemeRef) Validate(c context.Context) error {
	v := value.Value
	if v == nil {
		return foundUnresolvedRef(value.Ref)
	}
	return v.Validate(c)
}
