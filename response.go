package server

import (
	"github.com/gorilla/websocket"
)

type ResponseBody struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type Response struct {
	filled bool
	body   ResponseBody
	conn   *websocket.Conn
}

func NewResponse(conn *websocket.Conn) *Response {
	return &Response{
		conn: conn,
	}
}

func (r *Response) GetConn() *websocket.Conn {
	return r.conn
}

func (r *Response) GetResponseBody() ResponseBody {
	return r.body
}

func (r *Response) SetResponseBody(body ResponseBody) {
	r.filled = true
	r.body = body
}

func (r *Response) Success(object any) {
	r.write(0, "Success", object)
}

func (r *Response) Fail() {
	r.write(1, "Fail", nil)
}

func (r *Response) FailWithCode(code int) {
	r.write(code, "Fail", nil)
}

func (r *Response) FailWithMessage(message string) {
	r.write(1, message, nil)
}

func (r *Response) FailWithCodeAndMessage(code int, message string) {
	r.write(code, message, nil)
}

func (r *Response) write(code int, message string, object any) {
	if r.filled {
		return
	}
	r.filled = true

	r.body = ResponseBody{
		Code:    code,
		Message: message,
		Data:    object,
	}
}
