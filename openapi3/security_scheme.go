package openapi3

import (
	"context"
	"errors"
	"fmt"
	"github.com/ronniedada/kin-openapi/jsoninfo"
)

type SecurityScheme struct {
	ExtensionProps

	Type         string      `json:"type,omitempty"`
	Description  string      `json:"description,omitempty"`
	Name         string      `json:"name,omitempty"`
	In           string      `json:"in,omitempty"`
	Scheme       string      `json:"scheme,omitempty"`
	BearerFormat string      `json:"bearerFormat,omitempty"`
	Flow         *OAuthFlows `json:"flow,omitempty"`
}

func NewSecurityScheme() *SecurityScheme {
	return &SecurityScheme{}
}

func NewCSRFSecurityScheme() *SecurityScheme {
	return &SecurityScheme{
		Type: "apiKey",
		In:   "header",
		Name: "X-XSRF-TOKEN",
	}
}

func NewJWTSecurityScheme() *SecurityScheme {
	return &SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
	}
}

func (value *SecurityScheme) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStrictStruct(value)
}

func (value *SecurityScheme) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStrictStruct(data, value)
}

func (ss *SecurityScheme) WithType(value string) *SecurityScheme {
	ss.Type = value
	return ss
}

func (ss *SecurityScheme) WithDescription(value string) *SecurityScheme {
	ss.Description = value
	return ss
}

func (ss *SecurityScheme) WithName(value string) *SecurityScheme {
	ss.Name = value
	return ss
}

func (ss *SecurityScheme) WithIn(value string) *SecurityScheme {
	ss.In = value
	return ss
}

func (ss *SecurityScheme) WithScheme(value string) *SecurityScheme {
	ss.Scheme = value
	return ss
}

func (ss *SecurityScheme) WithBearerFormat(value string) *SecurityScheme {
	ss.BearerFormat = value
	return ss
}

func (securityScheme *SecurityScheme) Validate(c context.Context) error {
	hasIn := false
	hasBearerFormat := false
	hasFlow := false
	switch securityScheme.Type {
	case "apiKey":
		hasIn = true
		hasBearerFormat = true
	case "http":
		scheme := securityScheme.Scheme
		switch scheme {
		case "bearer":
			hasBearerFormat = true
		case "basic":
		default:
			return fmt.Errorf("Security scheme of type 'http' has invalid 'scheme' value '%s'", scheme)
		}
	case "oauth2":
		hasFlow = true
	case "openIdConnect":
		return fmt.Errorf("Support for security schemes with type '%v' has not been implemented", securityScheme.Type)
	default:
		return fmt.Errorf("Security scheme 'type' can't be '%v'", securityScheme.Type)
	}

	// Validate "in" and "name"
	if hasIn {
		switch securityScheme.In {
		case "query", "header":
		default:
			return fmt.Errorf("Security scheme of type 'apiKey' should have 'in'. It can be 'query' or 'header', not '%s'",
				securityScheme.In)
		}
		if securityScheme.Name == "" {
			return errors.New("Security scheme of type 'apiKey' should have 'name'")
		}
	} else if len(securityScheme.In) > 0 {
		return fmt.Errorf("Security scheme of type '%s' can't have 'in'", securityScheme.Type)
	} else if len(securityScheme.Name) > 0 {
		return errors.New("Security scheme of type 'apiKey' can't have 'name'")
	}

	// Validate "format"
	if hasBearerFormat {
		switch securityScheme.BearerFormat {
		case "", "JWT":
		default:
			return fmt.Errorf("Security scheme has unsupported 'bearerFormat' value '%s'", securityScheme.BearerFormat)
		}
	} else if len(securityScheme.BearerFormat) > 0 {
		return errors.New("Security scheme of type 'apiKey' can't have 'bearerFormat'")
	}

	// Validate "flow"
	if hasFlow {
		flow := securityScheme.Flow
		if flow == nil {
			return fmt.Errorf("Security scheme of type '%v' should have 'flow'", securityScheme.Type)
		}
		if err := flow.Validate(c); err != nil {
			return fmt.Errorf("Security scheme 'flow' is invalid: %v", err)
		}
	} else if securityScheme.Flow != nil {
		return fmt.Errorf("Security scheme of type '%s' can't have 'flow'",
			securityScheme.Type)
	}
	return nil
}

type OAuthFlows struct {
	ExtensionProps
	Implicit          *OAuthFlow `json:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty"`
}

func (value *OAuthFlows) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStrictStruct(value)
}

func (value *OAuthFlows) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStrictStruct(data, value)
}

func (flows *OAuthFlows) Validate(c context.Context) error {
	if v := flows.Implicit; v != nil {
		return v.Validate(c)
	}
	if v := flows.Password; v != nil {
		return v.Validate(c)
	}
	if v := flows.ClientCredentials; v != nil {
		return v.Validate(c)
	}
	if v := flows.AuthorizationCode; v != nil {
		return v.Validate(c)
	}
	return fmt.Errorf("No OAuth flow is defined")
}

type OAuthFlow struct {
	ExtensionProps
	AuthorizationURL string            `json:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty"`
	RefreshURL       string            `json:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes"`
}

func (value *OAuthFlow) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStrictStruct(value)
}

func (value *OAuthFlow) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStrictStruct(data, value)
}

func (flow *OAuthFlow) Validate(c context.Context) error {
	if v := flow.AuthorizationURL; v == "" {
		return fmt.Errorf("An OAuth flow is missing 'authorizationUrl'")
	}
	if v := flow.TokenURL; v == "" {
		return fmt.Errorf("An OAuth flow is missing 'tokenUrl'")
	}
	if v := flow.Scopes; v == nil || len(v) == 0 {
		return fmt.Errorf("An OAuth flow is missing 'scopes'")
	}
	return nil
}
