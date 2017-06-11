// Package openapi3filter validates that requests and inputs request an OpenAPI 3 specification file.
package openapi3filter

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func ValidateResponse(c context.Context, input *ResponseValidationInput) error {
	req := input.RequestValidationInput.Request
	switch req.Method {
	case "HEAD":
		return nil
	}
	status := input.Status
	if status < 100 {
		return &ResponseError{
			Input:  input,
			Reason: "illegal status code",
			Err:    fmt.Errorf("Status %d", status),
		}
	}

	// These status codes will never be validated.
	// TODO: The list is probably missing some.
	switch status {
	case http.StatusNotModified,
		http.StatusPermanentRedirect,
		http.StatusTemporaryRedirect,
		http.StatusMovedPermanently:
		return nil
	}
	route := input.RequestValidationInput.Route
	options := input.Options
	if options == nil {
		options = DefaultOptions
	}

	// Find input for the current status
	inputSpecs := route.Operation.Responses
	if inputSpecs != nil && len(inputSpecs) > 0 {
		inputSpec := inputSpecs.Get(status) // Response
		if inputSpec == nil {
			inputSpec = inputSpecs.Default() // Default input
		}
		if inputSpec == nil {
			// By default, status that is not documented is allowed
			if !options.IncludeResponseStatus {
				return nil
			}

			// Other.
			return &ResponseError{
				Input:  input,
				Reason: "status is not supported",
			}
		}
		content := inputSpec.Content
		if content != nil && len(content) > 0 && options.ExcludeResponseBody == false {
			inputMIME := input.Header.Get("Content-Type")
			mediaType := parseMediaType(inputMIME)
			contentType := content[mediaType]
			if contentType == nil {
				return &ResponseError{
					Input:  input,
					Reason: "input header 'Content-type' has unexpected value",
				}
			}
			schema := contentType.Schema
			if schema != nil && mediaType == "application/json" {
				// Read request body
				body := input.Body

				// Response would contain partial or empty input body
				// after we begin reading.
				// Ensure that this doesn't happen.
				input.Body = nil

				// Ensure we close the reader
				defer body.Close()

				// Read all
				data, err := ioutil.ReadAll(body)
				if err != nil {
					return &ResponseError{
						Input:  input,
						Reason: "reading the input body failed",
						Err:    err,
					}
				}

				// Put the data back into the request
				input.SetBodyBytes(data)

				// Decode JSON
				var value interface{}
				err = json.Unmarshal(data, &value)
				if err != nil {
					return err
				}
				if err != nil {
					return &ResponseError{
						Input:  input,
						Reason: "decoding JSON in the input body failed",
						Err:    err,
					}
				}

				// Validate JSON with the schema
				err = schema.VisitJSON(value)
				if err != nil {
					return &ResponseError{
						Input: input,
						Err:   err,
					}
				}
			}
		}
	}
	return nil
}
