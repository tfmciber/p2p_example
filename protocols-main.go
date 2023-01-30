package main

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/libp2p/go-libp2p/core/network"
)

func DisconnectHost(stream network.Stream) {

	for k := range SendStreams {

		if strings.Split(k, "-")[0] == strings.Split(stream.ID(), "-")[0] {

			delete(SendStreams, k)

		}

	}
	fmt.Printf("Hay %d conexiones", len(SendStreams))

}

func handleStream(stream network.Stream) {

	// Create a buffer stream for non blocking read and write.
	fmt.Print("handleStream stream.ID() ", stream.ID(), "\n")
	switch stream.Protocol() {

	case "/chat/1.1.0":
		ReceiveTexthandler(stream)
	}

}

//Function that reads data of size n from stream and calls f funcion
func readData(stream network.Stream, size uint16, f func(buff []byte, stream network.Stream)) {
	fmt.Print("readData stream.ID() ", stream.ID(), "\n")
	buff := make([]byte, size)
	for {

		r := bufio.NewReader(stream)

		_, err := r.Read(buff)

		if err != nil {
			fmt.Println("Error reading from buffer removing stream from list")
			fmt.Print("errr", err)
			DisconnectHost(stream)
			return

		}
		f(buff, stream)
	}
}

func WriteData() {

	for {

		data := <-textChan
		for _, stream := range SendStreams {
			fmt.Print(stream)
			w := bufio.NewWriter(stream)
			_, err := w.Write(data)
			if err != nil {
				fmt.Println("Error writing to buffer")

			}

			err = w.Flush()
			if err != nil {
				fmt.Println("Error flushing buffer")

			}

		}

	}
}
