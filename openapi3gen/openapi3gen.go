// Package openapi3gen generates OpenAPI 3 schemas for Go types.
package openapi3gen

import (
	"fmt"
	"github.com/jban332/kin-openapi/jsoninfo"
	"github.com/jban332/kin-openapi/openapi3"
	"reflect"
	"sync"
	"time"
)

// CycleError indicates that a type graph has one or more possible cycles.
type CycleError struct{}

func (err *CycleError) Error() string {
	return "Detected JSON cycle"
}

var globalSchemasMutex sync.Mutex
var globalSchemas = map[string]*openapi3.Schema{}

func SchemaFromInstance(value interface{}) (*openapi3.Schema, error) {
	return SchemaFromType(reflect.TypeOf(value))
}

func SchemaFromType(t reflect.Type) (*openapi3.Schema, error) {
	return schemaFromType(nil, t)
}

func schemaFromType(parents []*openapi3.Schema, t reflect.Type) (*openapi3.Schema, error) {
	// Get TypeInfo
	typeInfo := jsoninfo.GetTypeInfo(t)

	// Get schema from TypeInfo.
	// This is an atomic read so no need for synchronization.
	schema, _ := typeInfo.Schema.(*openapi3.Schema)
	if schema != nil {
		for _, parent := range parents {
			if schema == parent {
				return nil, &CycleError{}
			}
		}
		return schema, nil
	}

	// Doesn't exist.
	// Acquire a lock.
	typeInfo.SchemaMutex.Lock()
	defer typeInfo.SchemaMutex.Unlock()

	// Try get the schema again
	schema, _ = typeInfo.Schema.(*openapi3.Schema)
	if schema != nil {
		for _, parent := range parents {
			if schema == parent {
				panic(fmt.Errorf("Type '%s' has a cyclic schema", t.String()))
			}
		}
		return schema, nil
	}

	// Doesn't exist.
	// Create the schema.
	schema = openapi3.NewObjectSchema()
	typeInfo.Schema = schema
	if cap(parents) == 0 {
		parents = make([]*openapi3.Schema, 0, 4)
	}
	parents = append(parents, schema)

	// Add fields
	for _, field := range typeInfo.Fields {
		fieldSchema := &openapi3.Schema{}
		schema.Properties[field.JSONName] = &openapi3.SchemaRef{
			Value: fieldSchema,
		}
		t := field.Type
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		if t == timeType {
			fieldSchema.Type = "string"
			fieldSchema.Format = "datetime"
			continue
		}
		if _, ok := t.MethodByName("MarshalJSON"); ok {
			continue
		}
		if _, ok := t.MethodByName("UnmarshalJSON"); ok {
			continue
		}
		switch t.Kind() {
		case reflect.Bool:
			fieldSchema.Type = "bool"
		case reflect.Int, reflect.Int64:
			fieldSchema.Type = "number"
			fieldSchema.Format = "int64"
		case reflect.Float64:
			fieldSchema.Type = "number"
		case reflect.String:
			fieldSchema.Type = "string"
		case reflect.Slice:
			if t.Elem().Kind() == reflect.Uint8 {
				fieldSchema.Type = "string"
				fieldSchema.Format = "byte"
			} else {
				fieldSchema.Type = "array"
			}
		case reflect.Map:
			fieldSchema.Type = "object"
			valueSchema, err := schemaFromType(parents, t.Elem())
			if err != nil {
				return nil, err
			}
			fieldSchema.AdditionalProperties = &openapi3.SchemaRef{
				Value: valueSchema,
			}
		case reflect.Struct:
			fieldSchema.Type = "object"
			newFieldSchema, err := schemaFromType(parents, t)
			if err != nil {
				return nil, err
			}
			if newFieldSchema != nil {
				schema.Properties[field.JSONName] = &openapi3.SchemaRef{
					Value: newFieldSchema,
				}
			}
		}
	}
	return schema, nil
}

var timeType = reflect.TypeOf(time.Time{})
