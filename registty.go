package myrpc

import (
	"sync"
	"time"
)

type RpcRegistty struct {
	timeout time.Duration
	mu      sync.Mutex
	servers map[string]*ServerItem
}

type ServerItem struct {
	Addr  string
	start time.Time
}

const (
	defaultPath    = "_myrpc_ registry"
	defaultTimeout = time.Minute * 5
)

func New(timeout time.Duration) *RpcRegistty {
	return &RpcRegistty{
		servers: make(map[string]*ServerItem),
		timeout: timeout,
	}
}

var DefaultRegistry = New(defaultTimeout)
