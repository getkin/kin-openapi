package openapi3filter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"sort"

	"github.com/getkin/kin-openapi/openapi3"
)

func ValidateRequest(c context.Context, input *RequestValidationInput) error {
	options := input.Options
	if options == nil {
		options = DefaultOptions
	}
	route := input.Route
	if route == nil {
		return errors.New("invalid route")
	}
	operation := route.Operation
	if operation == nil {
		return errRouteMissingOperation
	}
	operationParameters := operation.Parameters
	pathItemParameters := route.PathItem.Parameters

	// For each parameter of the PathItem
	for _, parameterRef := range pathItemParameters {
		parameter := parameterRef.Value
		if operationParameters != nil {
			if override := operationParameters.GetByInAndName(parameter.In, parameter.Name); override != nil {
				continue
			}
			if err := ValidateParameter(c, input, parameter); err != nil {
				return err
			}
		}
	}

	// For each parameter of the Operation
	for _, parameter := range operationParameters {
		if err := ValidateParameter(c, input, parameter.Value); err != nil {
			return err
		}
	}

	// RequestBody
	requestBody := operation.RequestBody
	if requestBody != nil && !options.ExcludeRequestBody {
		if err := ValidateRequestBody(c, input, requestBody.Value); err != nil {
			return err
		}
	}

	// Security
	security := operation.Security
	if security != nil {
		if err := ValidateSecurityRequirements(c, input, *security); err != nil {
			return err
		}
	}
	return nil
}

func ValidateParameter(c context.Context, input *RequestValidationInput, parameter *openapi3.Parameter) error {
	value, err := decodeParameter(parameter, input)
	if err != nil {
		return err
	}

	// Validate a parameter's value.
	if value == nil {
		if parameter.Required {
			return &RequestError{
				Input:     input,
				Parameter: parameter,
				Reason:    "must have a value",
			}
		}
		return nil
	}
	if schemaRef := parameter.Schema; schemaRef != nil {
		// Only check schema if no transformation is needed
		if schema := schemaRef.Value; schema.Type == "string" {
			if err = schema.VisitJSON(value); err != nil {
				return &RequestError{Input: input, Parameter: parameter, Err: err}
			}
		}
	}
	return nil
}

func ValidateRequestBody(c context.Context, input *RequestValidationInput, requestBody *openapi3.RequestBody) error {
	req := input.Request
	content := requestBody.Content
	if content != nil && len(content) > 0 {
		inputMIME := req.Header.Get("Content-Type")
		mediaType := parseMediaType(inputMIME)
		contentType := requestBody.Content[mediaType]
		if contentType == nil {
			return &RequestError{
				Input:       input,
				RequestBody: requestBody,
				Reason:      fmt.Sprintf("header 'Content-Type' has unexpected value: %q", inputMIME),
			}
		}
		schemaRef := contentType.Schema
		if schemaRef != nil && isMediaTypeJSON(mediaType) {
			schema := schemaRef.Value
			body := req.Body
			defer body.Close()
			data, err := ioutil.ReadAll(body)
			if err != nil {
				return &RequestError{
					Input:       input,
					RequestBody: requestBody,
					Reason:      "reading failed",
					Err:         err,
				}
			}

			// Put the data back into the input
			req.Body = ioutil.NopCloser(bytes.NewReader(data))

			// Decode JSON
			var value interface{}
			if err := json.Unmarshal(data, &value); err != nil {
				return &RequestError{
					Input:       input,
					RequestBody: requestBody,
					Reason:      "decoding JSON failed",
					Err:         err,
				}
			}

			// Validate JSON with the schema
			if err := schema.VisitJSON(value); err != nil {
				return &RequestError{
					Input:       input,
					RequestBody: requestBody,
					Reason:      "doesn't input the schema",
					Err:         err,
				}
			}
		}
	}
	return nil
}

// ValidateSecurityRequirements validates a multiple OpenAPI 3 security requirements.
// Returns nil if one of them inputed.
// Otherwise returns an error describing the security failures.
func ValidateSecurityRequirements(c context.Context, input *RequestValidationInput, srs openapi3.SecurityRequirements) error {
	// Alternative requirements
	if len(srs) == 0 {
		return nil
	}

	doneChan := make(chan bool, len(srs))
	errs := make([]error, len(srs))

	// For each alternative
	for i, securityRequirement := range srs {
		// Capture index from iteration variable
		currentIndex := i
		currentSecurityRequirement := securityRequirement
		go func() {
			defer func() {
				v := recover()
				if v != nil {
					if err, ok := v.(error); ok {
						errs[currentIndex] = err
					} else {
						errs[currentIndex] = errors.New("Panicked")
					}
					doneChan <- false
				}
			}()
			if err := validateSecurityRequirement(c, input, currentSecurityRequirement); err == nil {
				doneChan <- true
			} else {
				errs[currentIndex] = err
				doneChan <- false
			}
		}()
	}

	// Wait for all
	for i := 0; i < len(srs); i++ {
		ok := <-doneChan
		if ok {
			close(doneChan)
			return nil
		}
	}
	return &SecurityRequirementsError{
		SecurityRequirements: srs,
		Errors:               errs,
	}
}

// validateSecurityRequirement validates a single OpenAPI 3 security requirement
func validateSecurityRequirement(c context.Context, input *RequestValidationInput, securityRequirement openapi3.SecurityRequirement) error {
	swagger := input.Route.Swagger
	if swagger == nil {
		return errRouteMissingSwagger
	}
	securitySchemes := swagger.Components.SecuritySchemes

	// Ensure deterministic order
	names := make([]string, 0, len(securityRequirement))
	for name := range securityRequirement {
		names = append(names, name)
	}
	sort.Strings(names)

	// Get authentication function
	options := input.Options
	if options == nil {
		options = DefaultOptions
	}
	f := options.AuthenticationFunc
	if f == nil {
		return ErrAuthenticationServiceMissing
	}

	if len(names) > 0 {
		name := names[0]
		var securityScheme *openapi3.SecurityScheme
		if securitySchemes != nil {
			if ref := securitySchemes[name]; ref != nil {
				securityScheme = ref.Value
			}
		}
		if securityScheme == nil {
			return &RequestError{
				Input: input,
				Err:   fmt.Errorf("Security scheme '%s' is not declared", name),
			}
		}
		scopes := securityRequirement[name]
		return f(c, &AuthenticationInput{
			RequestValidationInput: input,
			SecuritySchemeName:     name,
			SecurityScheme:         securityScheme,
			Scopes:                 scopes,
		})
	}
	return nil
}
