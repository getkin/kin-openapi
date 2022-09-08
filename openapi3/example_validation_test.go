package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExamplesValidation(t *testing.T) {
	spec := []byte(`
openapi: 3.0.3

paths:
  /user:
    post:
      description: User creation.
      operationId: createUser
      requestBody:
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/CreateUserRequest"
            examples:
              BadUser:
                $ref: '#/components/examples/BadUser'
        description: Created user object
        required: true
      responses:
        "201":
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/CreateUserResponse"
          description: successful operation
      summary: creates a new user
      tags:
        - user
components:
  schemas:
    CreateUserRequest:
      description: represents a new User
      example:
        username: good
        email: real@email
        password: validpassword
      required:
        - username
        - email
        - password
      properties:
        username:
          type: string
          pattern: "^[ a-zA-Z0-9_-]+$"
          minLength: 3
        email:
          type: string
          pattern: "^[A-Za-z0-9+_.-]+@(.+)$"
        password:
          type: string
          minLength: 7
      title: a User
      type: object
    CreateUserResponse:
      description: represents the response to a User creation
      properties:
        access_token:
          type: string
        user_id:
          format: int64
          type: integer
      type: object
  examples:
    BadUser:
      value:
        username: "]bad["
        email: bad
        password: short
info:
  title: An API
  version: 1.2.3.4
`)

	loader := NewLoader()

	doc, err := loader.LoadFromData(spec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	t.Logf("error: %v", err)
	require.Error(t, err)
}

func TestExampleValidation(t *testing.T) {
	spec := []byte(`
openapi: 3.0.3

paths:
  /user:
    post:
      description: User creation.
      operationId: createUser
      requestBody:
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/CreateUserRequest"
            example:
              username: "]bad["
              email: bad
              password: short
        description: Created user object
        required: true
      responses:
        "201":
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/CreateUserResponse"
          description: successful operation
      summary: creates a new user
      tags:
        - user
components:
  schemas:
    CreateUserRequest:
      description: represents a new User
      example:
        username: good
        email: real@email
        password: validpassword
      required:
        - username
        - email
        - password
      properties:
        username:
          type: string
          pattern: "^[ a-zA-Z0-9_-]+$"
          minLength: 3
        email:
          type: string
          pattern: "^[A-Za-z0-9+_.-]+@(.+)$"
        password:
          type: string
          minLength: 7
      title: a User
      type: object
    CreateUserResponse:
      description: represents the response to a User creation
      properties:
        access_token:
          type: string
        user_id:
          format: int64
          type: integer
      type: object
info:
  title: An API
  version: 1.2.3.4
`)

	loader := NewLoader()

	doc, err := loader.LoadFromData(spec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	t.Logf("error: %v", err)
	require.Error(t, err)
}

func TestSchemaExampleValidation(t *testing.T) {
	spec := []byte(`
openapi: 3.0.3

paths:
  /user:
    post:
      description: User creation.
      operationId: createUser
      requestBody:
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/CreateUserRequest"
        description: Created user object
        required: true
      responses:
        "201":
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/CreateUserResponse"
          description: successful operation
      summary: creates a new user
      tags:
        - user
components:
  schemas:
    CreateUserRequest:
      description: represents a new User
      example:
        username: good
        email: good@email.com
        # missing password
      required:
        - username
        - email
        - password
      properties:
        username:
          type: string
          pattern: "^[ a-zA-Z0-9_-]+$"
          minLength: 3
        email:
          type: string
          pattern: "^[A-Za-z0-9+_.-]+@(.+)$"
        password:
          type: string
          minLength: 7
      title: a User
      type: object
    CreateUserResponse:
      description: represents the response to a User creation
      properties:
        access_token:
          type: string
        user_id:
          format: int64
          type: integer
      type: object
info:
  title: An API
  version: 1.2.3.4
`)

	loader := NewLoader()

	doc, err := loader.LoadFromData(spec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	t.Logf("error: %v", err)
	require.Error(t, err)
}

// XXX should this be a validation error or skipped?
func TestBadExample(t *testing.T) {
	spec := []byte(`
openapi: 3.0.3
components:
  examples:
    User:
      description: empty user example
info:
  title: An API
  version: 1.2.3.4
`)

	loader := NewLoader()

	doc, err := loader.LoadFromData(spec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	t.Logf("error: %v", err)
	require.Error(t, err)
}
