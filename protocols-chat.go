package main

import (
	"fmt"

	"github.com/libp2p/go-libp2p/core/network"
)

func SendTextHandler(text string, rendezvous string) {

	fmt.Println("Sending text to:", rendezvous, "text:", text)
	WriteDataRend([]byte(text), "/chat/1.1.0", rendezvous)

}

func ReceiveTexthandler(stream network.Stream) {

	readData(stream, 2000, func(buff []byte, stream network.Stream) {

		fmt.Printf("%s = %s \n", stream.Conn().RemotePeer(), string(buff[:]))

		buff = nil

	})

}
