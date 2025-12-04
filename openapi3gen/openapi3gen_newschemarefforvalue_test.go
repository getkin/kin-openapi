package openapi3gen_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"github.com/getkin/kin-openapi/openapi3gen/internal/subpkg"
)

// Make sure that custom schema name generator is employed and results produced with it are properly used
func ExampleNewSchemaRefForValue_withSubPackages() {
	type Parent struct {
		Field1 string       `json:"field1"`
		Child  subpkg.Child `json:"child"`
	}

	// these schema names should be returned by name generator
	parentSchemaName := "PARENT_TYPE"
	childSchemaName := "CHILD_TYPE"

	// sample of a type name generator
	typeNameGenerator := func(t reflect.Type) string {
		switch t {
		case reflect.TypeOf(Parent{}):
			return parentSchemaName
		case reflect.TypeOf(subpkg.Child{}):
			return childSchemaName
		}
		return t.Name()
	}

	schemas := make(openapi3.Schemas)
	schemaRef, err := openapi3gen.NewSchemaRefForValue(
		&Parent{},
		schemas,
		openapi3gen.CreateComponentSchemas(openapi3gen.ExportComponentSchemasOptions{
			ExportComponentSchemas: true, ExportTopLevelSchema: true,
		}),
		openapi3gen.CreateTypeNameGenerator(typeNameGenerator),
		openapi3gen.UseAllExportedFields(),
	)
	if err != nil {
		panic(err)
	}

	var data []byte
	if data, err = json.MarshalIndent(&schemas, "", "  "); err != nil {
		panic(err)
	}
	fmt.Printf("schemas: %s\n", data)
	if data, err = json.MarshalIndent(&schemaRef, "", "  "); err != nil {
		panic(err)
	}

	fmt.Printf("schemaRef: %s\n", data)
	// Output:
	// schemas: {
	//   "CHILD_TYPE": {
	//     "properties": {
	//       "name": {
	//         "type": "string"
	//       }
	//     },
	//     "type": "object"
	//   },
	//   "PARENT_TYPE": {
	//     "properties": {
	//       "child": {
	//         "$ref": "#/components/schemas/CHILD_TYPE"
	//       },
	//       "field1": {
	//         "type": "string"
	//       }
	//     },
	//     "type": "object"
	//   }
	// }
	// schemaRef: {
	//   "$ref": "#/components/schemas/PARENT_TYPE"
	// }

}

func ExampleNewSchemaRefForValue_withExportingSchemas() {
	type Child struct {
		Age string `json:"age"`
	}
	type AnotherStruct struct {
		Field1 string `json:"field1"`
		Field2 string `json:"field2"`
		Field3 string `json:"field3"`
	}
	type RecursiveType struct {
		Field1        string        `json:"field1"`
		Field2        string        `json:"field2"`
		Field3        string        `json:"field3"`
		AnotherStruct AnotherStruct `json:"children,omitempty"`
		Child         subpkg.Child  `json:"child"`
		Child2        Child         `json:"child2"`
	}

	// sample of a type name generator
	typeNameGenerator := func(t reflect.Type) string {
		packages := strings.Split(t.PkgPath(), "/")
		return packages[len(packages)-1] + "_" + t.Name()
	}

	schemas := make(openapi3.Schemas)
	schemaRef, err := openapi3gen.NewSchemaRefForValue(
		&RecursiveType{},
		schemas,
		openapi3gen.CreateComponentSchemas(openapi3gen.ExportComponentSchemasOptions{
			ExportComponentSchemas: true, ExportTopLevelSchema: false,
		}),
		openapi3gen.CreateTypeNameGenerator(typeNameGenerator),
		openapi3gen.UseAllExportedFields(),
	)
	if err != nil {
		panic(err)
	}

	var schemasByte []byte
	if schemasByte, err = json.MarshalIndent(&schemas, "", "  "); err != nil {
		panic(err)
	}
	var schemaRefByte []byte
	if schemaRefByte, err = json.MarshalIndent(&schemaRef, "", "  "); err != nil {
		panic(err)
	}
	fmt.Printf("schemas: %s\nschemaRef: %s\n", schemasByte, schemaRefByte)
	// Output:
	// schemas: {
	//   "openapi3gen_test_AnotherStruct": {
	//     "properties": {
	//       "field1": {
	//         "type": "string"
	//       },
	//       "field2": {
	//         "type": "string"
	//       },
	//       "field3": {
	//         "type": "string"
	//       }
	//     },
	//     "type": "object"
	//   },
	//   "openapi3gen_test_Child": {
	//     "properties": {
	//       "age": {
	//         "type": "string"
	//       }
	//     },
	//     "type": "object"
	//   },
	//   "subpkg_Child": {
	//     "properties": {
	//       "name": {
	//         "type": "string"
	//       }
	//     },
	//     "type": "object"
	//   }
	// }
	// schemaRef: {
	//   "properties": {
	//     "child": {
	//       "$ref": "#/components/schemas/subpkg_Child"
	//     },
	//     "child2": {
	//       "$ref": "#/components/schemas/openapi3gen_test_Child"
	//     },
	//     "children": {
	//       "$ref": "#/components/schemas/openapi3gen_test_AnotherStruct"
	//     },
	//     "field1": {
	//       "type": "string"
	//     },
	//     "field2": {
	//       "type": "string"
	//     },
	//     "field3": {
	//       "type": "string"
	//     }
	//   },
	//   "type": "object"
	// }
}

func ExampleNewSchemaRefForValue_withExportingSchemasIgnoreTopLevelParent() {
	type AnotherStruct struct {
		Field1 string `json:"field1"`
		Field2 string `json:"field2"`
		Field3 string `json:"field3"`
	}
	type RecursiveType struct {
		Field1        string        `json:"field1"`
		Field2        string        `json:"field2"`
		Field3        string        `json:"field3"`
		AnotherStruct AnotherStruct `json:"children,omitempty"`
	}

	schemas := make(openapi3.Schemas)
	schemaRef, err := openapi3gen.NewSchemaRefForValue(&RecursiveType{}, schemas, openapi3gen.CreateComponentSchemas(openapi3gen.ExportComponentSchemasOptions{
		ExportComponentSchemas: true, ExportTopLevelSchema: false,
	}))
	if err != nil {
		panic(err)
	}

	var data []byte
	if data, err = json.MarshalIndent(&schemas, "", "  "); err != nil {
		panic(err)
	}
	fmt.Printf("schemas: %s\n", data)
	if data, err = json.MarshalIndent(&schemaRef, "", "  "); err != nil {
		panic(err)
	}
	fmt.Printf("schemaRef: %s\n", data)
	// Output:
	// schemas: {
	//   "AnotherStruct": {
	//     "properties": {
	//       "field1": {
	//         "type": "string"
	//       },
	//       "field2": {
	//         "type": "string"
	//       },
	//       "field3": {
	//         "type": "string"
	//       }
	//     },
	//     "type": "object"
	//   }
	// }
	// schemaRef: {
	//   "properties": {
	//     "children": {
	//       "$ref": "#/components/schemas/AnotherStruct"
	//     },
	//     "field1": {
	//       "type": "string"
	//     },
	//     "field2": {
	//       "type": "string"
	//     },
	//     "field3": {
	//       "type": "string"
	//     }
	//   },
	//   "type": "object"
	// }
}

func ExampleNewSchemaRefForValue_withExportingSchemasWithGeneric() {
	type Child struct {
		Age string `json:"age"`
	}
	type GenericStruct[T any] struct {
		GenericField T     `json:"genericField"`
		Child        Child `json:"child"`
	}
	type AnotherStruct struct {
		Field1 string `json:"field1"`
		Field2 string `json:"field2"`
		Field3 string `json:"field3"`
	}
	type RecursiveType struct {
		Field1        string                `json:"field1"`
		Field2        string                `json:"field2"`
		Field3        string                `json:"field3"`
		AnotherStruct AnotherStruct         `json:"children,omitempty"`
		Child         Child                 `json:"child"`
		GenericStruct GenericStruct[string] `json:"genericChild"`
	}

	schemas := make(openapi3.Schemas)
	schemaRef, err := openapi3gen.NewSchemaRefForValue(
		&RecursiveType{},
		schemas,
		openapi3gen.CreateComponentSchemas(openapi3gen.ExportComponentSchemasOptions{
			ExportComponentSchemas: true, ExportTopLevelSchema: true, ExportGenerics: false,
		}),
		openapi3gen.UseAllExportedFields(),
	)
	if err != nil {
		panic(err)
	}

	var data []byte
	if data, err = json.MarshalIndent(&schemas, "", "  "); err != nil {
		panic(err)
	}
	fmt.Printf("schemas: %s\n", data)
	if data, err = json.MarshalIndent(&schemaRef, "", "  "); err != nil {
		panic(err)
	}
	fmt.Printf("schemaRef: %s\n", data)
	// Output:
	// schemas: {
	//   "AnotherStruct": {
	//     "properties": {
	//       "field1": {
	//         "type": "string"
	//       },
	//       "field2": {
	//         "type": "string"
	//       },
	//       "field3": {
	//         "type": "string"
	//       }
	//     },
	//     "type": "object"
	//   },
	//   "Child": {
	//     "properties": {
	//       "age": {
	//         "type": "string"
	//       }
	//     },
	//     "type": "object"
	//   },
	//   "RecursiveType": {
	//     "properties": {
	//       "child": {
	//         "$ref": "#/components/schemas/Child"
	//       },
	//       "children": {
	//         "$ref": "#/components/schemas/AnotherStruct"
	//       },
	//       "field1": {
	//         "type": "string"
	//       },
	//       "field2": {
	//         "type": "string"
	//       },
	//       "field3": {
	//         "type": "string"
	//       },
	//       "genericChild": {
	//         "properties": {
	//           "child": {
	//             "$ref": "#/components/schemas/Child"
	//           },
	//           "genericField": {
	//             "type": "string"
	//           }
	//         },
	//         "type": "object"
	//       }
	//     },
	//     "type": "object"
	//   }
	// }
	// schemaRef: {
	//   "$ref": "#/components/schemas/RecursiveType"
	// }
}

func ExampleNewSchemaRefForValue_withExportingSchemasWithMap() {
	type Child struct {
		Age string `json:"age"`
	}
	type MyType struct {
		Field1 string           `json:"field1"`
		Field2 string           `json:"field2"`
		Map1   map[string]any   `json:"anymap"`
		Map2   map[string]Child `json:"anymapChild"`
	}

	schemas := make(openapi3.Schemas)
	schemaRef, err := openapi3gen.NewSchemaRefForValue(
		&MyType{},
		schemas,
		openapi3gen.CreateComponentSchemas(openapi3gen.ExportComponentSchemasOptions{
			ExportComponentSchemas: true, ExportTopLevelSchema: false, ExportGenerics: true,
		}),
		openapi3gen.UseAllExportedFields(),
	)
	if err != nil {
		panic(err)
	}

	var data []byte
	if data, err = json.MarshalIndent(&schemas, "", "  "); err != nil {
		panic(err)
	}
	fmt.Printf("schemas: %s\n", data)
	if data, err = json.MarshalIndent(&schemaRef, "", "  "); err != nil {
		panic(err)
	}
	fmt.Printf("schemaRef: %s\n", data)
	// Output:
	// schemas: {
	//   "Child": {
	//     "properties": {
	//       "age": {
	//         "type": "string"
	//       }
	//     },
	//     "type": "object"
	//   }
	// }
	// schemaRef: {
	//   "properties": {
	//     "anymap": {
	//       "additionalProperties": {},
	//       "type": "object"
	//     },
	//     "anymapChild": {
	//       "additionalProperties": {
	//         "$ref": "#/components/schemas/Child"
	//       },
	//       "type": "object"
	//     },
	//     "field1": {
	//       "type": "string"
	//     },
	//     "field2": {
	//       "type": "string"
	//     }
	//   },
	//   "type": "object"
	// }
}
