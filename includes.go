package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gen2brain/malgo"
	"github.com/libp2p/go-libp2p/core/peer"
)

var cmdChan = make(chan string)
var FoundPeersDHT <-chan peer.AddrInfo
var FoundPeersMDNS <-chan peer.AddrInfo
var hostctx context.Context
var discctx_m context.Context
var discctx_d context.Context
var ctx1 *malgo.AllocatedContext
var textChan = make(chan []byte)

func initCTX() {
	var cancel context.CancelFunc
	hostctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	var err1 error
	ctx1, err1 = malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {

	})
	if err1 != nil {
		fmt.Println(err1)
		os.Exit(1)
	}
	defer func() {
		_ = ctx1.Uninit()
		ctx1.Free()
	}()

}
