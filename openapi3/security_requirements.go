package openapi3

import (
	"context"
)

type SecurityRequirements []SecurityRequirement

func NewSecurityRequirements() *SecurityRequirements {
	return &SecurityRequirements{}
}

func (srs *SecurityRequirements) With(securityRequirement SecurityRequirement) *SecurityRequirements {
	*srs = append(*srs, securityRequirement)
	return srs
}

// Validate goes through the receiver value and its descendants and errors on any non compliance to the OpenAPIv3 specification.
func (srs SecurityRequirements) Validate(ctx context.Context) error {
	for _, item := range srs {
		if err := item.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

type SecurityRequirement map[string][]string

func NewSecurityRequirement() SecurityRequirement {
	return make(SecurityRequirement)
}

func (security SecurityRequirement) Authenticate(provider string, scopes ...string) SecurityRequirement {
	if len(scopes) == 0 {
		scopes = []string{} // Forces the variable to be encoded as an array instead of null
	}
	security[provider] = scopes
	return security
}

// Validate goes through the receiver value and its descendants and errors on any non compliance to the OpenAPIv3 specification.
func (security SecurityRequirement) Validate(ctx context.Context) error {
	return nil
}
