package openapi3

import (
	"context"
	"errors"

	"github.com/getkin/kin-openapi/jsoninfo"
)

// Info is specified by OpenAPI/Swagger standard version 3.0.
type Info struct {
	ExtensionProps
	Title          string   `json:"title"` // Required
	Description    string   `json:"description,omitempty"`
	TermsOfService string   `json:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty"`
	License        *License `json:"license"` // Required
	Version        string   `json:"version"` // Required
}

func (value *Info) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStrictStruct(value)
}

func (value *Info) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStrictStruct(data, value)
}

func (value *Info) Validate(c context.Context) error {
	if value.Title == "" {
		return errors.New("Variable 'title' must be a non-empty JSON string")
	}

	if value.Version == "" {
		return errors.New("Variable 'version' must be a non-empty JSON string")
	}

	if err := value.Contact.Validate(c); err != nil {
		return err
	}

	if err := value.License.Validate(c); err != nil {
		return err
	}

	return nil
}

// Contact is specified by OpenAPI/Swagger standard version 3.0.
type Contact struct {
	ExtensionProps
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

func (value *Contact) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStrictStruct(value)
}

func (value *Contact) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStrictStruct(data, value)
}

func (value *Contact) Validate(c context.Context) error {
	return nil
}

// License is specified by OpenAPI/Swagger standard version 3.0.
type License struct {
	ExtensionProps
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

func (value *License) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStrictStruct(value)
}

func (value *License) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStrictStruct(data, value)
}

func (value *License) Validate(c context.Context) error {
	if value.Name == "" {
		return errors.New("Variable 'name' must be a non-empty JSON string")
	}
	return nil
}
