package main

import (
	"fmt"

	"github.com/libp2p/go-libp2p/core/network"
)

func sendTextHandler(text string, rendezvous string) {

	writeDataRend([]byte(text), string(textproto), rendezvous, true)

}

func receiveTexthandler(stream network.Stream) {

	readData(stream, 2000, func(buff []byte, stream network.Stream) {

		fmt.Printf("%s = %s \n", stream.Conn().RemotePeer(), string(buff[:]))

		buff = nil

	})

}
