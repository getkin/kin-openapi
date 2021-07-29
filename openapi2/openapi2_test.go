package openapi2_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/ghodss/yaml"
)

func Example() {
	input, err := ioutil.ReadFile("testdata/swagger.json")
	if err != nil {
		panic(err)
	}

	var doc openapi2.T
	if err = json.Unmarshal(input, &doc); err != nil {
		panic(err)
	}
	if doc.ExternalDocs.Description != "Find out more about Swagger" {
		panic(`doc.ExternalDocs was parsed incorrectly!`)
	}

	outputJSON, err := json.Marshal(doc)
	if err != nil {
		panic(err)
	}
	var docAgainFromJSON openapi2.T
	if err = json.Unmarshal(outputJSON, &docAgainFromJSON); err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(doc, docAgainFromJSON) {
		fmt.Println("objects doc & docAgainFromJSON should be the same")
	}

	outputYAML, err := yaml.Marshal(doc)
	if err != nil {
		panic(err)
	}
	var docAgainFromYAML openapi2.T
	if err = yaml.Unmarshal(outputYAML, &docAgainFromYAML); err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(doc, docAgainFromYAML) {
		fmt.Println("objects doc & docAgainFromYAML should be the same")
	}

	// Output:
}
