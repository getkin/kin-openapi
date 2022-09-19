package openapi3filter

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	legacyrouter "github.com/getkin/kin-openapi/routers/legacy"
)

func TestValidatingRequestParameterAndSetDefault(t *testing.T) {
	const spec = `{
  "openapi": "3.0.3",
  "info": {
    "version": "1.0.0",
    "title": "title",
    "description": "desc",
    "contact": {
      "email": "email"
    }
  },
  "paths": {
    "/accounts": {
      "get": {
        "description": "Create a new account",
        "parameters": [
          {
            "in": "query",
            "name": "q1",
            "schema": {
              "type": "string",
              "default": "Q"
            }
          },
          {
            "in": "query",
            "name": "q2",
            "schema": {
              "type": "string",
              "default": "Q"
            }
          },
          {
            "in": "query",
            "name": "q3",
            "schema": {
              "type": "string"
            }
          },
          {
            "in": "header",
            "name": "h1",
            "schema": {
              "type": "boolean",
              "default": true
            }
          },
          {
            "in": "header",
            "name": "h2",
            "schema": {
              "type": "boolean",
              "default": true
            }
          },
          {
            "in": "header",
            "name": "h3",
            "schema": {
              "type": "boolean"
            }
          },
          {
            "in": "cookie",
            "name": "c1",
            "schema": {
              "type": "integer",
              "default": 128
            }
          },
          {
            "in": "cookie",
            "name": "c2",
            "schema": {
              "type": "integer",
              "default": 128
            }
          },
          {
            "in": "cookie",
            "name": "c3",
            "schema": {
              "type": "integer"
            }
          }
        ],
        "responses": {
          "201": {
            "description": "Successfully created a new account"
          },
          "400": {
            "description": "The server could not understand the request due to invalid syntax",
          }
        }
      }
    }
  }
}
`

	sl := openapi3.NewLoader()
	doc, err := sl.LoadFromData([]byte(spec))
	require.NoError(t, err)
	err = doc.Validate(sl.Context)
	require.NoError(t, err)
	router, err := legacyrouter.NewRouter(doc)
	require.NoError(t, err)

	httpReq, err := http.NewRequest(http.MethodGet, "/accounts", nil)
	require.NoError(t, err)

	params := &url.Values{
		"q2": []string{"from_request"},
	}
	httpReq.URL.RawQuery = params.Encode()
	httpReq.Header.Set("h2", "false")
	httpReq.AddCookie(&http.Cookie{Name: "c2", Value: "1024"})

	route, pathParams, err := router.FindRoute(httpReq)
	require.NoError(t, err)

	err = ValidateRequest(sl.Context, &RequestValidationInput{
		Request:    httpReq,
		PathParams: pathParams,
		Route:      route,
	})
	require.NoError(t, err)

	// Unset default values in URL were set
	require.Equal(t, "Q", httpReq.URL.Query().Get("q1"))
	// Unset default values in headers were set
	require.Equal(t, "true", httpReq.Header.Get("h1"))
	// Unset default values in cookies were set
	cookie, err := httpReq.Cookie("c1")
	require.NoError(t, err)
	require.Equal(t, "128", cookie.Value)

	// All values from request were retained
	require.Equal(t, "from_request", httpReq.URL.Query().Get("q2"))
	require.Equal(t, "false", httpReq.Header.Get("h2"))
	cookie, err = httpReq.Cookie("c2")
	require.NoError(t, err)
	require.Equal(t, "1024", cookie.Value)

	// Not set value to parameters without default value
	require.Equal(t, "", httpReq.URL.Query().Get("q3"))
	require.Equal(t, "", httpReq.Header.Get("h3"))
	_, err = httpReq.Cookie("c3")
	require.Equal(t, http.ErrNoCookie, err)
}

func TestValidateRequestBodyAndSetDefault(t *testing.T) {
	const spec = `{
  "openapi": "3.0.3",
  "info": {
    "version": "1.0.0",
    "title": "title",
    "description": "desc",
    "contact": {
      "email": "email"
    }
  },
  "paths": {
    "/accounts": {
      "post": {
        "description": "Create a new account",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["id"],
                "properties": {
                  "id": {
                    "type": "string",
                    "pattern": "[0-9a-v]+$",
                    "minLength": 20,
                    "maxLength": 20
                  },
                  "name": {
                    "type": "string",
                    "default": "default"
                  },
                  "code": {
                    "type": "integer",
                    "default": 123
                  },
                  "all": {
                    "type": "boolean",
                    "default": false
                  },
                  "page": {
                    "type": "object",
                    "properties": {
                      "num": {
                        "type": "integer",
                        "default": 1
                      },
                      "size": {
                        "type": "integer",
                        "default": 10
                      },
                      "order": {
                        "type": "string",
                        "enum": ["asc", "desc"],
                        "default": "desc"
                      }
                    }
                  },
                  "filters": {
                    "type": "array",
                    "nullable": true,
                    "items": {
                      "type": "object",
                      "properties": {
                        "field": {
                          "type": "string",
                          "default": "name"
                        },
                        "op": {
                          "type": "string",
                          "enum": ["eq", "ne"],
                          "default": "eq"
                        },
                        "value": {
                          "type": "integer",
                          "default": 123
                        }
                      }
                    }
                  },
                  "social_network": {
                    "oneOf": [
                      {
                        "type": "object",
                        "required": ["platform"],
                        "properties": {
                          "platform": {
                            "type": "string",
                            "enum": [
                              "twitter"
                            ]
                          },
                          "tw_link": {
                            "type": "string",
                            "default": "www.twitter.com"
                          }
                        }
                      },
                      {
                        "type": "object",
                        "required": ["platform"],
                        "properties": {
                          "platform": {
                            "type": "string",
                            "enum": [
                              "facebook"
                            ]
                          },
                          "fb_link": {
                            "type": "string",
                            "default": "www.facebook.com"
                          }
                        }
                      }
                    ]
                  },
                  "social_network_2": {
                    "anyOf": [
                      {
                        "type": "object",
                        "required": ["platform"],
                        "properties": {
                          "platform": {
                            "type": "string",
                            "enum": [
                              "twitter"
                            ]
                          },
                          "tw_link": {
                            "type": "string",
                            "default": "www.twitter.com"
                          }
                        }
                      },
                      {
                        "type": "object",
                        "required": ["platform"],
                        "properties": {
                          "platform": {
                            "type": "string",
                            "enum": [
                              "facebook"
                            ]
                          },
                          "fb_link": {
                            "type": "string",
                            "default": "www.facebook.com"
                          }
                        }
                      }
                    ]
                  }
                }
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "Successfully created a new account"
          },
          "400": {
            "description": "The server could not understand the request due to invalid syntax",
          }
        }
      }
    }
  }
}`
	sl := openapi3.NewLoader()
	doc, err := sl.LoadFromData([]byte(spec))
	require.NoError(t, err)
	err = doc.Validate(sl.Context)
	require.NoError(t, err)
	router, err := legacyrouter.NewRouter(doc)
	require.NoError(t, err)

	type page struct {
		Num   int    `json:"num,omitempty"`
		Size  int    `json:"size,omitempty"`
		Order string `json:"order,omitempty"`
	}
	type filter struct {
		Field string `json:"field,omitempty"`
		OP    string `json:"op,omitempty"`
		Value int    `json:"value,omitempty"`
	}
	type socialNetwork struct {
		Platform string `json:"platform,omitempty"`
		FBLink   string `json:"fb_link,omitempty"`
		TWLink   string `json:"tw_link,omitempty"`
	}
	type body struct {
		ID             string         `json:"id,omitempty"`
		Name           string         `json:"name,omitempty"`
		Code           int            `json:"code,omitempty"`
		All            bool           `json:"all,omitempty"`
		Page           *page          `json:"page,omitempty"`
		Filters        []filter       `json:"filters,omitempty"`
		SocialNetwork  *socialNetwork `json:"social_network,omitempty"`
		SocialNetwork2 *socialNetwork `json:"social_network_2,omitempty"`
	}

	testCases := []struct {
		name          string
		body          body
		bodyAssertion func(t *testing.T, body string)
	}{
		{
			name: "only id",
			body: body{
				ID: "bt6kdc3d0cvp6u8u3ft0",
			},
			bodyAssertion: func(t *testing.T, body string) {
				require.JSONEq(t, `{"id":"bt6kdc3d0cvp6u8u3ft0", "name": "default", "code": 123, "all": false}`, body)
			},
		},
		{
			name: "id & name",
			body: body{
				ID:   "bt6kdc3d0cvp6u8u3ft0",
				Name: "non-default",
			},
			bodyAssertion: func(t *testing.T, body string) {
				require.JSONEq(t, `{"id":"bt6kdc3d0cvp6u8u3ft0", "name": "non-default", "code": 123, "all": false}`, body)
			},
		},
		{
			name: "id & name & code",
			body: body{
				ID:   "bt6kdc3d0cvp6u8u3ft0",
				Name: "non-default",
				Code: 456,
			},
			bodyAssertion: func(t *testing.T, body string) {
				require.JSONEq(t, `{"id":"bt6kdc3d0cvp6u8u3ft0", "name": "non-default", "code": 456, "all": false}`, body)
			},
		},
		{
			name: "id & name & code & all",
			body: body{
				ID:   "bt6kdc3d0cvp6u8u3ft0",
				Name: "non-default",
				Code: 456,
				All:  true,
			},
			bodyAssertion: func(t *testing.T, body string) {
				require.JSONEq(t, `{"id":"bt6kdc3d0cvp6u8u3ft0", "name": "non-default", "code": 456, "all": true}`, body)
			},
		},
		{
			name: "id & page(num)",
			body: body{
				ID: "bt6kdc3d0cvp6u8u3ft0",
				Page: &page{
					Num: 10,
				},
			},
			bodyAssertion: func(t *testing.T, body string) {
				require.JSONEq(t, `
{
  "id": "bt6kdc3d0cvp6u8u3ft0",
  "name": "default",
  "code": 123,
  "all": false,
  "page": {
    "num": 10,
    "size": 10,
    "order": "desc"
  }
}
        `, body)
			},
		},
		{
			name: "id & page(num & order)",
			body: body{
				ID: "bt6kdc3d0cvp6u8u3ft0",
				Page: &page{
					Num:   10,
					Order: "asc",
				},
			},
			bodyAssertion: func(t *testing.T, body string) {
				require.JSONEq(t, `
{
  "id": "bt6kdc3d0cvp6u8u3ft0",
  "name": "default",
  "code": 123,
  "all": false,
  "page": {
    "num": 10,
    "size": 10,
    "order": "asc"
  }
}
        `, body)
			},
		},
		{
			name: "id & page & filters(one element and contains field)",
			body: body{
				ID: "bt6kdc3d0cvp6u8u3ft0",
				Page: &page{
					Num:   10,
					Order: "asc",
				},
				Filters: []filter{
					{
						Field: "code",
					},
				},
			},
			bodyAssertion: func(t *testing.T, body string) {
				require.JSONEq(t, `
{
  "id": "bt6kdc3d0cvp6u8u3ft0",
  "name": "default",
  "code": 123,
  "all": false,
  "page": {
    "num": 10,
    "size": 10,
    "order": "asc"
  },
  "filters": [
    {
      "field": "code",
      "op": "eq",
      "value": 123
    }
  ]
}
        `, body)
			},
		},
		{
			name: "id & page & filters(one element and contains field & op & value)",
			body: body{
				ID: "bt6kdc3d0cvp6u8u3ft0",
				Page: &page{
					Num:   10,
					Order: "asc",
				},
				Filters: []filter{
					{
						Field: "code",
						OP:    "ne",
						Value: 456,
					},
				},
			},
			bodyAssertion: func(t *testing.T, body string) {
				require.JSONEq(t, `
{
  "id": "bt6kdc3d0cvp6u8u3ft0",
  "name": "default",
  "code": 123,
  "all": false,
  "page": {
    "num": 10,
    "size": 10,
    "order": "asc"
  },
  "filters": [
    {
      "field": "code",
      "op": "ne",
      "value": 456
    }
  ]
}
        `, body)
			},
		},
		{
			name: "id & page & filters(multiple elements)",
			body: body{
				ID: "bt6kdc3d0cvp6u8u3ft0",
				Page: &page{
					Num:   10,
					Order: "asc",
				},
				Filters: []filter{
					{
						Value: 456,
					},
					{
						OP: "ne",
					},
					{
						Field: "code",
						Value: 456,
					},
					{
						OP:    "ne",
						Value: 789,
					},
					{
						Field: "code",
						OP:    "ne",
						Value: 456,
					},
				},
			},
			bodyAssertion: func(t *testing.T, body string) {
				require.JSONEq(t, `
{
  "id": "bt6kdc3d0cvp6u8u3ft0",
  "name": "default",
  "code": 123,
  "all": false,
  "page": {
    "num": 10,
    "size": 10,
    "order": "asc"
  },
  "filters": [
    {
      "field": "name",
      "op": "eq",
      "value": 456
    },
    {
      "field": "name",
      "op": "ne",
      "value": 123
    },
    {
      "field": "code",
      "op": "eq",
      "value": 456
    },
    {
      "field": "name",
      "op": "ne",
      "value": 789
    },
    {
      "field": "code",
      "op": "ne",
      "value": 456
    }
  ]
}
        `, body)
			},
		},
		{
			name: "social_network(oneOf)",
			body: body{
				ID: "bt6kdc3d0cvp6u8u3ft0",
				SocialNetwork: &socialNetwork{
					Platform: "facebook",
				},
			},
			bodyAssertion: func(t *testing.T, body string) {
				require.JSONEq(t, `
{
  "id": "bt6kdc3d0cvp6u8u3ft0",
  "name": "default",
  "code": 123,
  "all": false,
  "social_network": {
    "platform": "facebook",
    "fb_link": "www.facebook.com"
  }
}
        `, body)
			},
		},
		{
			name: "social_network_2(anyOf)",
			body: body{
				ID: "bt6kdc3d0cvp6u8u3ft0",
				SocialNetwork2: &socialNetwork{
					Platform: "facebook",
				},
			},
			bodyAssertion: func(t *testing.T, body string) {
				require.JSONEq(t, `
{
  "id": "bt6kdc3d0cvp6u8u3ft0",
  "name": "default",
  "code": 123,
  "all": false,
  "social_network_2": {
    "platform": "facebook",
    "fb_link": "www.facebook.com"
  }
}
        `, body)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.body)
			require.NoError(t, err)
			httpReq, err := http.NewRequest(http.MethodPost, "/accounts", bytes.NewReader(b))
			require.NoError(t, err)
			httpReq.Header.Add(headerCT, "application/json")

			route, pathParams, err := router.FindRoute(httpReq)
			require.NoError(t, err)

			err = ValidateRequest(sl.Context, &RequestValidationInput{
				Request:    httpReq,
				PathParams: pathParams,
				Route:      route,
			})
			require.NoError(t, err)

			validatedReqBody, err := ioutil.ReadAll(httpReq.Body)
			require.NoError(t, err)
			tc.bodyAssertion(t, string(validatedReqBody))
		})
	}
}
