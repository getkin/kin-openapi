package openapi3_test

import (
	"fmt"
	"github.com/ronniedada/kin-openapi/openapi3"
)

func ExampleSwaggerLoader() {
	source := `{"info":{"description":"An API"}}`
	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData([]byte(source))
	if err != nil {
		panic(err)
	}
	fmt.Print(swagger.Info.Description)
	// Output:
	// An API
}
