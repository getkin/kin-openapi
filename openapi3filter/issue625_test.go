package openapi3filter_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func TestIssue625(t *testing.T) {

	anyItemSpec := `
openapi: 3.0.0
info:
  version: 1.0.0
  title: Sample API
paths:
 /items:
  get:
    description: Returns a list of stuff
    parameters:
    - description: test object
      explode: false
      in: query
      name: test
      required: false
      schema:
        type: array
        items: {}   ###
    responses:
      '200':
        description: Successful response
`[1:]

	objectArraySpec := `
openapi: 3.0.0
info:
  version: 1.0.0
  title: Sample API
paths:
 /items:
  get:
    description: Returns a list of stuff
    parameters:
    - description: test object
      explode: false
      in: query
      name: test
      required: false
      schema:
       type: array
       items:
         type: object
         properties:
           name:
            type: string
    responses:
      '200':
        description: Successful response
`[1:]

	anyOfArraySpec := `
openapi: 3.0.0
info:
  version: 1.0.0
  title: Sample API
paths:
 /items:
  get:
    description: Returns a list of stuff
    parameters:
    - description: test object
      explode: false
      in: query
      name: test
      required: false
      schema:
        type: array
        items:
         anyOf:
          - type: integer
          - type: boolean
    responses:
      '200':
        description: Successful response
`[1:]

	allOfArraySpec := `
openapi: 3.0.0
info:
  version: 1.0.0
  title: Sample API
paths:
 /items:
  get:
    description: Returns a list of stuff
    parameters:
    - description: test object
      explode: false
      in: query
      name: test
      required: false
      schema:
        type: array
        items:
         allOf:
          - type: integer
          - type: number
    responses:
      '200':
        description: Successful response
`[1:]

	oneOfArraySpec := `
openapi: 3.0.0
info:
  version: 1.0.0
  title: Sample API
paths:
 /items:
  get:
    description: Returns a list of stuff
    parameters:
    - description: test object
      explode: false
      in: query
      name: test
      required: false
      schema:
        type: array
        items:
         oneOf:
          - type: integer
          - type: boolean
    responses:
      '200':
        description: Successful response
`[1:]

	tests := []struct {
		name  string
		spec  string
		req   string
		isErr bool
	}{
		{
			name:  "successful any item array",
			spec:  anyItemSpec,
			req:   "/items?test=3",
			isErr: false,
		},
		{
			name:  "successful any item object array",
			spec:  anyItemSpec,
			req:   `/items?test={"name": "test1"}`,
			isErr: false,
		},
		{
			name:  "successful object array",
			spec:  objectArraySpec,
			req:   `/items?test={"name": "test1"}`,
			isErr: false,
		},
		{
			name:  "failed object array",
			spec:  objectArraySpec,
			req:   "/items?test=3",
			isErr: true,
		},
		{
			name:  "success anyof object array",
			spec:  anyOfArraySpec,
			req:   "/items?test=3,7",
			isErr: false,
		},
		{
			name:  "failed anyof object array",
			spec:  anyOfArraySpec,
			req:   "/items?test=s1,s2",
			isErr: true,
		},

		{
			name:  "success allof object array",
			spec:  allOfArraySpec,
			req:   `/items?test=1,3`,
			isErr: false,
		},
		{
			name:  "failed allof object array",
			spec:  allOfArraySpec,
			req:   `/items?test=1.2,3.1`,
			isErr: true,
		},
		{
			name:  "success anyof object array",
			spec:  oneOfArraySpec,
			req:   `/items?test=true,3`,
			isErr: false,
		},
		{
			name:  "faled anyof object array",
			spec:  oneOfArraySpec,
			req:   `/items?test="val1","val2"`,
			isErr: true,
		},
	}

	for _, testcase := range tests {
		t.Run(testcase.name, func(t *testing.T) {
			loader := openapi3.NewLoader()
			ctx := loader.Context

			doc, err := loader.LoadFromData([]byte(testcase.spec))
			require.NoError(t, err)

			err = doc.Validate(ctx)
			require.NoError(t, err)

			router, err := gorillamux.NewRouter(doc)
			require.NoError(t, err)
			httpReq, err := http.NewRequest(http.MethodGet, testcase.req, nil)
			require.NoError(t, err)

			route, pathParams, err := router.FindRoute(httpReq)
			require.NoError(t, err)

			requestValidationInput := &openapi3filter.RequestValidationInput{
				Request:    httpReq,
				PathParams: pathParams,
				Route:      route,
			}
			err = openapi3filter.ValidateRequest(ctx, requestValidationInput)
			require.Equal(t, testcase.isErr, err != nil)
		},
		)
	}
}
