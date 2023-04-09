package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
)

//function that receives a peer addrinfo and tries to connect to it, and writes to stream if it is successful or not

func (c *P2Papp) receiveCommandhandler(stream network.Stream) {

	buff := make([]byte, 2000)
	n, err := stream.Read(buff)
	if err != nil {
		fmt.Println("Error reading stream: ", err)
		return
	}

	// split message by /
	message := strings.Split(string(buff[:n]), "$")
	if len(message) < 2 {
		fmt.Println("Invalid command")
		return
	}

	switch message[0] {
	case "connect":
		if len(message) < 4 {
			fmt.Println("Invalid connect command")
			return
		}
		peersstr := message[1]
		rendezvous := message[2]
		quicstr := message[3]
		quic, err := strconv.ParseBool(quicstr)
		if err != nil {
			fmt.Println("Error parsing quic bool: ", err)
			return
		}
		fmt.Println("quic: ", quic)

		// convert json string to peer.AddrInfo

		var addrinf peer.AddrInfo
		err = addrinf.UnmarshalJSON([]byte(peersstr))
		if err != nil {
			fmt.Println("Error unmarshalling json: ", err)
			return
		} else {

			fmt.Println("[*] Connecting to peer: ", addrinf.ID)

			ok := c.connectToPeer(c.ctx, addrinf, rendezvous, quic, true)
			if ok {
				fmt.Println("\t [*] Connected to peer: ", addrinf.ID)
			} else {
				fmt.Println("\t [*] Failed to connect to peer: ", addrinf.ID)
			}
		}

	case "rendezvous":
		fmt.Println("New user added: ", stream.Conn().RemotePeer())
		c.Add(message[1], stream.Conn().RemotePeer())
		client.Reserve(context.Background(), c.Host, c.Host.Network().Peerstore().PeerInfo(stream.Conn().RemotePeer()))

	case "request":
		if len(message) < 4 {
			fmt.Println("Invalid connect command")
			return
		}
		peers := strings.Split(string(message[1]), ",")
		rendezvous := message[2]
		quic, err := strconv.ParseBool(message[3])
		if err != nil {
			fmt.Println("Error parsing quic bool: ", err)
			return
		}
		c.receiveReqhandler(peers, rendezvous, quic, stream)
	default:
		fmt.Println("Invalid command")

	}

	stream.Close()

}

func (c *P2Papp) receiveReqhandler(peers []string, rendezvous string, quic bool, stream network.Stream) {

	// read message from stream

	fmt.Println("[*] Received request for connection to: ", peers[:len(peers)-1], "using quic: ", quic)
	addrinfostr, _ := c.Host.Peerstore().PeerInfo(stream.Conn().RemotePeer()).MarshalJSON()
	message := "connect$" + string(addrinfostr) + "$" + rendezvous + "$" + strconv.FormatBool(quic)

	conns := ""
	for _, p := range peers[:len(peers)-1] {
		peerID, err := peer.Decode(p)
		if err != nil {
			fmt.Println("Error decoding peer ID: ", err)
			goto sendresponse
		}

		if c.Host.Network().Connectedness(peerID) == network.Connected {
			fmt.Println("\t[*] Starting connection to: ", p)

			stream, err := c.Host.NewStream(context.Background(), peerID, c.cmdproto)
			if err != nil {
				fmt.Println("Error creating stream: ", err)
				goto sendresponse
			}
			fmt.Println("\t\t [*] Sending address")
			_, err = stream.Write([]byte(message))
			if err != nil {
				fmt.Println("Error writing to stream: ", err)
				goto sendresponse
			}
			fmt.Print("\r \t\t [*] Sent address")
			conns += peerID.String() + ","

		}
	}

	//write to stream the conn var
sendresponse:
	if conns == "" {
		conns = "0"
	}
	fmt.Println("[*] Sending response", conns)
	_, err := stream.Write([]byte(conns))
	if err != nil {
		fmt.Println("Error writing to stream: ", err)

	}
}

// func to select a random peer that is connected to the rendezvous and send a message with
// list of the peer ID from failed var and the quic bool
func (c *P2Papp) requestConnection(failed []peer.ID, rendezvous string, quic bool, ctx context.Context, ctx2 context.Context) []peer.ID {

	if len(failed) == 0 {
		return nil
	}

	peersID := make(chan []peer.ID, 0)

	// get random peer from rendezvous that is connected
	go func() {
		fmt.Println("[*] Starting connection request")
		Connectedpeers := c.Get(rendezvous)

		peers := make([]peer.AddrInfo, 0)
		for _, peer := range Connectedpeers {
			if c.Host.Network().Connectedness(peer) == network.Connected {
				peers = append(peers, c.Host.Peerstore().PeerInfo(peer))
			}
		}

		if len(peers) == 0 {

			peersID <- nil
			return
		}

		index := 0
	selectpeer:

		if index >= len(Connectedpeers) {
			peersID <- nil
			return
		}

		selpeer := peers[index]
		index++

		fmt.Println("\t[*] Selected peer: ", selpeer)
		// create stream to random peer
		stream, err := c.Host.NewStream(context.Background(), selpeer.ID, c.cmdproto)
		if err != nil {
			fmt.Println("Error creating stream: ", err)
			peersID = nil
			return
		}
		// create message with failed peers and quic bool
		var msg = "request$"
		for _, peerid := range failed {
			fmt.Println("\t\t[*] request for peer: ", peerid.String())
			msg += peerid.String() + ","
		}
		msg += "$" + rendezvous + "$" + strconv.FormatBool(quic)

		fmt.Println("\t\t[*] request message: ", msg)

		// write message to stream
		n, err := stream.Write([]byte(msg))
		fmt.Println("Wrote ", n, " bytes to stream, err: ", err)

		fmt.Println("\t [*] Waiting for response...")
		buff := make([]byte, 2000)
		n, err = stream.Read(buff)
		if err != nil {
			fmt.Println("Error reading stream hereee: ", err)
			peersID = nil
			return
		}
		fmt.Println("\t [*] Received response ", string(buff[:n]))

		if string(buff[:n]) == "0" {
			fmt.Println("Response indicates no peers can see Host")
			goto selectpeer
		} else {
			//get peers from response
			peers := strings.Split(string(buff[:n]), ",")
			fmt.Println("[*] Received online peers that we cant reach")

			//slice of string to slice of peer.ID
			peersIDaux := make([]peer.ID, 0)
			for _, p := range peers[:len(peers)-1] {
				peerID, err := peer.Decode(p)
				if err != nil {
					fmt.Println("Error decoding peer ID: ", err)
					peersID = nil
					return
				}
				peersIDaux = append(peersIDaux, peerID)
			}
			peersID <- peersIDaux

			return
		}

	}()
	select {
	case <-ctx.Done():
		fmt.Println("Context done")
		return nil
	case <-ctx2.Done():
		fmt.Println("Context done")
		return nil
	case peersID := <-peersID:
		fmt.Println("Request done")
		return peersID
	}

}
