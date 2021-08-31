// +build santhoshtekuri

package openapi3

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v4"
)

// // SchemaFormatValidationDisabled FIXME: drop
// var SchemaFormatValidationDisabled = true

// type schemaLoader = *gojsonschema.SchemaLoader

func (doc *T) compileSchemas(settings *schemaValidationSettings) (err error) {
	return
}

func (schema *Schema) visitData(doc *T, data interface{}, opts ...SchemaValidationOption) (err error) {
	return
}

func (schema *Schema) validate(ctx context.Context, stack []*Schema) (err error) {
	for _, existing := range stack {
		if existing == schema {
			return
		}
	}
	stack = append(stack, schema)

	if schema.ReadOnly && schema.WriteOnly {
		return errors.New("a property MUST NOT be marked as both readOnly and writeOnly being true")
	}

	for _, item := range schema.OneOf {
		v := item.Value
		if v == nil {
			return foundUnresolvedRef(item.Ref)
		}
		if err = v.validate(ctx, stack); err == nil {
			return
		}
	}

	for _, item := range schema.AnyOf {
		v := item.Value
		if v == nil {
			return foundUnresolvedRef(item.Ref)
		}
		if err = v.validate(ctx, stack); err != nil {
			return
		}
	}

	for _, item := range schema.AllOf {
		v := item.Value
		if v == nil {
			return foundUnresolvedRef(item.Ref)
		}
		if err = v.validate(ctx, stack); err != nil {
			return
		}
	}

	if ref := schema.Not; ref != nil {
		v := ref.Value
		if v == nil {
			return foundUnresolvedRef(ref.Ref)
		}
		if err = v.validate(ctx, stack); err != nil {
			return
		}
	}

	schemaType := schema.Type
	// NOTE: any format is valid, as per:
	// > However, to support documentation needs, the format property is an open string-valued property, and can have any value.
	switch schemaType {
	case "":
	case "boolean":
	case "number":
	case "integer":
	case "string":
	case "array":
		if schema.Items == nil {
			return errors.New("when schema type is 'array', schema 'items' must be non-null")
		}
	case "object":
	default:
		return fmt.Errorf("unsupported 'type' value %q", schemaType)
	}

	if pattern := schema.Pattern; pattern != "" {
		if _, err = regexp.Compile(pattern); err != nil {
			return &SchemaError{
				Schema:      schema,
				SchemaField: "pattern",
				Reason:      fmt.Sprintf("cannot compile pattern %q: %v", pattern, err),
			}
		}
	}

	if ref := schema.Items; ref != nil {
		v := ref.Value
		if v == nil {
			return foundUnresolvedRef(ref.Ref)
		}
		if err = v.validate(ctx, stack); err != nil {
			return
		}
	}

	for _, ref := range schema.Properties {
		v := ref.Value
		if v == nil {
			return foundUnresolvedRef(ref.Ref)
		}
		if err = v.validate(ctx, stack); err != nil {
			return
		}
	}

	if ref := schema.AdditionalProperties; ref != nil {
		v := ref.Value
		if v == nil {
			return foundUnresolvedRef(ref.Ref)
		}
		if err = v.validate(ctx, stack); err != nil {
			return
		}
	}

	return
}
