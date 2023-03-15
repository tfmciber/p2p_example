package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/net/swarm"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
	"github.com/multiformats/go-multiaddr"
	ma "github.com/multiformats/go-multiaddr"
)

func (c *P2Papp) listCons() {
	fmt.Println("Conns open:")
	for _, v := range c.Host.Network().Conns() {
		fmt.Println(v)
	}
}

// func to close all streams of a given protocol
func (c *P2Papp) closeStreams(protocol string) {
	for _, v := range c.Host.Network().Conns() {
		//list all streams for the connection
		for _, s := range v.GetStreams() {
			if string(s.Protocol()) == protocol {
				s.Reset()
			}
		}

	}
}

// func to list all streams open
func (c *P2Papp) listStreams() {
	fmt.Println("Streams open:")
	for _, v := range c.Host.Network().Conns() {
		//list all streams for the connection
		for _, s := range v.GetStreams() {
			fmt.Println(s.ID(), " ", s.Protocol())
		}

	}
}

// funct to list all curennt users
func (c *P2Papp) listallUSers() {

	fmt.Println("Users connected:")
	for str, peerr := range c.data {
		for _, p := range peerr.peers {

			//check if users is connected
			online := false
			if c.Host.Network().Connectedness(p) == network.Connected {
				online = true
			}
			fmt.Printf("Rendezvous %s peer ID %s, online %t \n", str, p.String(), online)
		}
	}

}
func (c *P2Papp) setPeersTRansport(ctx context.Context, rendezvous string, preferQUIC bool) bool {
	//if anyone is succesfully changed, return true
	ret := false
	//get all peers connected using rendezvous

	Peers := c.Get(rendezvous)

	for _, v := range Peers {
		aux := c.setTransport(v, preferQUIC)
		if aux {
			ret = true
		}

	}
	return ret

}

func (c *P2Papp) startStreams(rendezvous string, peerid peer.ID) {

	stream1 := c.streamStart(peerid, c.textproto)
	go c.receiveTexthandler(stream1)
	stream2 := c.streamStart(peerid, c.audioproto)
	go c.receiveAudioHandler(stream2)

}
func (c *P2Papp) closeConns(ID peer.ID) {
	for _, v := range c.Host.Network().ConnsToPeer(ID) {
		v.Close()
	}
}

func (c *P2Papp) onlinePeers(rendezvous string) []peer.ID {
	var peers []peer.ID
	rendezvousPeers := c.Get(rendezvous)
	for _, v := range rendezvousPeers {
		if c.Host.Network().Connectedness(v) == network.Connected {
			peers = append(peers, v)

		}
	}
	return peers
}

// func to get a stream with a peer of a given protcol
func (c *P2Papp) getStreamsFromPeerProto(peerID peer.ID, protocol protocol.ID) network.Stream {

	for _, v := range c.Host.Network().Conns() {
		if v.RemotePeer() == peerID {
			for _, s := range v.GetStreams() {

				if s.Protocol() == protocol {
					return s
				}
			}
		}
	}

	return nil
}
func selectAddrs(peeraddrs []multiaddr.Multiaddr, preferQUIC bool, preferTCP bool) []multiaddr.Multiaddr {
	var addrs []multiaddr.Multiaddr
	if preferQUIC {

		for _, v := range peeraddrs {
			if strings.Contains(v.String(), "quic") {
				addrs = append(addrs, v)
			}
		}

	}

	//if not prefer quic connect to tcp addresses and if there are no quic addreses
	if preferTCP {

		for _, v := range peeraddrs {
			if strings.Contains(v.String(), "tcp") {
				addrs = append(addrs, v)
			}
		}

	}
	return addrs

}
func (c *P2Papp) connecToPeersMDNS(peerChan <-chan peer.AddrInfo, rendezvous string, preferQUIC bool, preferTCP bool) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return
		default:

			for peer := range peerChan {

				if peer.ID == c.Host.ID() {
					continue
				}

				fmt.Println("\nConnecting to:", peer.ID.String(), " ", peer.Addrs)

				addrs := selectAddrs(peer.Addrs, preferQUIC, preferTCP)
				if len(addrs) > 0 {
					peer.Addrs = addrs
				}
				err := c.Host.Connect(c.ctx, peer)
				fmt.Println("Connected to:", peer.ID.String(), " ", peer.Addrs)
				if err != nil {
					fmt.Println("Error connecting to peer:", err)
				}

				c.Add(rendezvous, peer.ID)

			}

		}
	}
}

// func to disconnect from all peers and close connections
func (c *P2Papp) clear() {
	for _, v := range c.Host.Network().Conns() {

		for _, s := range v.GetStreams() {
			s.Close()
			s.Reset()

		}
		v.Close()

	}

	c.Clear()
}

// func to start a stream with a peer only if there is no stream open and return the stream in any cases
func (c *P2Papp) streamStart(peerid peer.ID, ProtocolID protocol.ID) network.Stream {

	stream := c.getStreamsFromPeerProto(peerid, ProtocolID)

	if stream == nil {
		var err error
		stream, err = c.Host.NewStream(c.ctx, peerid, ProtocolID)
		if err != nil {
			fmt.Println(err.Error())
			if err.Error() == "transient connection to peer" {
				stream, err = c.Host.NewStream(network.WithUseTransient(context.Background(), "relay"), peerid, ProtocolID)

			}
		}

		if err != nil {
			fmt.Println("stream to ", peerid, "failed", ProtocolID)
			fmt.Println("Stream failed:", err)

			return nil

		}

	}
	return stream

}

func (c *P2Papp) interrupts() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-quit
		fmt.Println("\r- Exiting Program")
		c.clear()
		c.Host.Close()
		os.Exit(0)
	}()
}

// func to connect to input peers using relay server
func (c *P2Papp) connectthrougRelays(peersid []peer.ID, rendezvous string, preferQUIC bool) {
	fmt.Println("[*] Connecting to peers through Relays")
	for _, server := range c.Get(rendezvous) {
		serverpeerinfo := c.Host.Network().Peerstore().PeerInfo(server)
		if serverpeerinfo.Addrs == nil {
			continue
		}
		fmt.Println("\t[*] Connecting using relay server:", serverpeerinfo.ID.String(), " ", serverpeerinfo.Addrs)

		for _, v := range peersid {

			//check if peer is already connected
			if c.Host.Network().Connectedness(v) == network.Connected {
				continue
			}
			relayaddr, err := ma.NewMultiaddr("/p2p/" + serverpeerinfo.ID.String() + "/p2p-circuit/p2p/" + v.String())
			if err != nil {
				fmt.Println(err)
				continue

			}
			fmt.Println("\t\t[*] Connecting to:", v.String(), " ", relayaddr)

			// Clear the backoff for the unreachable c.Host
			c.Host.Network().(*swarm.Swarm).Backoff().Clear(v)
			// Open a connection to the previously unreachable c.Host via the relay address
			peerrelayinfo := peer.AddrInfo{ID: v, Addrs: []ma.Multiaddr{relayaddr}}
			if c.connectToPeer(network.WithUseTransient(context.Background(), "relay"), peerrelayinfo, rendezvous, preferQUIC, true) {

				//delete peer from peers
				for i, p := range peersid {
					if p == v {
						peersid = append(peersid[:i], peersid[i+1:]...)
					}
				}
			}
		}

	}
	fmt.Println("[*] Connecting using Relays finished.")
}

// func to reserve circuit with relay server and return all successful connections
func (c *P2Papp) connectRelay(rendezvous string) {

	fmt.Println("[*] Reserving circuit with connected c.Hosts...")

	for _, v := range c.Get(rendezvous) {

		// check if peer is  connected
		if c.Host.Network().Connectedness(v) == network.Connected {

			_, err := client.Reserve(c.ctx, c.Host, c.Host.Network().Peerstore().PeerInfo(v))
			if err == nil {
				fmt.Println("\t[*] Reserved circuit with:", v.String())

			}
		}

	}
	fmt.Println("[*] Reservation finished.")

}

// go func for when a channel, "aux" is written create a new nuction that runs every 5 minutes appends the value written to the channel to a list and then runs the function
// for all the values in the list that wherent run in the last 5 minutes
func (c *P2Papp) dhtRoutine(quic bool, refresh uint, typem bool) {

	for {
		select {
		case <-time.After(60 * time.Second):

			for rendezvous, s := range c.data {

				if s.timer == 0 {
					c.rendezvousS <- rendezvous
				} else {
					c.SetTimer(rendezvous, s.timer-1)

				}
			}

		case aux := <-c.rendezvousS:

			fmt.Println("[*] Searching for peers at rendezvous:", aux, "...")
			FoundPeersDHT := c.discoverPeers(aux)
			Received := c.receivePeersDHT(FoundPeersDHT, aux)
			failed := c.connectToPeers(Received, aux, quic, true)
			failed = c.requestConnection(failed, aux, quic)
			time.Sleep(5 * time.Second)
			c.connectRelay(aux)
			if len(failed) > 0 {
				c.connectthrougRelays(failed, aux, quic)

			}
			c.SetTimer(aux, refresh)

		}
	}
}
