package openapi3filter

import (
	"encoding/json"
	"fmt"
	"github.com/jban332/kinapi/openapi2"
	"github.com/jban332/kinapi/openapi2conv"
	"github.com/jban332/kinapi/openapi3"
	"github.com/jban332/kinapi/pathpattern"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type Route struct {
	Swagger   *openapi3.Swagger
	Server    *openapi3.Server
	Path      string
	PathItem  *openapi3.PathItem
	Method    string
	Operation *openapi3.Operation

	// For developers who want use the router for handling too
	Handler http.Handler
}

// Routers maps a HTTP request to a Router.
type Routers []*Router

func (routers Routers) FindRoute(method string, url *url.URL) (*Router, *Route, map[string]string, error) {
	for _, router := range routers {
		// Skip routers that have DO NOT have servers
		if len(router.swagger.Servers) == 0 {
			continue
		}
		route, pathParams, err := router.FindRoute(method, url)
		if err == nil {
			return router, route, pathParams, nil
		}
	}
	for _, router := range routers {
		// Skip routers that DO have servers
		if len(router.swagger.Servers) > 0 {
			continue
		}
		route, pathParams, err := router.FindRoute(method, url)
		if err == nil {
			return router, route, pathParams, nil
		}
	}
	return nil, nil, nil, &RouteError{
		Reason: "None of the routers matches",
	}
}

// Router maps a HTTP request to an OpenAPI operation.
type Router struct {
	swagger  *openapi3.Swagger
	pathNode *pathpattern.Node
}

func NewRouterFromFile(path string) (*Router, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	swagger3 := &openapi3.Swagger{}
	err = json.Unmarshal(data, swagger3)
	if err == nil {
		return NewRouter(swagger3), nil
	}

	// Try version 2
	{
		swagger2 := &openapi2.Swagger{}
		err := json.Unmarshal(data, swagger2)
		if err == nil {
			swagger3 := openapi2conv.ToV3Swagger(swagger2)
			return NewRouter(swagger3), nil
		}
	}
	// Version 2 didn't work
	// Return an error
	return nil, fmt.Errorf("Error deserializing Swagger v2/v3 in '%s': %v", path, err)
}

// NewRouter creates a new router.
//
// If the given Swagger has servers, router will use them.
// All operations of the Swagger will be added to the router.
func NewRouter(swagger *openapi3.Swagger) *Router {
	router := &Router{
		swagger: swagger,
	}
	router.AddSwagger(swagger)
	return router
}

// AddSwagger adds all operations (but not servers) in the router.
func (router *Router) AddSwagger(swagger *openapi3.Swagger) {
	root := router.node()
	if paths := swagger.Paths; paths != nil {
		for path, pathItem := range paths {
			for method, operation := range pathItem.Operations() {
				method = strings.ToUpper(method)
				root.Add(method+" "+path, &Route{
					Swagger:   swagger,
					Path:      path,
					PathItem:  pathItem,
					Method:    method,
					Operation: operation,
				}, nil)
			}
		}
	}
}

// AddRoute adds a route in the router.
func (router *Router) AddRoute(route *Route) {
	method := route.Method
	if method == "" {
		panic("Route is missing method")
	}
	method = strings.ToUpper(method)
	path := route.Path
	if path == "" {
		panic("Route is missing path")
	}
	router.node().Add(method+" "+path, router, nil)
}

func (router *Router) node() *pathpattern.Node {
	root := router.pathNode
	if root == nil {
		root = &pathpattern.Node{}
		router.pathNode = root
	}
	return root
}

func (router *Router) FindRoute(method string, url *url.URL) (*Route, map[string]string, error) {
	swagger := router.swagger

	// Get server
	servers := swagger.Servers
	var server *openapi3.Server
	var remainingPath string
	var pathParams map[string]string
	if len(servers) == 0 {
		remainingPath = url.Path
	} else {
		var paramValues []string
		server, paramValues, remainingPath = servers.MatchURL(url)
		if server == nil {
			return nil, nil, &RouteError{
				Route: Route{
					Swagger: swagger,
				},
				Reason: "Does not match any server",
			}
		}
		pathParams = make(map[string]string, 8)
		paramNames, _ := server.ParameterNames()
		for i, value := range paramValues {
			name := paramNames[i]
			pathParams[name] = value
		}
	}

	// Get PathItem
	root := router.node()
	var route *Route
	node, paramValues := root.Match(method + " " + remainingPath)
	if node != nil {
		route, _ = node.Value.(*Route)
	}
	if route == nil {
		return nil, nil, &RouteError{
			Route: Route{
				Swagger: swagger,
				Server:  server,
			},
			Reason: "Path was not found",
		}
	}

	// Get operation
	pathItem := route.PathItem
	operation := pathItem.GetOperation(method)
	if operation == nil {
		return nil, nil, &RouteError{
			Route: Route{
				Swagger: swagger,
				Server:  server,
			},
			Reason: "Path doesn't support the HTTP method",
		}
	}
	if pathParams == nil {
		pathParams = make(map[string]string, len(paramValues))
	}
	paramKeys := node.VariableNames
	for i, value := range paramValues {
		key := paramKeys[i]
		if strings.HasSuffix(key, "*") {
			key = key[:len(key)-1]
		}
		pathParams[key] = value
	}
	return route, pathParams, nil
}
