package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/libp2p/go-libp2p/core/network"
)

var textChan = make(chan []byte)

func SendTextHandler() {

	go ReadStdin()

	go WriteData(textChan, "/chat/1.1.0")

}

func ReceiveTexthandler(stream network.Stream) {

	go readData(stream, 2000, func(buff []byte, stream network.Stream) {

		fmt.Printf("\x1b[32m%s\x1b[0m> ", string(buff[:]))

		buff = nil

	})

}

func ReadStdin() {

	stdReader := bufio.NewReader(os.Stdin)

	for {

		fmt.Print("> ")
		Data, err := stdReader.ReadString('\n')
		if err != nil {

			return

		}

		textChan <- []byte(Data)
	}

}
