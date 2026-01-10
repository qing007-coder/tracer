package tracer

import (
	"fmt"
	"net/http"
)
import "google.golang.org/grpc/metadata"

type Carrier interface {
	Set(key string, value interface{})
	Get(key string) interface{}
	Foreach(f func(key string, value interface{}))
}

type HttpCarrier struct {
	Header http.Header
}

func (c *HttpCarrier) Set(key string, value interface{}) {
	c.Header.Set(key, fmt.Sprintf("%v", value))
}

func (c *HttpCarrier) Get(key string) interface{} {
	return c.Header.Get(key)
}

func (c *HttpCarrier) Foreach(f func(key string, value interface{})) {
	for k, v := range c.Header {
		f(k, v)
	}
}

type GRPCCarrier struct {
	MD metadata.MD
}

func (c *GRPCCarrier) Set(key string, value interface{}) {
	c.MD.Set(key, fmt.Sprintf("%v", value))
}

func (c *GRPCCarrier) Get(key string) interface{} {
	return c.MD[key]
}

func (c *GRPCCarrier) Foreach(f func(key string, value interface{})) {
	for k, v := range c.MD {
		f(k, v)
	}
}
