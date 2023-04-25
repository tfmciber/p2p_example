package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

// func for removing and disconecting a peer
func (c *P2Papp) disconnectHost(stream network.Stream, err error, protocol string) {

	c.fmtPrintln("Host disconnected:", stream.Conn().RemotePeer(), err, protocol)

	//get addrinfo of peer
	peerinfo := c.Host.Peerstore().PeerInfo(stream.Conn().RemotePeer())
	c.fmtPrintln("Trying to reconnect")
	//try to reconnect to peer
	err = c.Host.Connect(c.ctx, peerinfo)
	if err != nil {
		c.fmtPrintln("Error reconnecting to peer:", err)
		c.fmtPrintln("Disconnecting host:", stream.Conn().RemotePeer(), err, protocol)
		c.Host.Network().ClosePeer(stream.Conn().RemotePeer())
		c.useradded <- true
	}

}

//funt to send data to all peers in a rendezvous in a effient way from a channel
func (c *P2Papp) writeDataRendFunc(ProtocolID protocol.ID, rendezvous string, f func(stream network.Stream)) {

	rend, _ := c.Get(rendezvous)

	var wg sync.WaitGroup
	for _, v := range rend {

		if c.Host.Network().Connectedness(v) == network.Connected {

			stream := c.streamStart(v, ProtocolID)

			if stream == nil {
				c.fmtPrintln("stream is nil")

			} else {

				//run go func f and wait until all it intances are done waitgroup
				wg.Add(1)
				go func(peerID peer.ID) {
					defer stream.Close()

					defer wg.Done()
					f(stream)
				}(v)

			}
		}

	}
	wg.Wait()

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
