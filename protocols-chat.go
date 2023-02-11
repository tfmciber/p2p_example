package main

import (
	"fmt"

	"github.com/libp2p/go-libp2p/core/network"
)

func SendTextHandler() {

	go WriteData(textChan, "/chat/1.1.0")

}

func ReceiveTexthandler(stream network.Stream) {

	go readData(stream, 2000, func(buff []byte, stream network.Stream) {

		fmt.Printf("%s = %s \n", stream.Conn().RemotePeer(), string(buff[:]))

		buff = nil

	})

}
