package server

import (
	"github.com/gorilla/websocket"
)

type ResponseBody struct {
	UUID    string `json:"uuid"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type Response struct {
	uuid   string
	filled bool
	body   ResponseBody
	conn   *websocket.Conn
}

func NewResponse(uuid string, conn *websocket.Conn) *Response {
	return &Response{
		uuid: uuid,
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
		UUID:    r.uuid,
		Code:    code,
		Message: message,
		Data:    object,
	}
}
