// Package openapi2conv converts an OpenAPI v2 specification to v3.
package openapi2conv

import (
	"github.com/jban332/kinapi/openapi2"
	"github.com/jban332/kinapi/openapi3"
	"net/url"
)

func ToV3Swagger(swagger *openapi2.Swagger) *openapi3.Swagger {
	result := &openapi3.Swagger{
		OpenAPI: "3.0",
		Info:    swagger.Info,
		Components: openapi3.Components{
			Tags: swagger.Tags,
		},
	}
	host := swagger.Host
	if len(host) > 0 {
		schemes := swagger.Schemes
		if len(schemes) == 0 {
			schemes = []string{
				"https://",
			}
		}
		basePath := swagger.BasePath
		for _, scheme := range schemes {
			result.AddServer(&openapi3.Server{
				URL: scheme + host + basePath,
			})
		}
	}
	if paths := swagger.Paths; paths != nil {
		resultPaths := make(map[string]*openapi3.PathItem, len(paths))
		for path, pathItem := range paths {
			resultPaths[path] = ToV3PathItem(swagger, pathItem)
		}
		result.Paths = resultPaths
	}
	if parameters := swagger.Parameters; parameters != nil {
		result.Components.Parameters = make(map[string]*openapi3.Parameter)
		result.Components.RequestBodies = make(map[string]*openapi3.RequestBody)
		for k, parameter := range parameters {
			resultParameter, resultRequestBody := ToV3Parameter(parameter)
			if resultParameter != nil {
				result.Components.Parameters[k] = resultParameter
			}
			if resultRequestBody != nil {
				result.Components.RequestBodies[k] = resultRequestBody
			}
		}
	}
	if responses := swagger.Responses; responses != nil {
		result.Components.Responses = make(map[string]*openapi3.Response, len(responses))
		for k, response := range responses {
			result.Components.Responses[k] = ToV3Response(response)
		}
	}
	result.Components.Schemas = swagger.Definitions
	if m := swagger.SecurityDefinitions; m != nil {
		resultSecuritySchemes := make(map[string]*openapi3.SecurityScheme)
		for k, v := range m {
			resultSecuritySchemes[k] = ToV3SecurityScheme(v)
		}
		result.Components.SecuritySchemes = resultSecuritySchemes
	}
	result.Security = ToV3SecurityRequirements(swagger.Security)
	return result
}

func ToV3PathItem(swagger *openapi2.Swagger, pathItem *openapi2.PathItem) *openapi3.PathItem {
	resultPathItem := &openapi3.PathItem{
		Delete:  ToV3Operation(swagger, pathItem, pathItem.Delete),
		Get:     ToV3Operation(swagger, pathItem, pathItem.Get),
		Head:    ToV3Operation(swagger, pathItem, pathItem.Head),
		Options: ToV3Operation(swagger, pathItem, pathItem.Options),
		Patch:   ToV3Operation(swagger, pathItem, pathItem.Patch),
		Post:    ToV3Operation(swagger, pathItem, pathItem.Post),
		Put:     ToV3Operation(swagger, pathItem, pathItem.Put),
	}
	if parameters := pathItem.Parameters; parameters != nil {
		for _, parameter := range parameters {
			v3Parameter, _ := ToV3Parameter(parameter)
			if v3Parameter != nil {
				resultPathItem.Parameters = append(resultPathItem.Parameters, v3Parameter)
			}
		}
	}
	return resultPathItem
}

func ToV3Operation(swagger *openapi2.Swagger, pathItem *openapi2.PathItem, operation *openapi2.Operation) *openapi3.Operation {
	if operation == nil {
		return nil
	}
	result := &openapi3.Operation{
		Description: operation.Description,
		Tags:        operation.Tags,
	}
	if v := operation.Security; v != nil {
		resultSecurity := ToV3SecurityRequirements(*v)
		result.Security = &resultSecurity
	}
	for _, parameter := range pathItem.Parameters {
		_, v3RequestBody := ToV3Parameter(parameter)
		if v3RequestBody != nil {
			result.RequestBody = v3RequestBody
		}
	}
	for _, parameter := range operation.Parameters {
		v3Parameter, v3RequestBody := ToV3Parameter(parameter)
		if v3RequestBody != nil {
			result.RequestBody = v3RequestBody
		} else if v3Parameter != nil {
			result.AddParameter(v3Parameter)
		}
	}
	if responses := operation.Responses; responses != nil {
		resultResponses := make(openapi3.Responses)
		result.Responses = resultResponses
		for k, response := range responses {
			resultResponses[k] = ToV3Response(response)
		}
	}
	return result
}

func ToV3Parameter(parameter *openapi2.Parameter) (*openapi3.Parameter, *openapi3.RequestBody) {
	if parameter == nil {
		return nil, nil
	}
	in := parameter.In
	if in == "body" {
		requestBody := &openapi3.RequestBody{
			Description: parameter.Description,
			Required:    parameter.Required,
		}
		if schema := parameter.Schema; schema != nil {
			// Assume it's JSON
			requestBody.WithJSONSchema(schema)
		}
		return nil, requestBody
	}
	resultParameter := &openapi3.Parameter{
		In:          in,
		Name:        parameter.Name,
		Description: parameter.Description,
		Required:    parameter.Required,
		Schema:      parameter.Schema,
	}
	if parameter.Type != "" {
		parameter.Schema = &openapi3.Schema{
			Type:         parameter.Type,
			Format:       parameter.Format,
			Enum:         parameter.Enum,
			Min:          parameter.Minimum,
			Max:          parameter.Maximum,
			ExclusiveMin: parameter.ExclusiveMinimum,
			ExclusiveMax: parameter.ExclusiveMinimum,
			MinLength:    parameter.MinLength,
			MaxLength:    parameter.MaxLength,
		}
	}
	return resultParameter, nil
}

func ToV3Response(response *openapi2.Response) *openapi3.Response {
	return &openapi3.Response{
		Description: response.Description,
	}
}

func ToV3SecurityRequirements(requirements openapi2.SecurityRequirements) openapi3.SecurityRequirements {
	if requirements == nil {
		return nil
	}
	result := make(openapi3.SecurityRequirements, len(requirements))
	for i, item := range requirements {
		result[i] = item
	}
	return result
}

func ToV3SecurityScheme(securityScheme *openapi2.SecurityScheme) *openapi3.SecurityScheme {
	if securityScheme == nil {
		return nil
	}
	result := &openapi3.SecurityScheme{
		Description: securityScheme.Description,
	}
	switch securityScheme.Type {
	case "basic":
		result.Type = "http"
		result.Scheme = "basic"
		return result
	case "apiKey":
		result.Type = "apiKey"
		result.In = securityScheme.In
		result.Name = securityScheme.Name
		return result
	case "oauth2":
		result.Type = "oauth2"
		flows := &openapi3.OAuthFlows{}
		result.Flow = flows
		scopesMap := make(map[string]string)
		for _, scope := range securityScheme.Scopes {
			scopesMap[scope] = ""
		}
		flow := &openapi3.OAuthFlow{
			AuthorizationURL: securityScheme.AuthorizationURL,
			TokenURL:         securityScheme.TokenURL,
			Scopes:           scopesMap,
		}
		switch securityScheme.Flow {
		case "implicit":
			flows.Implicit = flow
		case "accesscode":
			flows.AuthorizationCode = flow
		case "password":
			flows.Password = flow
		default:
			return nil
		}
		return result
	}
	return nil
}

func FromV3Swagger(swagger *openapi3.Swagger) *openapi2.Swagger {
	result := &openapi2.Swagger{
		Info: swagger.Info,
		Tags: swagger.Components.Tags,
	}
	isHTTPS := false
	isHTTP := false
	servers := swagger.Servers
	for i, server := range servers {
		parsedURL, err := url.Parse(server.URL)
		if err == nil {
			// See which schemes seem to be supported
			if parsedURL.Scheme == "https" {
				isHTTPS = true
			} else if parsedURL.Scheme == "http" {
				isHTTP = true
			}
			// The first server is assumed to provide the base path
			if i == 0 {
				result.Host = parsedURL.Host
				result.BasePath = parsedURL.Path
			}
		}
	}
	if isHTTPS {
		result.Schemes = append(result.Schemes, "https")
	}
	if isHTTP {
		result.Schemes = append(result.Schemes, "http")
	}
	if paths := swagger.Paths; paths != nil {
		for path, pathItem := range paths {
			if pathItem == nil {
				continue
			}
			for method, operation := range pathItem.Operations() {
				if operation == nil {
					continue
				}
				result.AddOperation(path, method, FromV3Operation(swagger, operation))
			}
		}
	}
	if m := swagger.Components.SecuritySchemes; m != nil {
		resultSecuritySchemes := make(map[string]*openapi2.SecurityScheme)
		for id, securityScheme := range m {
			resultSecuritySchemes[id] = FromV3SecurityScheme(swagger, securityScheme)
		}
		result.SecurityDefinitions = resultSecuritySchemes
	}
	result.Security = FromV3SecurityRequirements(swagger.Security)
	return result
}

func FromV3SecurityRequirements(requirements openapi3.SecurityRequirements) openapi2.SecurityRequirements {
	if requirements == nil {
		return nil
	}
	result := make([]map[string][]string, len(requirements))
	for i, item := range requirements {
		result[i] = item
	}
	return result
}

func FromV3PathItem(swagger *openapi3.Swagger, pathItem *openapi3.PathItem) *openapi2.PathItem {
	resultPathItem := &openapi2.PathItem{
		Delete:  FromV3Operation(swagger, pathItem.Delete),
		Get:     FromV3Operation(swagger, pathItem.Get),
		Head:    FromV3Operation(swagger, pathItem.Head),
		Options: FromV3Operation(swagger, pathItem.Options),
		Patch:   FromV3Operation(swagger, pathItem.Patch),
		Post:    FromV3Operation(swagger, pathItem.Post),
		Put:     FromV3Operation(swagger, pathItem.Put),
	}
	for _, parameter := range pathItem.Parameters {
		resultPathItem.Parameters = append(resultPathItem.Parameters,
			FromV3Parameter(parameter))
	}
	return resultPathItem
}

func findNameForRequestBody(operation *openapi3.Operation) string {
nameSearch:
	for _, name := range attemptedBodyParameterNames {
		for _, p := range operation.Parameters {
			if p.Name == name {
				continue nameSearch
			}
		}
		return name
	}
	return ""
}

func FromV3Operation(swagger *openapi3.Swagger, operation *openapi3.Operation) *openapi2.Operation {
	if operation == nil {
		return nil
	}
	result := &openapi2.Operation{
		Description: operation.Description,
		Tags:        operation.Tags,
	}
	if v := operation.Security; v != nil {
		resultSecurity := FromV3SecurityRequirements(*v)
		result.Security = &resultSecurity
	}
	for _, parameter := range operation.Parameters {
		result.Parameters = append(result.Parameters, FromV3Parameter(parameter))
	}
	if requestBody := operation.RequestBody; requestBody != nil {
		// Find parameter name that we can use for the body
		name := findNameForRequestBody(operation)

		// If found an available name
		if name != "" {
			resultParameter := &openapi2.Parameter{
				In:          "body",
				Name:        name,
				Description: requestBody.Description,
				Required:    requestBody.Required,
			}

			// Add JSON schema
			contentType := requestBody.GetContentType("application/json")
			if contentType != nil {
				resultParameter.Schema = contentType.Schema
			}

			// OK
			result.Parameters = append(result.Parameters, resultParameter)
		}
	}
	if responses := operation.Responses; responses != nil {
		resultResponses := make(map[string]*openapi2.Response, len(responses))
		result.Responses = resultResponses
		for k, response := range responses {
			resultResponses[k] = FromV3Response(response)
		}
	}
	return result
}

func FromV3Parameter(parameter *openapi3.Parameter) *openapi2.Parameter {
	if parameter == nil {
		return nil
	}
	result := &openapi2.Parameter{
		Description: parameter.Description,
		In:          parameter.In,
		Name:        parameter.Name,
		Required:    parameter.Required,
		Schema:      parameter.Schema,
	}
	if schema := parameter.Schema; schema != nil {
		result.Type = schema.Type
		result.Format = schema.Format
		result.Enum = schema.Enum
		result.Minimum = schema.Min
		result.Maximum = schema.Min
		result.ExclusiveMinimum = schema.ExclusiveMin
		result.ExclusiveMaximum = schema.ExclusiveMax
		result.MinLength = schema.MinLength
		result.MaxLength = schema.MaxLength
		result.Pattern = schema.Pattern
	}
	return result
}

func FromV3Response(response *openapi3.Response) *openapi2.Response {
	result := &openapi2.Response{
		Description: response.Description,
	}
	if content := response.Content; content != nil {
		if ct := content["application/json"]; ct != nil {
			result.Schema = ct.Schema
		}
	}
	return result
}

func FromV3SecurityScheme(swagger *openapi3.Swagger, securityScheme *openapi3.SecurityScheme) *openapi2.SecurityScheme {
	if securityScheme == nil {
		return nil
	}
	result := &openapi2.SecurityScheme{
		Description: securityScheme.Description,
	}
	switch securityScheme.Type {
	case "http":
		switch securityScheme.Scheme {
		case "basic":
			result.Type = "basic"
			return result
		default:
			result.Type = "apiKey"
			result.In = "header"
			result.Name = "Authorization"
			return result
		}
	case "apiKey":
		result.Type = "apiKey"
		result.In = securityScheme.In
		result.Name = securityScheme.Name
		return result
	case "oauth2":
		result.Type = "oauth2"
		flows := securityScheme.Flow
		if flows == nil {
			return nil
		}
		var flow *openapi3.OAuthFlow
		// TODO: Is this the right priority? What if multiple defined?
		if flow = flows.Implicit; flow != nil {
			result.Flow = "implicit"
		} else if flow = flows.AuthorizationCode; flow != nil {
			result.Flow = "accesscode"
		} else if flow = flows.Password; flow != nil {
			result.Flow = "password"
		} else {
			return nil
		}
		for scope := range flow.Scopes {
			result.Scopes = append(result.Scopes, scope)
		}
	}
	return nil
}

var attemptedBodyParameterNames = []string{
	"body",
	"requestBody",
}
