package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
	"github.com/multiformats/go-multiaddr"
)

//function that receives a peer addrinfo and tries to connect to it, and writes to stream if it is successful or not

func receiveCommandhandler(stream network.Stream) {

	buff := make([]byte, 2000)
	n, err := stream.Read(buff)
	if err != nil {
		fmt.Println("Error reading stream: ", err)
		return
	}

	// split message by /
	message := strings.Split(string(buff[:n]), "/")
	if len(message) < 2 {
		fmt.Println("Invalid command")
		return
	}
	switch message[0] {
	case "connect":
		fmt.Println("Received message: ", string(buff[:n]))

		peers := strings.Split(string(message[1]), ",")
		quic, err := strconv.ParseBool(peers[len(peers)-1])
		ma := multiaddr.StringCast(peers[0])
		peerinfo, err := peer.AddrInfoFromP2pAddr(ma)
		if err != nil {
			fmt.Println("Error parsing addrinfo: ", err)
			return
		}
		fmt.Println("[*] Received request to connect to: ", peerinfo, "using quic: ", quic)

	case "rendezvous":
		rendezvous := message[1]
		if !contains(Ren[rendezvous], stream.Conn().RemotePeer()) {
			log.Println("New peer:", stream.Conn().RemotePeer(), "added to rendezvous:", rendezvous)
			Ren[rendezvous] = append(Ren[rendezvous], stream.Conn().RemotePeer())

		}
		client.Reserve(context.Background(), Host, Host.Network().Peerstore().PeerInfo(stream.Conn().RemotePeer()))
	case "request":

		fmt.Println("Request")
		fmt.Println("Received message: ", string(buff[:n]))
		peers := strings.Split(string(message[1]), ",")
		quic, err := strconv.ParseBool(peers[len(peers)-1])
		if err != nil {
			fmt.Println("Error parsing quic bool: ", err)
			return
		}
		receiveReqhandler(peers, quic, stream)
	default:
		fmt.Println("Invalid command")

	}

	stream.Close()

}

func receiveReqhandler(peers []string, quic bool, stream network.Stream) {

	// read message from stream

	fmt.Println("[*] Received request for connection to: ", peers[:len(peers)-1], "using quic: ", quic)

	addrinfostr := Host.Peerstore().PeerInfo(stream.Conn().RemotePeer()).String()
	message := "connect/" + addrinfostr + "," + strconv.FormatBool(quic)

	conns := ""
	for _, p := range peers[:len(peers)-1] {
		peerID, err := peer.Decode(p)
		if err == nil {
			fmt.Println("\t[*] Trying connection to: ", p, len(p))

			if Host.Network().Connectedness(peerID) == network.Connected {
				fmt.Println("\t[*] Starting connection to: ", p)

				stream, err := Host.NewStream(context.Background(), peerID, cmdproto)
				if err != nil {
					fmt.Println("Error creating stream: ", err)
					return
				}
				//send message containing quic bool and incomming peer addrinfo
				n, err := stream.Write([]byte(message))
				fmt.Println("Wrote ", n, " bytes to stream, err: ", err)

				//wait for response, containing true (1) or false (0)

				buff := make([]byte, 1)
				n, err = stream.Read(buff)
				if err != nil {
					fmt.Println("Error reading stream: ", err)
					return
				} else {
					fmt.Println("Received message: ", string(buff[:n]))
					if (buff[0]) == 1 { //conntacted peer can see Host

						conns += p + ","
					}
				}
			}
		}
		//write to stream the conn var
		if conns == "" {
			conns = "0"
		}
		_, err = stream.Write([]byte(conns))
		if err != nil {
			fmt.Println("Error writing to stream: ", err)

		}
	}

}

// func to select a random peer that is connected to the rendezvous and send a message with
// list of the peer ID from failed var and the quic bool
func requestConnection(failed []peer.ID, rendezvous string, quic bool) []peer.ID {

	// get random peer from rendezvous that is connected

	Connectedpeers := Ren[rendezvous]

	peers := make([]peer.AddrInfo, 0)
	for _, peer := range Connectedpeers {
		if Host.Network().Connectedness(peer) == network.Connected {
			peers = append(peers, Host.Peerstore().PeerInfo(peer))
		}
	}

	if len(peers) == 0 {
		return nil
	}
	fmt.Println("[*] Starting connection request")
	rand.Seed(time.Now().UnixNano())
	randpeer := peers[rand.Intn(len(peers))]
	fmt.Println("\t[*] Selected peer: ", randpeer)
	// create stream to random peer
	stream, err := Host.NewStream(context.Background(), randpeer.ID, cmdproto)
	if err != nil {
		fmt.Println("Error creating stream: ", err)
		return nil
	}
	// create message with failed peers and quic bool
	var msg = "request/"
	for _, peerid := range failed {
		fmt.Println("\t\t[*] request for peer: ", peerid)
		msg += peerid.String() + ","
	}
	msg += strconv.FormatBool(quic)
	fmt.Println("\t\t[*] request message: ", msg)

	// write message to stream
	n, err := stream.Write([]byte(msg))
	fmt.Println("Wrote ", n, " bytes to stream, err: ", err)

	fmt.Println("\t [*] Waiting for response...")
	buff := make([]byte, 2000)
	n, err = stream.Read(buff)
	if err != nil {
		fmt.Println("Error reading stream hereee: ", err)
		return nil
	}
	fmt.Println("\t [*] Received response ", string(buff[:n]))

	if string(buff[:n]) == "0" {
		fmt.Println("Response indicates no peers can see Host")
	} else {
		//get peers from response
		peers := strings.Split(string(buff[:n]), ",")
		quic, err := strconv.ParseBool(peers[len(peers)-1])
		if err != nil {
			fmt.Println("Error parsing bool: ", err)
			return nil
		}
		fmt.Println("[*] Received response for connection to: ", peers[:len(peers)-1], "using quic: ", quic)

		//slice of string to slice of peer.ID
		peersID := make([]peer.ID, 0)
		for _, p := range peers[:len(peers)-1] {
			peerID, err := peer.Decode(p)
			if err != nil {
				fmt.Println("Error decoding peer ID: ", err)
				return nil
			}
			peersID = append(peersID, peerID)
		}
		return peersID

	}
	return nil

}
