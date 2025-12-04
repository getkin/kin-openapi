package openapi3filter_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func Example() {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile("./testdata/petstore.yaml")
	if err != nil {
		panic(err)
	}

	if err = doc.Validate(loader.Context); err != nil {
		panic(err)
	}

	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		panic(err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route, pathParams, err := router.FindRoute(r)
		if err != nil {
			fmt.Println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = openapi3filter.ValidateRequest(r.Context(), &openapi3filter.RequestValidationInput{
			Request:    r,
			PathParams: pathParams,
			Route:      route,
			Options: &openapi3filter.Options{
				MultiError: true,
			},
		})
		switch err := err.(type) {
		case nil:
		case openapi3.MultiError:
			issues := convertError(err)
			names := make([]string, 0, len(issues))
			for k := range issues {
				names = append(names, k)
			}
			sort.Strings(names)
			for _, k := range names {
				msgs := issues[k]
				fmt.Println("===== Start New Error =====")
				fmt.Println(k + ":")
				for _, msg := range msgs {
					fmt.Printf("\t%s\n", msg)
				}
			}
			w.WriteHeader(http.StatusBadRequest)
		default:
			fmt.Println(err.Error())
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer ts.Close()

	// (note invalid type for name and invalid status)
	body := strings.NewReader(`{"name": 100, "photoUrls": [], "status": "invalidStatus"}`)
	req, err := http.NewRequest("POST", ts.URL+"/pet?num=0", body)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	fmt.Printf("response: %d %s\n", resp.StatusCode, resp.Body)

	// Output:
	// ===== Start New Error =====
	// @body.name:
	// 	Error at "/name": value must be a string
	// Schema:
	//   {
	//     "example": "doggie",
	//     "type": "string"
	//   }
	//
	// Value:
	//   100
	//
	// ===== Start New Error =====
	// @body.status:
	// 	Error at "/status": value is not one of the allowed values ["available","pending","sold"]
	// Schema:
	//   {
	//     "description": "pet status in the store",
	//     "enum": [
	//       "available",
	//       "pending",
	//       "sold"
	//     ],
	//     "type": "string"
	//   }
	//
	// Value:
	//   "invalidStatus"
	//
	// ===== Start New Error =====
	// query.num:
	// 	parameter "num" in query has an error: number must be at least 1
	// Schema:
	//   {
	//     "minimum": 1,
	//     "type": "integer"
	//   }
	//
	// Value:
	//   0
	//
	// response: 400 {}
}

func convertError(me openapi3.MultiError) map[string][]string {
	issues := make(map[string][]string)
	for _, err := range me {
		const prefixBody = "@body"
		switch err := err.(type) {
		case *openapi3.SchemaError:
			// Can inspect schema validation errors here, e.g. err.Value
			field := prefixBody
			if path := err.JSONPointer(); len(path) > 0 {
				field = fmt.Sprintf("%s.%s", field, strings.Join(path, "."))
			}
			issues[field] = append(issues[field], err.Error())
		case *openapi3filter.RequestError: // possible there were multiple issues that failed validation

			// check if invalid HTTP parameter
			if err.Parameter != nil {
				prefix := err.Parameter.In
				name := fmt.Sprintf("%s.%s", prefix, err.Parameter.Name)
				issues[name] = append(issues[name], err.Error())
				continue
			}

			if err, ok := err.Err.(openapi3.MultiError); ok {
				for k, v := range convertError(err) {
					issues[k] = append(issues[k], v...)
				}
				continue
			}

			// check if requestBody
			if err.RequestBody != nil {
				issues[prefixBody] = append(issues[prefixBody], err.Error())
				continue
			}
		default:
			const unknown = "@unknown"
			issues[unknown] = append(issues[unknown], err.Error())
		}
	}
	return issues
}
