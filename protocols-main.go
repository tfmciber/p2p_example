package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/libp2p/go-libp2p/core/network"
)

//func for removing and disconecting a peer
func DisconnectHost(stream network.Stream, err error) {
	fmt.Println("Disconnecting host:", err)
	Host.Network().ClosePeer(stream.Conn().RemotePeer())
	//func to get all keys from a map

}

func handleStream(stream network.Stream) {

	// check if a stream with the same protocol and peer is already open
	fmt.Print("New stream from:", stream.Conn().RemotePeer(), "Protocol:", stream.Protocol(), "Transport:", stream.Conn().ConnState().Transport)
	if GetStreamsFromPeerProto(stream.Conn().RemotePeer(), string(stream.Protocol())) != nil {

		switch stream.Protocol() {

		case "/chat/1.1.0":
			ReceiveTexthandler(stream)
		case "/audio/1.1.0":

			ReceiveAudioHandler(stream)
		case "/file/1.1.0":

			ReceiveFilehandler(stream)
		case "/bench/1.1.0":

			ReceiveBenchhandler(stream)

		}
	} else {
		stream.Close()
	}
}

//Function that reads data of size n from stream and calls f funcion
func readData(stream network.Stream, size uint16, f func(buff []byte, stream network.Stream)) {

	for {
		buff := make([]byte, size)

		_, err := stream.Read(buff)

		if err != nil {

			DisconnectHost(stream, err)
			return

		}
		f(buff, stream)

	}
}

func WriteDataRend(data []byte, ProtocolID string, rendezvous string) {

	// loop over Peers map
	for _, v := range Ren[rendezvous] {
		aux := Peers[v]
		if aux.online == true {
		restart:

			stream := streamStart(hostctx, v, ProtocolID)

			if stream == nil {
				fmt.Println("stream is nil")
				aux.online = false
				Peers[v] = aux

			} else {

				_, err := stream.Write(data)

				if err != nil {
					fmt.Println("Write failed: restarting ", err)
					stream.Close()
					goto restart

				}
			}
		}
	}
}

//function to send data over a stream

//Function that reads data from stdin and sends it to the cmd channel
func ReadStdin() {

	for {

		fmt.Print("$ ")

		var Data string
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		Data = scanner.Text()

		cmdChan <- Data

	}

}
