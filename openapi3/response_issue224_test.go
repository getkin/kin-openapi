package openapi3

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmptyResponsesAreInvalid(t *testing.T) {
	spec := `{
    "openapi": "3.0.0",
    "servers": [
        {
            "url": "http://petstore.swagger.io/v2"
        }
    ],
    "info": {
        "description": ":dog: :cat: :rabbit: This is a sample server Petstore server.  You can find out more about Swagger at [http://swagger.io](http://swagger.io) or on [irc.freenode.net, #swagger](http://swagger.io/irc/).  For this sample, you can use the api key to test the authorization filters.",
        "version": "1.0.0",
        "title": "Swagger Petstore",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "email": "apiteam@swagger.io"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        }
    },
    "tags": [
        {
            "name": "pet",
            "description": "Everything about your Pets",
            "externalDocs": {
                "description": "Find out more",
                "url": "http://swagger.io"
            }
        },
        {
            "name": "store",
            "description": "Access to Petstore orders"
        },
        {
            "name": "user",
            "description": "Operations about user",
            "externalDocs": {
                "description": "Find out more about our store",
                "url": "http://swagger.io"
            }
        }
    ],
    "paths": {
        "/pet": {
            "post": {
                "tags": [
                    "pet"
                ],
                "summary": "Add a new pet to the store",
                "description": "",
                "operationId": "addPet",
                "responses": {

                },
                "security": [
                    {
                        "petstore_auth": [
                            "write:pets",
                            "read:pets"
                        ]
                    }
                ],
                "requestBody": {
                    "$ref": "#/components/requestBodies/Pet"
                },
                "parameters": []
            },
            "put": {
                "tags": [
                    "pet"
                ],
                "summary": "Update an existing pet",
                "description": "",
                "operationId": "updatePet",
                "responses": {

                },
                "security": [
                    {
                        "petstore_auth": [
                            "write:pets",
                            "read:pets"
                        ]
                    }
                ],
                "requestBody": {
                    "$ref": "#/components/requestBodies/Pet"
                },
                "parameters": []
            }
        },
        "/pet/{petId}": {
            "get": {
                "tags": [
                    "pet"
                ],
                "summary": "Find pet by ID",
                "description": "Returns a single pet",
                "operationId": "getPetById",
                "parameters": [
                    {
                        "name": "petId",
                        "in": "path",
                        "description": "ID of pet to return",
                        "required": true,
                        "schema": {
                            "type": "integer",
                            "format": "int64"
                        }
                    }
                ],
                "responses": {

                },
                "security": [
                    {
                        "api_key": []
                    }
                ]
            },
            "post": {
                "tags": [
                    "pet"
                ],
                "summary": "Updates a pet in the store with form data",
                "description": "",
                "operationId": "updatePetWithForm",
                "parameters": [
                    {
                        "name": "petId",
                        "in": "path",
                        "description": "ID of pet that needs to be updated",
                        "required": true,
                        "schema": {
                            "type": "integer",
                            "format": "int64"
                        }
                    }
                ],
                "responses": {

                },
                "security": [
                    {
                        "petstore_auth": [
                            "write:pets",
                            "read:pets"
                        ]
                    }
                ],
                "requestBody": {
                    "content": {
                        "application/x-www-form-urlencoded": {
                            "schema": {
                                "type": "object",
                                "properties": {
                                    "name": {
                                        "description": "Updated name of the pet",
                                        "type": "string"
                                    },
                                    "status": {
                                        "description": "Updated status of the pet",
                                        "type": "string"
                                    }
                                }
                            }
                        }
                    }
                }
            },
            "delete": {
                "tags": [
                    "pet"
                ],
                "summary": "Deletes a pet",
                "description": "",
                "operationId": "deletePet",
                "parameters": [
                    {
                        "name": "api_key",
                        "in": "header",
                        "required": false,
                        "schema": {
                            "type": "string"
                        }
                    },
                    {
                        "name": "petId",
                        "in": "path",
                        "description": "Pet id to delete",
                        "required": true,
                        "schema": {
                            "type": "integer",
                            "format": "int64"
                        }
                    }
                ],
                "responses": {

                },
                "security": [
                    {
                        "petstore_auth": [
                            "write:pets",
                            "read:pets"
                        ]
                    }
                ]
            }
        }
    },
    "externalDocs": {
        "description": "See AsyncAPI example",
        "url": "https://mermade.github.io/shins/asyncapi.html"
    },
    "components": {
        "schemas": {
            "Order": {
                "type": "object",
                "properties": {
                    "id": {
                        "type": "integer",
                        "format": "int64"
                    },
                    "petId": {
                        "type": "integer",
                        "format": "int64"
                    },
                    "quantity": {
                        "type": "integer",
                        "format": "int32"
                    },
                    "shipDate": {
                        "type": "string",
                        "format": "date-time"
                    },
                    "status": {
                        "type": "string",
                        "description": "Order Status",
                        "enum": [
                            "placed",
                            "approved",
                            "delivered"
                        ]
                    },
                    "complete": {
                        "type": "boolean",
                        "default": false
                    }
                },
                "xml": {
                    "name": "Order"
                }
            },
            "Category": {
                "type": "object",
                "properties": {
                    "id": {
                        "type": "integer",
                        "format": "int64"
                    },
                    "name": {
                        "type": "string"
                    }
                },
                "xml": {
                    "name": "Category"
                }
            },
            "User": {
                "type": "object",
                "properties": {
                    "id": {
                        "type": "integer",
                        "format": "int64"
                    },
                    "username": {
                        "type": "string"
                    },
                    "firstName": {
                        "type": "string"
                    },
                    "lastName": {
                        "type": "string"
                    },
                    "email": {
                        "type": "string"
                    },
                    "password": {
                        "type": "string"
                    },
                    "phone": {
                        "type": "string"
                    },
                    "userStatus": {
                        "type": "integer",
                        "format": "int32",
                        "description": "User Status"
                    }
                },
                "xml": {
                    "name": "User"
                }
            },
            "Tag": {
                "type": "object",
                "properties": {
                    "id": {
                        "type": "integer",
                        "format": "int64"
                    },
                    "name": {
                        "type": "string"
                    }
                },
                "xml": {
                    "name": "Tag"
                }
            },
            "Pet": {
                "type": "object",
                "required": [
                    "name",
                    "photoUrls"
                ],
                "properties": {
                    "id": {
                        "type": "integer",
                        "format": "int64"
                    },
                    "category": {
                        "$ref": "#/components/schemas/Category"
                    },
                    "name": {
                        "type": "string",
                        "example": "doggie"
                    },
                    "photoUrls": {
                        "type": "array",
                        "xml": {
                            "name": "photoUrl",
                            "wrapped": true
                        },
                        "items": {
                            "type": "string"
                        }
                    },
                    "tags": {
                        "type": "array",
                        "xml": {
                            "name": "tag",
                            "wrapped": true
                        },
                        "items": {
                            "$ref": "#/components/schemas/Tag"
                        }
                    },
                    "status": {
                        "type": "string",
                        "description": "pet status in the store",
                        "enum": [
                            "available",
                            "pending",
                            "sold"
                        ]
                    }
                },
                "xml": {
                    "name": "Pet"
                }
            },
            "ApiResponse": {
                "type": "object",
                "properties": {
                    "code": {
                        "type": "integer",
                        "format": "int32"
                    },
                    "type": {
                        "type": "string"
                    },
                    "message": {
                        "type": "string"
                    }
                }
            }
        },
        "requestBodies": {
            "Pet": {
                "content": {
                    "application/json": {
                        "schema": {
                            "$ref": "#/components/schemas/Pet"
                        }
                    },
                    "application/xml": {
                        "schema": {
                            "$ref": "#/components/schemas/Pet"
                        }
                    }
                },
                "description": "Pet object that needs to be added to the store",
                "required": true
            },
            "UserArray": {
                "content": {
                    "application/json": {
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/components/schemas/User"
                            }
                        }
                    }
                },
                "description": "List of user object",
                "required": true
            }
        },
        "securitySchemes": {
            "petstore_auth": {
                "type": "oauth2",
                "flows": {
                    "implicit": {
                        "authorizationUrl": "http://petstore.swagger.io/oauth/dialog",
                        "scopes": {
                            "write:pets": "modify pets in your account",
                            "read:pets": "read your pets"
                        }
                    }
                }
            },
            "api_key": {
                "type": "apiKey",
                "name": "api_key",
                "in": "header"
            }
        },
        "links": {},
        "callbacks": {}
    },
    "security": []
}
`

	doc, err := NewSwaggerLoader().LoadSwaggerFromData([]byte(spec))
	require.NoError(t, err)
	err = doc.Validate(context.Background())
	require.EqualError(t, err, `invalid paths: the responses object MUST contain at least one response code`)
}
