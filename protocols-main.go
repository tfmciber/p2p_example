package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
)

// func for removing and disconecting a peer
func (c *P2Papp) disconnectHost(stream network.Stream, err error, protocol string) {
	fmt.Println("Disconnecting host:", stream.Conn().RemotePeer(), err, protocol)
	c.Host.Network().ClosePeer(stream.Conn().RemotePeer())
	//func to get all keys from a map

}

// Function that reads data of size n from stream and calls f funcion
func (c *P2Papp) readData(stream network.Stream, size uint16, f func(buff []byte, stream network.Stream)) {

	for {
		buff := make([]byte, size)

		_, err := stream.Read(buff)

		if err != nil {

			c.disconnectHost(stream, err, string(stream.Protocol()))
			return

		}
		f(buff, stream)

	}
}

func (c *P2Papp) writeDataRend(data []byte, ProtocolID protocol.ID, rendezvous string, verbose bool) {

	for _, v := range c.Get(rendezvous) {

		if c.Host.Network().Connectedness(v) == network.Connected {
			if verbose {
				fmt.Println("Sending data to:", v)
			}

		restart:
			stream := c.streamStart(v, ProtocolID)

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
func readStdin(cmdChan chan string) {

	for {

		fmt.Print("$ ")

		var Data string
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		Data = scanner.Text()

		cmdChan <- Data

	}

}
