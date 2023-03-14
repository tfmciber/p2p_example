package main

import (
	"sync"

	"github.com/libp2p/go-libp2p/core/peer"
)

type safePeerSlice struct {
	mu   sync.Mutex
	data map[string][]peer.ID
}

var Ren = safePeerSlice{data: make(map[string][]peer.ID)}

func (c *safePeerSlice) Add(key string, value peer.ID) {
	c.mu.Lock()
	for _, a := range c.data[key] {
		if a == value {
			return
		}
	}
	c.data[key] = append(c.data[key], value)

	c.mu.Unlock()
}
func (c *safePeerSlice) Clear() {
	c.mu.Lock()
	c.data = make(map[string][]peer.ID)
	c.mu.Unlock()
}
func (c *safePeerSlice) Get(key string) []peer.ID {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.data[key]
}
