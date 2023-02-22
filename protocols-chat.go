package main

import (
	"fmt"
	"strings"

	"github.com/libp2p/go-libp2p/core/network"
)

func SendTextHandler(text string, rendezvous string) {

	WriteDataRend([]byte(text), "/chat/1.1.0", rendezvous, true)

}

func ReceiveTexthandler(stream network.Stream) {

	readData(stream, 2000, func(buff []byte, stream network.Stream) {

		//if buff starts with /cmd/ then it is a command
		if string(buff[:5]) == "/cmd/" {
			//rest of the string until / is the command

			rendezvous := string(buff[5:])
			rendezvous = strings.Split(rendezvous, "/")[0]
			if !Contains(Ren[rendezvous], stream.Conn().RemotePeer()) {
				fmt.Print("New peer:", stream.Conn().RemotePeer(), "added to rendezvous:", rendezvous)
				Ren[rendezvous] = append(Ren[rendezvous], stream.Conn().RemotePeer())

			}

		} else {

			fmt.Printf("%s = %s \n", stream.Conn().RemotePeer(), string(buff[:]))

			buff = nil
		}

	})

}
