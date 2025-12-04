package openapi3_test

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

func ExampleSetSchemaErrorMessageCustomizer() {
	loader := openapi3.NewLoader()
	spc := `
components:
  schemas:
    Something:
      type: object
      properties:
        field:
          title: Some field
          type: string
`[1:]

	doc, err := loader.LoadFromData([]byte(spc))
	if err != nil {
		panic(err)
	}

	opt := openapi3.SetSchemaErrorMessageCustomizer(func(err *openapi3.SchemaError) string {
		return fmt.Sprintf(`field "%s" should be string`, err.Schema.Title)
	})

	err = doc.Components.Schemas["Something"].Value.Properties["field"].Value.VisitJSON(123, opt)

	fmt.Println(err.Error())

	// Output: field "Some field" should be string
}
