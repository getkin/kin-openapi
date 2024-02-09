package openapi3filter

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func ExampleAuthenticationFunc() {
	const spec = `
openapi: 3.0.0
info:
  title: 'Validator'
  version: 0.0.1
components:
  securitySchemes:
    OAuth2:
      type: oauth2
      flows:
        clientCredentials:
          tokenUrl: /oauth2/token
          scopes:
            secrets.read: Ability to read secrets
            secrets.write: Ability to write secrets
paths:
  /secret:
    post:
      security:
        - OAuth2:
          - secrets.write
      responses:
        '200':
          description: Ok
        '401':
          description: Unauthorized
`
	var (
		errUnauthenticated = errors.New("login required")
		errForbidden       = errors.New("permission denied")
	)

	userScopes := map[string][]string{
		"Alice": {"secrets.read"},
		"Bob":   {"secrets.read", "secrets.write"},
	}

	authenticationFunc := func(_ context.Context, ai *AuthenticationInput) error {
		user := ai.RequestValidationInput.Request.Header.Get("X-User")
		if user == "" {
			return errUnauthenticated
		}

		for _, requiredScope := range ai.Scopes {
			var allowed bool
			for _, scope := range userScopes[user] {
				if scope == requiredScope {
					allowed = true
					break
				}
			}
			if !allowed {
				return errForbidden
			}
		}

		return nil
	}

	loader := openapi3.NewLoader()
	doc, _ := loader.LoadFromData([]byte(spec))
	router, _ := gorillamux.NewRouter(doc)

	validateRequest := func(req *http.Request) {
		route, pathParams, _ := router.FindRoute(req)

		validationInput := &RequestValidationInput{
			Request:    req,
			PathParams: pathParams,
			Route:      route,
			Options: &Options{
				AuthenticationFunc: authenticationFunc,
			},
		}
		err := ValidateRequest(context.TODO(), validationInput)
		switch {
		case errors.Is(err, errUnauthenticated):
			fmt.Println("username is required")
		case errors.Is(err, errForbidden):
			fmt.Println("user is not allowed to perform this action")
		case err == nil:
			fmt.Println("ok")
		default:
			log.Fatal(err)
		}
	}

	req1, _ := http.NewRequest(http.MethodPost, "/secret", nil)
	req1.Header.Set("X-User", "Alice")

	req2, _ := http.NewRequest(http.MethodPost, "/secret", nil)
	req2.Header.Set("X-User", "Bob")

	req3, _ := http.NewRequest(http.MethodPost, "/secret", nil)

	validateRequest(req1)
	validateRequest(req2)
	validateRequest(req3)
	// output:
	// user is not allowed to perform this action
	// ok
	// username is required
}
