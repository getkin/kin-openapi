// Package openapi3gen generates OpenAPIv3 JSON schemas from Go types.
package openapi3gen

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/jsoninfo"
	"github.com/getkin/kin-openapi/openapi3"
)

// CycleError indicates that a type graph has one or more possible cycles.
type CycleError struct{}

func (err *CycleError) Error() string { return "detected cycle" }

// Option allows tweaking SchemaRef generation
type Option func(*generatorOpt)

// SchemaCustomizerFn is a callback function, allowing
// the OpenAPI schema definition to be updated with additional
// properties during the generation process, based on the
// name of the field, the Go type, and the struct tags.
// name will be "_root" for the top level object, and tag will be ""
type SchemaCustomizerFn func(name string, t reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error

type generatorOpt struct {
	useAllExportedFields bool
	throwErrorOnCycle    bool
	schemaCustomizer     SchemaCustomizerFn
}

// UseAllExportedFields changes the default behavior of only
// generating schemas for struct fields with a JSON tag.
func UseAllExportedFields() Option {
	return func(x *generatorOpt) { x.useAllExportedFields = true }
}

// ThrowErrorOnCycle changes the default behavior of creating cycle
// refs to instead error if a cycle is detected.
func ThrowErrorOnCycle() Option {
	return func(x *generatorOpt) { x.throwErrorOnCycle = true }
}

// SchemaCustomizer allows customization of the schema that is generated
// for a field, for example to support an additional tagging scheme
func SchemaCustomizer(sc SchemaCustomizerFn) Option {
	return func(x *generatorOpt) { x.schemaCustomizer = sc }
}

// NewSchemaRefForValue uses reflection on the given value to produce a SchemaRef.
func NewSchemaRefForValue(value interface{}, opts ...Option) (*openapi3.SchemaRef, map[*openapi3.SchemaRef]int, error) {
	g := NewGenerator(opts...)
	ref, err := g.GenerateSchemaRef(reflect.TypeOf(value))
	for ref := range g.SchemaRefs {
		ref.Ref = ""
	}
	return ref, g.SchemaRefs, err
}

type Generator struct {
	opts generatorOpt

	Types map[reflect.Type]*openapi3.SchemaRef

	// SchemaRefs contains all references and their counts.
	// If count is 1, it's not ne
	// An OpenAPI identifier has been assigned to each.
	SchemaRefs map[*openapi3.SchemaRef]int
}

func NewGenerator(opts ...Option) *Generator {
	gOpt := &generatorOpt{}
	for _, f := range opts {
		f(gOpt)
	}
	return &Generator{
		Types:      make(map[reflect.Type]*openapi3.SchemaRef),
		SchemaRefs: make(map[*openapi3.SchemaRef]int),
		opts:       *gOpt,
	}
}

func (g *Generator) GenerateSchemaRef(t reflect.Type) (*openapi3.SchemaRef, error) {
	//check generatorOpt consistency here
	return g.generateSchemaRefFor(nil, t, "_root", "")
}

func (g *Generator) generateSchemaRefFor(parents []*jsoninfo.TypeInfo, t reflect.Type, name string, tag reflect.StructTag) (*openapi3.SchemaRef, error) {
	if ref := g.Types[t]; ref != nil && g.opts.schemaCustomizer == nil {
		g.SchemaRefs[ref]++
		return ref, nil
	}
	ref, err := g.generateWithoutSaving(parents, t, name, tag)
	if ref != nil {
		g.Types[t] = ref
		g.SchemaRefs[ref]++
	}
	return ref, err
}

func (g *Generator) generateWithoutSaving(parents []*jsoninfo.TypeInfo, t reflect.Type, name string, tag reflect.StructTag) (*openapi3.SchemaRef, error) {
	typeInfo := jsoninfo.GetTypeInfo(t)
	for _, parent := range parents {
		if parent == typeInfo {
			return nil, &CycleError{}
		}
	}

	if cap(parents) == 0 {
		parents = make([]*jsoninfo.TypeInfo, 0, 4)
	}
	parents = append(parents, typeInfo)

	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if strings.HasSuffix(t.Name(), "Ref") {
		_, a := t.FieldByName("Ref")
		v, b := t.FieldByName("Value")
		if a && b {
			vs, err := g.generateSchemaRefFor(parents, v.Type, name, tag)
			if err != nil {
				if _, ok := err.(*CycleError); ok && !g.opts.throwErrorOnCycle {
					g.SchemaRefs[vs]++
					return vs, nil
				}
				return nil, err
			}
			refSchemaRef := RefSchemaRef
			g.SchemaRefs[refSchemaRef]++
			ref := openapi3.NewSchemaRef(t.Name(), &openapi3.Schema{
				OneOf: []*openapi3.SchemaRef{
					refSchemaRef,
					vs,
				},
			})
			g.SchemaRefs[ref]++
			return ref, nil
		}
	}

	schema := &openapi3.Schema{}

	switch t.Kind() {
	case reflect.Func, reflect.Chan:
		return nil, nil // ignore

	case reflect.Bool:
		schema.Type = "boolean"

	case reflect.Int:
		schema.Type = "integer"
	case reflect.Int8:
		schema.Type = "integer"
		schema.Min = &minInt8
		schema.Max = &maxInt8
	case reflect.Int16:
		schema.Type = "integer"
		schema.Min = &minInt16
		schema.Max = &maxInt16
	case reflect.Int32:
		schema.Type = "integer"
		schema.Format = "int32"
	case reflect.Int64:
		schema.Type = "integer"
		schema.Format = "int64"
	case reflect.Uint:
		schema.Type = "integer"
		schema.Min = &zeroInt
	case reflect.Uint8:
		schema.Type = "integer"
		schema.Min = &zeroInt
		schema.Max = &maxUint8
	case reflect.Uint16:
		schema.Type = "integer"
		schema.Min = &zeroInt
		schema.Max = &maxUint16
	case reflect.Uint32:
		schema.Type = "integer"
		schema.Min = &zeroInt
		schema.Max = &maxUint32
	case reflect.Uint64:
		schema.Type = "integer"
		schema.Min = &zeroInt
		schema.Max = &maxUint64

	case reflect.Float32:
		schema.Type = "number"
		schema.Format = "float"
	case reflect.Float64:
		schema.Type = "number"
		schema.Format = "double"

	case reflect.String:
		schema.Type = "string"

	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			if t == rawMessageType {
				return &openapi3.SchemaRef{Value: schema}, nil
			}
			schema.Type = "string"
			schema.Format = "byte"
		} else {
			schema.Type = "array"
			items, err := g.generateSchemaRefFor(parents, t.Elem(), name, tag)
			if err != nil {
				if _, ok := err.(*CycleError); ok && !g.opts.throwErrorOnCycle {
					items = g.generateCycleSchemaRef(t.Elem(), schema)
				} else {
					return nil, err
				}
			}
			if items != nil {
				g.SchemaRefs[items]++
				schema.Items = items
			}
		}

	case reflect.Map:
		schema.Type = "object"
		additionalProperties, err := g.generateSchemaRefFor(parents, t.Elem(), name, tag)
		if err != nil {
			if _, ok := err.(*CycleError); ok && !g.opts.throwErrorOnCycle {
				additionalProperties = g.generateCycleSchemaRef(t.Elem(), schema)
			} else {
				return nil, err
			}
		}
		if additionalProperties != nil {
			g.SchemaRefs[additionalProperties]++
			schema.AdditionalProperties = additionalProperties
		}

	case reflect.Struct:
		if t == timeType {
			schema.Type = "string"
			schema.Format = "date-time"
		} else {
			for _, fieldInfo := range typeInfo.Fields {
				// Only fields with JSON tag are considered (by default)
				if !fieldInfo.HasJSONTag && !g.opts.useAllExportedFields {
					continue
				}
				// If asked, try to use yaml tag
				fieldName, fType := fieldInfo.JSONName, fieldInfo.Type
				if !fieldInfo.HasJSONTag && g.opts.useAllExportedFields {
					// Handle anonymous fields/embedded structs
					if t.Field(fieldInfo.Index[0]).Anonymous {
						ref, err := g.generateSchemaRefFor(parents, fType, fieldName, tag)
						if err != nil {
							if _, ok := err.(*CycleError); ok && !g.opts.throwErrorOnCycle {
								ref = g.generateCycleSchemaRef(fType, schema)
							} else {
								return nil, err
							}
						}
						if ref != nil {
							g.SchemaRefs[ref]++
							schema.WithPropertyRef(fieldName, ref)
						}
					} else {
						ff := t.Field(fieldInfo.Index[len(fieldInfo.Index)-1])
						if tag, ok := ff.Tag.Lookup("yaml"); ok && tag != "-" {
							fieldName, fType = tag, ff.Type
						}
					}
				}

				// extract the field tag if we have a customizer
				var fieldTag reflect.StructTag
				if g.opts.schemaCustomizer != nil {
					ff := t.Field(fieldInfo.Index[len(fieldInfo.Index)-1])
					fieldTag = ff.Tag
				}

				ref, err := g.generateSchemaRefFor(parents, fType, fieldName, fieldTag)
				if err != nil {
					if _, ok := err.(*CycleError); ok && !g.opts.throwErrorOnCycle {
						ref = g.generateCycleSchemaRef(fType, schema)
					} else {
						return nil, err
					}
				}
				if ref != nil {
					g.SchemaRefs[ref]++
					schema.WithPropertyRef(fieldName, ref)
				}
			}

			// Object only if it has properties
			if schema.Properties != nil {
				schema.Type = "object"
			}
		}
	}

	if g.opts.schemaCustomizer != nil {
		if err := g.opts.schemaCustomizer(name, t, tag, schema); err != nil {
			return nil, err
		}
	}

	return openapi3.NewSchemaRef(t.Name(), schema), nil
}

func (g *Generator) generateCycleSchemaRef(t reflect.Type, schema *openapi3.Schema) *openapi3.SchemaRef {
	var typeName string
	switch t.Kind() {
	case reflect.Ptr:
		return g.generateCycleSchemaRef(t.Elem(), schema)
	case reflect.Slice:
		ref := g.generateCycleSchemaRef(t.Elem(), schema)
		sliceSchema := openapi3.NewSchema()
		sliceSchema.Type = "array"
		sliceSchema.Items = ref
		return openapi3.NewSchemaRef("", sliceSchema)
	case reflect.Map:
		ref := g.generateCycleSchemaRef(t.Elem(), schema)
		mapSchema := openapi3.NewSchema()
		mapSchema.Type = "object"
		mapSchema.AdditionalProperties = ref
		return openapi3.NewSchemaRef("", mapSchema)
	default:
		typeName = t.Name()
	}

	return openapi3.NewSchemaRef(fmt.Sprintf("#/components/schemas/%s", typeName), schema)
}

var RefSchemaRef = openapi3.NewSchemaRef("Ref",
	openapi3.NewObjectSchema().WithProperty("$ref", openapi3.NewStringSchema().WithMinLength(1)))

var (
	timeType       = reflect.TypeOf(time.Time{})
	rawMessageType = reflect.TypeOf(json.RawMessage{})

	zeroInt   = float64(0)
	maxInt8   = float64(math.MaxInt8)
	minInt8   = float64(math.MinInt8)
	maxInt16  = float64(math.MaxInt16)
	minInt16  = float64(math.MinInt16)
	maxUint8  = float64(math.MaxUint8)
	maxUint16 = float64(math.MaxUint16)
	maxUint32 = float64(math.MaxUint32)
	maxUint64 = float64(math.MaxUint64)
)
