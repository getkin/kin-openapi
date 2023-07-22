package openapi3

import (
	"context"
	"encoding/json"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMerge_MultipleOf(t *testing.T) {

	//todo - more tests

	schema := Schema{
		AllOf: SchemaRefs{
			&SchemaRef{
				Value: &Schema{
					Type:       "object",
					MultipleOf: Float64Ptr(10.0),
				},
			},
			&SchemaRef{
				Value: &Schema{
					Type:       "object",
					MultipleOf: Float64Ptr(2.0),
				},
			},
		},
	}

	merged, err := Merge(schema)
	require.Equal(t, nil, err)
	require.Equal(t, float64(10), *merged.MultipleOf)

	schema = Schema{
		AllOf: SchemaRefs{
			&SchemaRef{
				Value: &Schema{
					Type:       "object",
					MultipleOf: Float64Ptr(11.0),
				},
			},
			&SchemaRef{
				Value: &Schema{
					Type:       "object",
					MultipleOf: Float64Ptr(0.7),
				},
			},
		},
	}

	merged, err = Merge(schema)
	require.Equal(t, nil, err)
	require.Equal(t, float64(77), *merged.MultipleOf)
}

func TestMerge_Enum(t *testing.T) {
	obj1Enum := make([]interface{}, 3)
	obj1Enum[0] = "1"
	obj1Enum[1] = nil
	obj1Enum[2] = 1

	obj2Enum := make([]interface{}, 1)
	obj2Enum[0] = struct{}{}

	schema := Schema{
		AllOf: SchemaRefs{
			&SchemaRef{
				Value: &Schema{
					Type: "object",
					Enum: obj1Enum,
				},
			},
			&SchemaRef{
				Value: &Schema{
					Type: "object",
					Enum: obj2Enum,
				},
			},
		},
	}
	concatenated := append(obj1Enum, obj2Enum...)
	merged, err := Merge(schema)
	require.Equal(t, nil, err)
	require.Equal(t, concatenated, merged.Enum)
}

func TestMerge_MinMax(t *testing.T) {
	schema := Schema{
		AllOf: SchemaRefs{
			&SchemaRef{
				Value: &Schema{
					Type: "object",
					Min:  Float64Ptr(10),
					Max:  Float64Ptr(40),
				},
			},
			&SchemaRef{
				Value: &Schema{
					Type: "object",
					Min:  Float64Ptr(5),
					Max:  Float64Ptr(25),
				},
			},
		},
	}
	merged, err := Merge(schema)
	require.Equal(t, nil, err)
	require.Equal(t, float64(10), *merged.Min)
	require.Equal(t, float64(25), *merged.Max)
}

func TestMerge_MaxLength(t *testing.T) {
	schema := Schema{
		AllOf: SchemaRefs{
			&SchemaRef{
				Value: &Schema{
					Type:      "object",
					MaxLength: Uint64Ptr(10),
				},
			},
			&SchemaRef{
				Value: &Schema{
					Type:      "object",
					MaxLength: Uint64Ptr(20),
				},
			},
		},
	}
	merged, err := Merge(schema)
	require.Equal(t, nil, err)
	require.Equal(t, uint64(10), *merged.MaxLength)
}

func TestMerge_MinLength(t *testing.T) {
	schema := Schema{
		AllOf: SchemaRefs{
			&SchemaRef{
				Value: &Schema{
					Type:      "object",
					MinLength: 10,
				},
			},
			&SchemaRef{
				Value: &Schema{
					Type:      "object",
					MinLength: 20,
				},
			},
		},
	}
	merged, err := Merge(schema)
	require.Equal(t, nil, err)
	require.Equal(t, uint64(20), merged.MinLength)
}

func TestMerge_Description(t *testing.T) {
	schema := Schema{
		AllOf: SchemaRefs{
			&SchemaRef{
				Value: &Schema{
					Type:        "object",
					Description: "desc1",
				},
			},
			&SchemaRef{
				Value: &Schema{
					Type:        "object",
					Description: "desc2",
				},
			},
		},
	}
	merged, err := Merge(schema)
	require.Equal(t, nil, err)
	require.Equal(t, "desc1", merged.Description)
}

func TestMerge_Type(t *testing.T) {
	schema := Schema{
		AllOf: SchemaRefs{
			&SchemaRef{
				Value: &Schema{
					Type: "object",
				},
			},
			&SchemaRef{
				Value: &Schema{
					Type: "object",
				},
			},
		},
	}
	merged, err := Merge(schema)
	require.Equal(t, nil, err)
	require.Equal(t, "object", merged.Type)
}

func TestMerge_Title(t *testing.T) {
	schema := Schema{
		AllOf: SchemaRefs{
			&SchemaRef{
				Value: &Schema{
					Type:  "object",
					Title: "first",
				},
			},
			&SchemaRef{
				Value: &Schema{
					Type:  "object",
					Title: "second",
				},
			},
		},
	}
	merged, err := Merge(schema)
	require.Equal(t, nil, err)
	require.Equal(t, "first", merged.Title)
}

func TestMerge_Format(t *testing.T) {
	schema := Schema{
		AllOf: SchemaRefs{
			&SchemaRef{
				Value: &Schema{
					Type:   "object",
					Format: "number",
				},
			},
			&SchemaRef{
				Value: &Schema{
					Type:   "object",
					Format: "number",
				},
			},
		},
	}

	merged, err := Merge(schema)
	require.Equal(t, nil, err)
	require.Equal(t, "number", merged.Format)
}

func TestMerge_EmptySchema(t *testing.T) {
	schema := Schema{}
	merged, err := Merge(schema)
	require.Equal(t, nil, err)
	require.Equal(t, &schema, merged) //todo &schema
}

func TestMerge_NoAllOf(t *testing.T) {
	schema := Schema{
		Title: "test",
	}
	merged, err := Merge(schema)
	require.Equal(t, nil, err)
	require.Equal(t, &schema, merged) //todo &schema
}

func TestMerge_TwoObjects(t *testing.T) {

	obj1 := Schemas{}
	obj1["description"] = &SchemaRef{
		Value: &Schema{
			Type: "string",
		},
	}

	obj2 := Schemas{}
	obj2["name"] = &SchemaRef{
		Value: &Schema{
			Type: "string",
		},
	}

	schema := Schema{
		AllOf: SchemaRefs{
			&SchemaRef{
				Value: &Schema{
					Type:       "object",
					Properties: obj1,
				},
			},
			&SchemaRef{
				Value: &Schema{
					Type:       "object",
					Properties: obj2,
				},
			},
		},
	}

	merged, err := Merge(schema)
	require.Equal(t, nil, err)
	require.Len(t, merged.AllOf, 0)
	require.Len(t, merged.Properties, 2)
	require.Equal(t, obj1["description"], merged.Properties["description"])
	require.Equal(t, obj2["name"], merged.Properties["name"])
}

func TestMerge_OneObjectOneProp(t *testing.T) {

	object := Schemas{}
	object["description"] = &SchemaRef{
		Value: &Schema{
			Type: "string",
		},
	}

	schema := Schema{
		AllOf: SchemaRefs{
			&SchemaRef{
				Value: &Schema{
					Type:       "object",
					Properties: object,
				},
			},
		},
	}

	merged, err := Merge(schema)
	require.Equal(t, nil, err)
	require.Len(t, merged.Properties, 1)
	require.Equal(t, object["description"], merged.Properties["description"])
}

func TestMerge_OneObjectNoProps(t *testing.T) {

	schema := Schema{
		AllOf: SchemaRefs{
			&SchemaRef{
				Value: &Schema{
					Type:       "object",
					Properties: Schemas{},
				},
			},
		},
	}

	merged, err := Merge(schema)
	require.Equal(t, nil, err)
	require.Len(t, merged.Properties, 0)
}

func TestMerge_OverlappingProps(t *testing.T) {

	obj1 := Schemas{}
	obj1["description"] = &SchemaRef{
		Value: &Schema{
			Title: "first",
			// Type: "string",   TODO: decide on Type conflict resolution strategy
		},
	}

	obj2 := Schemas{}
	obj2["description"] = &SchemaRef{
		Value: &Schema{
			Title: "second",
			// Type: "int",      TODO: decide on Type conflict resolution strategy
		},
	}

	schema := Schema{
		AllOf: SchemaRefs{
			&SchemaRef{
				Value: &Schema{
					Type:       "object",
					Properties: obj1,
				},
			},
			&SchemaRef{
				Value: &Schema{
					Type:       "object",
					Properties: obj2,
				},
			},
		},
	}
	merged, err := Merge(schema)
	require.Equal(t, nil, err)
	require.Len(t, merged.AllOf, 0)
	require.Len(t, merged.Properties, 1)
	require.Equal(t, (*obj1["description"].Value), (*merged.Properties["description"].Value))
}

func TestMergeAllOf_Pattern(t *testing.T) {
	schema := Schema{
		AllOf: SchemaRefs{
			&SchemaRef{
				Value: &Schema{
					Type:    "object",
					Pattern: "foo",
				},
			},
			&SchemaRef{
				Value: &Schema{
					Type:    "object",
					Pattern: "bar",
				},
			},
		},
	}
	merged, err := Merge(schema)
	require.Equal(t, nil, err)
	require.Equal(t, "(?=foo)(?=bar)", merged.Pattern)
}
func TestMerge_Required(t *testing.T) {

	ctx := context.Background()

	tests := []struct {
		filename string
	}{
		{"testdata/mergeSchemas/properties.yml"},
	}

	for _, test := range tests {
		t.Run(test.filename, func(t *testing.T) {
			// Load in the reference spec from the testdata
			sl := NewLoader()
			sl.IsExternalRefsAllowed = true
			doc, err := sl.LoadFromFile(test.filename)
			require.NoError(t, err, "loading test file")
			err = doc.Validate(ctx)
			require.NoError(t, err, "validating spec")
			merged, err := Merge(*doc.Paths["/products"].Get.Responses["200"].Value.Content["application/json"].Schema.Value)
			require.Equal(t, nil, err)

			props := merged.Properties
			require.Len(t, props, 3)
			require.Contains(t, props, "id")
			require.Contains(t, props, "createdAt")
			require.Contains(t, props, "otherId")

			required := merged.Required
			require.Len(t, required, 2)
			require.Contains(t, required, "id")
			require.Contains(t, required, "otherId")
		})
	}
}

/* temporary */
func PrettyPrintJSON(rawJSON []byte) error {
	var data interface{}

	err := json.Unmarshal(rawJSON, &data)
	if err != nil {
		return err
	}

	prettyJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	log.Println(string(prettyJSON))
	return nil
}

// func TestMerge_NestedAllOf(t *testing.T) {
// 	obj2 := Schemas{}
// 	obj2["description"] = &SchemaRef{
// 		Value: &Schema{
// 			Type:  "object",
// 			Title: "description",
// 		},
// 	}
// 	obj2["abcdefg"] = &SchemaRef{
// 		Value: &Schema{
// 			Type:  "object",
// 			Title: "abc",
// 		},
// 	}

// 	obj1 := Schemas{}
// 	obj1["description"] = &SchemaRef{
// 		Value: &Schema{
// 			Type:  "object",
// 			Title: "object2",
// 		},
// 	}
// 	obj1["test"] = &SchemaRef{
// 		Value: &Schema{
// 			AllOf: SchemaRefs{
// 				&SchemaRef{
// 					Value: &Schema{
// 						Type:  "object",
// 						Title: "abc",
// 					},
// 				},
// 				&SchemaRef{
// 					Value: &Schema{
// 						Type:       "object",
// 						Properties: obj2,
// 					},
// 				},
// 			},
// 		},
// 	}

// 	schema := Schema{
// 		AllOf: SchemaRefs{
// 			&SchemaRef{
// 				Value: &Schema{
// 					Type:       "object",
// 					Properties: obj1,
// 				},
// 			},
// 		},
// 	}

// 	d, _ := schema.MarshalJSON()
// 	PrettyPrintJSON(d)
// 	//todo add tests.
// }
