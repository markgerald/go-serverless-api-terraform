package docs

import (
	"github.com/swaggo/swag"
)

// This minimal docs package registers a basic Swagger 2.0 spec so that
// gin-swagger can serve /swagger without requiring code generation.
// You can later replace this with generated files from `swag init`.

type swaggerDoc struct{}

func (s *swaggerDoc) ReadDoc() string {
	// Minimal Swagger 2.0 specification that documents key endpoints and models.
	// Update as needed or replace with generated docs from swag.
	return `{
	  "swagger": "2.0",
	  "info": {
	    "title": "Orders API",
	    "description": "API for managing orders and order items",
	    "version": "1.0"
	  },
	  "basePath": "/",
	  "schemes": ["https", "http"],
	  "paths": {
	    "/orders": {
	      "get": {
	        "summary": "List orders",
	        "responses": {
	          "200": {
	            "description": "OK",
	            "schema": {"type": "array", "items": {"$ref": "#/definitions/models.Order"}}
	          }
	        }
	      },
	      "post": {
	        "summary": "Create order",
	        "parameters": [
	          {
	            "in": "body",
	            "name": "order",
	            "required": true,
	            "schema": {"$ref": "#/definitions/handlers.createOrderReq"}
	          }
	        ],
	        "responses": {
	          "201": {"description": "Created", "schema": {"$ref": "#/definitions/models.Order"}},
	          "400": {"description": "Bad Request"}
	        }
	      }
	    },
	    "/orders/{orderId}": {
	      "parameters": [{"name":"orderId","in":"path","required":true,"type":"string"}],
	      "get": {
	        "summary": "Get order",
	        "responses": {"200": {"description": "OK", "schema": {"$ref": "#/definitions/models.Order"}}}
	      },
	      "put": {
	        "summary": "Update order",
	        "parameters": [
	          {"name":"orderId","in":"path","required":true,"type":"string"},
	          {"in": "body", "name": "order", "required": true, "schema": {"$ref": "#/definitions/handlers.updateOrderReq"}}
	        ],
	        "responses": {"200": {"description": "OK", "schema": {"$ref": "#/definitions/models.Order"}}}
	      },
	      "delete": {
	        "summary": "Delete order",
	        "responses": {"204": {"description": "No Content"}}
	      }
	    },
	    "/orders/{orderId}/items": {
	      "parameters": [{"name":"orderId","in":"path","required":true,"type":"string"}],
	      "get": {
	        "summary": "List items",
	        "responses": {"200": {"description": "OK", "schema": {"type": "array", "items": {"$ref": "#/definitions/models.OrderItem"}}}}
	      },
	      "post": {
	        "summary": "Create item",
	        "parameters": [
	          {"name":"orderId","in":"path","required":true,"type":"string"},
	          {"in": "body", "name": "item", "required": true, "schema": {"$ref": "#/definitions/handlers.createItemReq"}}
	        ],
	        "responses": {"201": {"description": "Created", "schema": {"$ref": "#/definitions/models.OrderItem"}}}
	      }
	    },
	    "/orders/{orderId}/items/{itemId}": {
	      "parameters": [
	        {"name":"orderId","in":"path","required":true,"type":"string"},
	        {"name":"itemId","in":"path","required":true,"type":"string"}
	      ],
	      "get": {"summary": "Get item", "responses": {"200": {"description": "OK", "schema": {"$ref": "#/definitions/models.OrderItem"}}}},
	      "put": {
	        "summary": "Update item",
	        "parameters": [
	          {"name":"orderId","in":"path","required":true,"type":"string"},
	          {"name":"itemId","in":"path","required":true,"type":"string"},
	          {"in": "body", "name": "item", "required": true, "schema": {"$ref": "#/definitions/handlers.updateItemReq"}}
	        ],
	        "responses": {"200": {"description": "OK", "schema": {"$ref": "#/definitions/models.OrderItem"}}}
	      },
	      "delete": {"summary": "Delete item", "responses": {"204": {"description": "No Content"}}}
	    }
	  },
	  "definitions": {
	    "models.Order": {
	      "type": "object",
	      "properties": {
	        "id": {"type": "string", "example": "b5e1c2f4-1234-4a7e-8c1a-abcdef012345"},
	        "customer_name": {"type": "string", "example": "Alice"},
	        "status": {"type": "string", "example": "new"},
	        "created_at": {"type": "string", "example": "2024-01-01T12:00:00Z"},
	        "updated_at": {"type": "string", "example": "2024-01-01T12:00:00Z"}
	      }
	    },
	    "models.OrderItem": {
	      "type": "object",
	      "properties": {
	        "order_id": {"type": "string", "example": "b5e1c2f4-1234-4a7e-8c1a-abcdef012345"},
	        "id": {"type": "string", "example": "1f2e3d4c-5678-4b3a-9c0d-abcdef012345"},
	        "product_name": {"type": "string", "example": "Keyboard"},
	        "quantity": {"type": "integer", "format": "int32", "example": 2},
	        "price": {"type": "number", "format": "double", "example": 99.99},
	        "created_at": {"type": "string", "example": "2024-01-01T12:00:00Z"},
	        "updated_at": {"type": "string", "example": "2024-01-01T12:00:00Z"}
	      }
	    },
	    "handlers.createOrderReq": {
	      "type": "object",
	      "required": ["customer_name"],
	      "properties": {
	        "customer_name": {"type": "string"},
	        "status": {"type": "string"}
	      }
	    },
	    "handlers.updateOrderReq": {
	      "type": "object",
	      "properties": {
	        "customer_name": {"type": "string"},
	        "status": {"type": "string"}
	      }
	    },
	    "handlers.createItemReq": {
	      "type": "object",
	      "required": ["product_name", "quantity", "price"],
	      "properties": {
	        "product_name": {"type": "string"},
	        "quantity": {"type": "integer", "format": "int32"},
	        "price": {"type": "number", "format": "double"}
	      }
	    },
	    "handlers.updateItemReq": {
	      "type": "object",
	      "properties": {
	        "product_name": {"type": "string"},
	        "quantity": {"type": "integer", "format": "int32"},
	        "price": {"type": "number", "format": "double"}
	      }
	    }
	  }
	}`
}

func init() {
	// Register default swagger doc instance used by gin-swagger.
	swag.Register(swag.Name, &swaggerDoc{})
}
