// +build xeipuuv

package openapi3

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

// SchemaFormatValidationDisabled FIXME: drop
var SchemaFormatValidationDisabled = true

type schemaLoader = *gojsonschema.SchemaLoader

func (doc *T) compileSchemas(settings *schemaValidationSettings) (err error) {
	docSchemas := doc.Components.Schemas
	schemas := make(schemasJSON, len(docSchemas))
	for name, docSchema := range docSchemas {
		schemas[name] = docSchema.Value.fromOpenAPISchema(settings)
	}
	//FIXME merge loops
	refd := gojsonschema.NewSchemaLoader()
	for name, schema := range schemas {
		absRef := "#/components/schemas/" + name
		sl := gojsonschema.NewGoLoader(schema)
		if err = refd.AddSchema(absRef, sl); err != nil {
			return
		}
	}

	switch {
	case settings.asreq:
		doc.refdAsReq = refd
	case settings.asrep:
		doc.refdAsRep = refd
	default:
		doc.refd = refd
	}
	return
}

func (schema *Schema) visitData(doc *T, data interface{}, opts ...SchemaValidationOption) (err error) {
	settings := newSchemaValidationSettings(opts...)
	ls := gojsonschema.NewGoLoader(schema.fromOpenAPISchema(settings))
	ld := gojsonschema.NewGoLoader(data)

	var res *gojsonschema.Result
	if doc != nil {
		if doc.refdAsReq == nil || doc.refdAsRep == nil || doc.refd == nil {
			panic(`func (*T) CompileSchemas() error must be called first`)
		}
		var whole *gojsonschema.Schema
		switch {
		case settings.asreq:
			whole, err = doc.refdAsReq.Compile(ls)
		case settings.asrep:
			whole, err = doc.refdAsRep.Compile(ls)
		default:
			whole, err = doc.refd.Compile(ls)
		}
		if err != nil {
			return
		}
		res, err = whole.Validate(ld)
	} else {
		res, err = gojsonschema.Validate(ls, ld)
	}
	if err != nil {
		return
	}

	if !res.Valid() {
		err := SchemaValidationError(res.Errors())
		if settings.multiError {
			return err.asMultiError()
		}
		return err
	}
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

// SchemaValidationError is a collection of errors
type SchemaValidationError []gojsonschema.ResultError

var _ error = (*SchemaValidationError)(nil)

func (e SchemaValidationError) Error() string {
	var buff strings.Builder
	for i, re := range []gojsonschema.ResultError(e) {
		if i != 0 {
			buff.WriteString("\n")
		}
		buff.WriteString(re.String())
	}
	return buff.String()
}

// Errors unwraps into much detailed errors.
// See https://pkg.go.dev/github.com/xeipuuv/gojsonschema#ResultError
func (e SchemaValidationError) Errors() []gojsonschema.ResultError {
	return e
}

// JSONPointer returns a dot (.) delimited "JSON path" to the context of the first error.
func (e SchemaValidationError) JSONPointer() string {
	return []gojsonschema.ResultError(e)[0].Field()
}

func (e SchemaValidationError) asMultiError() MultiError {
	errs := make([]error, 0, len(e))
	for _, re := range e {
		errs = append(errs, errors.New(re.String()))
	}
	return errs
}

type SchemaError struct {
	Value       interface{}
	reversePath []string //FIXME
	Schema      *Schema
	SchemaField string
	Reason      string
	Origin      error //FIXME
}

func (err *SchemaError) JSONPointer() []string {
	return nil //FIXME
}

func (err *SchemaError) Error() string {
	// if err.Origin != nil {
	// 	return err.Origin.Error()
	// }

	buf := bytes.NewBuffer(make([]byte, 0, 256))
	// if len(err.reversePath) > 0 {
	// 	buf.WriteString(`Error at "`)
	// 	reversePath := err.reversePath
	// 	for i := len(reversePath) - 1; i >= 0; i-- {
	// 		buf.WriteByte('/')
	// 		buf.WriteString(reversePath[i])
	// 	}
	// 	buf.WriteString(`": `)
	// }
	reason := err.Reason
	if reason == "" {
		buf.WriteString(`Doesn't match schema "`)
		buf.WriteString(err.SchemaField)
		buf.WriteString(`"`)
	} else {
		buf.WriteString(reason)
	}
	{ // if !SchemaErrorDetailsDisabled {
		buf.WriteString("\nSchema:\n  ")
		encoder := json.NewEncoder(buf)
		encoder.SetIndent("  ", "  ")
		if err := encoder.Encode(err.Schema); err != nil {
			panic(err)
		}
		buf.WriteString("\nValue:\n  ")
		if err := encoder.Encode(err.Value); err != nil {
			panic(err)
		}
	}
	return buf.String()
}
