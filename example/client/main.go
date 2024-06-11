package main

import (
	"context"
	"fmt"
	server "github.com/jqqjj/ws-server"
	"log"
	"math/rand"
	"time"
	"unsafe"
)

func main() {
	var (
		ch          = make(chan []byte)
		uri         = fmt.Sprintf("ws://%s/", "localhost:8089")
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*10)
	)
	defer cancel()

	client := server.NewClient(uri, "0.1", time.Second*15)
	go client.Run(ctx)

	client.Subscribe(ctx, "haha", ch)

	go func() {
		for v := range ch {
			log.Println("收到推送", string(v))
		}
	}()

	for {
		data, err := client.Send(ctx, "api/pd/test", struct {
			Username string `json:"username"`
		}{
			Username: "admin",
		})

		if err != nil {
			log.Println("错误", err)
			time.Sleep(time.Second)
			continue
		}

		log.Println("收到", string(data))

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
