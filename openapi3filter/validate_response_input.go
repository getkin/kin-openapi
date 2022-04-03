package openapi3filter

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
)

type ResponseValidationInput struct {
	Body                   io.ReadCloser
	RequestValidationInput *RequestValidationInput
	Header                 http.Header
	Options                *Options
	Status                 int
}

func (input *ResponseValidationInput) SetBodyBytes(value []byte) *ResponseValidationInput {
	input.Body = ioutil.NopCloser(bytes.NewReader(value))
	return input
}

var JSONPrefixes = []string{
	")]}',\n",
}

// TrimJSONPrefix trims one of the possible prefixes
func TrimJSONPrefix(data []byte) []byte {
search:
	for _, prefix := range JSONPrefixes {
		if len(data) < len(prefix) {
			continue
		}
		for i, b := range data[:len(prefix)] {
			if b != prefix[i] {
				continue search
			}
		}
		return data[len(prefix):]
	}
	return data
}
