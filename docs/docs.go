// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/vehicle/{tokenId}/trips": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Lists vehicle trips.",
                "produces": [
                    "application/json"
                ],
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Vehicle token id",
                        "name": "tokenId",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "Page of trips to retrieve. Defaults to 1.",
                        "name": "page",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_DIMO-Network_trips-api_internal_helper.VehicleTrips"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "github_com_DIMO-Network_trips-api_internal_helper.PointTime": {
            "type": "object",
            "properties": {
                "latitude": {
                    "type": "number",
                    "example": 37.7749
                },
                "longitude": {
                    "type": "number",
                    "example": -122.4194
                },
                "time": {
                    "type": "string",
                    "example": "2023-05-04T09:00:00Z"
                }
            }
        },
        "github_com_DIMO-Network_trips-api_internal_helper.TripDetails": {
            "type": "object",
            "properties": {
                "end": {
                    "$ref": "#/definitions/github_com_DIMO-Network_trips-api_internal_helper.TripEvent"
                },
                "id": {
                    "type": "string",
                    "example": "2Y83IHPItgk0uHD7hybGnA776Bo"
                },
                "start": {
                    "$ref": "#/definitions/github_com_DIMO-Network_trips-api_internal_helper.TripEvent"
                }
            }
        },
        "github_com_DIMO-Network_trips-api_internal_helper.TripEvent": {
            "type": "object",
            "properties": {
                "actual": {
                    "$ref": "#/definitions/github_com_DIMO-Network_trips-api_internal_helper.PointTime"
                },
                "estimate": {
                    "$ref": "#/definitions/github_com_DIMO-Network_trips-api_internal_helper.PointTime"
                }
            }
        },
        "github_com_DIMO-Network_trips-api_internal_helper.VehicleTrips": {
            "type": "object",
            "properties": {
                "currentPage": {
                    "type": "integer",
                    "example": 1
                },
                "totalPages": {
                    "type": "integer",
                    "example": 1
                },
                "trips": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/github_com_DIMO-Network_trips-api_internal_helper.TripDetails"
                    }
                }
            }
        }
    },
    "securityDefinitions": {
        "BearerAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "",
	BasePath:         "/v1",
	Schemes:          []string{},
	Title:            "DIMO Segment API",
	Description:      "segments",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
