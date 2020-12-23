package openapi2conv

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestConvOpenAPIV3ToV2(t *testing.T) {
	var doc3 openapi3.Swagger
	err := json.Unmarshal([]byte(exampleV3), &doc3)
	require.NoError(t, err)
	{
		// Refs need resolving before we can Validate
		sl := openapi3.NewSwaggerLoader()
		err = sl.ResolveRefsIn(&doc3, nil)
		require.NoError(t, err)
		err = doc3.Validate(context.Background())
		require.NoError(t, err)
	}

	spec2, err := FromV3Swagger(&doc3)
	require.NoError(t, err)
	data, err := json.Marshal(spec2)
	require.NoError(t, err)
	require.JSONEq(t, exampleV2, string(data))
}

func TestConvOpenAPIV2ToV3(t *testing.T) {
	var doc2 openapi2.Swagger
	err := json.Unmarshal([]byte(exampleV2), &doc2)
	require.NoError(t, err)

	spec3, err := ToV3Swagger(&doc2)
	require.NoError(t, err)
	err = spec3.Validate(context.Background())
	require.NoError(t, err)
	data, err := json.Marshal(spec3)
	require.NoError(t, err)
	require.JSONEq(t, exampleV3, string(data))
}

const exampleV2 = `
{
	"basePath": "/v2",
	"consumes": [
		"application/json",
		"application/xml"
	],
	"definitions": {
		"Error": {
			"description": "Error response.",
			"properties": {
				"message": {
					"type": "string"
				}
			},
			"required": [
				"message"
			],
			"type": "object"
		},
		"Item": {
			"additionalProperties": true,
			"properties": {
				"foo": {
					"type": "string"
				},
				"quux": {
					"$ref": "#/definitions/ItemExtension"
				}
			},
			"type": "object"
		},
		"ItemExtension": {
			"description": "It could be anything.",
			"type": "boolean"
		}
	},
	"externalDocs": {
		"description": "Example Documentation",
		"url": "https://example/doc/"
	},
	"host": "test.example.com",
	"info": {
		"title": "MyAPI",
		"version": "0.1",
		"x-info": "info extension"
	},
	"parameters": {
		"banana": {
			"in": "path",
			"name": "banana",
			"required": true,
			"type": "string"
		},
		"post_form_ref": {
			"description": "param description",
			"in": "formData",
			"name": "fileUpload2",
			"required": true,
			"type": "file",
			"x-formData-name": "fileUpload2",
			"x-mimetype": "text/plain"
		},
		"put_body": {
			"in": "body",
			"name": "banana",
			"required": true,
			"schema": {
				"type": "string"
			},
			"x-originalParamName": "banana"
		}
	},
	"paths": {
		"/another/{banana}/{id}": {
			"parameters": [
				{
					"$ref": "#/parameters/banana"
				},
				{
					"in": "path",
					"name": "id",
					"required": true,
					"type": "integer"
				}
			]
		},
		"/example": {
			"delete": {
				"description": "example delete",
				"operationId": "example-delete",
				"parameters": [
					{
						"description": "Only return results that intersect the provided bounding box.",
						"in": "query",
						"items": {
							"type": "number"
						},
						"maxItems": 4,
						"minItems": 4,
						"name": "bbox",
						"type": "array"
					},
					{
						"in": "query",
						"name": "x",
						"type": "string",
						"x-parameter": "parameter extension 1"
					},
					{
						"default": 250,
						"description": "The y parameter",
						"in": "query",
						"maximum": 10000,
						"minimum": 1,
						"name": "y",
						"type": "integer"
					}
				],
				"responses": {
					"200": {
						"description": "ok",
						"schema": {
							"items": {
								"$ref": "#/definitions/Item"
							},
							"type": "array"
						}
					},
					"404": {
						"description": "404 response"
					},
					"default": {
						"description": "default response",
						"x-response": "response extension 1"
					}
				},
				"security": [
					{
						"get_security_0": [
							"scope0",
							"scope1"
						],
						"get_security_1": []
					}
				],
				"summary": "example get",
				"tags": [
					"Example"
				]
			},
			"get": {
				"description": "example get",
				"responses": {
					"403": {
						"$ref": "#/responses/ForbiddenError"
					},
					"404": {
						"description": "404 response"
					},
					"default": {
						"description": "default response"
					}
				},
				"x-operation": "operation extension 1"
			},
			"head": {
				"description": "example head",
				"responses": {
					"default": {
						"description": "default response"
					}
				}
			},
			"options": {
				"description": "example options",
				"responses": {
					"default": {
						"description": "default response"
					}
				}
			},
			"patch": {
				"consumes": [
					"application/json",
					"application/xml"
				],
				"description": "example patch",
				"parameters": [
					{
						"in": "body",
						"name": "patch_body",
						"schema": {
							"allOf": [
								{
									"$ref": "#/definitions/Item"
								}
							]
						},
						"x-originalParamName": "patch_body",
						"x-requestBody": "requestbody extension 1"
					}
				],
				"responses": {
					"default": {
						"description": "default response"
					}
				}
			},
			"post": {
				"consumes": [
					"multipart/form-data"
				],
				"description": "example post",
				"parameters": [
					{
						"$ref": "#/parameters/post_form_ref"
					},
					{
						"description": "param description",
						"in": "formData",
						"name": "fileUpload",
						"type": "file",
						"x-formData-name": "fileUpload",
						"x-mimetype": "text/plain"
					},
					{
						"description": "File Id",
						"in": "query",
						"name": "id",
						"type": "integer"
					},
					{
						"description": "Description of file contents",
						"in": "formData",
						"name": "note",
						"type": "integer",
						"x-formData-name": "note"
					}
				],
				"responses": {
					"default": {
						"description": "default response"
					}
				}
			},
			"put": {
				"description": "example put",
				"parameters": [
					{
						"$ref": "#/parameters/put_body"
					}
				],
				"responses": {
					"default": {
						"description": "default response"
					}
				}
			},
			"x-path": "path extension 1",
			"x-path2": "path extension 2"
		}
	},
	"responses": {
		"ForbiddenError": {
			"description": "Insufficient permission to perform the requested action.",
			"schema": {
				"$ref": "#/definitions/Error"
			}
		}
	},
	"schemes": [
		"https"
	],
	"security": [
		{
			"default_security_0": [
				"scope0",
				"scope1"
			],
			"default_security_1": []
		}
	],
	"swagger": "2.0",
	"tags": [
		{
			"description": "An example tag.",
			"name": "Example"
		}
	],
	"x-root": "root extension 1",
	"x-root2": "root extension 2"
}
`

const exampleV3 = `
{
	"components": {
		"parameters": {
			"banana": {
				"in": "path",
				"name": "banana",
				"required": true,
				"schema": {
					"type": "string"
				}
			}
		},
		"requestBodies": {
			"put_body": {
				"content": {
					"application/json": {
						"schema": {
							"type": "string"
						}
					},
					"application/xml": {
						"schema": {
							"type": "string"
						}
					}
				},
				"required": true,
				"x-originalParamName": "banana"
			}
		},
		"responses": {
			"ForbiddenError": {
				"content": {
					"application/json": {
						"schema": {
							"$ref": "#/components/schemas/Error"
						}
					}
				},
				"description": "Insufficient permission to perform the requested action."
			}
		},
		"schemas": {
			"Error": {
				"description": "Error response.",
				"properties": {
					"message": {
						"type": "string"
					}
				},
				"required": [
					"message"
				],
				"type": "object"
			},
			"Item": {
				"additionalProperties": true,
				"properties": {
					"foo": {
						"type": "string"
					},
					"quux": {
						"$ref": "#/components/schemas/ItemExtension"
					}
				},
				"type": "object"
			},
			"ItemExtension": {
				"description": "It could be anything.",
				"type": "boolean"
			},
			"post_form_ref": {
				"description": "param description",
				"format": "binary",
				"required": [
					"fileUpload2"
				],
				"type": "string",
				"x-formData-name": "fileUpload2",
				"x-mimetype": "text/plain"
			}
		}
	},
	"externalDocs": {
		"description": "Example Documentation",
		"url": "https://example/doc/"
	},
	"info": {
		"title": "MyAPI",
		"version": "0.1",
		"x-info": "info extension"
	},
	"openapi": "3.0.3",
	"paths": {
		"/another/{banana}/{id}": {
			"parameters": [
				{
					"$ref": "#/components/parameters/banana"
				},
				{
					"in": "path",
					"name": "id",
					"required": true,
					"schema": {
						"type": "integer"
					}
				}
			]
		},
		"/example": {
			"delete": {
				"description": "example delete",
				"operationId": "example-delete",
				"parameters": [
					{
						"description": "Only return results that intersect the provided bounding box.",
						"in": "query",
						"name": "bbox",
						"schema": {
							"items": {
								"type": "number"
							},
							"maxItems": 4,
							"minItems": 4,
							"type": "array"
						}
					},
					{
						"in": "query",
						"name": "x",
						"schema": {
							"type": "string"
						},
						"x-parameter": "parameter extension 1"
					},
					{
						"description": "The y parameter",
						"in": "query",
						"name": "y",
						"schema": {
							"default": 250,
							"maximum": 10000,
							"minimum": 1,
							"type": "integer"
						}
					}
				],
				"responses": {
					"200": {
						"content": {
							"application/json": {
								"schema": {
									"items": {
										"$ref": "#/components/schemas/Item"
									},
									"type": "array"
								}
							}
						},
						"description": "ok"
					},
					"404": {
						"description": "404 response"
					},
					"default": {
						"description": "default response",
						"x-response": "response extension 1"
					}
				},
				"security": [
					{
						"get_security_0": [
							"scope0",
							"scope1"
						],
						"get_security_1": []
					}
				],
				"summary": "example get",
				"tags": [
					"Example"
				]
			},
			"get": {
				"description": "example get",
				"responses": {
					"403": {
						"$ref": "#/components/responses/ForbiddenError"
					},
					"404": {
						"description": "404 response"
					},
					"default": {
						"description": "default response"
					}
				},
				"x-operation": "operation extension 1"
			},
			"head": {
				"description": "example head",
				"responses": {
					"default": {
						"description": "default response"
					}
				}
			},
			"options": {
				"description": "example options",
				"responses": {
					"default": {
						"description": "default response"
					}
				}
			},
			"patch": {
				"description": "example patch",
				"requestBody": {
					"content": {
						"application/json": {
							"schema": {
								"allOf": [
									{
										"$ref": "#/components/schemas/Item"
									}
								]
							}
						},
						"application/xml": {
							"schema": {
								"allOf": [
									{
										"$ref": "#/components/schemas/Item"
									}
								]
							}
						}
					},
					"x-originalParamName": "patch_body",
					"x-requestBody": "requestbody extension 1"
				},
				"responses": {
					"default": {
						"description": "default response"
					}
				}
			},
			"post": {
				"description": "example post",
				"parameters": [
					{
						"description": "File Id",
						"in": "query",
						"name": "id",
						"schema": {
							"type": "integer"
						}
					}
				],
				"requestBody": {
					"content": {
						"multipart/form-data": {
							"schema": {
								"properties": {
									"fileUpload": {
										"description": "param description",
										"format": "binary",
										"type": "string",
										"x-formData-name": "fileUpload",
										"x-mimetype": "text/plain"
									},
									"fileUpload2": {
										"$ref": "#/components/schemas/post_form_ref"
									},
									"note": {
										"description": "Description of file contents",
										"type": "integer",
										"x-formData-name": "note"
									}
								},
								"required": [
									"fileUpload2"
								],
								"type": "object"
							}
						}
					}
				},
				"responses": {
					"default": {
						"description": "default response"
					}
				}
			},
			"put": {
				"description": "example put",
				"requestBody": {
					"$ref": "#/components/requestBodies/put_body"
				},
				"responses": {
					"default": {
						"description": "default response"
					}
				}
			},
			"x-path": "path extension 1",
			"x-path2": "path extension 2"
		}
	},
	"security": [
		{
			"default_security_0": [
				"scope0",
				"scope1"
			],
			"default_security_1": []
		}
	],
	"servers": [
		{
			"url": "https://test.example.com/v2"
		}
	],
	"tags": [
		{
			"description": "An example tag.",
			"name": "Example"
		}
	],
	"x-root": "root extension 1",
	"x-root2": "root extension 2"
}
`
