package server

import (
	"github.com/gorilla/websocket"
)

type Response struct {
	uuid    string
	replied bool
	conn    *websocket.Conn
}

func NewResponse(uuid string, conn *websocket.Conn) *Response {
	return &Response{
		uuid: uuid,
		conn: conn,
	}
}

func (r *Response) Success(object any) error {
	return r.response(0, "Success", object)
}

func (r *Response) Fail() error {
	return r.response(1, "Fail", nil)
}

func (r *Response) FailWithCode(code int) error {
	return r.response(code, "Fail", nil)
}

func (r *Response) FailWithMessage(message string) error {
	return r.response(1, message, nil)
}

func (r *Response) FailWithCodeAndMessage(code int, message string) error {
	return r.response(code, message, nil)
}

func (r *Response) response(code int, message string, object any) error {
	if r.replied {
		return ErrDuplicate
	}
	r.replied = true

	body := struct {
		UUID    string `json:"uuid"`
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    any    `json:"data"`
	}{
		UUID:    r.uuid,
		Code:    code,
		Message: message,
		Data:    object,
	}

	resp := struct {
		Type string `json:"type"` //push response
		Body any    `json:"body"`
	}{
		Type: "response",
		Body: body,
	}

	return r.conn.WriteJSON(resp)
}
