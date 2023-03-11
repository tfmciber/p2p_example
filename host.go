package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"strings"

	"github.com/gen2brain/malgo"
	"github.com/libp2p/go-libp2p"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"

	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/libp2p/go-libp2p/p2p/muxer/mplex"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"

	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	quic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	tcp "github.com/libp2p/go-libp2p/p2p/transport/tcp"
)

// Host is the libp2p host
var Host host.Host

// Ren variable to store the rendezvous string and the peers associated to it
var Ren = make(map[string][]peer.ID) //map of peers associated to a rendezvous string

var textproto = protocol.ID("/text/1.0.0")
var audioproto = protocol.ID("/audio/1.0.0")
var benchproto = protocol.ID("/bench/1.0.0")
var cmdproto = protocol.ID("/cmd/1.0.0")
var fileproto = protocol.ID("/file/1.0.0")

// function to create a host with a private key and a resource manager to limit the number of connections and streams per peer and per protocol
func newHost(ctx context.Context, priv crypto.PrivKey) (host.Host, network.ResourceManager) {

	limiter := rcmgr.InfiniteLimits

	rcm, err := rcmgr.NewResourceManager(rcmgr.NewFixedLimiter(limiter))
	if err != nil {
		log.Fatal(err)
	}
	var DefaultTransports = libp2p.ChainOptions(

		libp2p.Transport(tcp.NewTCPTransport),

		libp2p.Transport(quic.NewTransport),
	)

	fmt.Println("[*] Creating Host")
	h, err := libp2p.New(
		// Use the keypair we generated
		libp2p.Identity(priv),
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/udp/0/quic", "/ip4/0.0.0.0/tcp/0"),

		// support TLS connections
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		// support noise connections
		libp2p.Security(noise.ID, noise.New),
		// support any other default transports (TCP,quic)
		DefaultTransports,
		libp2p.DefaultConnectionManager,
		libp2p.ResourceManager(rcm),
		libp2p.UserAgent("P2P_Example"),
		libp2p.NATPortMap(),
		libp2p.EnableRelay(),
		libp2p.EnableHolePunching(),
		libp2p.EnableNATService(),

		libp2p.ChainOptions(
			libp2p.Muxer("/yamux/1.0.0", yamux.DefaultTransport),
			libp2p.Muxer("/mplex/6.7.0", mplex.DefaultTransport),
		),
	)
	if err != nil {
		panic(err)
	}
	var protocols = []protocol.ID{audioproto, textproto, fileproto, benchproto, cmdproto}
	for _, ProtocolID := range protocols {

		h.SetStreamHandler(ProtocolID, handleStream)
	}
	fmt.Println("\t[*] Host listening on: ", h.Addrs())
	fmt.Println("\t[*] Starting Relay system")

	_, err1 := relay.New(h)
	if err1 != nil {
		log.Printf("Failed to instantiate the relay: %v", err)

	}

	fmt.Println("[*] Host Created")
	return h, rcm
}

// func to connect to peers found in rendezvous
func connecToPeers(ctx context.Context, peerChan <-chan peer.AddrInfo, rendezvous string, preferQUIC bool, start bool) []peer.ID {
	fmt.Println("[*] Connecting to peers")
	var peersFound []peer.AddrInfo
	var failed []peer.ID
	fmt.Println("\t[*] Receiving peers")
	for peer := range peerChan {

		if peer.ID == Host.ID() {
			continue
		}

		//check if peer is already connected
		if Host.Network().Connectedness(peer.ID) == network.Connected {
			continue
		}
		peersFound = append(peersFound, peer)
		fmt.Println("\t\t[*] New peer Found:", peer.ID)

	}
	fmt.Println("\t[*] Receiving peers finished")
	fmt.Println("\t[*] Connecting to peers")

	for _, peeraddr := range peersFound {

		fmt.Println("\t\t[*] Connecting to: ", peeraddr.ID)

		err := Host.Connect(ctx, peeraddr)

		if err == nil {

			fmt.Println("\t\t\t Successfully connected to ", peeraddr.ID, peeraddr.Addrs)
			if !contains(Ren[rendezvous], peeraddr.ID) {
				Ren[rendezvous] = append(Ren[rendezvous], peeraddr.ID)
			}

			setTransport(ctx, peeraddr.ID, preferQUIC)
			stream, err1 := Host.NewStream(ctx, peeraddr.ID, cmdproto)
			if err1 != nil {
				fmt.Println("Error creating stream: ", err1)
				continue
			}
			if start {
				startStreams(rendezvous, peeraddr.ID)
			}

			n, err := stream.Write([]byte("rendezvous/" + rendezvous))
			fmt.Println("Wrote ", n, " bytes to stream, err: ", err)

		} else {
			fmt.Println("\t\t\tError connecting to ", peeraddr.ID)
			failed = append(failed, peeraddr.ID)
		}

	}
	fmt.Println("[*] Finished peer discovery, ", len(peersFound), " peers found, ", len(Ren[rendezvous]), " peers connected")
	return failed

}

func setTransport(ctx context.Context, peerid peer.ID, preferQUIC bool) bool {

	// get peer addrinfo from id
	peeraddr := Host.Peerstore().PeerInfo(peerid)
	var addrs []multiaddr.Multiaddr

	conn := Host.Network().ConnsToPeer(peeraddr.ID)
	for _, c := range conn {
		fmt.Println("Currently Using : ", c.ConnState().Transport)

		// if we want to use quic and the connection is quic, return
		if c.ConnState().Transport == "quic" && preferQUIC {
			fmt.Println("[*] Using QUIC protocol")
			return true
		}
		// if we want to use tcp and the connection is tcp, return
		if c.ConnState().Transport == "tcp" && !preferQUIC {
			fmt.Println("[*] Using TCP protocol")

			return true
		}
		//conn is tcp and we want quic
		if c.ConnState().Transport == "tcp" {
			fmt.Println("[*] Changing to QUIC protocol")
			addrs = selectAddrs(peeraddr.Addrs, true, false)

		}
		//conn is quic and we want tcp
		if c.ConnState().Transport == "quic" {

			fmt.Println("[*] Changing to TCP protocol")
			addrs = selectAddrs(peeraddr.Addrs, false, true)

		}

	}
	if len(addrs) == 0 {

		fmt.Print("Not Supported", peeraddr.Addrs, preferQUIC, "current transport", conn)
		return false

	}

	Host.Peerstore().ClearAddrs(peeraddr.ID)
	//Host.Peerstore().AddAddrs(peeraddr.ID, addrs, time.Hour*1)
	closeConns(peeraddr.ID)

	var NewPeer peer.AddrInfo
	NewPeer.ID = peeraddr.ID
	NewPeer.Addrs = addrs

	err := Host.Connect(ctx, NewPeer)

	if err != nil {

		fmt.Println("Error connecting to ", addrs, err)
		return false
	}

	singleconn := Host.Network().ConnsToPeer(peeraddr.ID)[0]
	if preferQUIC && singleconn.ConnState().Transport == "quic" {
		fmt.Println("[*] Succesfully changed to QUIC")
		return true

	} else if !preferQUIC && singleconn.ConnState().Transport == "tcp" {
		fmt.Println("[*] Succesfully changed to TCP")
		return true

	} else {
		fmt.Println("Error changing transport")
		return false
	}

}
func disconnectPeer(peerid string) {

	for _, conn := range Host.Network().ConnsToPeer(peer.ID(peerid)) {
		for _, stream := range conn.GetStreams() {
			fmt.Println("Closing stream", stream, stream.Protocol())
			stream.Close()
			stream.Reset()
		}
		conn.Close()

	}

}

func execCommnad(ctx context.Context, ctxmalgo *malgo.AllocatedContext, quic bool, cmdChan chan string) {
	var quitchan chan bool

	for {

		cmds := strings.SplitN(<-cmdChan, "$", 5)
		var cmd string
		var rendezvous string
		var param2 string

		if len(cmds) > 0 {
			cmd = cmds[0]
		}
		if len(cmds) > 1 {
			rendezvous = cmds[1]
		}
		if len(cmds) > 2 {
			param2 = cmds[2]
		}

		switch {

		case cmd == "mdns":

			FoundPeersMDNS := findPeersMDNS(rendezvous)
			go connecToPeersMDNS(ctx, FoundPeersMDNS, rendezvous, quic, false)

		case cmd == "dht":

			rendezvousS <- rendezvous
		case cmd == "remove":
			deleteRendezvous <- rendezvous
		case cmd == "clear":
			clear()
		case cmd == "text":
			sendTextHandler(param2, rendezvous)
		case cmd == "file":
			go sendFile(rendezvous, param2)
		case cmd == "call":
			initAudio(ctxmalgo)
			sendAudioHandler(rendezvous)
		case cmd == "stopcall":
			quitAudio(ctxmalgo)
		case cmd == "audio": //iniciar audio
			recordAudio(ctxmalgo, rendezvous, quitchan)
		case cmd == "stopaudio": //para audio y enviar
			quitchan <- true
		case cmd == "users":
			listallUSers()
		case cmd == "conns":
			listCons()
		case cmd == "streams":
			listStreams()
		case cmd == "disconn":
			disconnectPeer(param2)
		case cmd == "id":
			fmt.Println("ID: ", Host.ID())
		case cmd == "benchmark":
			nMess := 2048
			nBytes := 1024
			times := 100
			benchTCPQUIC(ctx, rendezvous, times, nBytes, nMess)
		case cmd == "help":
			fmt.Println("Commands:  \n mdns$rendezvous : Discover peers using Multicast DNS \n dht$rendezvous : Discover peers using DHT \n remove$rendezvous : Remove rendezvous from DHT \n clear : Disconnect all peers \n text$rendezvous$text : Send text to peers \n file$rendezvous$filepath : Send file to peers \n call$rendezvous : Call peers \n stopcall : Stop call \n audio$rendezvous : Record audio and send to peer \n stopaudio : Stop recording audio \n users : List all users \n conns : List all connections \n streams : List all streams \n disconn$peerid : Disconnect peer \n benchmark$times$nMessages$nBytes : Benchmark TCP/QUIC \n help : Show this help")
		default:
			fmt.Printf("Command %s not valid, see help for valid commands \n", cmd)

		}
	}
}

func hostStats(rcm network.ResourceManager) {
	for {
		<-time.After(1 * time.Minute)
		rcm.ViewSystem(func(scope network.ResourceScope) error {
			stat := scope.Stat()
			fmt.Println("System:",
				"\n\t memory", stat.Memory,
				"\n\t numFD", stat.NumFD,
				"\n\t connsIn", stat.NumConnsInbound,
				"\n\t connsOut", stat.NumConnsOutbound,
				"\n\t streamIn", stat.NumStreamsInbound,
				"\n\t streamOut", stat.NumStreamsOutbound)
			return nil
		})
		rcm.ViewTransient(func(scope network.ResourceScope) error {
			stat := scope.Stat()
			fmt.Println("Transient:",
				"\n\t memory:", stat.Memory,
				"\n\t numFD:", stat.NumFD,
				"\n\t connsIn:", stat.NumConnsInbound,
				"\n\t connsOut:", stat.NumConnsOutbound,
				"\n\t streamIn:", stat.NumStreamsInbound,
				"\n\t streamOut:", stat.NumStreamsOutbound)
			return nil
		})

	}
}
