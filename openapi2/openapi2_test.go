package openapi2_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"

	"github.com/getkin/kin-openapi/openapi2"
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

	output, err := json.Marshal(doc)
	if err != nil {
		panic(err)
	}

	var docAgain openapi2.T
	if err = json.Unmarshal(output, &docAgain); err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(doc, docAgain) {
		fmt.Println("Objects doc & docAgain should be the same")
	}
	// Output:
}
