package main

import (
	"fmt"
	"strings"

	"github.com/libp2p/go-libp2p/core/network"
)

func (c *P2Papp) sendTextHandler(text string, rendezvous string) {

	c.writeDataRendFunc(c.textproto, rendezvous, func(stream network.Stream) {

		n, err := stream.Write([]byte(rendezvous + "$" + text))
		fmt.Println(n, err)

		stream.Close()
		stream.Reset()
		return

	})

}

func (c *P2Papp) receiveTexthandler(stream network.Stream) {
	fmt.Println("receiveTexthandler")

	buff := make([]byte, 2000)

	_, err := stream.Read(buff)

	if err != nil {

		//c.disconnectHost(stream, err, string(stream.Protocol()))
		return

	}

	data := strings.SplitN(string(buff[:]), "$", 2)
	var rendezvous string
	var text string
	if len(data) > 1 {
		rendezvous = data[0]
		text = data[1]

	}
	fmt.Printf("[%s] %s = %s \n", rendezvous, stream.Conn().RemotePeer(), text)

	buff = nil
	stream.Close()
	stream.Reset()
	return

}
