package main

import (
	"bufio"
	"fmt"

	"github.com/libp2p/go-libp2p/core/network"
)

func DisconnectHost(stream network.Stream) {

	delete(streams, stream.ID())
	falta eliminar el stream
	fmt.Printf("Hay %d Clientes conectados \n", len(streams))
	for k, v := range streams {
		fmt.Println(k, "value is", v)
	}

}

func handleStream(stream network.Stream) {

	// Create a buffer stream for non blocking read and write.

	switch stream.Protocol() {

	case "/chat/1.1.0":
		chat_func(stream)
	}

}

func readData(stream network.Stream, buff []byte, f func(buff []byte, stream network.Stream)) {
	for {
		r := bufio.NewReader(stream)

		n, err := r.Read(buff)

		if err != nil {
			fmt.Println("Error reading from buffer removing stream from list")
			fmt.Printf("\x1b[32m%s,%d\x1b[0m> ", string(buff[:]), n)
			fmt.Print("errr", err)
			DisconnectHost(stream)
		}
		f(buff, stream)

	}

}
