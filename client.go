package server

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"sync"
	"time"
)

type clientWSEnt struct {
	Version string `json:"version"`
	UUID    string `json:"uuid"`
	Command string `json:"command"`
	Payload any    `json:"payload"`
}

type clientWSRequest struct {
	body      clientWSEnt
	callbacks []func(bool, []byte)
	tryLeft   int
	cancel    context.CancelFunc
}

type Client struct {
	addr     string
	version  string
	writeCh  chan clientWSRequest
	querying sync.Map
	pubSub   *PubSub[string, []byte]

	onDialError func()
	onConnected func(conn *websocket.Conn)
	onClosed    func()
}

func NewClient(addr, version string) *Client {
	return &Client{
		addr:    addr,
		version: version,
		writeCh: make(chan clientWSRequest, 100),
		pubSub:  NewPubSub[string, []byte](),
	}
}

func (c *Client) Run(ctx context.Context) {
	var (
		err   error
		fails int
		conn  *websocket.Conn

		dial = websocket.Dialer{
			HandshakeTimeout:  time.Second * 30,
			EnableCompression: false,
		}
		ticker = time.NewTicker(time.Millisecond)
	)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		ticker.Stop()

		if conn, _, err = dial.DialContext(ctx, c.addr, nil); err != nil {
			if c.onDialError != nil {
				c.onDialError()
			}
			fails++
			ticker.Reset(time.Duration(fails) * 2 * time.Second)
			continue
		}
		fails = 0

		//connected事件
		if c.onConnected != nil {
			c.onConnected(conn)
		}

		c.loop(ctx, conn)

		//closed事件
		if c.onClosed != nil {
			c.onClosed()
		}

		ticker.Reset(time.Millisecond)
	}
}

func (c *Client) loop(ctx context.Context, conn *websocket.Conn) {
	var (
		err  error
		data []byte

		wg sync.WaitGroup

		subCtx, subCancel = context.WithCancel(ctx)
	)
	defer wg.Wait()
	defer subCancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(time.Second * 30)
		defer ticker.Stop()
		defer conn.Close() //处理外部退出时关闭conn
		for {
			select {
			case <-subCtx.Done():
				return
			case <-ticker.C:
			}
			c.AddRequest("ping", nil) //主动发心跳包
		}
	}()

	//建立单线程发送任务
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.runSending(subCtx, conn)
	}()

	//把正在发送中的队列重发
	histories := make([]clientWSRequest, 0)
	c.querying.Range(func(key, value any) bool {
		histories = append(histories, value.(clientWSRequest))
		c.querying.Delete(key)
		return true
	})
	for _, v := range histories {
		select {
		case <-ctx.Done():
			c.querying.Store(v.body.UUID, v)
		case c.writeCh <- v:
		}
	}

	for {
		var resp struct {
			UUID    string          `json:"uuid"`
			Type    string          `json:"type"` //response push
			Command string          `json:"command"`
			Body    json.RawMessage `json:"body"`
		}
		if _, data, err = conn.ReadMessage(); err != nil {
			return
		}
		if err = json.Unmarshal(data, &resp); err != nil {
			return
		}
		if resp.Type == "push" {
			c.pubSub.Publish(resp.Command, resp.Body)
			continue
		}

		if val, ok := c.querying.LoadAndDelete(resp.UUID); ok {
			ent := val.(clientWSRequest)
			if ent.cancel != nil {
				ent.cancel() //取消超时重试的任务
			}
			wg.Add(1)
			go func() { //处理回调函数
				defer wg.Done()
				for _, callback := range ent.callbacks {
					callback(true, resp.Body)
				}
			}()
		}
	}
}

func (c *Client) runSending(ctx context.Context, conn *websocket.Conn) {
	var (
		err error
		wg  sync.WaitGroup
		req clientWSRequest
	)
	defer wg.Wait()

	//接收请求向ws写入
	for {
		select {
		case <-ctx.Done():
			return
		case req = <-c.writeCh:
		}

		//没有重试机会的立即失败回调
		if req.tryLeft <= 0 {
			wg.Add(1)
			go func() { //处理回调函数(失败)
				defer wg.Done()
				for _, callback := range req.callbacks {
					callback(false, nil)
				}
			}()
			continue
		}

		//发送
		if err = conn.WriteJSON(req.body); err == nil {
			req.tryLeft-- //只有真正写入conn，才扣除重试次数
		}

		//保存到发送中队列
		subCtx, subCancel := context.WithCancel(context.Background())
		req.cancel = subCancel
		c.querying.Store(req.body.UUID, req)
		//设定定时器，超时继续重试
		wg.Add(1)
		go func(uuid string) {
			defer wg.Done()
			ticker := time.NewTimer(time.Second * 30)
			defer ticker.Stop()
			select {
			case <-ctx.Done(): //外部退出执行这里
				return
			case <-subCtx.Done(): //成功收到发送响应，会执行这里
				return
			case <-ticker.C:
			}
			//执行重试
			if val, ok := c.querying.LoadAndDelete(uuid); ok {
				select {
				case <-ctx.Done(): //外部退出时，重回querying列表
					c.querying.Store(uuid, val.(clientWSRequest))
				case c.writeCh <- val.(clientWSRequest):
				}
			}
		}(req.body.UUID)
	}
}

func (c *Client) AddRequest(command string, object any, callbacks ...func(bool, []byte)) error {
	return c.AddRequestWithTryTimes(1, command, object, callbacks...)
}

func (c *Client) AddRequestWithTryTimes(times int, command string, object any, callbacks ...func(bool, []byte)) error {
	var (
		reqId  = uuid.NewV4().String()
		ticker = time.NewTimer(time.Second * 3)
	)
	defer ticker.Stop()

	select {
	case <-ticker.C:
		return errors.New("timeout")
	case c.writeCh <- clientWSRequest{
		body: clientWSEnt{
			UUID:    reqId,
			Command: command,
			Payload: object,
		},
		callbacks: callbacks,
		tryLeft:   times,
	}:
		return nil
	}
}

func (c *Client) Subscribe(ctx context.Context, topic string, ch chan<- []byte) {
	c.pubSub.Subscribe(ctx, topic, ch)
}
