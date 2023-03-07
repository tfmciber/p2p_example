package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/libp2p/go-libp2p/core/network"
)

// func for removing and disconecting a peer
func disconnectHost(stream network.Stream, err error) {
	fmt.Println("Disconnecting host:", stream.Conn().RemotePeer(), err)
	Host.Network().ClosePeer(stream.Conn().RemotePeer())
	//func to get all keys from a map

}

func handleStream(stream network.Stream) {

	if getStreamsFromPeerProto(stream.Conn().RemotePeer(), string(stream.Protocol())) != nil {

		switch stream.Protocol() {

		case "/chat/1.1.0":
			receiveTexthandler(stream)
		case "/audio/1.1.0":

			receiveAudioHandler(stream)
		case "/file/1.1.0":

			receiveFilehandler(stream)
		case "/bench/1.1.0":

			receiveBenchhandler(stream)

		}
	} else {
		stream.Close()
	}
}

// Function that reads data of size n from stream and calls f funcion
func readData(stream network.Stream, size uint16, f func(buff []byte, stream network.Stream)) {

	for {
		buff := make([]byte, size)

		_, err := stream.Read(buff)

		if err != nil {

			disconnectHost(stream, err)
			return

		}
		f(buff, stream)

	}
}

func writeDataRend(data []byte, ProtocolID string, rendezvous string, verbose bool) {

	for _, v := range Ren[rendezvous] {

		if Host.Network().Connectedness(v) == network.Connected {
			if verbose {
				fmt.Println("Sending data to:", v)
			}

		restart:
			stream := streamStart(hostctx, v, ProtocolID)

			if stream == nil {
				fmt.Println("stream is nil")

			} else {

				_, err := stream.Write(data)

				if err != nil {
					fmt.Println("Write failed: restarting ", err)
					stream.Close()
					goto restart

				}
				if verbose {
					if err != nil {
						fmt.Println("Write failed: restarting ", err)
					} else {
						fmt.Println("Data sent to:", v)
					}
				}
			}
		}
	}

}

//function to send data over a stream

// Function that reads data from stdin and sends it to the cmd channel
func readStdin() {

	for {

		fmt.Print("$ ")

		var Data string
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		Data = scanner.Text()

		cmdChan <- Data

	}

}
