package server

import "github.com/gorilla/websocket"

type Push struct {
	conn *websocket.Conn
}

func NewPush(conn *websocket.Conn) *Push {
	return &Push{
		conn: conn,
	}
}

func (p *Push) GetConn() *websocket.Conn {
	return p.conn
}

func (p *Push) Write(command string, object any) error {
	body := struct {
		Command string `json:"command"`
		Data    any    `json:"data"`
	}{
		Command: command,
		Data:    object,
	}

	resp := struct {
		Type string `json:"type"`
		Body any    `json:"body"`
	}{
		Type: "push",
		Body: body,
	}

	return p.conn.WriteJSON(resp)
}
