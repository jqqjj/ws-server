package server

import (
	"context"
	"github.com/gorilla/websocket"
)

type Conn struct {
	ctx  context.Context
	conn *websocket.Conn
}

func NewConn(ctx context.Context, conn *websocket.Conn) *Conn {
	return &Conn{ctx, conn}
}

func (c *Conn) Close() error {
	return c.conn.Close()
}

func (c *Conn) ReadMessage() (messageType int, p []byte, err error) {
	return c.conn.ReadMessage()
}

func (c *Conn) SendJSON(v any) error {
	return c.conn.WriteJSON(v)
}
