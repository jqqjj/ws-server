package server

import (
	"context"
	"github.com/gorilla/websocket"
	"sync"
)

type Conn struct {
	mux  sync.Mutex
	ctx  context.Context
	conn *websocket.Conn

	queueJson []any
	trigger   chan struct{}
}

func NewConn(ctx context.Context, conn *websocket.Conn) *Conn {
	c := &Conn{ctx: ctx, conn: conn, queueJson: make([]any, 0), trigger: make(chan struct{}, 1)}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-c.trigger:
				c.flush()
			}
		}
	}()
	return c
}

func (c *Conn) flush() {
	c.mux.Lock()
	arr := make([]any, 0, len(c.queueJson))
	arr = append(arr, c.queueJson...)
	c.queueJson = c.queueJson[0:0]
	c.mux.Unlock()

	for _, v := range arr {
		c.conn.WriteJSON(v)
	}
}

func (c *Conn) Close() error {
	return c.conn.Close()
}

func (c *Conn) ReadMessage() (messageType int, p []byte, err error) {
	return c.conn.ReadMessage()
}

func (c *Conn) SendJSON(v any) {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.queueJson = append(c.queueJson, v)
	select {
	case c.trigger <- struct{}{}:
	default:
	}
}
