package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/net/swarm"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
	"github.com/multiformats/go-multiaddr"
	ma "github.com/multiformats/go-multiaddr"
)

var rendezvousS = make(chan string, 1)
var deleteRendezvous = make(chan string, 1)

func listCons() {
	fmt.Println("Conns open:")
	for _, v := range Host.Network().Conns() {
		fmt.Println(v)
	}
}

// func to close all streams of a given protocol
func closeStreams(protocol string) {
	for _, v := range Host.Network().Conns() {
		//list all streams for the connection
		for _, s := range v.GetStreams() {
			if string(s.Protocol()) == protocol {
				s.Reset()
			}
		}

	}
}

// func to list all streams open
func listStreams() {
	fmt.Println("Streams open:")
	for _, v := range Host.Network().Conns() {
		//list all streams for the connection
		for _, s := range v.GetStreams() {
			fmt.Println(s.ID(), " ", s.Protocol())
		}

	}
}

// func to get peer.Addrinfo from peer.ID
func getPeerInfo(id peer.ID) peer.AddrInfo {
	return Host.Network().Peerstore().PeerInfo(id)
}

// funct to list all curennt users
func listallUSers() {

	fmt.Println("Users connected:")
	for str, peerr := range Ren {
		for _, p := range peerr {

			//check if users is connected
			online := false
			if Host.Network().Connectedness(p) == network.Connected {
				online = true
			}
			fmt.Printf("Rendezvous %s peer ID %s, online %t \n", str, p, online)
		}
	}

}
func setPeersTRansport(ctx context.Context, preferQUIC bool) bool {
	//if anyone is succesfully changed, return true
	ret := false
	//get all peers connected using rendezvous

	Peers := getPeersFromRendezvous()

	for _, v := range Peers {
		aux := setTransport(ctx, v, preferQUIC)
		if aux {
			ret = true
		}

	}
	return ret

}

// get all unique peers connected to a rendezvous that are online
func getPeersFromRendezvous() []peer.ID {
	var Peers []peer.ID
	for _, v := range Ren {
		for _, p := range v {
			if Host.Network().Connectedness(p) == network.Connected && !containsPeer(Peers, p) {
				if !containsPeer(Peers, p) {
					Peers = append(Peers, p)
				}
			}
		}
	}
	return Peers
}

func startStreams(rendezvous string, peeraddr peer.AddrInfo, stream network.Stream) {

	go receiveTexthandler(stream)
	stream2 := streamStart(hostctx, peeraddr.ID, "/audio/1.1.0")
	go receiveAudioHandler(stream2)

}
func closeConns(ID peer.ID) {
	for _, v := range Host.Network().ConnsToPeer(ID) {
		v.Close()
	}
}

func onlinePeers(rendezvous string) []peer.ID {
	var peers []peer.ID
	rendezvousPeers := Ren[rendezvous]
	for _, v := range rendezvousPeers {
		if Host.Network().Connectedness(v) == network.Connected {
			peers = append(peers, v)

		}
	}
	return peers
}

// func to get all streams with a peer of a given protcol
func getStreamsFromPeerProto(peerID peer.ID, protocol string) network.Stream {

	for _, v := range Host.Network().Conns() {
		if v.RemotePeer() == peerID {
			for _, s := range v.GetStreams() {

				if string(s.Protocol()) == protocol {
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
func connecToPeersMDNS(ctx context.Context, peerChan <-chan peer.AddrInfo, rendezvous string, preferQUIC bool, preferTCP bool) {
	for {
		select {
		case <-ctx.Done():
			return
		default:

			for peer := range peerChan {

				if peer.ID == Host.ID() {
					continue
				}

				fmt.Println("\nConnecting to:", peer.ID.String(), " ", peer.Addrs)

				addrs := selectAddrs(peer.Addrs, preferQUIC, preferTCP)
				if len(addrs) > 0 {
					peer.Addrs = addrs
				}
				err := Host.Connect(hostctx, peer)
				fmt.Println("Connected to:", peer.ID.String(), " ", peer.Addrs)
				if err != nil {
					fmt.Println("Error connecting to peer:", err)
				}
				//in peer.ID not in Ren[rendezvous] add to Ren[rendezvous]
				if !contains(Ren[rendezvous], peer.ID) {

					Ren[rendezvous] = append(Ren[rendezvous], peer.ID)
				}

				//start stream of text and audio

			}

		}
	}
}

// func to disconnect from all peers and close connections
func disconnectAll() {
	for _, v := range Host.Network().Conns() {

		for _, s := range v.GetStreams() {
			s.Close()
			s.Reset()

		}
		v.Close()

	}
	//clear Ren
	Ren = make(map[string][]peer.ID)

}

// func for getting all streams from host of a given protocol
func getStreamsFromProtocol(protocol string) []network.Stream {
	var streams []network.Stream
	for _, c := range Host.Network().Conns() {
		for _, s := range c.GetStreams() {
			if string(s.Protocol()) == protocol {
				streams = append(streams, s)
			}

		}

	}
	return streams
}

// func for checking if host has any stream open of a protocol
func hasStreamOfProtocol(protocol string) bool {
	for _, c := range Host.Network().Conns() {
		for _, s := range c.GetStreams() {

			if string(s.Protocol()) == protocol {
				return true
			}
		}
	}
	return false
}

// func for checking if host has any connection open
func hasConnection() bool {
	for _, c := range Host.Network().Conns() {
		if c != nil {
			return true
		}
	}
	return false
}

// func to start a stream with a peer only if there is no stream open and return the stream in any cases
func streamStart(ctx context.Context, peerid peer.ID, ProtocolID string) network.Stream {

	stream := getStreamsFromPeerProto(peerid, ProtocolID)

	if stream == nil {
		var err error
		stream, err = Host.NewStream(ctx, peerid, protocol.ID(ProtocolID))

		if err != nil {
			fmt.Println("stream to ", peerid, "failed", ProtocolID)
			fmt.Println("Stream failed:", err)
			return nil

		}

	}
	return stream

}

func interrupts() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-quit
		fmt.Println("\r- Exiting Program")
		disconnectAll()
		Host.Close()
		os.Exit(0)
	}()
}

// func to connect to input peers using relay server
func connectthrougRelays(peers []peer.AddrInfo, rendezvous string) {

	for _, server := range Ren[rendezvous] {
		serverpeerinfo := getPeerInfo(server)
		if serverpeerinfo.Addrs == nil {
			continue
		}
		for _, v := range peers {

			//check if peer is already connected or is self
			if containsPeer(Host.Network().Peers(), v.ID) || v.ID == Host.ID() {
				continue
			}
			relayaddr, err := ma.NewMultiaddr("/p2p/" + serverpeerinfo.ID.String() + "/p2p-circuit/p2p/" + v.ID.String())
			if err != nil {
				fmt.Println(err)

			}
			// Clear the backoff for the unreachable host
			Host.Network().(*swarm.Swarm).Backoff().Clear(v.ID)
			// Open a connection to the previously unreachable host via the relay address
			peerrelayinfo := peer.AddrInfo{ID: v.ID, Addrs: []ma.Multiaddr{relayaddr}}
			if err := Host.Connect(context.Background(), peerrelayinfo); err != nil {
				fmt.Println(err)

			}
		}
	}

}

// func to reserve circuit with relay server and return all successful connections
func connectRelay(rendezvous string) {

	fmt.Println("[*] Reserving circuit with connected hosts...")

	for _, v := range Ren[rendezvous] {

		// check if peer is  connected
		if containsPeer(Host.Network().Peers(), v) {

			_, err := client.Reserve(context.Background(), Host, getPeerInfo(v))
			if err == nil {
				fmt.Println("\t[*] Reserved circuit with:", v.String())

			}
		}

	}
	fmt.Println("[*] Reservation finished.")

}

func containsPeer(peers []peer.ID, peer peer.ID) bool {
	for _, v := range peers {
		if v == peer {
			return true
		}
	}
	return false
}

// func to check if there are any peers online at a given rendezvous
func hasPeer(rendezvous string) bool {
	for _, v := range Ren[rendezvous] {
		if containsPeer(Host.Network().Peers(), v) {
			return true
		}
	}
	return false
}

// go func for when a channel, "aux" is written create a new nuction that runs every 5 minutes appends the value written to the channel to a list and then runs the function
// for all the values in the list that wherent run in the last 5 minutes
func dhtRoutine(ctx context.Context, rendezvousS chan string, kademliaDHT *dht.IpfsDHT, quic bool) {
	var allRedenzvous = map[string]int{}
	for {
		select {
		case <-time.After(1 * time.Minute):

			for rendezvous, time := range allRedenzvous {
				if time == 0 {
					rendezvousS <- rendezvous
				} else {
					allRedenzvous[rendezvous]--
					fmt.Println("Rendezvous", rendezvous, "restarting in ", allRedenzvous[rendezvous])
				}
			}

		case aux := <-rendezvousS:

			FoundPeersDHT := discoverPeers(ctx, kademliaDHT, Host, aux)
			failed := connecToPeers(ctx, FoundPeersDHT, aux, quic, true)
			connectRelay(aux)
			connectthrougRelays(failed, aux)
			allRedenzvous[aux] = 5

		case aux := <-deleteRendezvous:
			delete(allRedenzvous, aux)
		}
	}
}
