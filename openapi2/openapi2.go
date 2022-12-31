package openapi2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"github.com/getkin/kin-openapi/openapi3"
)

// T is the root of an OpenAPI v2 document
type T struct {
	Extensions map[string]interface{} `json:"-" yaml:"-"`

	Swagger             string                         `json:"swagger" yaml:"swagger"` // required
	Info                openapi3.Info                  `json:"info" yaml:"info"`       // required
	ExternalDocs        *openapi3.ExternalDocs         `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	Schemes             []string                       `json:"schemes,omitempty" yaml:"schemes,omitempty"`
	Consumes            []string                       `json:"consumes,omitempty" yaml:"consumes,omitempty"`
	Produces            []string                       `json:"produces,omitempty" yaml:"produces,omitempty"`
	Host                string                         `json:"host,omitempty" yaml:"host,omitempty"`
	BasePath            string                         `json:"basePath,omitempty" yaml:"basePath,omitempty"`
	Paths               map[string]*PathItem           `json:"paths,omitempty" yaml:"paths,omitempty"`
	Definitions         map[string]*openapi3.SchemaRef `json:"definitions,omitempty" yaml:"definitions,omitempty"`
	Parameters          map[string]*Parameter          `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Responses           map[string]*Response           `json:"responses,omitempty" yaml:"responses,omitempty"`
	SecurityDefinitions map[string]*SecurityScheme     `json:"securityDefinitions,omitempty" yaml:"securityDefinitions,omitempty"`
	Security            SecurityRequirements           `json:"security,omitempty" yaml:"security,omitempty"`
	Tags                openapi3.Tags                  `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// MarshalJSON returns the JSON encoding of T.
func (doc T) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{}, 15+len(doc.Extensions))
	for k, v := range doc.Extensions {
		m[k] = v
	}
	m["swagger"] = doc.Swagger
	m["info"] = doc.Info
	if x := doc.ExternalDocs; x != nil {
		m["externalDocs"] = x
	}
	if x := doc.Schemes; len(x) != 0 {
		m["schemes"] = x
	}
	if x := doc.Consumes; len(x) != 0 {
		m["consumes"] = x
	}
	if x := doc.Produces; len(x) != 0 {
		m["produces"] = x
	}
	if x := doc.Host; x != "" {
		m["host"] = x
	}
	if x := doc.BasePath; x != "" {
		m["basePath"] = x
	}
	if x := doc.Paths; len(x) != 0 {
		m["paths"] = x
	}
	if x := doc.Definitions; len(x) != 0 {
		m["definitions"] = x
	}
	if x := doc.Parameters; len(x) != 0 {
		m["parameters"] = x
	}
	if x := doc.Responses; len(x) != 0 {
		m["responses"] = x
	}
	if x := doc.SecurityDefinitions; len(x) != 0 {
		m["securityDefinitions"] = x
	}
	if x := doc.Security; len(x) != 0 {
		m["security"] = x
	}
	if x := doc.Tags; len(x) != 0 {
		m["tags"] = x
	}
	return json.Marshal(m)
}

// UnmarshalJSON sets T to a copy of data.
func (doc *T) UnmarshalJSON(data []byte) error {
	type TBis T
	var x TBis
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	_ = json.Unmarshal(data, &x.Extensions)
	delete(x.Extensions, "swagger")
	delete(x.Extensions, "info")
	delete(x.Extensions, "externalDocs")
	delete(x.Extensions, "schemes")
	delete(x.Extensions, "consumes")
	delete(x.Extensions, "produces")
	delete(x.Extensions, "host")
	delete(x.Extensions, "basePath")
	delete(x.Extensions, "paths")
	delete(x.Extensions, "definitions")
	delete(x.Extensions, "parameters")
	delete(x.Extensions, "responses")
	delete(x.Extensions, "securityDefinitions")
	delete(x.Extensions, "security")
	delete(x.Extensions, "tags")
	*doc = T(x)
	return nil
}

func (doc *T) AddOperation(path string, method string, operation *Operation) {
	if doc.Paths == nil {
		doc.Paths = make(map[string]*PathItem)
	}
	pathItem := doc.Paths[path]
	if pathItem == nil {
		pathItem = &PathItem{}
		doc.Paths[path] = pathItem
	}
	pathItem.SetOperation(method, operation)
}

type PathItem struct {
	Extensions map[string]interface{} `json:"-" yaml:"-"`

	Ref string `json:"$ref,omitempty" yaml:"$ref,omitempty"`

	Delete     *Operation `json:"delete,omitempty" yaml:"delete,omitempty"`
	Get        *Operation `json:"get,omitempty" yaml:"get,omitempty"`
	Head       *Operation `json:"head,omitempty" yaml:"head,omitempty"`
	Options    *Operation `json:"options,omitempty" yaml:"options,omitempty"`
	Patch      *Operation `json:"patch,omitempty" yaml:"patch,omitempty"`
	Post       *Operation `json:"post,omitempty" yaml:"post,omitempty"`
	Put        *Operation `json:"put,omitempty" yaml:"put,omitempty"`
	Parameters Parameters `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

// MarshalJSON returns the JSON encoding of PathItem.
func (pathItem PathItem) MarshalJSON() ([]byte, error) {
	if ref := pathItem.Ref; ref != "" {
		return json.Marshal(openapi3.Ref{Ref: ref})
	}

	m := make(map[string]interface{}, 8+len(pathItem.Extensions))
	for k, v := range pathItem.Extensions {
		m[k] = v
	}
	if x := pathItem.Delete; x != nil {
		m["delete"] = x
	}
	if x := pathItem.Get; x != nil {
		m["get"] = x
	}
	if x := pathItem.Head; x != nil {
		m["head"] = x
	}
	if x := pathItem.Options; x != nil {
		m["options"] = x
	}
	if x := pathItem.Patch; x != nil {
		m["patch"] = x
	}
	if x := pathItem.Post; x != nil {
		m["post"] = x
	}
	if x := pathItem.Put; x != nil {
		m["put"] = x
	}
	if x := pathItem.Parameters; len(x) != 0 {
		m["parameters"] = x
	}
	return json.Marshal(m)
}

// UnmarshalJSON sets PathItem to a copy of data.
func (pathItem *PathItem) UnmarshalJSON(data []byte) error {
	type PathItemBis PathItem
	var x PathItemBis
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	_ = json.Unmarshal(data, &x.Extensions)
	delete(x.Extensions, "$ref")
	delete(x.Extensions, "delete")
	delete(x.Extensions, "get")
	delete(x.Extensions, "head")
	delete(x.Extensions, "options")
	delete(x.Extensions, "patch")
	delete(x.Extensions, "post")
	delete(x.Extensions, "put")
	delete(x.Extensions, "parameters")
	*pathItem = PathItem(x)
	return nil
}

func (pathItem *PathItem) Operations() map[string]*Operation {
	operations := make(map[string]*Operation)
	if v := pathItem.Delete; v != nil {
		operations[http.MethodDelete] = v
	}
	if v := pathItem.Get; v != nil {
		operations[http.MethodGet] = v
	}
	if v := pathItem.Head; v != nil {
		operations[http.MethodHead] = v
	}
	if v := pathItem.Options; v != nil {
		operations[http.MethodOptions] = v
	}
	if v := pathItem.Patch; v != nil {
		operations[http.MethodPatch] = v
	}
	if v := pathItem.Post; v != nil {
		operations[http.MethodPost] = v
	}
	if v := pathItem.Put; v != nil {
		operations[http.MethodPut] = v
	}
	return operations
}

func (pathItem *PathItem) GetOperation(method string) *Operation {
	switch method {
	case http.MethodDelete:
		return pathItem.Delete
	case http.MethodGet:
		return pathItem.Get
	case http.MethodHead:
		return pathItem.Head
	case http.MethodOptions:
		return pathItem.Options
	case http.MethodPatch:
		return pathItem.Patch
	case http.MethodPost:
		return pathItem.Post
	case http.MethodPut:
		return pathItem.Put
	default:
		panic(fmt.Errorf("unsupported HTTP method %q", method))
	}
}

func (pathItem *PathItem) SetOperation(method string, operation *Operation) {
	switch method {
	case http.MethodDelete:
		pathItem.Delete = operation
	case http.MethodGet:
		pathItem.Get = operation
	case http.MethodHead:
		pathItem.Head = operation
	case http.MethodOptions:
		pathItem.Options = operation
	case http.MethodPatch:
		pathItem.Patch = operation
	case http.MethodPost:
		pathItem.Post = operation
	case http.MethodPut:
		pathItem.Put = operation
	default:
		panic(fmt.Errorf("unsupported HTTP method %q", method))
	}
}

type Operation struct {
	Extensions map[string]interface{} `json:"-" yaml:"-"`

	Summary      string                 `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description  string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Deprecated   bool                   `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	ExternalDocs *openapi3.ExternalDocs `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	Tags         []string               `json:"tags,omitempty" yaml:"tags,omitempty"`
	OperationID  string                 `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Parameters   Parameters             `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Responses    map[string]*Response   `json:"responses" yaml:"responses"`
	Consumes     []string               `json:"consumes,omitempty" yaml:"consumes,omitempty"`
	Produces     []string               `json:"produces,omitempty" yaml:"produces,omitempty"`
	Schemes      []string               `json:"schemes,omitempty" yaml:"schemes,omitempty"`
	Security     *SecurityRequirements  `json:"security,omitempty" yaml:"security,omitempty"`
}

// MarshalJSON returns the JSON encoding of Operation.
func (operation Operation) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{}, 12+len(operation.Extensions))
	for k, v := range operation.Extensions {
		m[k] = v
	}
	if x := operation.Summary; x != "" {
		m["summary"] = x
	}
	if x := operation.Description; x != "" {
		m["description"] = x
	}
	if x := operation.Deprecated; x {
		m["deprecated"] = x
	}
	if x := operation.ExternalDocs; x != nil {
		m["externalDocs"] = x
	}
	if x := operation.Tags; len(x) != 0 {
		m["tags"] = x
	}
	if x := operation.OperationID; x != "" {
		m["operationId"] = x
	}
	if x := operation.Parameters; len(x) != 0 {
		m["parameters"] = x
	}
	m["responses"] = operation.Responses
	if x := operation.Consumes; len(x) != 0 {
		m["consumes"] = x
	}
	if x := operation.Produces; len(x) != 0 {
		m["produces"] = x
	}
	if x := operation.Schemes; len(x) != 0 {
		m["schemes"] = x
	}
	if x := operation.Security; x != nil {
		m["security"] = x
	}
	return json.Marshal(m)
}

// UnmarshalJSON sets Operation to a copy of data.
func (operation *Operation) UnmarshalJSON(data []byte) error {
	type OperationBis Operation
	var x OperationBis
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	_ = json.Unmarshal(data, &x.Extensions)
	delete(x.Extensions, "summary")
	delete(x.Extensions, "description")
	delete(x.Extensions, "deprecated")
	delete(x.Extensions, "externalDocs")
	delete(x.Extensions, "tags")
	delete(x.Extensions, "operationId")
	delete(x.Extensions, "parameters")
	delete(x.Extensions, "responses")
	delete(x.Extensions, "consumes")
	delete(x.Extensions, "produces")
	delete(x.Extensions, "schemes")
	delete(x.Extensions, "security")
	*operation = Operation(x)
	return nil
}

type Parameters []*Parameter

var _ sort.Interface = Parameters{}

func (ps Parameters) Len() int      { return len(ps) }
func (ps Parameters) Swap(i, j int) { ps[i], ps[j] = ps[j], ps[i] }
func (ps Parameters) Less(i, j int) bool {
	if ps[i].Name != ps[j].Name {
		return ps[i].Name < ps[j].Name
	}
	if ps[i].In != ps[j].In {
		return ps[i].In < ps[j].In
	}
	return ps[i].Ref < ps[j].Ref
}

type Parameter struct {
	Extensions map[string]interface{} `json:"-" yaml:"-"`

	Ref string `json:"$ref,omitempty" yaml:"$ref,omitempty"`

	In               string              `json:"in,omitempty" yaml:"in,omitempty"`
	Name             string              `json:"name,omitempty" yaml:"name,omitempty"`
	Description      string              `json:"description,omitempty" yaml:"description,omitempty"`
	CollectionFormat string              `json:"collectionFormat,omitempty" yaml:"collectionFormat,omitempty"`
	Type             string              `json:"type,omitempty" yaml:"type,omitempty"`
	Format           string              `json:"format,omitempty" yaml:"format,omitempty"`
	Pattern          string              `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	AllowEmptyValue  bool                `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
	Required         bool                `json:"required,omitempty" yaml:"required,omitempty"`
	UniqueItems      bool                `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`
	ExclusiveMin     bool                `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`
	ExclusiveMax     bool                `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
	Schema           *openapi3.SchemaRef `json:"schema,omitempty" yaml:"schema,omitempty"`
	Items            *openapi3.SchemaRef `json:"items,omitempty" yaml:"items,omitempty"`
	Enum             []interface{}       `json:"enum,omitempty" yaml:"enum,omitempty"`
	MultipleOf       *float64            `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`
	Minimum          *float64            `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	Maximum          *float64            `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	MaxLength        *uint64             `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	MaxItems         *uint64             `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	MinLength        uint64              `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	MinItems         uint64              `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	Default          interface{}         `json:"default,omitempty" yaml:"default,omitempty"`
}

// MarshalJSON returns the JSON encoding of Parameter.
func (parameter Parameter) MarshalJSON() ([]byte, error) {
	if ref := parameter.Ref; ref != "" {
		return json.Marshal(openapi3.Ref{Ref: ref})
	}

	m := make(map[string]interface{}, 24+len(parameter.Extensions))
	for k, v := range parameter.Extensions {
		m[k] = v
	}

	if x := parameter.In; x != "" {
		m["in"] = x
	}
	if x := parameter.Name; x != "" {
		m["name"] = x
	}
	if x := parameter.Description; x != "" {
		m["description"] = x
	}
	if x := parameter.CollectionFormat; x != "" {
		m["collectionFormat"] = x
	}
	if x := parameter.Type; x != "" {
		m["type"] = x
	}
	if x := parameter.Format; x != "" {
		m["format"] = x
	}
	if x := parameter.Pattern; x != "" {
		m["pattern"] = x
	}
	if x := parameter.AllowEmptyValue; x {
		m["allowEmptyValue"] = x
	}
	if x := parameter.Required; x {
		m["required"] = x
	}
	if x := parameter.UniqueItems; x {
		m["uniqueItems"] = x
	}
	if x := parameter.ExclusiveMin; x {
		m["exclusiveMinimum"] = x
	}
	if x := parameter.ExclusiveMax; x {
		m["exclusiveMaximum"] = x
	}
	if x := parameter.Schema; x != nil {
		m["schema"] = x
	}
	if x := parameter.Items; x != nil {
		m["items"] = x
	}
	if x := parameter.Enum; x != nil {
		m["enum"] = x
	}
	if x := parameter.MultipleOf; x != nil {
		m["multipleOf"] = x
	}
	if x := parameter.Minimum; x != nil {
		m["minimum"] = x
	}
	if x := parameter.Maximum; x != nil {
		m["maximum"] = x
	}
	if x := parameter.MaxLength; x != nil {
		m["maxLength"] = x
	}
	if x := parameter.MaxItems; x != nil {
		m["maxItems"] = x
	}
	if x := parameter.MinLength; x != 0 {
		m["minLength"] = x
	}
	if x := parameter.MinItems; x != 0 {
		m["minItems"] = x
	}
	if x := parameter.Default; x != nil {
		m["default"] = x
	}

	return json.Marshal(m)
}

// UnmarshalJSON sets Parameter to a copy of data.
func (parameter *Parameter) UnmarshalJSON(data []byte) error {
	type ParameterBis Parameter
	var x ParameterBis
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	_ = json.Unmarshal(data, &x.Extensions)
	delete(x.Extensions, "$ref")

	delete(x.Extensions, "in")
	delete(x.Extensions, "name")
	delete(x.Extensions, "description")
	delete(x.Extensions, "collectionFormat")
	delete(x.Extensions, "type")
	delete(x.Extensions, "format")
	delete(x.Extensions, "pattern")
	delete(x.Extensions, "allowEmptyValue")
	delete(x.Extensions, "required")
	delete(x.Extensions, "uniqueItems")
	delete(x.Extensions, "exclusiveMinimum")
	delete(x.Extensions, "exclusiveMaximum")
	delete(x.Extensions, "schema")
	delete(x.Extensions, "items")
	delete(x.Extensions, "enum")
	delete(x.Extensions, "multipleOf")
	delete(x.Extensions, "minimum")
	delete(x.Extensions, "maximum")
	delete(x.Extensions, "maxLength")
	delete(x.Extensions, "maxItems")
	delete(x.Extensions, "minLength")
	delete(x.Extensions, "minItems")
	delete(x.Extensions, "default")

	*parameter = Parameter(x)
	return nil
}

type Response struct {
	Extensions map[string]interface{} `json:"-" yaml:"-"`

	Ref string `json:"$ref,omitempty" yaml:"$ref,omitempty"`

	Description string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Schema      *openapi3.SchemaRef    `json:"schema,omitempty" yaml:"schema,omitempty"`
	Headers     map[string]*Header     `json:"headers,omitempty" yaml:"headers,omitempty"`
	Examples    map[string]interface{} `json:"examples,omitempty" yaml:"examples,omitempty"`
}

// MarshalJSON returns the JSON encoding of Response.
func (response Response) MarshalJSON() ([]byte, error) {
	if ref := response.Ref; ref != "" {
		return json.Marshal(openapi3.Ref{Ref: ref})
	}

	m := make(map[string]interface{}, 4+len(response.Extensions))
	for k, v := range response.Extensions {
		m[k] = v
	}
	if x := response.Description; x != "" {
		m["description"] = x
	}
	if x := response.Schema; x != nil {
		m["schema"] = x
	}
	if x := response.Headers; len(x) != 0 {
		m["headers"] = x
	}
	if x := response.Examples; len(x) != 0 {
		m["examples"] = x
	}
	return json.Marshal(m)
}

// UnmarshalJSON sets Response to a copy of data.
func (response *Response) UnmarshalJSON(data []byte) error {
	type ResponseBis Response
	var x ResponseBis
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	_ = json.Unmarshal(data, &x.Extensions)
	delete(x.Extensions, "$ref")
	delete(x.Extensions, "description")
	delete(x.Extensions, "schema")
	delete(x.Extensions, "headers")
	delete(x.Extensions, "examples")
	*response = Response(x)
	return nil
}

type Header struct {
	Parameter
}

// MarshalJSON returns the JSON encoding of Header.
func (header Header) MarshalJSON() ([]byte, error) {
	return header.Parameter.MarshalJSON()
}

// UnmarshalJSON sets Header to a copy of data.
func (header *Header) UnmarshalJSON(data []byte) error {
	return header.Parameter.UnmarshalJSON(data)
}

type SecurityRequirements []map[string][]string

type SecurityScheme struct {
	Extensions map[string]interface{} `json:"-" yaml:"-"`

	Ref string `json:"$ref,omitempty" yaml:"$ref,omitempty"`

	Description      string            `json:"description,omitempty" yaml:"description,omitempty"`
	Type             string            `json:"type,omitempty" yaml:"type,omitempty"`
	In               string            `json:"in,omitempty" yaml:"in,omitempty"`
	Name             string            `json:"name,omitempty" yaml:"name,omitempty"`
	Flow             string            `json:"flow,omitempty" yaml:"flow,omitempty"`
	AuthorizationURL string            `json:"authorizationUrl,omitempty" yaml:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty" yaml:"tokenUrl,omitempty"`
	Scopes           map[string]string `json:"scopes,omitempty" yaml:"scopes,omitempty"`
	Tags             openapi3.Tags     `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// MarshalJSON returns the JSON encoding of SecurityScheme.
func (securityScheme SecurityScheme) MarshalJSON() ([]byte, error) {
	if ref := securityScheme.Ref; ref != "" {
		return json.Marshal(openapi3.Ref{Ref: ref})
	}

	m := make(map[string]interface{}, 10+len(securityScheme.Extensions))
	for k, v := range securityScheme.Extensions {
		m[k] = v
	}
	if x := securityScheme.Description; x != "" {
		m["description"] = x
	}
	if x := securityScheme.Type; x != "" {
		m["type"] = x
	}
	if x := securityScheme.In; x != "" {
		m["in"] = x
	}
	if x := securityScheme.Name; x != "" {
		m["name"] = x
	}
	if x := securityScheme.Flow; x != "" {
		m["flow"] = x
	}
	if x := securityScheme.AuthorizationURL; x != "" {
		m["authorizationUrl"] = x
	}
	if x := securityScheme.TokenURL; x != "" {
		m["tokenUrl"] = x
	}
	if x := securityScheme.Scopes; len(x) != 0 {
		m["scopes"] = x
	}
	if x := securityScheme.Tags; len(x) != 0 {
		m["tags"] = x
	}
	return json.Marshal(m)
}

// UnmarshalJSON sets SecurityScheme to a copy of data.
func (securityScheme *SecurityScheme) UnmarshalJSON(data []byte) error {
	type SecuritySchemeBis SecurityScheme
	var x SecuritySchemeBis
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	_ = json.Unmarshal(data, &x.Extensions)
	delete(x.Extensions, "$ref")
	delete(x.Extensions, "description")
	delete(x.Extensions, "type")
	delete(x.Extensions, "in")
	delete(x.Extensions, "name")
	delete(x.Extensions, "flow")
	delete(x.Extensions, "authorizationUrl")
	delete(x.Extensions, "tokenUrl")
	delete(x.Extensions, "scopes")
	delete(x.Extensions, "tags")
	*securityScheme = SecurityScheme(x)
	return nil
}
