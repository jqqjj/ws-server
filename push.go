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
	resp := struct {
		Type    string `json:"type"`
		Command string `json:"command"`
		Body    any    `json:"body"`
	}{
		Type:    "push",
		Command: command,
		Body:    object,
	}

	return p.conn.WriteJSON(resp)
}
