package openapi3

import (
	"context"
	"github.com/jban332/kin-openapi/jsoninfo"
)

type CallbackRef struct {
	Ref   string
	Value *Callback
}

func (value *CallbackRef) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalRef(value.Ref, value.Value)
}

func (value *CallbackRef) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalRef(data, &value.Ref, &value.Value)
}

func (value *CallbackRef) Validate(c context.Context) error {
	if v := value.Value; v != nil {
		return v.Validate(c)
	}
	return nil
}

type ExampleRef struct {
	Ref   string
	Value *interface{}
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

func (value *HeaderRef) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalRef(value.Ref, value.Value)
}

func (value *HeaderRef) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalRef(data, &value.Ref, &value.Value)
}

func (value *HeaderRef) Validate(c context.Context) error {
	if v := value.Value; v != nil {
		return v.Validate(c)
	}
	return nil
}

type LinkRef struct {
	Ref   string
	Value *Link
}

func (value *LinkRef) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalRef(value.Ref, value.Value)
}

func (value *LinkRef) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalRef(data, &value.Ref, &value.Value)
}

func (value *LinkRef) Validate(c context.Context) error {
	if v := value.Value; v != nil {
		return v.Validate(c)
	}
	return nil
}

type ParameterRef struct {
	Ref   string
	Value *Parameter
}

func (value *ParameterRef) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalRef(value.Ref, value.Value)
}

func (value *ParameterRef) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalRef(data, &value.Ref, &value.Value)
}

func (value *ParameterRef) Validate(c context.Context) error {
	if v := value.Value; v != nil {
		return v.Validate(c)
	}
	return nil
}

type ResponseRef struct {
	Ref   string
	Value *Response
}

func (value *ResponseRef) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalRef(value.Ref, value.Value)
}

func (value *ResponseRef) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalRef(data, &value.Ref, &value.Value)
}

func (value *ResponseRef) Validate(c context.Context) error {
	if v := value.Value; v != nil {
		return v.Validate(c)
	}
	return nil
}

type RequestBodyRef struct {
	Ref   string
	Value *RequestBody
}

func (value *RequestBodyRef) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalRef(value.Ref, value.Value)
}

func (value *RequestBodyRef) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalRef(data, &value.Ref, &value.Value)
}

func (value *RequestBodyRef) Validate(c context.Context) error {
	if v := value.Value; v != nil {
		return v.Validate(c)
	}
	return nil
}

type SchemaRef struct {
	Ref   string
	Value *Schema
}

func (value *SchemaRef) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalRef(value.Ref, value.Value)
}

func (value *SchemaRef) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalRef(data, &value.Ref, &value.Value)
}

func (value *SchemaRef) Validate(c context.Context) error {
	if v := value.Value; v != nil {
		return v.Validate(c)
	}
	return nil
}

type SecuritySchemeRef struct {
	Ref   string
	Value *SecurityScheme
}

func (value *SecuritySchemeRef) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalRef(value.Ref, value.Value)
}

func (value *SecuritySchemeRef) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalRef(data, &value.Ref, &value.Value)
}

func (value *SecuritySchemeRef) Validate(c context.Context) error {
	if v := value.Value; v != nil {
		return v.Validate(c)
	}
	return nil
}
