// Package openapi3gen generates OpenAPIv3 JSON schemas from Go types.
package openapi3gen

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

// CycleError indicates that a type graph has one or more possible cycles.
type CycleError struct{}

func (err *CycleError) Error() string { return "detected cycle" }

// ExcludeSchemaSentinel indicates that the schema for a specific field should not be included in the final output.
type ExcludeSchemaSentinel struct{}

func (err *ExcludeSchemaSentinel) Error() string { return "schema excluded" }

// Option allows tweaking SchemaRef generation
type Option func(*generatorOpt)

// SchemaCustomizerFn is a callback function, allowing
// the OpenAPI schema definition to be updated with additional
// properties during the generation process, based on the
// name of the field, the Go type, and the struct tags.
// name will be "_root" for the top level object, and tag will be "".
// A SchemaCustomizerFn can return an ExcludeSchemaSentinel error to
// indicate that the schema for this field should not be included in
// the final output
type SchemaCustomizerFn func(name string, t reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error

// SetSchemar allows client to set their own schema definition according to
// their specification. Useful when some custom datatype is needed and/or some custom logic
// is needed on how the schema values would be generated
type SetSchemar interface {
	SetSchema(*openapi3.Schema)
}

type ExportComponentSchemasOptions struct {
	ExportComponentSchemas bool
	ExportTopLevelSchema   bool
	ExportGenerics         bool
}

type TypeNameGenerator func(t reflect.Type) string

type generatorOpt struct {
	useAllExportedFields   bool
	throwErrorOnCycle      bool
	schemaCustomizer       SchemaCustomizerFn
	exportComponentSchemas ExportComponentSchemasOptions
	typeNameGenerator      TypeNameGenerator
}

// UseAllExportedFields changes the default behavior of only
// generating schemas for struct fields with a JSON tag.
func UseAllExportedFields() Option {
	return func(x *generatorOpt) { x.useAllExportedFields = true }
}

func CreateTypeNameGenerator(tngnrt TypeNameGenerator) Option {
	return func(x *generatorOpt) { x.typeNameGenerator = tngnrt }
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

// CreateComponents changes the default behavior
// to add all schemas as components
// Reduces duplicate schemas in routes
func CreateComponentSchemas(exso ExportComponentSchemasOptions) Option {
	return func(x *generatorOpt) { x.exportComponentSchemas = exso }
}

// NewSchemaRefForValue is a shortcut for NewGenerator(...).NewSchemaRefForValue(...)
func NewSchemaRefForValue(value any, schemas openapi3.Schemas, opts ...Option) (*openapi3.SchemaRef, error) {
	g := NewGenerator(opts...)
	return g.NewSchemaRefForValue(value, schemas)
}

type Generator struct {
	opts generatorOpt

	Types map[reflect.Type]*openapi3.SchemaRef

	// SchemaRefs contains all references and their counts.
	// If count is 1, it's not ne
	// An OpenAPI identifier has been assigned to each.
	SchemaRefs map[*openapi3.SchemaRef]int

	// componentSchemaRefs is a set of schemas that must be defined in the components to avoid cycles
	// or if we have specified create components schemas
	componentSchemaRefs map[string]struct{}
}

func NewGenerator(opts ...Option) *Generator {
	gOpt := &generatorOpt{}
	for _, f := range opts {
		f(gOpt)
	}
	return &Generator{
		Types:               make(map[reflect.Type]*openapi3.SchemaRef),
		SchemaRefs:          make(map[*openapi3.SchemaRef]int),
		componentSchemaRefs: make(map[string]struct{}),
		opts:                *gOpt,
	}
}

func (g *Generator) GenerateSchemaRef(t reflect.Type) (*openapi3.SchemaRef, error) {
	//check generatorOpt consistency here
	return g.generateSchemaRefFor(nil, t, "_root", "")
}

// NewSchemaRefForValue uses reflection on the given value to produce a SchemaRef, and updates a supplied map with any dependent component schemas if they lead to cycles
func (g *Generator) NewSchemaRefForValue(value any, schemas openapi3.Schemas) (*openapi3.SchemaRef, error) {
	ref, err := g.GenerateSchemaRef(reflect.TypeOf(value))
	if err != nil {
		return nil, err
	}
	for ref := range g.SchemaRefs {
		refName := ref.Ref
		if g.opts.exportComponentSchemas.ExportComponentSchemas && strings.HasPrefix(refName, "#/components/schemas/") {
			refName = strings.TrimPrefix(refName, "#/components/schemas/")
		}

		if _, ok := g.componentSchemaRefs[refName]; ok && schemas != nil {
			if ref.Value != nil && ref.Value.Properties != nil {
				schemas[refName] = &openapi3.SchemaRef{
					Value: ref.Value,
				}
			}
		}
		if strings.HasPrefix(ref.Ref, "#/components/schemas/") {
			ref.Value = nil
		} else {
			ref.Ref = ""
		}
	}
	return ref, nil
}

func (g *Generator) generateSchemaRefFor(parents []*theTypeInfo, t reflect.Type, name string, tag reflect.StructTag) (*openapi3.SchemaRef, error) {
	if ref := g.Types[t]; ref != nil && g.opts.schemaCustomizer == nil {
		g.SchemaRefs[ref]++
		return ref, nil
	}
	ref, err := g.generateWithoutSaving(parents, t, name, tag)
	if _, ok := err.(*ExcludeSchemaSentinel); ok {
		// This schema should not be included in the final output
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if ref != nil {
		g.Types[t] = ref
		g.SchemaRefs[ref]++
	}
	return ref, nil
}

func getStructField(t reflect.Type, fieldInfo theFieldInfo) reflect.StructField {
	var ff reflect.StructField
	// fieldInfo.Index is an array of indexes starting from the root of the type
	for i := 0; i < len(fieldInfo.Index); i++ {
		ff = t.Field(fieldInfo.Index[i])
		t = ff.Type
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
	}
	return ff
}

func (g *Generator) generateWithoutSaving(parents []*theTypeInfo, t reflect.Type, name string, tag reflect.StructTag) (*openapi3.SchemaRef, error) {
	typeInfo := getTypeInfo(t)
	for _, parent := range parents {
		if parent == typeInfo {
			return nil, &CycleError{}
		}
	}
	isRoot := cap(parents) == 0
	if isRoot {
		parents = make([]*theTypeInfo, 0, 4)
	}
	parents = append(parents, typeInfo)

	isNullable := false
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
		isNullable = !isRoot
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
	schema.Nullable = isNullable

	switch t.Kind() {
	case reflect.Func, reflect.Chan:
		return nil, nil // ignore

	case reflect.Bool:
		schema.Type = &openapi3.Types{"boolean"}

	case reflect.Int:
		schema.Type = &openapi3.Types{"integer"}
	case reflect.Int8:
		schema.Type = &openapi3.Types{"integer"}
		schema.Min = &minInt8
		schema.Max = &maxInt8
	case reflect.Int16:
		schema.Type = &openapi3.Types{"integer"}
		schema.Min = &minInt16
		schema.Max = &maxInt16
	case reflect.Int32:
		schema.Type = &openapi3.Types{"integer"}
		schema.Format = "int32"
	case reflect.Int64:
		schema.Type = &openapi3.Types{"integer"}
		schema.Format = "int64"
	case reflect.Uint:
		schema.Type = &openapi3.Types{"integer"}
		schema.Min = &zeroInt
	case reflect.Uint8:
		schema.Type = &openapi3.Types{"integer"}
		schema.Min = &zeroInt
		schema.Max = &maxUint8
	case reflect.Uint16:
		schema.Type = &openapi3.Types{"integer"}
		schema.Min = &zeroInt
		schema.Max = &maxUint16
	case reflect.Uint32:
		schema.Type = &openapi3.Types{"integer"}
		schema.Min = &zeroInt
		schema.Max = &maxUint32
	case reflect.Uint64:
		schema.Type = &openapi3.Types{"integer"}
		schema.Min = &zeroInt
		schema.Max = &maxUint64

	case reflect.Float32:
		schema.Type = &openapi3.Types{"number"}
		schema.Format = "float"
	case reflect.Float64:
		schema.Type = &openapi3.Types{"number"}
		schema.Format = "double"

	case reflect.String:
		schema.Type = &openapi3.Types{"string"}

	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			if t == rawMessageType {
				return &openapi3.SchemaRef{Value: schema}, nil
			}
			schema.Type = &openapi3.Types{"string"}
			schema.Format = "byte"
		} else {
			schema.Type = &openapi3.Types{"array"}
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
		schema.Type = &openapi3.Types{"object"}
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
			schema.AdditionalProperties = openapi3.AdditionalProperties{Schema: additionalProperties}
		}

	case reflect.Struct:
		if t == timeType {
			schema.Type = &openapi3.Types{"string"}
			schema.Format = "date-time"
		} else {
			typeName := g.generateTypeName(t)

			if _, ok := g.componentSchemaRefs[typeName]; ok && g.opts.exportComponentSchemas.ExportComponentSchemas {
				// Check if we have already parsed this component schema ref based on the name of the struct
				// and use that if so
				return openapi3.NewSchemaRef(fmt.Sprintf("#/components/schemas/%s", typeName), schema), nil
			}

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
						ff := getStructField(t, fieldInfo)
						if tag, ok := ff.Tag.Lookup("yaml"); ok && tag != "-" {
							fieldName, fType = tag, ff.Type
						}
					}
				}

				// extract the field tag if we have a customizer
				var fieldTag reflect.StructTag
				if g.opts.schemaCustomizer != nil {
					ff := getStructField(t, fieldInfo)
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
				schema.Type = &openapi3.Types{"object"}
			}
		}

	default:
		// Object has their own schema's implementation, so we'll use those
		if v := reflect.New(t); v.CanInterface() {
			if v, ok := v.Interface().(SetSchemar); ok {
				v.SetSchema(schema)
			}
		}

	}

	if g.opts.schemaCustomizer != nil {
		if err := g.opts.schemaCustomizer(name, t, tag, schema); err != nil {
			return nil, err
		}
	}

	if !g.opts.exportComponentSchemas.ExportComponentSchemas || t.Kind() != reflect.Struct {
		return openapi3.NewSchemaRef(t.Name(), schema), nil
	}

	// Best way I could find to check that
	// this current type is a generic
	isGeneric, err := regexp.Match(`^.*\[.*\]$`, []byte(t.Name()))
	if err != nil {
		return nil, err
	}

	if isGeneric && !g.opts.exportComponentSchemas.ExportGenerics {
		return openapi3.NewSchemaRef(t.Name(), schema), nil
	}

	// For structs we add the schemas to the component schemas
	if len(parents) > 1 || g.opts.exportComponentSchemas.ExportTopLevelSchema {
		typeName := g.generateTypeName(t)

		g.componentSchemaRefs[typeName] = struct{}{}
		return openapi3.NewSchemaRef(fmt.Sprintf("#/components/schemas/%s", typeName), schema), nil
	}

	return openapi3.NewSchemaRef(t.Name(), schema), nil
}

func (g *Generator) generateTypeName(t reflect.Type) string {
	if g.opts.typeNameGenerator != nil {
		return g.opts.typeNameGenerator(t)
	}

	return t.Name()
}

func (g *Generator) generateCycleSchemaRef(t reflect.Type, schema *openapi3.Schema) *openapi3.SchemaRef {
	var typeName string
	switch t.Kind() {
	case reflect.Ptr:
		return g.generateCycleSchemaRef(t.Elem(), schema)
	case reflect.Slice:
		ref := g.generateCycleSchemaRef(t.Elem(), schema)
		sliceSchema := openapi3.NewSchema()
		sliceSchema.Type = &openapi3.Types{"array"}
		sliceSchema.Items = ref
		return openapi3.NewSchemaRef("", sliceSchema)
	case reflect.Map:
		ref := g.generateCycleSchemaRef(t.Elem(), schema)
		mapSchema := openapi3.NewSchema()
		mapSchema.Type = &openapi3.Types{"object"}
		mapSchema.AdditionalProperties = openapi3.AdditionalProperties{Schema: ref}
		return openapi3.NewSchemaRef("", mapSchema)
	default:
		typeName = g.generateTypeName(t)
	}

	g.componentSchemaRefs[typeName] = struct{}{}
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
