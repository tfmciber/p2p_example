package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/libp2p/go-libp2p/core/network"
)

func SendTextHandler() {

	go WriteData(textChan, "/chat/1.1.0")

}

func ReceiveTexthandler(stream network.Stream) {

	go readData(stream, 2000, func(buff []byte, stream network.Stream) {

		fmt.Printf("\x1b[32m%s\x1b[0m> \n", string(buff[:]))

		buff = nil

	})

}

func ReadStdin() {

	stdReader := bufio.NewReader(os.Stdin)

	for {

		fmt.Print("$ ")
		fmt.Println(len(SendStreams))
		Data, err := stdReader.ReadString('\n')
		Data = strings.TrimSuffix(Data, "\n")
		if err != nil {

			return

		}
		if strings.HasPrefix(Data, "/") {
			Data = strings.TrimPrefix(Data, "/")

			cmdChan <- Data

		} else if len(SendStreams) > 0 {
			fmt.Println("enviando mensaje a")

			textChan <- []byte(Data)
		}
	}

}
