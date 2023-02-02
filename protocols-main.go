package main

import (
	"bufio"
	"fmt"
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

func handleStream(stream network.Stream) {

	switch stream.Protocol() {

	case "/chat/1.1.0":
		ReceiveTexthandler(stream)
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

func WriteData() {

	for {

		data := <-textChan
		for _, stream := range SendStreams {

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
