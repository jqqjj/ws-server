package main

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/jqqjj/ws-server"
	"net/http"
)

func main() {
	srv := wsServer()
	ctx := context.Background()

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		var (
			err  error
			conn *websocket.Conn

			u = websocket.Upgrader{
				ReadBufferSize:  1024 * 4,
				WriteBufferSize: 1024 * 4,
				CheckOrigin:     func(*http.Request) bool { return true },
			}
		)

		if conn, err = u.Upgrade(writer, request, nil); err != nil {
			writer.Write([]byte("error when upgrade to websocket"))
			return
		}
		defer conn.Close()

		srv.Process(ctx, conn)
	})

	http.ListenAndServe("0.0.0.0:8089", nil)
}

func wsServer() *server.Server {
	s := server.NewServer()

	s.Use(func(next server.HandleFunc, r *server.Request, w *server.Response) {
		fmt.Println("s1")
		next(r, w)
	}, func(next server.HandleFunc, r *server.Request, w *server.Response) {
		fmt.Println("s2")
		next(r, w)
	})

	g := s.Group("api/", func(next server.HandleFunc, r *server.Request, w *server.Response) {
		fmt.Println("gm1")
		next(r, w)
	}, func(next server.HandleFunc, r *server.Request, w *server.Response) {
		fmt.Println("gm2")
		//w.FailWithMessage("error: gm2")
		next(r, w)
	})

	g.SetHandle("sea", func(r *server.Request, w *server.Response) {
		fmt.Printf("sea data:%s, command:%s, ip:%s, uuid:%s\n", r.Payload, r.Command, r.ClientIP, r.UUID)
		w.Success("sea")
	})

	g.Use(func(next server.HandleFunc, r *server.Request, w *server.Response) {
		fmt.Println("g1")
		next(r, w)
	}, func(next server.HandleFunc, r *server.Request, w *server.Response) {
		fmt.Println("g2")
		next(r, w)
	})

	subG := g.Group("pd/")

	subG.SetHandle("test", func(r *server.Request, w *server.Response) {
		//panic("panic  aaa")
		fmt.Printf("data:%s, command:%s, ip:%s, uuid:%s\n", r.Payload, r.Command, r.ClientIP, r.UUID)
		w.Success("test")
	}, func(next server.HandleFunc, r *server.Request, w *server.Response) {
		fmt.Println("c1")
		next(r, w)
	})

	return s
}
