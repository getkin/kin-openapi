// Package openapi2conv converts an OpenAPI v2 specification to v3.
package openapi2conv

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi3"
)

// ToV3Swagger converts an OpenAPIv2 spec to an OpenAPIv3 spec
func ToV3Swagger(swagger *openapi2.Swagger) (*openapi3.Swagger, error) {
	stripNonCustomExtensions(swagger.Extensions)

	result := &openapi3.Swagger{
		OpenAPI:        "3.0.3",
		Info:           &swagger.Info,
		Components:     openapi3.Components{},
		Tags:           swagger.Tags,
		ExtensionProps: swagger.ExtensionProps,
		ExternalDocs:   swagger.ExternalDocs,
	}

	if host := swagger.Host; host != "" {
		schemes := swagger.Schemes
		if len(schemes) == 0 {
			schemes = []string{"https://"}
		}
		basePath := swagger.BasePath
		for _, scheme := range schemes {
			u := url.URL{
				Scheme: scheme,
				Host:   host,
				Path:   basePath,
			}
			result.AddServer(&openapi3.Server{URL: u.String()})
		}
	}

	result.Components.Schemas = make(map[string]*openapi3.SchemaRef)
	if parameters := swagger.Parameters; len(parameters) != 0 {
		result.Components.Parameters = make(map[string]*openapi3.ParameterRef)
		result.Components.RequestBodies = make(map[string]*openapi3.RequestBodyRef)
		for k, parameter := range parameters {
			v3Parameter, v3RequestBody, v3SchemaMap, err := ToV3Parameter(&result.Components, parameter, swagger.Consumes)
			switch {
			case err != nil:
				return nil, err
			case v3RequestBody != nil:
				result.Components.RequestBodies[k] = v3RequestBody
			case v3SchemaMap != nil:
				for _, v3Schema := range v3SchemaMap {
					result.Components.Schemas[k] = v3Schema
				}
			default:
				result.Components.Parameters[k] = v3Parameter
			}
		}
	}

	if paths := swagger.Paths; len(paths) != 0 {
		resultPaths := make(map[string]*openapi3.PathItem, len(paths))
		for path, pathItem := range paths {
			r, err := ToV3PathItem(swagger, &result.Components, pathItem, swagger.Consumes)
			if err != nil {
				return nil, err
			}
			resultPaths[path] = r
		}
		result.Paths = resultPaths
	}

	if responses := swagger.Responses; len(responses) != 0 {
		result.Components.Responses = make(map[string]*openapi3.ResponseRef, len(responses))
		for k, response := range responses {
			r, err := ToV3Response(response)
			if err != nil {
				return nil, err
			}
			result.Components.Responses[k] = r
		}
	}

	for key, schema := range ToV3Schemas(swagger.Definitions) {
		result.Components.Schemas[key] = schema
	}

	if m := swagger.SecurityDefinitions; len(m) != 0 {
		resultSecuritySchemes := make(map[string]*openapi3.SecuritySchemeRef)
		for k, v := range m {
			r, err := ToV3SecurityScheme(v)
			if err != nil {
				return nil, err
			}
			resultSecuritySchemes[k] = r
		}
		result.Components.SecuritySchemes = resultSecuritySchemes
	}

	result.Security = ToV3SecurityRequirements(swagger.Security)
	{
		sl := openapi3.NewSwaggerLoader()
		if err := sl.ResolveRefsIn(result, nil); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func ToV3PathItem(swagger *openapi2.Swagger, components *openapi3.Components, pathItem *openapi2.PathItem, consumes []string) (*openapi3.PathItem, error) {
	stripNonCustomExtensions(pathItem.Extensions)
	result := &openapi3.PathItem{
		ExtensionProps: pathItem.ExtensionProps,
	}
	for method, operation := range pathItem.Operations() {
		resultOperation, err := ToV3Operation(swagger, components, pathItem, operation, consumes)
		if err != nil {
			return nil, err
		}
		result.SetOperation(method, resultOperation)
	}
	for _, parameter := range pathItem.Parameters {
		v3Parameter, v3RequestBody, v3Schema, err := ToV3Parameter(components, parameter, consumes)
		switch {
		case err != nil:
			return nil, err
		case v3RequestBody != nil:
			return nil, errors.New("pathItem must not have a body parameter")
		case v3Schema != nil:
			return nil, errors.New("pathItem must not have a schema parameter")
		default:
			result.Parameters = append(result.Parameters, v3Parameter)
		}
	}
	return result, nil
}

func ToV3Operation(swagger *openapi2.Swagger, components *openapi3.Components, pathItem *openapi2.PathItem, operation *openapi2.Operation, consumes []string) (*openapi3.Operation, error) {
	if operation == nil {
		return nil, nil
	}
	stripNonCustomExtensions(operation.Extensions)
	result := &openapi3.Operation{
		OperationID:    operation.OperationID,
		Summary:        operation.Summary,
		Description:    operation.Description,
		Tags:           operation.Tags,
		ExtensionProps: operation.ExtensionProps,
	}
	if v := operation.Security; v != nil {
		resultSecurity := ToV3SecurityRequirements(*v)
		result.Security = &resultSecurity
	}

	if len(operation.Consumes) > 0 {
		consumes = operation.Consumes
	}

	var reqBodies []*openapi3.RequestBodyRef
	formDataSchemas := make(map[string]*openapi3.SchemaRef)
	for _, parameter := range operation.Parameters {
		v3Parameter, v3RequestBody, v3SchemaMap, err := ToV3Parameter(components, parameter, consumes)
		switch {
		case err != nil:
			return nil, err
		case v3RequestBody != nil:
			reqBodies = append(reqBodies, v3RequestBody)
		case v3SchemaMap != nil:
			for key, v3Schema := range v3SchemaMap {
				formDataSchemas[key] = v3Schema
			}
		default:
			result.Parameters = append(result.Parameters, v3Parameter)
		}
	}
	var err error
	if result.RequestBody, err = onlyOneReqBodyParam(reqBodies, formDataSchemas, components, consumes); err != nil {
		return nil, err
	}

	if responses := operation.Responses; responses != nil {
		resultResponses := make(openapi3.Responses, len(responses))
		for k, response := range responses {
			result, err := ToV3Response(response)
			if err != nil {
				return nil, err
			}
			resultResponses[k] = result
		}
		result.Responses = resultResponses
	}
	return result, nil
}

func getParameterNameFromOldRef(ref string) string {
	cleanPath := strings.TrimPrefix(ref, "#/parameters/")
	pathSections := strings.SplitN(cleanPath, "/", 1)

	return pathSections[0]
}

func ToV3Parameter(components *openapi3.Components, parameter *openapi2.Parameter, consumes []string) (*openapi3.ParameterRef, *openapi3.RequestBodyRef, map[string]*openapi3.SchemaRef, error) {
	if ref := parameter.Ref; ref != "" {
		if strings.HasPrefix(ref, "#/parameters/") {
			name := getParameterNameFromOldRef(ref)
			if _, ok := components.RequestBodies[name]; ok {
				v3Ref := strings.Replace(ref, "#/parameters/", "#/components/requestBodies/", 1)
				return nil, &openapi3.RequestBodyRef{Ref: v3Ref}, nil, nil
			} else if schema, ok := components.Schemas[name]; ok {
				schemaRefMap := make(map[string]*openapi3.SchemaRef)
				if val, ok := schema.Value.Extensions["x-formData-name"]; ok {
					name = val.(string)
				}
				v3Ref := strings.Replace(ref, "#/parameters/", "#/components/schemas/", 1)
				schemaRefMap[name] = &openapi3.SchemaRef{Ref: v3Ref}
				return nil, nil, schemaRefMap, nil
			}
		}
		return &openapi3.ParameterRef{Ref: ToV3Ref(ref)}, nil, nil, nil
	}
	stripNonCustomExtensions(parameter.Extensions)

	switch parameter.In {
	case "body":
		result := &openapi3.RequestBody{
			Description:    parameter.Description,
			Required:       parameter.Required,
			ExtensionProps: parameter.ExtensionProps,
		}
		if parameter.Name != "" {
			result.Extensions["x-originalParamName"] = parameter.Name
		}

		if schemaRef := parameter.Schema; schemaRef != nil {
			// Assuming JSON
			result.WithSchemaRef(ToV3SchemaRef(schemaRef), consumes)
		}
		return nil, &openapi3.RequestBodyRef{Value: result}, nil, nil

	case "formData":
		format, typ := parameter.Format, parameter.Type
		if typ == "file" {
			format, typ = "binary", "string"
		}
		if parameter.ExtensionProps.Extensions == nil {
			parameter.ExtensionProps.Extensions = make(map[string]interface{})
		}
		parameter.ExtensionProps.Extensions["x-formData-name"] = parameter.Name
		var required []string
		if parameter.Required {
			required = []string{parameter.Name}
		}
		schemaRef := &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Description:     parameter.Description,
				Type:            typ,
				ExtensionProps:  parameter.ExtensionProps,
				Format:          format,
				Enum:            parameter.Enum,
				Min:             parameter.Minimum,
				Max:             parameter.Maximum,
				ExclusiveMin:    parameter.ExclusiveMin,
				ExclusiveMax:    parameter.ExclusiveMax,
				MinLength:       parameter.MinLength,
				MaxLength:       parameter.MaxLength,
				Default:         parameter.Default,
				Items:           parameter.Items,
				MinItems:        parameter.MinItems,
				MaxItems:        parameter.MaxItems,
				Pattern:         parameter.Pattern,
				AllowEmptyValue: parameter.AllowEmptyValue,
				UniqueItems:     parameter.UniqueItems,
				MultipleOf:      parameter.MultipleOf,
				Required:        required,
			},
		}
		schemaRefMap := make(map[string]*openapi3.SchemaRef)
		schemaRefMap[parameter.Name] = schemaRef
		return nil, nil, schemaRefMap, nil

	default:
		required := parameter.Required
		if parameter.In == openapi3.ParameterInPath {
			required = true
		}

		result := &openapi3.Parameter{
			In:             parameter.In,
			Name:           parameter.Name,
			Description:    parameter.Description,
			Required:       required,
			ExtensionProps: parameter.ExtensionProps,
			Schema: ToV3SchemaRef(&openapi3.SchemaRef{Value: &openapi3.Schema{
				Type:            parameter.Type,
				Format:          parameter.Format,
				Enum:            parameter.Enum,
				Min:             parameter.Minimum,
				Max:             parameter.Maximum,
				ExclusiveMin:    parameter.ExclusiveMin,
				ExclusiveMax:    parameter.ExclusiveMax,
				MinLength:       parameter.MinLength,
				MaxLength:       parameter.MaxLength,
				Default:         parameter.Default,
				Items:           parameter.Items,
				MinItems:        parameter.MinItems,
				MaxItems:        parameter.MaxItems,
				Pattern:         parameter.Pattern,
				AllowEmptyValue: parameter.AllowEmptyValue,
				UniqueItems:     parameter.UniqueItems,
				MultipleOf:      parameter.MultipleOf,
			}}),
		}
		return &openapi3.ParameterRef{Value: result}, nil, nil, nil
	}
}

func formDataBody(bodies map[string]*openapi3.SchemaRef, reqs map[string]bool, consumes []string) *openapi3.RequestBodyRef {
	if len(bodies) != len(reqs) {
		panic(`request bodies and them being required must match`)
	}
	requireds := make([]string, 0, len(reqs))
	for propName, req := range reqs {
		if _, ok := bodies[propName]; !ok {
			panic(`request bodies and them being required must match`)
		}
		if req {
			requireds = append(requireds, propName)
		}
	}
	schema := &openapi3.Schema{
		Type:       "object",
		Properties: ToV3Schemas(bodies),
		Required:   requireds,
	}
	return &openapi3.RequestBodyRef{
		Value: openapi3.NewRequestBody().WithSchema(schema, consumes),
	}
}

func getParameterNameFromNewRef(ref string) string {
	cleanPath := strings.TrimPrefix(ref, "#/components/schemas/")
	pathSections := strings.SplitN(cleanPath, "/", 1)

	return pathSections[0]
}

func onlyOneReqBodyParam(bodies []*openapi3.RequestBodyRef, formDataSchemas map[string]*openapi3.SchemaRef, components *openapi3.Components, consumes []string) (*openapi3.RequestBodyRef, error) {
	if len(bodies) > 1 {
		return nil, errors.New("multiple body parameters cannot exist for the same operation")
	}

	if len(bodies) != 0 && len(formDataSchemas) != 0 {
		return nil, errors.New("body and form parameters cannot exist together for the same operation")
	}

	for _, requestBodyRef := range bodies {
		return requestBodyRef, nil
	}

	if len(formDataSchemas) > 0 {
		formDataParams := make(map[string]*openapi3.SchemaRef, len(formDataSchemas))
		formDataReqs := make(map[string]bool, len(formDataSchemas))
		for formDataName, formDataSchema := range formDataSchemas {
			if formDataSchema.Ref != "" {
				name := getParameterNameFromNewRef(formDataSchema.Ref)
				if schema := components.Schemas[name]; schema != nil && schema.Value != nil {
					if tempName, ok := schema.Value.Extensions["x-formData-name"]; ok {
						name = tempName.(string)
					}
					formDataParams[name] = formDataSchema
					formDataReqs[name] = false
					for _, req := range schema.Value.Required {
						if name == req {
							formDataReqs[name] = true
						}
					}
				}
			} else if formDataSchema.Value != nil {
				formDataParams[formDataName] = formDataSchema
				formDataReqs[formDataName] = false
				for _, req := range formDataSchema.Value.Required {
					if formDataName == req {
						formDataReqs[formDataName] = true
					}
				}
			}
		}

		return formDataBody(formDataParams, formDataReqs, consumes), nil
	}

	return nil, nil
}

func ToV3Response(response *openapi2.Response) (*openapi3.ResponseRef, error) {
	if ref := response.Ref; ref != "" {
		return &openapi3.ResponseRef{Ref: ToV3Ref(ref)}, nil
	}
	stripNonCustomExtensions(response.Extensions)
	result := &openapi3.Response{
		Description:    &response.Description,
		ExtensionProps: response.ExtensionProps,
	}
	if schemaRef := response.Schema; schemaRef != nil {
		result.WithJSONSchemaRef(ToV3SchemaRef(schemaRef))
	}
	return &openapi3.ResponseRef{
		Value: result,
	}, nil
}

func ToV3Schemas(defs map[string]*openapi3.SchemaRef) map[string]*openapi3.SchemaRef {
	schemas := make(map[string]*openapi3.SchemaRef, len(defs))
	for name, schema := range defs {
		schemas[name] = ToV3SchemaRef(schema)
	}
	return schemas
}

func ToV3SchemaRef(schema *openapi3.SchemaRef) *openapi3.SchemaRef {
	if ref := schema.Ref; ref != "" {
		return &openapi3.SchemaRef{Ref: ToV3Ref(ref)}
	}
	if schema.Value == nil {
		return schema
	}
	if schema.Value.Items != nil {
		schema.Value.Items = ToV3SchemaRef(schema.Value.Items)
	}
	for k, v := range schema.Value.Properties {
		schema.Value.Properties[k] = ToV3SchemaRef(v)
	}
	if v := schema.Value.AdditionalProperties; v != nil {
		schema.Value.AdditionalProperties = ToV3SchemaRef(v)
	}
	for i, v := range schema.Value.AllOf {
		schema.Value.AllOf[i] = ToV3SchemaRef(v)
	}
	return schema
}

var ref2To3 = map[string]string{
	"#/definitions/": "#/components/schemas/",
	"#/responses/":   "#/components/responses/",
	"#/parameters/":  "#/components/parameters/",
}

func ToV3Ref(ref string) string {
	for old, new := range ref2To3 {
		if strings.HasPrefix(ref, old) {
			ref = strings.Replace(ref, old, new, 1)
		}
	}
	return ref
}

func FromV3Ref(ref string) string {
	for new, old := range ref2To3 {
		if strings.HasPrefix(ref, old) {
			ref = strings.Replace(ref, old, new, 1)
		} else if strings.HasPrefix(ref, "#/components/requestBodies/") {
			ref = strings.Replace(ref, "#/components/requestBodies/", "#/parameters/", 1)
		}
	}
	return ref
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

func ToV3SecurityScheme(securityScheme *openapi2.SecurityScheme) (*openapi3.SecuritySchemeRef, error) {
	if securityScheme == nil {
		return nil, nil
	}
	stripNonCustomExtensions(securityScheme.Extensions)
	result := &openapi3.SecurityScheme{
		Description:    securityScheme.Description,
		ExtensionProps: securityScheme.ExtensionProps,
	}
	switch securityScheme.Type {
	case "basic":
		result.Type = "http"
		result.Scheme = "basic"
	case "apiKey":
		result.Type = "apiKey"
		result.In = securityScheme.In
		result.Name = securityScheme.Name
	case "oauth2":
		result.Type = "oauth2"
		flows := &openapi3.OAuthFlows{}
		result.Flows = flows
		scopesMap := make(map[string]string)
		for scope, desc := range securityScheme.Scopes {
			scopesMap[scope] = desc
		}
		flow := &openapi3.OAuthFlow{
			AuthorizationURL: securityScheme.AuthorizationURL,
			TokenURL:         securityScheme.TokenURL,
			Scopes:           scopesMap,
		}
		switch securityScheme.Flow {
		case "implicit":
			flows.Implicit = flow
		case "accessCode":
			flows.AuthorizationCode = flow
		case "password":
			flows.Password = flow
		default:
			return nil, fmt.Errorf("Unsupported flow '%s'", securityScheme.Flow)
		}
	}
	return &openapi3.SecuritySchemeRef{
		Ref:   ToV3Ref(securityScheme.Ref),
		Value: result,
	}, nil
}

// FromV3Swagger converts an OpenAPIv3 spec to an OpenAPIv2 spec
func FromV3Swagger(swagger *openapi3.Swagger) (*openapi2.Swagger, error) {
	resultResponses, err := FromV3Responses(swagger.Components.Responses, &swagger.Components)
	if err != nil {
		return nil, err
	}
	stripNonCustomExtensions(swagger.Extensions)
	schemas, parameters := FromV3Schemas(swagger.Components.Schemas, &swagger.Components)
	result := &openapi2.Swagger{
		Swagger:        "2.0",
		Info:           *swagger.Info,
		Definitions:    schemas,
		Parameters:     parameters,
		Responses:      resultResponses,
		Tags:           swagger.Tags,
		ExtensionProps: swagger.ExtensionProps,
		ExternalDocs:   swagger.ExternalDocs,
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
	for path, pathItem := range swagger.Paths {
		if pathItem == nil {
			continue
		}
		result.AddOperation(path, "GET", nil)
		stripNonCustomExtensions(pathItem.Extensions)
		addPathExtensions(result, path, pathItem.ExtensionProps)
		for method, operation := range pathItem.Operations() {
			if operation == nil {
				continue
			}
			resultOperation, err := FromV3Operation(swagger, operation)
			if err != nil {
				return nil, err
			}
			result.AddOperation(path, method, resultOperation)
		}
		params := openapi2.Parameters{}
		for _, param := range pathItem.Parameters {
			p, err := FromV3Parameter(param, &swagger.Components)
			if err != nil {
				return nil, err
			}
			params = append(params, p)
		}
		sort.Sort(params)
		result.Paths[path].Parameters = params
	}

	for name, param := range swagger.Components.Parameters {
		if result.Parameters[name], err = FromV3Parameter(param, &swagger.Components); err != nil {
			return nil, err
		}
	}

	for name, requestBodyRef := range swagger.Components.RequestBodies {
		bodyOrRefParameters, formDataParameters, consumes, err := fromV3RequestBodies(name, requestBodyRef, &swagger.Components)
		if err != nil {
			return nil, err
		}
		if len(formDataParameters) != 0 {
			for _, param := range formDataParameters {
				result.Parameters[param.Name] = param
			}
		} else if len(bodyOrRefParameters) != 0 {
			for _, param := range bodyOrRefParameters {
				result.Parameters[name] = param
			}
		}

		if len(consumes) != 0 {
			result.Consumes = consumesToArray(consumes)
		}
	}

	if m := swagger.Components.SecuritySchemes; m != nil {
		resultSecuritySchemes := make(map[string]*openapi2.SecurityScheme)
		for id, securityScheme := range m {
			v, err := FromV3SecurityScheme(swagger, securityScheme)
			if err != nil {
				return nil, err
			}
			resultSecuritySchemes[id] = v
		}
		result.SecurityDefinitions = resultSecuritySchemes
	}
	result.Security = FromV3SecurityRequirements(swagger.Security)
	return result, nil
}

func consumesToArray(consumes map[string]struct{}) []string {
	consumesArr := make([]string, 0, len(consumes))
	for key := range consumes {
		consumesArr = append(consumesArr, key)
	}
	sort.Strings(consumesArr)
	return consumesArr
}

func fromV3RequestBodies(name string, requestBodyRef *openapi3.RequestBodyRef, components *openapi3.Components) (
	bodyOrRefParameters openapi2.Parameters,
	formParameters openapi2.Parameters,
	consumes map[string]struct{},
	err error,
) {
	if ref := requestBodyRef.Ref; ref != "" {
		bodyOrRefParameters = append(bodyOrRefParameters, &openapi2.Parameter{Ref: FromV3Ref(ref)})
		return
	}

	//Only select one formData or request body for an individual requesstBody as swagger 2 does not support multiples
	if requestBodyRef.Value != nil {
		for contentType, mediaType := range requestBodyRef.Value.Content {
			if consumes == nil {
				consumes = make(map[string]struct{})
			}
			consumes[contentType] = struct{}{}
			if formParams := FromV3RequestBodyFormData(mediaType); len(formParams) != 0 {
				formParameters = formParams
			} else {
				paramName := name
				if originalName, ok := requestBodyRef.Value.Extensions["x-originalParamName"]; ok {
					json.Unmarshal(originalName.(json.RawMessage), &paramName)
				}

				var r *openapi2.Parameter
				if r, err = FromV3RequestBody(paramName, requestBodyRef, mediaType, components); err != nil {
					return
				}
				bodyOrRefParameters = append(bodyOrRefParameters, r)
			}
		}
	}
	return
}

func FromV3Schemas(schemas map[string]*openapi3.SchemaRef, components *openapi3.Components) (map[string]*openapi3.SchemaRef, map[string]*openapi2.Parameter) {
	v2Defs := make(map[string]*openapi3.SchemaRef)
	v2Params := make(map[string]*openapi2.Parameter)
	for name, schema := range schemas {
		schemaConv, parameterConv := FromV3SchemaRef(schema, components)
		if schemaConv != nil {
			v2Defs[name] = schemaConv
		} else if parameterConv != nil {
			if parameterConv.Name == "" {
				parameterConv.Name = name
			}
			v2Params[name] = parameterConv
		}
	}
	return v2Defs, v2Params
}

func FromV3SchemaRef(schema *openapi3.SchemaRef, components *openapi3.Components) (*openapi3.SchemaRef, *openapi2.Parameter) {
	if ref := schema.Ref; ref != "" {
		name := getParameterNameFromNewRef(ref)
		if val, ok := components.Schemas[name]; ok {
			if val.Value.Format == "binary" {
				v2Ref := strings.Replace(ref, "#/components/schemas/", "#/parameters/", 1)
				return nil, &openapi2.Parameter{Ref: v2Ref}
			}
		}

		return &openapi3.SchemaRef{Ref: FromV3Ref(ref)}, nil
	}
	if schema.Value == nil {
		return schema, nil
	}

	if schema.Value != nil {
		if schema.Value.Type == "string" && schema.Value.Format == "binary" {
			paramType := "file"
			required := false

			value, _ := schema.Value.Extensions["x-formData-name"]
			var originalName string
			json.Unmarshal(value.(json.RawMessage), &originalName)
			for _, prop := range schema.Value.Required {
				if originalName == prop {
					required = true
				}
			}
			return nil, &openapi2.Parameter{
				In:              "formData",
				Name:            originalName,
				Description:     schema.Value.Description,
				Type:            paramType,
				Enum:            schema.Value.Enum,
				Minimum:         schema.Value.Min,
				Maximum:         schema.Value.Max,
				ExclusiveMin:    schema.Value.ExclusiveMin,
				ExclusiveMax:    schema.Value.ExclusiveMax,
				MinLength:       schema.Value.MinLength,
				MaxLength:       schema.Value.MaxLength,
				Default:         schema.Value.Default,
				Items:           schema.Value.Items,
				MinItems:        schema.Value.MinItems,
				MaxItems:        schema.Value.MaxItems,
				AllowEmptyValue: schema.Value.AllowEmptyValue,
				UniqueItems:     schema.Value.UniqueItems,
				MultipleOf:      schema.Value.MultipleOf,
				ExtensionProps:  schema.Value.ExtensionProps,
				Required:        required,
			}
		}
	}
	if v := schema.Value.Items; v != nil {
		schema.Value.Items, _ = FromV3SchemaRef(v, components)
	}
	keys := make([]string, 0, len(schema.Value.Properties))
	for k := range schema.Value.Properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		schema.Value.Properties[key], _ = FromV3SchemaRef(schema.Value.Properties[key], components)
	}
	if v := schema.Value.AdditionalProperties; v != nil {
		schema.Value.AdditionalProperties, _ = FromV3SchemaRef(v, components)
	}
	for i, v := range schema.Value.AllOf {
		schema.Value.AllOf[i], _ = FromV3SchemaRef(v, components)
	}
	return schema, nil
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

func FromV3PathItem(swagger *openapi3.Swagger, pathItem *openapi3.PathItem) (*openapi2.PathItem, error) {
	stripNonCustomExtensions(pathItem.Extensions)
	result := &openapi2.PathItem{
		ExtensionProps: pathItem.ExtensionProps,
	}
	for method, operation := range pathItem.Operations() {
		r, err := FromV3Operation(swagger, operation)
		if err != nil {
			return nil, err
		}
		result.SetOperation(method, r)
	}
	for _, parameter := range pathItem.Parameters {
		p, err := FromV3Parameter(parameter, &swagger.Components)
		if err != nil {
			return nil, err
		}
		result.Parameters = append(result.Parameters, p)
	}
	return result, nil
}

func findNameForRequestBody(operation *openapi3.Operation) string {
nameSearch:
	for _, name := range attemptedBodyParameterNames {
		for _, parameterRef := range operation.Parameters {
			parameter := parameterRef.Value
			if parameter != nil && parameter.Name == name {
				continue nameSearch
			}
		}
		return name
	}
	return ""
}

func FromV3RequestBodyFormData(mediaType *openapi3.MediaType) openapi2.Parameters {
	parameters := openapi2.Parameters{}
	for propName, schemaRef := range mediaType.Schema.Value.Properties {
		if ref := schemaRef.Ref; ref != "" {
			v2Ref := strings.Replace(ref, "#/components/schemas/", "#/parameters/", 1)
			parameters = append(parameters, &openapi2.Parameter{Ref: v2Ref})
			continue
		}
		val := schemaRef.Value
		typ := val.Type
		if val.Format == "binary" {
			typ = "file"
		}
		required := false
		for _, name := range val.Required {
			if name == propName {
				required = true
				break
			}
		}
		parameter := &openapi2.Parameter{
			Name:           propName,
			Description:    val.Description,
			Type:           typ,
			In:             "formData",
			ExtensionProps: val.ExtensionProps,
			Enum:           val.Enum,
			ExclusiveMin:   val.ExclusiveMin,
			ExclusiveMax:   val.ExclusiveMax,
			MinLength:      val.MinLength,
			MaxLength:      val.MaxLength,
			Default:        val.Default,
			Items:          val.Items,
			MinItems:       val.MinItems,
			MaxItems:       val.MaxItems,
			Maximum:        val.Max,
			Minimum:        val.Min,
			Pattern:        val.Pattern,
			// CollectionFormat: val.CollectionFormat,
			// Format:          val.Format,
			AllowEmptyValue: val.AllowEmptyValue,
			Required:        required,
			UniqueItems:     val.UniqueItems,
			MultipleOf:      val.MultipleOf,
		}
		parameters = append(parameters, parameter)
	}
	return parameters
}

func FromV3Operation(swagger *openapi3.Swagger, operation *openapi3.Operation) (*openapi2.Operation, error) {
	if operation == nil {
		return nil, nil
	}
	stripNonCustomExtensions(operation.Extensions)
	result := &openapi2.Operation{
		OperationID:    operation.OperationID,
		Summary:        operation.Summary,
		Description:    operation.Description,
		Tags:           operation.Tags,
		ExtensionProps: operation.ExtensionProps,
	}
	if v := operation.Security; v != nil {
		resultSecurity := FromV3SecurityRequirements(*v)
		result.Security = &resultSecurity
	}
	for _, parameter := range operation.Parameters {
		r, err := FromV3Parameter(parameter, &swagger.Components)
		if err != nil {
			return nil, err
		}
		result.Parameters = append(result.Parameters, r)
	}
	if v := operation.RequestBody; v != nil {
		// Find parameter name that we can use for the body
		name := findNameForRequestBody(operation)
		if name == "" {
			return nil, errors.New("could not find a name for request body")
		}

		bodyOrRefParameters, formDataParameters, consumes, err := fromV3RequestBodies(name, v, &swagger.Components)
		if err != nil {
			return nil, err
		}
		if len(formDataParameters) != 0 {
			result.Parameters = append(result.Parameters, formDataParameters...)
		} else if len(bodyOrRefParameters) != 0 {
			for _, param := range bodyOrRefParameters {
				result.Parameters = append(result.Parameters, param)
				break // add a single request body
			}

		}

		if len(consumes) != 0 {
			result.Consumes = consumesToArray(consumes)
		}
	}
	sort.Sort(result.Parameters)

	if responses := operation.Responses; responses != nil {
		resultResponses, err := FromV3Responses(responses, &swagger.Components)
		if err != nil {
			return nil, err
		}
		result.Responses = resultResponses
	}
	return result, nil
}

func FromV3RequestBody(name string, requestBodyRef *openapi3.RequestBodyRef, mediaType *openapi3.MediaType, components *openapi3.Components) (*openapi2.Parameter, error) {
	requestBody := requestBodyRef.Value

	stripNonCustomExtensions(requestBody.Extensions)
	result := &openapi2.Parameter{
		In:             "body",
		Name:           name,
		Description:    requestBody.Description,
		Required:       requestBody.Required,
		ExtensionProps: requestBody.ExtensionProps,
	}

	if mediaType != nil {
		result.Schema, _ = FromV3SchemaRef(mediaType.Schema, components)
	}
	return result, nil
}

func FromV3Parameter(ref *openapi3.ParameterRef, components *openapi3.Components) (*openapi2.Parameter, error) {
	if ref := ref.Ref; ref != "" {
		return &openapi2.Parameter{Ref: FromV3Ref(ref)}, nil
	}
	parameter := ref.Value
	if parameter == nil {
		return nil, nil
	}
	stripNonCustomExtensions(parameter.Extensions)
	result := &openapi2.Parameter{
		Description:    parameter.Description,
		In:             parameter.In,
		Name:           parameter.Name,
		Required:       parameter.Required,
		ExtensionProps: parameter.ExtensionProps,
	}
	if schemaRef := parameter.Schema; schemaRef != nil {
		schemaRef, _ = FromV3SchemaRef(schemaRef, components)
		schema := schemaRef.Value
		result.Type = schema.Type
		result.Format = schema.Format
		result.Enum = schema.Enum
		result.Minimum = schema.Min
		result.Maximum = schema.Max
		result.ExclusiveMin = schema.ExclusiveMin
		result.ExclusiveMax = schema.ExclusiveMax
		result.MinLength = schema.MinLength
		result.MaxLength = schema.MaxLength
		result.Pattern = schema.Pattern
		result.Default = schema.Default
		result.Items = schema.Items
		result.MinItems = schema.MinItems
		result.MaxItems = schema.MaxItems
		result.AllowEmptyValue = schema.AllowEmptyValue
		// result.CollectionFormat = schema.CollectionFormat
		result.UniqueItems = schema.UniqueItems
		result.MultipleOf = schema.MultipleOf
	}
	return result, nil
}

func FromV3Responses(responses map[string]*openapi3.ResponseRef, components *openapi3.Components) (map[string]*openapi2.Response, error) {
	v2Responses := make(map[string]*openapi2.Response, len(responses))
	for k, response := range responses {
		r, err := FromV3Response(response, components)
		if err != nil {
			return nil, err
		}
		v2Responses[k] = r
	}
	return v2Responses, nil
}

func FromV3Response(ref *openapi3.ResponseRef, components *openapi3.Components) (*openapi2.Response, error) {
	if ref := ref.Ref; ref != "" {
		return &openapi2.Response{Ref: FromV3Ref(ref)}, nil
	}

	response := ref.Value
	if response == nil {
		return nil, nil
	}
	description := ""
	if desc := response.Description; desc != nil {
		description = *desc
	}
	stripNonCustomExtensions(response.Extensions)
	result := &openapi2.Response{
		Description:    description,
		ExtensionProps: response.ExtensionProps,
	}
	if content := response.Content; content != nil {
		if ct := content["application/json"]; ct != nil {
			result.Schema, _ = FromV3SchemaRef(ct.Schema, components)
		}
	}
	return result, nil
}

func FromV3SecurityScheme(swagger *openapi3.Swagger, ref *openapi3.SecuritySchemeRef) (*openapi2.SecurityScheme, error) {
	securityScheme := ref.Value
	if securityScheme == nil {
		return nil, nil
	}
	stripNonCustomExtensions(securityScheme.Extensions)
	result := &openapi2.SecurityScheme{
		Ref:            FromV3Ref(ref.Ref),
		Description:    securityScheme.Description,
		ExtensionProps: securityScheme.ExtensionProps,
	}
	switch securityScheme.Type {
	case "http":
		switch securityScheme.Scheme {
		case "basic":
			result.Type = "basic"
		default:
			result.Type = "apiKey"
			result.In = "header"
			result.Name = "Authorization"
		}
	case "apiKey":
		result.Type = "apiKey"
		result.In = securityScheme.In
		result.Name = securityScheme.Name
	case "oauth2":
		result.Type = "oauth2"
		flows := securityScheme.Flows
		if flows != nil {
			var flow *openapi3.OAuthFlow
			// TODO: Is this the right priority? What if multiple defined?
			if flow = flows.Implicit; flow != nil {
				result.Flow = "implicit"
			} else if flow = flows.AuthorizationCode; flow != nil {
				result.Flow = "accessCode"
			} else if flow = flows.Password; flow != nil {
				result.Flow = "password"
			} else {
				return nil, nil
			}
			for scope, desc := range flow.Scopes {
				result.Scopes[scope] = desc
			}
		}
	default:
		return nil, fmt.Errorf("Unsupported security scheme type '%s'", securityScheme.Type)
	}
	return result, nil
}

var attemptedBodyParameterNames = []string{
	"body",
	"requestBody",
}

func stripNonCustomExtensions(extensions map[string]interface{}) {
	for extName := range extensions {
		if !strings.HasPrefix(extName, "x-") {
			delete(extensions, extName)
		}
	}
}

func addPathExtensions(swagger *openapi2.Swagger, path string, extensionProps openapi3.ExtensionProps) {
	paths := swagger.Paths
	if paths == nil {
		paths = make(map[string]*openapi2.PathItem, 8)
		swagger.Paths = paths
	}
	pathItem := paths[path]
	if pathItem == nil {
		pathItem = &openapi2.PathItem{}
		paths[path] = pathItem
	}
	pathItem.ExtensionProps = extensionProps
}
