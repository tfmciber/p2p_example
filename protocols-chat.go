package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/libp2p/go-libp2p/core/network"
)

var textChan = make(chan string)

func chat_func(stream network.Stream) {

	go ReadStdin()
	aux := make([]byte, 2000)
	go readData(stream, aux, func(buff []byte, stream network.Stream) {

		fmt.Printf("\x1b[32m%s\x1b[0m> ", string(buff[:]))
	})
	go writeToStreams()

}

func ReadStdin() {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		Data, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin")
			panic(err)
		}

		textChan <- Data

	}
}

func writeToStreams() {

	for {
		data := <-textChan
		for _, stream := range streams {
			w := bufio.NewWriter(stream)
			_, err := w.WriteString(fmt.Sprintf("%s\n", data))
			if err != nil {
				fmt.Println("Error writing to buffer")
				panic(err)
			}

			err = w.Flush()
			if err != nil {
				fmt.Println("Error flushing buffer")
				panic(err)
			}

		}

	}
}

func writeData(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin")
			panic(err)
		}

		_, err = rw.WriteString(fmt.Sprintf("%s\n", sendData))
		if err != nil {
			fmt.Println("Error writing to buffer")
			panic(err)
		}
		err = rw.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer")
			panic(err)
		}
	}
}
