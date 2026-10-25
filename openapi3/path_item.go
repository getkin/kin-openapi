package openapi3

import (
	"context"
	"encoding/json"
	"maps"
	"net/http"
	"regexp"
	"strings"
)

// PathItem is specified by OpenAPI/Swagger standard version 3.
// See https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md#path-item-object
type PathItem struct {
	Extensions map[string]any `json:"-" yaml:"-"`
	Origin     *Origin        `json:"-" yaml:"-"`

	Ref         string     `json:"$ref,omitempty,case:strict" yaml:"$ref,omitempty"`
	Summary     string     `json:"summary,omitempty,case:strict" yaml:"summary,omitempty"`
	Description string     `json:"description,omitempty,case:strict" yaml:"description,omitempty"`
	Connect     *Operation `json:"connect,omitempty,case:strict" yaml:"connect,omitempty"`
	Delete      *Operation `json:"delete,omitempty,case:strict" yaml:"delete,omitempty"`
	Get         *Operation `json:"get,omitempty,case:strict" yaml:"get,omitempty"`
	Head        *Operation `json:"head,omitempty,case:strict" yaml:"head,omitempty"`
	Options     *Operation `json:"options,omitempty,case:strict" yaml:"options,omitempty"`
	Patch       *Operation `json:"patch,omitempty,case:strict" yaml:"patch,omitempty"`
	Post        *Operation `json:"post,omitempty,case:strict" yaml:"post,omitempty"`
	Put         *Operation `json:"put,omitempty,case:strict" yaml:"put,omitempty"`
	Trace       *Operation `json:"trace,omitempty,case:strict" yaml:"trace,omitempty"`
	Query       *Operation `json:"query,omitempty,case:strict" yaml:"query,omitempty"` // OpenAPI >=3.2
	Servers     Servers    `json:"servers,omitempty,case:strict" yaml:"servers,omitempty"`
	Parameters  Parameters `json:"parameters,omitempty,case:strict" yaml:"parameters,omitempty"`

	// AdditionalOperations maps custom HTTP method names to operations.
	// Keys are case-sensitive HTTP method tokens and must not duplicate
	// the methods covered by the fixed fields above.
	AdditionalOperations map[string]*Operation `json:"additionalOperations,omitempty,case:strict" yaml:"additionalOperations,omitempty"` // OpenAPI >=3.2

	// refSiblings holds the fields written next to Ref, as the document had
	// them: once the loader has merged the referenced path item in, they can no
	// longer be told apart from it, and marshaling writes back only these.
	refSiblings *PathItem
	// refResolved records that the loader has merged the referenced path item
	// in, so that it does not do it twice.
	refResolved bool
}

// MethodQuery is the HTTP QUERY method, introduced as a fixed Path Item
// Object field by OpenAPI 3.2. The net/http package has no constant for it.
const MethodQuery = "QUERY"

// fixedPathItemMethods are the HTTP methods that have a dedicated PathItem
// field; they must not appear as AdditionalOperations keys.
var fixedPathItemMethods = map[string]struct{}{
	http.MethodConnect: {},
	http.MethodDelete:  {},
	http.MethodGet:     {},
	http.MethodHead:    {},
	http.MethodOptions: {},
	http.MethodPatch:   {},
	http.MethodPost:    {},
	http.MethodPut:     {},
	http.MethodTrace:   {},
	MethodQuery:        {},
}

// PathItemMethods returns the HTTP methods that have a dedicated Path Item
// Object field, sorted. Any other method an operation is registered under
// lives in AdditionalOperations, so a caller that treats the two groups
// differently can tell them apart without listing the fixed methods itself.
func PathItemMethods() []string {
	return componentNames(fixedPathItemMethods)
}

// httpTokenRe matches an RFC 9110 token, the grammar HTTP method names
// must follow.
var httpTokenRe = regexp.MustCompile("^[!#$%&'*+\\-.^_`|~0-9A-Za-z]+$")

// MarshalJSON returns the JSON encoding of PathItem.
func (pathItem PathItem) MarshalJSON() ([]byte, error) {
	x, err := pathItem.MarshalYAML()
	if err != nil {
		return nil, err
	}
	return json.Marshal(x)
}

// MarshalYAML returns the YAML encoding of PathItem.
func (pathItem PathItem) MarshalYAML() (any, error) {
	if ref := pathItem.Ref; ref != "" {
		if pathItem.refResolved {
			if pathItem.refSiblings == nil {
				return Ref{Ref: ref}, nil
			}
			pathItem = *pathItem.refSiblings
		}
		pathItem.Ref = ""
		m, err := pathItem.MarshalYAML()
		if err != nil {
			return nil, err
		}
		m.(map[string]any)["$ref"] = ref
		return m, nil
	}

	m := make(map[string]any, 15+len(pathItem.Extensions))
	maps.Copy(m, pathItem.Extensions)
	if x := pathItem.Summary; x != "" {
		m["summary"] = x
	}
	if x := pathItem.Description; x != "" {
		m["description"] = x
	}
	if x := pathItem.Connect; x != nil {
		m["connect"] = x
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
	if x := pathItem.Trace; x != nil {
		m["trace"] = x
	}
	if x := pathItem.Query; x != nil {
		m["query"] = x
	}
	if x := pathItem.AdditionalOperations; len(x) != 0 {
		m["additionalOperations"] = x
	}
	if x := pathItem.Servers; len(x) != 0 {
		m["servers"] = x
	}
	if x := pathItem.Parameters; len(x) != 0 {
		m["parameters"] = x
	}
	return m, nil
}

// UnmarshalJSON sets PathItem to a copy of data.
func (pathItem *PathItem) UnmarshalJSON(data []byte) error {
	type PathItemBis PathItem
	var x PathItemBis
	if err := unmarshalJSON(data, &x); err != nil {
		return unmarshalError(err)
	}
	_ = unmarshalJSON(data, &x.Extensions)
	delete(x.Extensions, "$ref")
	delete(x.Extensions, "summary")
	delete(x.Extensions, "description")
	delete(x.Extensions, "connect")
	delete(x.Extensions, "delete")
	delete(x.Extensions, "get")
	delete(x.Extensions, "head")
	delete(x.Extensions, "options")
	delete(x.Extensions, "patch")
	delete(x.Extensions, "post")
	delete(x.Extensions, "put")
	delete(x.Extensions, "trace")
	delete(x.Extensions, "query")
	delete(x.Extensions, "additionalOperations")
	delete(x.Extensions, "servers")
	delete(x.Extensions, "parameters")
	if len(x.Extensions) == 0 {
		x.Extensions = nil
	}
	*pathItem = PathItem(x)
	return nil
}

// operationEntry pairs an operation with the method it is registered under
// and the JSON pointer suffix that addresses it inside a PathItem.
type operationEntry struct {
	method    string
	operation *Operation
	// pointerSuffix is "/get" for a fixed field and
	// "/additionalOperations/COPY" for a custom method.
	pointerSuffix string
}

// operationEntries returns pathItem's operations in a deterministic order
// (fixed fields first, then AdditionalOperations, each sorted by method),
// paired with the JSON pointer suffix addressing them. Custom methods
// shadowed by a fixed field are skipped, matching Operations().
func (pathItem *PathItem) operationEntries() []operationEntry {
	fixed := pathItem.fixedOperations()
	entries := make([]operationEntry, 0, len(fixed)+len(pathItem.AdditionalOperations))
	for _, method := range componentNames(fixed) {
		entries = append(entries, operationEntry{
			method:        method,
			operation:     fixed[method],
			pointerSuffix: "/" + strings.ToLower(method),
		})
	}
	for _, method := range componentNames(pathItem.AdditionalOperations) {
		operation := pathItem.AdditionalOperations[method]
		if operation == nil {
			continue
		}
		if _, ok := fixed[method]; ok {
			continue
		}
		entries = append(entries, operationEntry{
			method:        method,
			operation:     operation,
			pointerSuffix: "/additionalOperations/" + escapeRefString(method),
		})
	}
	return entries
}

// fixedOperations returns the operations held by PathItem's fixed fields.
func (pathItem *PathItem) fixedOperations() map[string]*Operation {
	operations := make(map[string]*Operation, len(fixedPathItemMethods))
	if v := pathItem.Connect; v != nil {
		operations[http.MethodConnect] = v
	}
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
	if v := pathItem.Trace; v != nil {
		operations[http.MethodTrace] = v
	}
	if v := pathItem.Query; v != nil {
		operations[MethodQuery] = v
	}
	return operations
}

// Operations returns pathItem's operations keyed by HTTP method, including
// the OpenAPI 3.2 QUERY field and any AdditionalOperations. When an
// AdditionalOperations key collides with a fixed field, the fixed field
// wins (Validate reports the duplicate separately).
func (pathItem *PathItem) Operations() map[string]*Operation {
	operations := pathItem.fixedOperations()
	for method, operation := range pathItem.AdditionalOperations {
		if operation == nil {
			continue
		}
		if _, ok := operations[method]; ok {
			continue
		}
		operations[method] = operation
	}
	return operations
}

func (pathItem *PathItem) GetOperation(method string) *Operation {
	switch method {
	case http.MethodConnect:
		return pathItem.Connect
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
	case http.MethodTrace:
		return pathItem.Trace
	case MethodQuery:
		return pathItem.Query
	default:
		return pathItem.AdditionalOperations[method]
	}
}

func (pathItem *PathItem) SetOperation(method string, operation *Operation) {
	switch method {
	case http.MethodConnect:
		pathItem.Connect = operation
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
	case http.MethodTrace:
		pathItem.Trace = operation
	case MethodQuery:
		pathItem.Query = operation
	default:
		// Custom (OpenAPI >=3.2) methods live in AdditionalOperations.
		if operation == nil {
			delete(pathItem.AdditionalOperations, method)
			return
		}
		if pathItem.AdditionalOperations == nil {
			pathItem.AdditionalOperations = make(map[string]*Operation, 1)
		}
		pathItem.AdditionalOperations[method] = operation
	}
}

// Validate returns an error if PathItem does not comply with the OpenAPI spec.
func (pathItem *PathItem) Validate(ctx context.Context, opts ...ValidationOption) error {
	ctx = WithValidationOptions(ctx, opts...)
	me := newErrCollector(ctx)

	if !getValidationOptions(ctx).isOpenAPI32OrLater {
		if pathItem.Query != nil {
			if err := me.emit(errFieldFor32Plus("query", pathItem.Origin)); err != nil {
				return err
			}
		}
		if len(pathItem.AdditionalOperations) != 0 {
			if err := me.emit(errFieldFor32Plus("additionalOperations", pathItem.Origin)); err != nil {
				return err
			}
		}
	}

	for _, method := range componentNames(pathItem.AdditionalOperations) {
		if _, ok := fixedPathItemMethods[method]; ok {
			if err := me.emit(newAdditionalOperationsDuplicateMethod(method, pathItem.Origin)); err != nil {
				return err
			}
			continue
		}
		if fixedMethod := strings.ToUpper(method); fixedMethod != method {
			if _, ok := fixedPathItemMethods[fixedMethod]; ok {
				if err := me.emit(newAdditionalOperationsFixedMethodCase(method, fixedMethod, pathItem.Origin)); err != nil {
					return err
				}
				continue
			}
		}
		if !httpTokenRe.MatchString(method) {
			if err := me.emit(newAdditionalOperationsInvalidMethod(method, pathItem.Origin)); err != nil {
				return err
			}
		}
	}

	operations := pathItem.Operations()

	for _, method := range componentNames(operations) {
		operation := operations[method]
		wrapOp := func(e error) error { return &OperationValidationError{Method: method, Cause: e} }
		if err := me.emitWrapped(wrapOp, operation.Validate(ctx)); err != nil {
			return err
		}
	}

	if v := pathItem.Parameters; v != nil {
		if err := me.emit(v.Validate(ctx)); err != nil {
			return err
		}
	}

	return me.finalize(validateExtensions(ctx, pathItem.Extensions, pathItem.Origin))
}

// isEmpty's introduced in 546590b1
// mergeRefSiblings sets on pathItem, a copy of the item its $ref points to,
// the fields that siblings, the item holding that $ref, sets.
//
// The specification only defines the result when no field is set on both: "In
// case a Path Item Object field appears both in the defined object and the
// referenced object, the behavior is undefined". The sibling then wins, like
// the summary and description of a Reference Object do over those of the
// object it points to, which is what OpenAPI 3.1 says this $ref is likely to
// be aligned with. It never modifies the maps of the referenced item, which
// other path items may share.
func (pathItem *PathItem) mergeRefSiblings(siblings *PathItem) {
	if len(siblings.Extensions) != 0 {
		extensions := make(map[string]any, len(pathItem.Extensions)+len(siblings.Extensions))
		maps.Copy(extensions, pathItem.Extensions)
		maps.Copy(extensions, siblings.Extensions)
		pathItem.Extensions = extensions
	}
	if x := siblings.Summary; x != "" {
		pathItem.Summary = x
	}
	if x := siblings.Description; x != "" {
		pathItem.Description = x
	}
	if x := siblings.Connect; x != nil {
		pathItem.Connect = x
	}
	if x := siblings.Delete; x != nil {
		pathItem.Delete = x
	}
	if x := siblings.Get; x != nil {
		pathItem.Get = x
	}
	if x := siblings.Head; x != nil {
		pathItem.Head = x
	}
	if x := siblings.Options; x != nil {
		pathItem.Options = x
	}
	if x := siblings.Patch; x != nil {
		pathItem.Patch = x
	}
	if x := siblings.Post; x != nil {
		pathItem.Post = x
	}
	if x := siblings.Put; x != nil {
		pathItem.Put = x
	}
	if x := siblings.Trace; x != nil {
		pathItem.Trace = x
	}
	if x := siblings.Query; x != nil {
		pathItem.Query = x
	}
	if x := siblings.AdditionalOperations; len(x) != 0 {
		pathItem.AdditionalOperations = x
	}
	if x := siblings.Servers; len(x) != 0 {
		pathItem.Servers = x
	}
	if x := siblings.Parameters; len(x) != 0 {
		pathItem.Parameters = x
	}
}

func (pathItem *PathItem) isEmpty() bool {
	// NOTE: ignores pathItem.Extensions
	// NOTE: ignores pathItem.Ref
	return pathItem.Summary == "" &&
		pathItem.Description == "" &&
		pathItem.Connect == nil &&
		pathItem.Delete == nil &&
		pathItem.Get == nil &&
		pathItem.Head == nil &&
		pathItem.Options == nil &&
		pathItem.Patch == nil &&
		pathItem.Post == nil &&
		pathItem.Put == nil &&
		pathItem.Trace == nil &&
		pathItem.Query == nil &&
		len(pathItem.AdditionalOperations) == 0 &&
		len(pathItem.Servers) == 0 &&
		len(pathItem.Parameters) == 0
}
