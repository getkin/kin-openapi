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
	doc, err := openapi3.NewLoader().LoadFromFile("./testdata/petstore.yaml")
	if err != nil {
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
	req, err := http.NewRequest("POST", ts.URL+"/pet", body)
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
	// 	Error at "/name": Field must be set to string or not be present
	// Schema:
	//   {
	//     "example": "doggie",
	//     "type": "string"
	//   }
	//
	// Value:
	//   "number, integer"
	//
	// ===== Start New Error =====
	// @body.status:
	// 	Error at "/status": value is not one of the allowed values
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
	// response: 400 {}
}

const (
	prefixBody = "@body"
	unknown    = "@unknown"
)

func convertError(me openapi3.MultiError) map[string][]string {
	issues := make(map[string][]string)
	for _, err := range me {
		switch err := err.(type) {
		case *openapi3.SchemaError:
			// Can inspect schema validation errors here, e.g. err.Value
			field := prefixBody
			if path := err.JSONPointer(); len(path) > 0 {
				field = fmt.Sprintf("%s.%s", field, strings.Join(path, "."))
			}
			if _, ok := issues[field]; !ok {
				issues[field] = make([]string, 0, 3)
			}
			issues[field] = append(issues[field], err.Error())
		case *openapi3filter.RequestError: // possible there were multiple issues that failed validation
			if err, ok := err.Err.(openapi3.MultiError); ok {
				for k, v := range convertError(err) {
					if _, ok := issues[k]; !ok {
						issues[k] = make([]string, 0, 3)
					}
					issues[k] = append(issues[k], v...)
				}
				continue
			}

			// check if invalid HTTP parameter
			if err.Parameter != nil {
				prefix := err.Parameter.In
				name := fmt.Sprintf("%s.%s", prefix, err.Parameter.Name)
				if _, ok := issues[name]; !ok {
					issues[name] = make([]string, 0, 3)
				}
				issues[name] = append(issues[name], err.Error())
				continue
			}

			// check if requestBody
			if err.RequestBody != nil {
				if _, ok := issues[prefixBody]; !ok {
					issues[prefixBody] = make([]string, 0, 3)
				}
				issues[prefixBody] = append(issues[prefixBody], err.Error())
				continue
			}
		default:
			reasons, ok := issues[unknown]
			if !ok {
				reasons = make([]string, 0, 3)
			}
			reasons = append(reasons, err.Error())
			issues[unknown] = reasons
		}
	}
	return issues
}
