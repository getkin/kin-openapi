package openapi3

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// conflicting uniqueItems values are merged successfully
func TestMerge_UniqueItemsTrue(t *testing.T) {
	merged, err := Merge(&Schema{
		AllOf: SchemaRefs{
			&SchemaRef{
				Value: &Schema{
					Type:        "object",
					UniqueItems: true,
				},
			},
			&SchemaRef{
				Value: &Schema{
					Type:        "object",
					UniqueItems: false,
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, true, merged.UniqueItems)
}

// non-conflicting uniqueItems values are merged successfully
func TestMerge_UniqueItemsFalse(t *testing.T) {
	merged, err := Merge(&Schema{
		AllOf: SchemaRefs{
			&SchemaRef{
				Value: &Schema{
					Type:        "object",
					UniqueItems: false,
				},
			},
			&SchemaRef{
				Value: &Schema{
					Type:        "object",
					UniqueItems: false,
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, false, merged.UniqueItems)
}

// Item merge fails due to conflicting item types.
func TestMerge_Items_Failure(t *testing.T) {
	_, err := Merge(&Schema{
		AllOf: SchemaRefs{
			&SchemaRef{
				Value: &Schema{
					Type: "object",
					Properties: Schemas{
						"test": &SchemaRef{
							Value: &Schema{
								Type: "array",
								Items: &SchemaRef{
									Value: &Schema{
										Type: "integer",
									},
								},
							},
						},
					},
				},
			},
			&SchemaRef{
				Value: &Schema{
					Type: "object",
					Properties: Schemas{
						"test": &SchemaRef{
							Value: &Schema{
								Type: "array",
								Items: &SchemaRef{
									Value: &Schema{
										Type: "string",
									},
								},
							},
						},
					},
				},
			},
		},
	})
	require.EqualError(t, err, TypeErrorMessage)
}

// items are merged successfully when there are no conflicts
func TestMerge_Items(t *testing.T) {
	merged, err := Merge(&Schema{
		AllOf: SchemaRefs{
			&SchemaRef{
				Value: &Schema{
					Type: "object",
					Properties: Schemas{
						"test": &SchemaRef{
							Value: &Schema{
								Type: "array",
								Items: &SchemaRef{
									Value: &Schema{
										Type: "integer",
									},
								},
							},
						},
					},
				},
			},
			&SchemaRef{
				Value: &Schema{
					Type: "object",
					Properties: Schemas{
						"test": &SchemaRef{
							Value: &Schema{
								Type: "array",
								Items: &SchemaRef{
									Value: &Schema{
										Type: "integer",
									},
								},
							},
						},
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Nil(t, merged.AllOf)
	require.Equal(t, "array", merged.Properties["test"].Value.Type)
	require.Equal(t, "integer", merged.Properties["test"].Value.Items.Value.Type)
}

func TestMerge_MultipleOfContained(t *testing.T) {

	//todo - more tests
	merged, err := Merge(&Schema{
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
	})
	require.NoError(t, err)
	require.Equal(t, float64(10), *merged.MultipleOf)
}

func TestMerge_MultipleOfNotContained(t *testing.T) {
	merged, err := Merge(&Schema{
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
	})
	require.NoError(t, err)
	require.Equal(t, float64(77), *merged.MultipleOf)
}

func TestMerge_EnumContained(t *testing.T) {
	merged, err := Merge(&Schema{
		AllOf: SchemaRefs{
			&SchemaRef{
				Value: &Schema{
					Type: "object",
					Enum: []interface{}{"1", nil, 1},
				},
			},
			&SchemaRef{
				Value: &Schema{
					Type: "object",
					Enum: []interface{}{"1"},
				},
			},
		},
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []interface{}{"1"}, merged.Enum)
}

func TestMerge_EnumNoIntersection(t *testing.T) {
	merged, err := Merge(&Schema{
		AllOf: SchemaRefs{
			&SchemaRef{
				Value: &Schema{
					Type: "object",
					Enum: []interface{}{"1", nil},
				},
			},
			&SchemaRef{
				Value: &Schema{
					Type: "object",
					Enum: []interface{}{"2"},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Empty(t, merged.Enum)
}

// Properties range is the most restrictive
func TestMerge_RangeProperties(t *testing.T) {
	merged, err := Merge(&Schema{
		AllOf: SchemaRefs{
			&SchemaRef{
				Value: &Schema{
					Type:     "object",
					MinProps: 10,
					MaxProps: Uint64Ptr(40),
				},
			},
			&SchemaRef{
				Value: &Schema{
					Type:     "object",
					MinProps: 5,
					MaxProps: Uint64Ptr(25),
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, uint64(10), merged.MinProps)
	require.Equal(t, uint64(25), *merged.MaxProps)
}

// Items range is the most restrictive
func TestMerge_RangeItems(t *testing.T) {

	merged, err := Merge(&Schema{
		AllOf: SchemaRefs{
			&SchemaRef{
				Value: &Schema{
					Type:     "object",
					MinItems: 10,
					MaxItems: Uint64Ptr(40),
				},
			},
			&SchemaRef{
				Value: &Schema{
					Type:     "object",
					MinItems: 5,
					MaxItems: Uint64Ptr(25),
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, uint64(10), merged.MinItems)
	require.Equal(t, uint64(25), *merged.MaxItems)
}

func TestMerge_Range(t *testing.T) {
	merged, err := Merge(&Schema{
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
	})
	require.NoError(t, err)
	require.Equal(t, float64(10), *merged.Min)
	require.Equal(t, float64(25), *merged.Max)
}

func TestMerge_MaxLength(t *testing.T) {
	merged, err := Merge(&Schema{
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
	})
	require.NoError(t, err)
	require.Equal(t, uint64(10), *merged.MaxLength)
}

func TestMerge_MinLength(t *testing.T) {
	merged, err := Merge(&Schema{
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
	})
	require.NoError(t, err)
	require.Equal(t, uint64(20), merged.MinLength)
}

func TestMerge_Description(t *testing.T) {
	merged, err := Merge(&Schema{
		Description: "desc0",
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
	})
	require.NoError(t, err)
	require.Equal(t, "desc0", merged.Description)
}

func TestMerge_Type(t *testing.T) {
	merged, err := Merge(&Schema{
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
	})
	require.NoError(t, err)
	require.Equal(t, "object", merged.Type)
}

func TestMerge_Title(t *testing.T) {
	merged, err := Merge(&Schema{
		Title: "base schema",
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
	})
	require.NoError(t, err)
	require.Equal(t, "base schema", merged.Title)
}

func TestMerge_Format(t *testing.T) {
	merged, err := Merge(&Schema{
		AllOf: SchemaRefs{
			&SchemaRef{
				Value: &Schema{
					Type:   "object",
					Format: "date",
				},
			},
			&SchemaRef{
				Value: &Schema{
					Type:   "object",
					Format: "date",
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, "date", merged.Format)
}

func TestMerge_Format_Failure(t *testing.T) {
	_, err := Merge(&Schema{
		AllOf: SchemaRefs{
			&SchemaRef{
				Value: &Schema{
					Type:   "object",
					Format: "date",
				},
			},
			&SchemaRef{
				Value: &Schema{
					Type:   "object",
					Format: "byte",
				},
			},
		},
	})
	require.EqualError(t, err, FormatErrorMessage)
}

func TestMerge_EmptySchema(t *testing.T) {
	schema := Schema{}
	merged, err := Merge(&schema)
	require.NoError(t, err)
	require.Equal(t, &schema, merged) //todo &schema
}

func TestMerge_NoAllOf(t *testing.T) {
	schema := Schema{
		Title: "test",
	}
	merged, err := Merge(&schema)
	require.NoError(t, err)
	require.Equal(t, &schema, merged) //todo &schema
}

func TestMerge_TwoObjects(t *testing.T) {

	obj1 := Schemas{
		"description": &SchemaRef{
			Value: &Schema{
				Type: "string",
			},
		},
	}

	obj2 := Schemas{
		"name": &SchemaRef{
			Value: &Schema{
				Type: "string",
			},
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

	merged, err := Merge(&schema)
	require.NoError(t, err)
	require.Len(t, merged.AllOf, 0)
	require.Len(t, merged.Properties, 2)
	require.Equal(t, obj1["description"].Value.Type, merged.Properties["description"].Value.Type)
	require.Equal(t, obj2["name"].Value.Type, merged.Properties["name"].Value.Type)
}

func TestMerge_OneObjectOneProp(t *testing.T) {

	object := Schemas{
		"description": &SchemaRef{
			Value: &Schema{
				Type: "string",
			},
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

	merged, err := Merge(&schema)
	require.NoError(t, err)
	require.Len(t, merged.Properties, 1)
	require.Equal(t, object["description"].Value.Type, merged.Properties["description"].Value.Type)
}

func TestMerge_OneObjectNoProps(t *testing.T) {

	merged, err := Merge(&Schema{
		AllOf: SchemaRefs{
			&SchemaRef{
				Value: &Schema{
					Type:       "object",
					Properties: Schemas{},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, merged.Properties, 0)
}

func TestMerge_OverlappingProps(t *testing.T) {

	obj1 := Schemas{
		"description": &SchemaRef{
			Value: &Schema{
				Title: "first",
				// Type: "string",   TODO: decide on Type conflict resolution strategy
			},
		},
	}

	obj2 := Schemas{
		"description": &SchemaRef{
			Value: &Schema{
				Title: "second",
				// Type: "int",      TODO: decide on Type conflict resolution strategy
			},
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
	merged, err := Merge(&schema)
	require.NoError(t, err)
	require.Len(t, merged.AllOf, 0)
	require.Len(t, merged.Properties, 1)
	require.Equal(t, (*obj1["description"].Value), (*merged.Properties["description"].Value))
}

func TestMergeAllOf_Pattern(t *testing.T) {

	merged, err := Merge(&Schema{
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
	})
	require.NoError(t, err)
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
			merged, err := Merge(doc.Paths["/products"].Get.Responses["200"].Value.Content["application/json"].Schema.Value)
			require.NoError(t, err)

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