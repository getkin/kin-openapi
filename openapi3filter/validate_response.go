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
	responses := route.Operation.Responses
	if responses != nil && len(responses) > 0 {
		responseRef := responses.Get(status) // Response
		if responseRef == nil {
			responseRef = responses.Default() // Default input
		}
		if responseRef == nil {
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
		response := responseRef.Value
		if response == nil {
			return &ResponseError{
				Input:  input,
				Reason: "response has not been resolved",
			}
		}
		content := response.Content
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
			schemaRef := contentType.Schema
			if schemaRef != nil && mediaType == "application/json" {
				schema := schemaRef.Value

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
