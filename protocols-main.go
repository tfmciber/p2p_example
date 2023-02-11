package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/libp2p/go-libp2p/core/network"
)

//func for removing and disconecting a peer
func DisconnectHost(stream network.Stream, err error) {
	fmt.Println("Disconnecting host:", err)
	Host.Network().ClosePeer(stream.Conn().RemotePeer())
}
func handleStream(stream network.Stream) {

	switch stream.Protocol() {

	case "/chat/1.1.0":
		ReceiveTexthandler(stream)
	case "/audio/1.1.0":

		ReceiveAudioHandler(stream)
	case "/file/1.1.0":

		ReceiveFilehandler(stream)

	}

}

//func to close all strems of a given protocol
func CloseStreams(ProtocolID string) {

	for _, peerid := range Host.Network().Peers() {

		for _, conns := range Host.Network().ConnsToPeer(peerid) {

			for _, stream := range conns.GetStreams() {

				if string(stream.Protocol()) == ProtocolID {

					stream.Close()
				}
			}
		}

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

func WriteData(Chan chan []byte, ProtocolID string) {
	for {
		data := <-Chan

		for _, peerid := range Host.Network().Peers() {
		restart:
			s := streamStart(context.Background(), peerid, ProtocolID)

			_, err := s.Write(data)

			if err != nil {
				fmt.Println("Write failed:", err)
				s.Close()
				goto restart
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

		if strings.HasPrefix(Data, "/") { // READ COMMANDS
			Data = strings.TrimPrefix(Data, "/")

			cmdChan <- Data

		} else {

			fmt.Println("Please enter a valid command")
		}
	}

}
