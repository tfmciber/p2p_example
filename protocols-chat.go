package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
)

func sendTextHandler(text string, rendezvous string) {

	writeDataRend([]byte(text), "/chat/1.1.0", rendezvous, true)

}

func receiveTexthandler(stream network.Stream) {

	readData(stream, 2000, func(buff []byte, stream network.Stream) {

		//if buff starts with /cmd/ then it is a command

		if string(buff[:5]) == "/cmd/" {
			//rest of the string until / is the command
			client.Reserve(context.Background(), Host, Host.Network().Peerstore().PeerInfo(stream.Conn().RemotePeer()))

			rendezvous := string(buff[5:])
			rendezvous = strings.Split(rendezvous, "/")[0]
			if !contains(Ren[rendezvous], stream.Conn().RemotePeer()) {
				log.Println("New peer:", stream.Conn().RemotePeer(), "added to rendezvous:", rendezvous)
				Ren[rendezvous] = append(Ren[rendezvous], stream.Conn().RemotePeer())

			}

		} else {

			fmt.Printf("%s = %s \n", stream.Conn().RemotePeer(), string(buff[:]))

			buff = nil
		}

	})

}
