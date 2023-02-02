package main

import (
	"sync"

	"github.com/libp2p/go-libp2p/core/peer"
)

func merge(cs ...<-chan peer.AddrInfo) <-chan peer.AddrInfo {
	out := make(chan peer.AddrInfo)
	var wg sync.WaitGroup
	wg.Add(len(cs))
	for _, c := range cs {
		go func(c <-chan peer.AddrInfo) {
			for v := range c {
				out <- v
			}
		}(c)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
func chk(err error) {
	if err != nil {
		panic(err)
	}
}
