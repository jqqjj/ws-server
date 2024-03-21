package main

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"time"
	"unsafe"
)

func main() {
	var (
		err error
		uri string

		conn *websocket.Conn

		dial = websocket.Dialer{
			HandshakeTimeout:  time.Second * 15,
			EnableCompression: false,
		}
	)

	uri = fmt.Sprintf("ws://%s/", "localhost:8089")

	if conn, _, err = dial.DialContext(context.Background(), uri, nil); err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	req := struct {
		UUID    string `json:"uuid"`
		Command string `json:"command"`
		Payload any    `json:"payload"`
	}{
		UUID:    "1234",
		Command: "test",
		Payload: struct {
			Username string `json:"username"`
		}{
			Username: "admin",
		},
	}

	for {
		req.UUID = StringRandom(6)
		conn.WriteJSON(req)

		var data []byte
		if _, data, err = conn.ReadMessage(); err != nil {
			log.Fatalln(err)
		}
		log.Printf("收到：%s\n", data)

		time.Sleep(time.Second)
	}
}

func StringRandom(n int) string {
	letters := "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	src := rand.NewSource(time.Now().UnixNano())
	letterIdBits := 6
	var letterIdMask int64 = 1<<letterIdBits - 1
	letterIdMax := 63 / letterIdBits

	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdMax letters!
	for i, cache, remain := n-1, src.Int63(), letterIdMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdMax
		}
		if idx := int(cache & letterIdMask); idx < len(letters) {
			b[i] = letters[idx]
			i--
		}
		cache >>= letterIdBits
		remain--
	}
	return *(*string)(unsafe.Pointer(&b))
}
