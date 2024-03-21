package server

import (
	"encoding/json"
	"sync"
)

type Request struct {
	UUID     string
	Command  string
	Payload  json.RawMessage
	ClientIP string

	meta sync.Map
}

func (r *Request) Set(key string, value any) {
	r.meta.Store(key, value)
}

func (r *Request) Get(key string) (any, bool) {
	return r.meta.Load(key)
}

func (r *Request) MustGet(key string) any {
	if o, ok := r.Get(key); ok {
		return o
	}
	panic("object not found: " + key)
}
