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

	for {

		fmt.Print("$ ")
		fmt.Print(len(SendStreams))
		var Data string
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		Data = scanner.Text()

		if strings.HasPrefix(Data, "/") {
			Data = strings.TrimPrefix(Data, "/")

			cmdChan <- Data

		} else if len(SendStreams) > 0 {

			textChan <- []byte(Data)
		} else {

			fmt.Println("No command selected and not connected to any host")
		}
	}

}
