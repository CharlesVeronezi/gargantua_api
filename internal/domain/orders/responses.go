package orders

import (
	"encoding/json"
	"encoding/xml"
	"github.com/go-chi/render"
	"net/http"
)

type Response struct {
	body        interface{}
	Code        int
	contentType string
}

type Error struct {
	Message string `json:"message"`
}

type CreateOrderResponse struct {
	ID string `json:"id"`
}

type GetOrdersResponse struct {
	Orders []CreateOrderRequest `json:"orders"`
}

func (resp *Response) Render(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", resp.contentType)
	render.Status(r, resp.Code)
	return nil
}

func (resp *Response) Status(code int) *Response {
	resp.Code = code
	return resp
}

func (resp *Response) ContentType(contentType string) *Response {
	resp.contentType = contentType
	return resp
}

func (resp *Response) MarshalJSON() ([]byte, error) {
	return json.Marshal(resp.body)
}

func (resp *Response) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.Encode(resp.body)
}

func ErrorJSON400Response(body Error) *Response {
	return &Response{
		body:        body,
		Code:        400,
		contentType: "application/json",
	}
}

func CreateOrderJSON201Response(body CreateOrderResponse) *Response {
	return &Response{
		body:        body,
		Code:        201,
		contentType: "application/json",
	}
}

func OrdersJSON200Response(body GetOrdersResponse) *Response {
	return &Response{
		body:        body,
		Code:        200,
		contentType: "application/json",
	}
}

func UniqueOrderJSON200Response(body CreateOrderRequest) *Response {
	return &Response{
		body:        body,
		Code:        200,
		contentType: "application/json",
	}
}
