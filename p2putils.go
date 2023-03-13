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
			fmt.Printf("Rendezvous %s peer ID %s, online %t \n", str, p.String(), online)
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

func startStreams(rendezvous string, peerid peer.ID) {

	stream1 := streamStart(context.Background(), peerid, string(textproto))
	go receiveTexthandler(stream1)
	stream2 := streamStart(context.Background(), peerid, string(audioproto))
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
				err := Host.Connect(context.Background(), peer)
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
func clear() {
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

// func to start a stream with a peer only if there is no stream open and return the stream in any cases
func streamStart(ctx context.Context, peerid peer.ID, ProtocolID string) network.Stream {

	stream := getStreamsFromPeerProto(peerid, ProtocolID)

	if stream == nil {
		var err error
		stream, err = Host.NewStream(ctx, peerid, protocol.ID(ProtocolID))
		if err != nil {
			fmt.Println(err.Error())
			if err.Error() == "transient connection to peer" {
				stream, err = Host.NewStream(network.WithUseTransient(context.Background(), "relay"), peerid, protocol.ID(ProtocolID))

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

func interrupts() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-quit
		fmt.Println("\r- Exiting Program")
		clear()
		Host.Close()
		os.Exit(0)
	}()
}

// func to connect to input peers using relay server
func connectthrougRelays(peersid []peer.ID, rendezvous string, preferQUIC bool) {
	fmt.Println("[*] Connecting to peers through Relays")
	for _, server := range Ren[rendezvous] {
		serverpeerinfo := Host.Network().Peerstore().PeerInfo(server)
		if serverpeerinfo.Addrs == nil {
			continue
		}
		fmt.Println("\t[*] Connecting using relay server:", serverpeerinfo.ID.String(), " ", serverpeerinfo.Addrs)

		for _, v := range peersid {

			//check if peer is already connected
			if Host.Network().Connectedness(v) == network.Connected {
				continue
			}
			relayaddr, err := ma.NewMultiaddr("/p2p/" + serverpeerinfo.ID.String() + "/p2p-circuit/p2p/" + v.String())
			if err != nil {
				fmt.Println(err)
				continue

			}
			fmt.Println("\t\t[*] Connecting to:", v.String(), " ", relayaddr)

			// Clear the backoff for the unreachable host
			Host.Network().(*swarm.Swarm).Backoff().Clear(v)
			// Open a connection to the previously unreachable host via the relay address
			peerrelayinfo := peer.AddrInfo{ID: v, Addrs: []ma.Multiaddr{relayaddr}}
			if connectToPeer(network.WithUseTransient(context.Background(), "relay"), peerrelayinfo, rendezvous, preferQUIC, true) {

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
func connectRelay(rendezvous string) {

	fmt.Println("[*] Reserving circuit with connected hosts...")

	for _, v := range Ren[rendezvous] {

		// check if peer is  connected
		if containsPeer(Host.Network().Peers(), v) {

			_, err := client.Reserve(context.Background(), Host, Host.Network().Peerstore().PeerInfo(v))
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
func dhtRoutine(ctx context.Context, rendezvousS chan string, kademliaDHT *dht.IpfsDHT, quic bool, refresh uint, typem bool) {
	var allRedenzvous = map[string]uint{}
	for {
		select {
		case <-time.After(1 * time.Minute):

			for rendezvous, time := range allRedenzvous {
				if time == 0 {
					rendezvousS <- rendezvous
				} else {
					allRedenzvous[rendezvous]--
				}
			}

		case aux := <-rendezvousS:

			fmt.Println("[*] Searching for peers at rendezvous:", aux, "...")
			FoundPeersDHT := discoverPeers(ctx, kademliaDHT, Host, aux)
			Received := receivePeersDHT(ctx, FoundPeersDHT, aux)
			failed := connectToPeers(ctx, Received, aux, quic, true)
			var failed2 []peer.ID

			if typem {

				peerid, _ := peer.Decode("QmXtW5fXrrvmHWPhq3FLHdm4zKnC5FZdhTRynSQT57Yrmd")
				fmt.Println("peerid: ", peerid)
				listallUSers()
				fmt.Println("peerid: ", peerid)
				time.Sleep(5 * time.Second)
				disconnectPeer("QmXtW5fXrrvmHWPhq3FLHdm4zKnC5FZdhTRynSQT57Yrmd")
				failed2 = append(failed2, peer.ID(peerid))
				listallUSers()

				//remove peerid from failed

				fmt.Println("failed2: ", failed2)
				failed = requestConnection(failed2, aux, quic)
			}

			fmt.Println("finished request connection")

			if len(failed) > 0 {
				connectRelay(aux)
				connectthrougRelays(failed, aux, quic)
				allRedenzvous[aux] = refresh

			}

		case aux := <-deleteRendezvous:
			fmt.Println("[*] Deleting rendezvous:", aux, "...")
			delete(allRedenzvous, aux)
		}
	}
}
