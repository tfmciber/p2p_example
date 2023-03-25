package main

import (
	"fmt"
	"strings"

	"github.com/libp2p/go-libp2p/core/network"
)

func (c *P2Papp) sendTextHandler(text string, rendezvous string) {

	c.writeDataRendFunc(c.textproto, rendezvous, func(stream network.Stream) {

		stream.Write([]byte(rendezvous + "$" + text))

	})

}

func (c *P2Papp) receiveTexthandler(stream network.Stream) {

	c.readData(stream, 2000, func(buff []byte, stream network.Stream) {

		data := strings.SplitN(string(buff[:]), "$", 2)
		var rendezvous string
		var text string
		if len(data) > 1 {
			rendezvous = data[0]
			text = data[1]

		}
		fmt.Printf("[%s] %s = %s \n", rendezvous, stream.Conn().RemotePeer(), text)

		buff = nil

	})

}
