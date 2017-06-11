package openapi3filter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jban332/kinapi/openapi3"
	"io/ioutil"
	"net/http"
	"sort"
)

func ValidateRequest(c context.Context, input *RequestValidationInput) error {
	options := input.Options
	if options == nil {
		options = DefaultOptions
	}
	route := input.Route
	operation := route.Operation
	if operation == nil {
		return errRouteMissingOperation
	}
	operationParameters := operation.Parameters
	pathItemParameters := route.PathItem.Parameters

	// For each parameter of the PathItem
	if pathItemParameters != nil {
		for _, parameter := range pathItemParameters {
			if operationParameters != nil {
				override := operationParameters.GetByInAndName(parameter.In, parameter.Name)
				if override != nil {
					continue
				}
			}
			err := ValidateParameter(c, input, parameter)
			if err != nil {
				return err
			}
		}
	}

	// For each parameter of the Operation
	if operationParameters != nil {
		for _, parameter := range operationParameters {
			err := ValidateParameter(c, input, parameter)
			if err != nil {
				return err
			}
		}
	}

	// RequestBody
	inputBody := operation.RequestBody
	if inputBody != nil && options.ExcludeRequestBody == false {
		err := ValidateRequestBody(c, input, inputBody)
		if err != nil {
			return err
		}
	}

	// Security
	security := operation.Security
	if security != nil {
		err := ValidateSecurityRequirements(c, input, *security)
		if err != nil {
			return err
		}
	}
	return nil
}

func ValidateParameter(c context.Context, input *RequestValidationInput, parameter *openapi3.Parameter) error {
	req := input.Request
	name := parameter.Name
	var value string
	var found bool
	switch parameter.In {
	case openapi3.ParameterInPath:
		pathParams := input.PathParams
		if pathParams != nil {
			value, found = pathParams[name]
		}
	case openapi3.ParameterInQuery:
		values := input.GetQueryParams()[name]
		if len(values) > 0 {
			value = values[0]
			found = true
		}
	case openapi3.ParameterInHeader:
		var values []string
		values, found = req.Header[http.CanonicalHeaderKey(name)]
		if len(values) > 0 {
			value = values[0]
		}
	case openapi3.ParameterInCookie:
		cookie, err := req.Cookie(name)
		if err == nil {
			value = cookie.Value
			found = true
		} else {
			if err != http.ErrNoCookie {
				return &RequestError{
					Input:     input,
					Parameter: parameter,
					Reason:    "parsing failed",
					Err:       err,
				}
			}
		}
	default:
		return &RequestError{
			Input:     input,
			Parameter: parameter,
			Reason:    "unsupported 'in'",
		}
	}
	if !found {
		if parameter.Required {
			return &RequestError{
				Input:     input,
				Parameter: parameter,
				Reason:    "must have a value",
			}
		}
		return nil
	}
	schema := parameter.Schema
	if schema != nil {
		// Only check schema if no transformation is needed
		if schema.TypesContains("string") {
			err := schema.VisitJSONString(value)
			if err != nil {
				return &RequestError{
					Input:     input,
					Parameter: parameter,
					Err:       err,
				}
			}
		}
	}
	return nil
}

func ValidateRequestBody(c context.Context, input *RequestValidationInput, inputBody *openapi3.RequestBody) error {
	req := input.Request
	content := inputBody.Content
	if content != nil && len(content) > 0 {
		inputMIME := req.Header.Get("Content-Type")
		mediaType := parseMediaType(inputMIME)
		contentType := inputBody.Content[mediaType]
		if contentType == nil {
			return &RequestError{
				Input:       input,
				RequestBody: inputBody,
				Reason:      "header 'Content-type' has unexpected value",
			}
		}
		schema := contentType.Schema
		if schema != nil && mediaType == "application/json" {
			body := req.Body
			defer body.Close()
			data, err := ioutil.ReadAll(body)
			if err != nil {
				return &RequestError{
					Input:       input,
					RequestBody: inputBody,
					Reason:      "reading failed",
					Err:         err,
				}
			}

			// Put the data back into the input
			req.Body = ioutil.NopCloser(bytes.NewReader(data))

			// Decode JSON
			var value interface{}
			err = json.Unmarshal(data, &value)
			if err != nil {
				return &RequestError{
					Input:       input,
					RequestBody: inputBody,
					Reason:      "decoding JSON failed",
					Err:         err,
				}
			}

			// Validate JSON with the schema
			err = schema.VisitJSON(value)
			if err != nil {
				return &RequestError{
					Input:       input,
					RequestBody: inputBody,
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
						errs[currentIndex] = fmt.Errorf("Panicked")
					}
					doneChan <- false
				}
			}()
			err := validateSecurityRequirement(c, input, currentSecurityRequirement)
			if err == nil {
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

	// Visit all requirements
	for _, name := range names {
		var securityScheme *openapi3.SecurityScheme
		if securitySchemes != nil {
			securityScheme = securitySchemes[name]
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
