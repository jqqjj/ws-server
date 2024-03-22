package server

import (
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"net"
)

type Server struct {
	*GroupServer
}

func NewServer() *Server {
	return &Server{
		GroupServer: NewGroup(),
	}
}

func (s *Server) Process(ctx context.Context, conn *websocket.Conn) {
	var (
		err  error
		data []byte

		done = make(chan struct{})
	)
	defer close(done)

	go func() {
		select {
		case <-ctx.Done():
		case <-done:
		}
		conn.Close()
	}()

	addr := conn.RemoteAddr().(*net.TCPAddr)
	ip := addr.IP.String()

	for {
		var req struct {
			Version string          `json:"version"`
			UUID    string          `json:"uuid"`
			Command string          `json:"command"`
			Payload json.RawMessage `json:"payload"`
		}
		if _, data, err = conn.ReadMessage(); err != nil {
			return
		}
		if err = json.Unmarshal(data, &req); err != nil {
			NewResponse("", conn).FailWithCodeAndMessage(404, "error to parse request")
			continue
		}

		handleEntity, ok := s.handles[req.Command]
		if !ok {
			NewResponse(req.UUID, conn).FailWithCodeAndMessage(404, "command not found")
			continue
		}

		next := handleEntity.handler
		for i := len(handleEntity.middlewares) - 1; i >= 0; i-- {
			nextFunc := handleEntity.middlewares[i]
			next = func(n HandleFunc) HandleFunc {
				return func(r *Request, w *Response) {
					nextFunc(n, r, w)
				}
			}(next)
		}

		reqEntity := Request{
			Version:  req.Version,
			UUID:     req.UUID,
			Command:  req.Command,
			Payload:  req.Payload,
			ClientIP: ip,
		}
		next(&reqEntity, NewResponse(req.UUID, conn))
	}
}
