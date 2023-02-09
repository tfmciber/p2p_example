package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/libp2p/go-libp2p/core/network"
)

// Removes host,stream from list and prints the error
func DisconnectHost(stream network.Stream, err error) {

	for k := range SendStreams {

		if strings.Split(k, "-")[0] == strings.Split(stream.ID(), "-")[0] {

			fmt.Printf("Removed %s stream(%s,%s) %s due to %s \n", stream.Conn().LocalPeer().String(), stream.Protocol(), stream.Conn().ConnState(), stream.ID(), err)
			delete(SendStreams, k)

		}

	}
	fmt.Printf("%d  Conections left\n", len(SendStreams))

}

func listUSers() {
	users := make(map[string]bool)
	for k := range SendStreams {
		user := strings.Split(k, "-")[0]
		if !users[user] {
			users[user] = true
			fmt.Println(user)
		}
	}

}
func isconnected(stream network.Stream) bool {
	for k := range SendStreams {

		if strings.Split(k, "-")[0] == strings.Split(stream.ID(), "-")[0] {
			return true
		}

	}
	return false
}

func handleStream(stream network.Stream) {

	//check if bidirectional stream exists
	if !isconnected(stream) {

		streamStart(context.Background(), stream.Conn().RemotePeer())

	}

	switch stream.Protocol() {

	case "/chat/1.1.0":
		ReceiveTexthandler(stream)
	case "/audio/1.1.0":

		ReceiveAudioHandler(stream)
	case "/file/1.1.0":

		ReceiveFilehandler(stream)

	}

}

//Function that reads data of size n from stream and calls f funcion
func readData(stream network.Stream, size uint16, f func(buff []byte, stream network.Stream)) {

	for {
		buff := make([]byte, size)
		r := bufio.NewReader(stream)

		_, err := r.Read(buff)

		if err != nil {

			DisconnectHost(stream, err)
			return

		}
		f(buff, stream)

	}
}
func SimplereadData(stream network.Stream, size uint16) []byte {

	buff := make([]byte, size)
	r := bufio.NewReader(stream)

	_, err := r.Read(buff)
	fmt.Print("leemos ", string(buff))
	if err != nil {

		DisconnectHost(stream, err)
		return nil

	}

	return buff

}
func WriteData(Chan chan []byte, protocol string) {

	for {

		data := <-Chan

		for _, stream := range SendStreams {

			if string(stream.Protocol()) == protocol {
				w := bufio.NewWriter(stream)
				_, err := w.Write(data)
				if err != nil {
					fmt.Println("Error writing to buffer")

				}
				err = w.Flush()
				if err != nil {
					fmt.Println("Error flushing buffer", err)

				}
			}
		}
	}
}

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

		} else if strings.HasPrefix(Data, "!") { // Send File
			Data = strings.TrimPrefix(Data, "!")
			sendFile(Data)

		} else if len(SendStreams) > 0 {

			textChan <- []byte(Data)
		} else {

			fmt.Println("No command selected and not connected to any host")
		}
	}

}
