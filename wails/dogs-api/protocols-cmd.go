package main

import (
	"context"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
)

//function that receives a peer addrinfo and tries to connect to it, and writes to stream if it is successful or not

func (c *P2Papp) receiveCommandhandler(stream network.Stream) {

	buff := make([]byte, 2000)
	n, err := stream.Read(buff)
	c.fmtPrintln("receiveCommandhandler: ", string(buff[:n]))
	var message []string
	if err != nil {
		if err != io.EOF {
			c.fmtPrintln("Error reading stream: ", err)
			goto end
		}

	}

	// split message by /
	message = strings.Split(string(buff[:n]), "$")
	if len(message) < 2 {
		c.fmtPrintln("Invalid command")
		goto end
	}

	switch message[0] {
	case "connect":
		if len(message) < 4 {

			c.fmtPrintln("Invalid connect command", message)
			goto end
		}
		peersstr := message[1]
		rendezvous := message[2]
		quicstr := message[3]
		quic, err := strconv.ParseBool(quicstr)
		if err != nil {
			c.fmtPrintln("Error parsing quic bool: ", err)
			goto end
		}
		c.fmtPrintln("quic: ", quic)

		// convert json string to peer.AddrInfo

		var addrinf peer.AddrInfo
		err = addrinf.UnmarshalJSON([]byte(peersstr))
		if err != nil {
			c.fmtPrintln("Error unmarshalling json: ", err)
			goto end
		} else {

			c.fmtPrintln("[*] Connecting to peer: ", addrinf.ID)

			ok := c.connectToPeer(c.ctx, addrinf, rendezvous, quic, true)
			if ok {
				c.fmtPrintln("\t [*] Connected to peer: ", addrinf.ID)
			} else {
				c.fmtPrintln("\t [*] Failed to connect to peer: ", addrinf.ID)
			}
		}

	case "rendezvous":
		if message[1] != "" {
			c.fmtPrintln("New user added: ", stream.Conn().RemotePeer())
			c.Add(message[1], stream.Conn().RemotePeer())
			client.Reserve(context.Background(), c.Host, c.Host.Network().Peerstore().PeerInfo(stream.Conn().RemotePeer()))
		}

	case "request":
		if len(message) < 4 {
			c.fmtPrintln("Invalid connect command")
			goto end
		}
		peers := strings.Split(string(message[1]), ",")
		rendezvous := message[2]
		quic, err := strconv.ParseBool(message[3])
		if err != nil {
			c.fmtPrintln("Error parsing quic bool: ", err)
			goto end
		}
		c.receiveReqhandler(peers, rendezvous, quic, stream)
	case "leave":
		c.fmtPrintln("User ", stream.Conn().RemotePeer(), " left chat", message[1])
		rendezvous := message[1]
		peerid, err := peer.Decode(message[1])
		c.fmtPrintln("Peerid: ", peerid, "err: ", err)
		isrend := c.checkRend(rendezvous)
		user := stream.Conn().RemotePeer()
		c.fmtPrintln("Isrend: ", isrend)

		if err == nil {
			c.fmtPrintln("removing DM")
			if peerid == c.Host.ID() {
				c.fmtPrintln("Removing DM deleted ")
				c.deleteDm(user)
				c.EmitEvent("dmLeft")
			} else {
				c.fmtPrintln("User is not the owner of the DM")
				goto end
			}
		} else if isrend {
			c.fmtPrintln("Removing"+user, "from ", rendezvous)
			aux := deleteValue(c.data[rendezvous].peers, user)
			c.SetPeers(rendezvous, aux)
			time := time.Now().Format("2006-01-02 15:04:05")
			c.EmitEvent("userLeft", rendezvous, time, user)
			c.useradded <- true

		} else {
			c.fmtPrintln("Error removing user")
			goto end
		}

	default:
		c.fmtPrintln("Invalid command")

	}
end:
	stream.Close()

}

func (c *P2Papp) receiveReqhandler(peers []string, rendezvous string, quic bool, stream network.Stream) {

	// read message from stream

	c.fmtPrintln("[*] Received request for connection to: ", peers, "using quic: ", quic)
	addrinfostr, _ := c.Host.Peerstore().PeerInfo(stream.Conn().RemotePeer()).MarshalJSON()
	message := "connect$" + string(addrinfostr) + "$" + rendezvous + "$" + strconv.FormatBool(quic)

	conns := ""

	for _, p := range peers[:len(peers)-1] {
		peerID, err := peer.Decode(p)
		if err != nil {
			c.fmtPrintln("Error decoding", c.GetPeerIDfromstring(p), p, peerID, "ID: ", err)

		} else {
			c.fmtPrintln("[*] Checking connection to: ", peerID)
			if c.Host.Network().Connectedness(peerID) == network.Connected {
				c.fmtPrintln("\t[*] Starting connection to: ", p)

				stream, err := c.Host.NewStream(context.Background(), peerID, c.cmdproto)
				if err != nil {
					c.fmtPrintln("Error creating stream: ", err)

				}
				c.fmtPrintln("\t\t [*] Sending address")
				_, err = stream.Write([]byte(message))
				if err != nil {
					c.fmtPrintln("Error writing to stream: ", err)

				}
				c.fmtPrintln("\r \t\t [*] Sent address")
				conns += peerID.String() + ","

			}
		}
	}

	//write to stream the conn var

	if conns == "" {
		conns = "0"
	}
	c.fmtPrintln("[*] Sending response", conns)
	_, err := stream.Write([]byte(conns))
	if err != nil {
		c.fmtPrintln("Error writing to stream: ", err)

	}
}

// func to select a random peer that is connected to the rendezvous and send a message with
// list of the peer ID from failed var and the quic bool
func (c *P2Papp) requestConnection(failed []peer.ID, rendezvous string, quic bool, ctx context.Context, ctx2 context.Context) []peer.ID {

	c.fmtPrintln("[*] Requesting connection to: ", failed, "using quic: ", quic)
	if len(failed) == 0 {
		return nil
	}

	peersID := make(chan []peer.ID, 0)

	// get random peer from rendezvous that is connected
	go func() {
		c.fmtPrintln("[*] Starting connection request")
		Connectedpeers, _ := c.Get(rendezvous)

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

		if index >= len(peers) {
			peersID <- nil
			return
		}

		selpeer := peers[index]
		index++

		c.fmtPrintln("\t[*] Selected peer: ", selpeer)
		// create stream to random peer
		stream, err := c.Host.NewStream(context.Background(), selpeer.ID, c.cmdproto)
		defer stream.Close()
		if err != nil {
			c.fmtPrintln("Error creating stream: ", err)
			peersID <- nil
			return
		}
		// create message with failed peers and quic bool
		var msg = "request$"
		for _, peerid := range failed {
			c.fmtPrintln("\t\t[*] request for peer: ", peerid.String())
			msg += peerid.String() + ","
		}
		msg += "$" + rendezvous + "$" + strconv.FormatBool(quic)

		c.fmtPrintln("\t\t[*] request message: ", msg)

		// write message to stream
		n, err := stream.Write([]byte(msg))
		c.fmtPrintln("Wrote ", n, " bytes to stream, err: ", err)

		c.fmtPrintln("\t [*] Waiting for response...")
		buff := make([]byte, 2000)
		n, err = stream.Read(buff)
		if err != nil {
			c.fmtPrintln("Error reading stream hereee: ", err, n, string(buff[:n]))
			if n == 0 {
				goto selectpeer
			}
		}
		c.fmtPrintln("\t [*] Received response ", string(buff[:n]))

		if string(buff[:n]) == "0" {
			c.fmtPrintln("Response indicates no peers can see Host")
			goto selectpeer
		} else {
			//get peers from response
			peers := strings.Split(string(buff[:n]), ",")
			c.fmtPrintln("[*] Received online peers that we cant reach")

			//slice of string to slice of peer.ID
			peersIDaux := make([]peer.ID, 0)
			for _, p := range peers {
				peerID, err := peer.Decode(p)
				if err != nil {
					c.fmtPrintln("Error decoding peer ID: ", err, p)
					continue
				} else {
					peersIDaux = append(peersIDaux, peerID)
				}

			}
			peersID <- peersIDaux

			return
		}

	}()
	select {
	case <-ctx.Done():
		c.fmtPrintln("Context done")
		return nil
	case <-ctx2.Done():
		c.fmtPrintln("Context done")
		return nil
	case peersID := <-peersID:
		c.fmtPrintln("Request done")
		return peersID
	}

}
