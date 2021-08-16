package legacy_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
)

const spec = `
openapi: 3.0.0
info:
  title: My API
  version: 0.0.1
paths:
  /:
    post:
      responses:
        default:
          description: ''
      requestBody:
        required: true
        content:
          application/json:
            schema:
              oneOf:
              - $ref: '#/components/schemas/Cat'
              - $ref: '#/components/schemas/Dog'
              discriminator:
                propertyName: pet_type

components:
  schemas:
    Pet:
      type: object
      required: [pet_type]
      properties:
        pet_type:
          type: string
      discriminator:
        propertyName: pet_type

    Dog:
      allOf:
      - $ref: '#/components/schemas/Pet'
      - type: object
        properties:
          breed:
            type: string
            enum: [Dingo, Husky, Retriever, Shepherd]
    Cat:
      allOf:
      - $ref: '#/components/schemas/Pet'
      - type: object
        properties:
          hunts:
            type: boolean
          age:
            type: integer
`

func Example() {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	if err != nil {
		panic(err)
	}
	if err := doc.Validate(loader.Context); err != nil {
		panic(err)
	}

	router, err := legacy.NewRouter(doc)
	if err != nil {
		panic(err)
	}

	p, err := json.Marshal(map[string]interface{}{
		"pet_type": "Cat",
		"breed":    "Dingo",
		"bark":     true,
	})
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewReader(p))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")

	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		panic(err)
	}

	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	}
	if err := openapi3filter.ValidateRequest(loader.Context, requestValidationInput); err != nil {
		fmt.Println(err)
	}
	// Output:
	// request body has an error: doesn't match the schema: input matches more than one oneOf schemas

}
