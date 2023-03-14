package main

import (
	"fmt"

	"github.com/libp2p/go-libp2p/core/network"
)

func (c *P2Papp) sendTextHandler(text string, rendezvous string) {

	c.writeDataRend([]byte(text), c.textproto, rendezvous, true)

}

func (c *P2Papp) receiveTexthandler(stream network.Stream) {

	c.readData(stream, 2000, func(buff []byte, stream network.Stream) {

		fmt.Printf("%s = %s \n", stream.Conn().RemotePeer(), string(buff[:]))

		buff = nil

	})

}
