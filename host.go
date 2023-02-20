package main

import (
	"context"
	"fmt"
	"strconv"
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
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	quic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	tcp "github.com/libp2p/go-libp2p/p2p/transport/tcp"
)

var Host host.Host

type Peer struct {
	peer   []peer.AddrInfo
	online bool
}

var Ren = make(map[string][]peer.ID) //map of peers associated to a rendezvous string
var Peers = make(map[peer.ID]Peer)   //map of peers
var found chan bool

//function executes terminal typed in comands

func execCommnad(ctx context.Context, ctxmalgo *malgo.AllocatedContext, priv crypto.PrivKey) {
	var quitchan chan bool
	for {

		cmds := strings.SplitN(<-cmdChan, ":", 4)
		var cmd string
		var rendezvous string
		var param2, param3, param4 string

		if len(cmds) > 0 {
			cmd = cmds[0]
		}
		if len(cmds) > 1 {
			rendezvous = cmds[1]
		}
		if len(cmds) > 2 {
			param2 = cmds[2]
		}
		if len(cmds) > 3 {
			param3 = cmds[3]
		}
		if len(cmds) > 4 {
			param4 = cmds[4]
		}
		quic := true
		//crear/llamar a las funciones para iniciar texto/audio/listar usuarios conectados/desactivar mic/sileciar/salir
		switch {

		case cmd == "mdns":

			fmt.Println("MDNS")
			FoundPeersMDNS = FindPeersMDNS(rendezvous)
			go ConnecToPeersMDNS(ctx, FoundPeersMDNS, rendezvous, quic)
			go Notifyonconnect()

		case cmd == "dht":

			FoundPeersDHT = discoverPeers(ctx, Host, rendezvous)
			ConnecToPeers(ctx, FoundPeersDHT, rendezvous, quic)
			go Notifyonconnect()

		case cmd == "text":

			SendTextHandler(param2, rendezvous)

		case cmd == "file":

			go sendFile(rendezvous, param2)

		case cmd == "call":

			initAudio(ctxmalgo)
			SendAudioHandler(rendezvous)

		case cmd == "stopcall":
			quitAudio(ctxmalgo)
		case cmd == "audio": //iniciar audio
			recordAudio(ctxmalgo, rendezvous, quitchan)
		case cmd == "stopaudio": //para audio y enviar
			quitchan <- true
		case cmd == "users":

			listUSers(rendezvous)
		case cmd == "allusers":

			listallUSers()

		case cmd == "conns":

			listCons()

		case cmd == "streams":

			listStreams()
		case cmd == "bench":
			sendBench(1000, 1000, rendezvous)

		case cmd == "benchmark":

			nMess := 2048
			nBytes := 1024
			mdns := "mdns"

			if param2 != "" {
				mdns = param2
			}

			if param3 != "" {
				nMess, _ = strconv.Atoi(param3)
			}
			if param4 != "" {
				nBytes, _ = strconv.Atoi(param4)

			}

			fmt.Println("QUIC benchmark")

			DisconnectAll()
			if mdns == "mdns" {
				FoundPeersMDNS = FindPeersMDNS(rendezvous)
				go ConnecToPeersMDNS(ctx, FoundPeersMDNS, rendezvous, true)
			} else {
				FoundPeersDHT = discoverPeers(ctx, Host, rendezvous)
				ConnecToPeers(ctx, FoundPeersDHT, rendezvous, true)

			}
			go Notifyonconnect()

			<-time.After(4 * time.Second)

			for j := 64; j < nBytes+1; j += 64 {
				for i := 0; i < 1000; i++ {
					fmt.Println("Benchmarking with ", nMess, " messages of ", j, " bytes")
					sendBench(nMess, j, rendezvous)
					time.Sleep(100 * time.Millisecond)
				}
			}

			fmt.Println("TCP benchmark")
			DisconnectAll()

			if mdns == "mdns" {
				FoundPeersMDNS = FindPeersMDNS(rendezvous)
				go ConnecToPeersMDNS(ctx, FoundPeersMDNS, rendezvous, false)
			} else {
				FoundPeersDHT = discoverPeers(ctx, Host, rendezvous)
				ConnecToPeers(ctx, FoundPeersDHT, rendezvous, false)
			}
			go Notifyonconnect()

			<-time.After(4 * time.Second)

			for j := 64; j < nBytes+1; j += 64 {
				for i := 0; i < 1000; i++ {
					fmt.Println("Benchmarking with ", nMess, " messages of ", j, " bytes")
					sendBench(nMess, j, rendezvous)
					time.Sleep(100 * time.Millisecond)
				}
			}

		default:
			fmt.Printf("Comnad %s not valid \n", cmd)
			fmt.Print("Valid commands are: \n")
			fmt.Print("mdns:rendezvous \n")
			fmt.Print("dht:rendezvous \n")
			fmt.Print("text:rendezvous:text \n")
			fmt.Print("file:rendezvous:file \n")
			fmt.Print("call:rendezvous \n")
			fmt.Print("stopcall \n")
			fmt.Print("audio:rendezvous \n")
			fmt.Print("stopaudio \n")
			fmt.Print("users:rendezvous \n")
			fmt.Print("allusers \n")
			fmt.Print("conns \n")
			fmt.Print("streams \n")
			fmt.Print("bench:rendezvous \n")
			fmt.Print("benchmark:rendezvous:mdns|dht:number of messages:number of bytes \n")

		}
	}
}

//create a new stream to a peer using tcp transport protocol

//function to create a host with a private key and a resource manager to limit the number of connections and streams per peer and per protocol
func NewHost(ctx context.Context, priv crypto.PrivKey) (host.Host, network.ResourceManager) {

	limiterCfg := `{
    "System":  {
      "StreamsInbound": 4096,
      "StreamsOutbound": 32768,
      "Conns": 64000,
      "ConnsInbound": 512,
      "ConnsOutbound": 32768,
      "FD": 64000
    },
    "Transient": {
      "StreamsInbound": 4096,
      "StreamsOutbound": 32768,
      "ConnsInbound": 512,
      "ConnsOutbound": 32768,
      "FD": 64000
    },

    "ProtocolDefault":{
      "StreamsInbound": 1024,
      "StreamsOutbound": 32768
    },

    "ServiceDefault":{
      "StreamsInbound": 2048,
      "StreamsOutbound": 32768
    }
  }`

	limiter, err := rcmgr.NewDefaultLimiterFromJSON(strings.NewReader(limiterCfg))
	if err != nil {
		panic(err)
	}
	rcm, err := rcmgr.NewResourceManager(limiter)
	if err != nil {
		panic(err)
	}

	var DefaultTransports = libp2p.ChainOptions(

		libp2p.Transport(tcp.NewTCPTransport),

		libp2p.Transport(quic.NewTransport),
	)

	//var idht *dht.IpfsDHT

	if err != nil {
		panic(err)
	}

	h, err := libp2p.New(
		// Use the keypair we generated
		libp2p.Identity(priv),

		// Multiple listen addresses
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/udp/0/quic", "/ip4/0.0.0.0/tcp/0"),

		// support TLS connections
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		// support noise connections
		libp2p.Security(noise.ID, noise.New),

		// support any other default transports (TCP,quic)
		DefaultTransports,
		libp2p.ResourceManager(rcm),
		libp2p.NATPortMap(),

		libp2p.ChainOptions(
			libp2p.Muxer("/yamux/1.0.0", yamux.DefaultTransport),
			libp2p.Muxer("/mplex/6.7.0", mplex.DefaultTransport),
		),
		libp2p.EnableNATService(),
	)
	if err != nil {
		panic(err)
	}
	var protocols = []string{"/audio/1.1.0", "/chat/1.1.0", "/file/1.1.0", "/bench/1.1.0"}
	for _, ProtocolID := range protocols {

		h.SetStreamHandler(protocol.ID(ProtocolID), handleStream)
	}

	return h, rcm
}

func listCons() {
	fmt.Println("Conns open:")
	for _, v := range Host.Network().Conns() {
		fmt.Println(v)
	}
}

//func to close all streams of a given protocol
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

// func t o notify on disconection

func Notifyonconnect() {
	Host.Network().Notify(&network.NotifyBundle{
		ConnectedF: func(net network.Network, conn network.Conn) {
			fmt.Println("Connected to:", conn.RemotePeer(), " ", conn.RemoteMultiaddr())
			var addrs []peer.AddrInfo

			addrs = append(addrs, net.Peerstore().PeerInfo(conn.RemotePeer()))
			Peers[conn.RemotePeer()] = Peer{peer: addrs, online: true}

		},
	})
}

// funct to list all curennt users in rendezvous
func listUSers(rendezvous string) {

	fmt.Println("Users connected:")
	for _, v := range Ren[rendezvous] {
		fmt.Println(v, Peers[v])
	}
}

// funct to list all curennt users
func listallUSers() {

	fmt.Println("Users connected:")
	for str, peerr := range Ren {
		for _, p := range peerr {
			fmt.Printf("Rendezvous %s peer ID %s ", str, p)
		}
	}
}

func ConnecToPeers(ctx context.Context, peerChan <-chan peer.AddrInfo, rendezvous string, preferQUIC bool) {

	var peersFound []peer.AddrInfo
	for peer := range peerChan {

		if peer.ID == Host.ID() {
			continue
		}
		peersFound = append(peersFound, peer)
		fmt.Println("found:", peer.ID.Pretty())

	}
	fmt.Print("Bootstrap finished")
	DisconnectAll()

	for _, peer := range peersFound {
		fmt.Println("Connecting to:", peer.ID.Pretty(), " ", peer.Addrs)
		//connect to the peer to the tcp address
		//connect only to quic addresses
		if preferQUIC {
			var addrs []multiaddr.Multiaddr
			for _, v := range peer.Addrs {
				if strings.Contains(v.String(), "quic") {
					addrs = append(addrs, v)
				}
			}
			peer.Addrs = addrs
		}

		//if not prefer quic connect to tcp addresses and if there are no quic addreses
		if !preferQUIC || len(peer.Addrs) == 0 {
			var addrs []multiaddr.Multiaddr
			for _, v := range peer.Addrs {
				if strings.Contains(v.String(), "tcp") {
					addrs = append(addrs, v)
				}
			}
			peer.Addrs = addrs
		}

		err := Host.Connect(ctx, peer)

		fmt.Println(err)
		if err != nil {
			fmt.Println("Error connecting to peer:", err)
		}
		if err == nil {

			if !Contains(Ren[rendezvous], peer.ID) {
				Ren[rendezvous] = append(Ren[rendezvous], peer.ID)
			}

			stream := streamStart(hostctx, peer.ID, "/chat/1.1.0")
			go ReceiveTexthandler(stream)

			stream2 := streamStart(hostctx, peer.ID, "/audio/1.1.0")
			go ReceiveAudioHandler(stream2)
		}

	}
	fmt.Println("Connected to all peers")

}

//func to get all streams with a peer of a given protcol
func GetStreamsFromPeerProto(peerID peer.ID, protocol string) network.Stream {

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

func ConnecToPeersMDNS(ctx context.Context, peerChan <-chan peer.AddrInfo, rendezvous string, preferQUIC bool) {

	for peer := range peerChan {

		if peer.ID == Host.ID() {
			continue
		}

		fmt.Println("Connecting to:", peer.ID.Pretty(), " ", peer.Addrs)

		//connect only to quic addresses
		if preferQUIC {
			fmt.Println("quic prerence")
			var addrs []multiaddr.Multiaddr
			for _, v := range peer.Addrs {
				if strings.Contains(v.String(), "quic") {
					addrs = append(addrs, v)
				}
			}
			peer.Addrs = addrs
		}

		//if not prefer quic connect to tcp addresses and if there are no quic addreses
		if !preferQUIC || len(peer.Addrs) == 0 {
			fmt.Println("tcp prerence")
			var addrs []multiaddr.Multiaddr
			for _, v := range peer.Addrs {
				if strings.Contains(v.String(), "tcp") {
					addrs = append(addrs, v)
				}
			}
			peer.Addrs = addrs
		}

		err := Host.Connect(ctx, peer)
		if err != nil {
			fmt.Println("Error connecting to peer:", err)
		}
		//in peer.ID not in Ren[rendezvous] add to Ren[rendezvous]
		if !Contains(Ren[rendezvous], peer.ID) {

			Ren[rendezvous] = append(Ren[rendezvous], peer.ID)
		}
		found <- true
		//start stream of text and audio

	}

}

//func to disconnect from all peers and close connections
func DisconnectAll() {
	for _, v := range Host.Network().Conns() {

		for _, s := range v.GetStreams() {
			s.Close()
			s.Reset()

		}
		v.Close()

	}
}

//func for getting all streams from host of a given protocol
func GetStreamsFromProtocol(protocol string) []network.Stream {
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

//func for checking if host has any stream open of a protocol
func HasStreamOfProtocol(protocol string) bool {
	for _, c := range Host.Network().Conns() {
		for _, s := range c.GetStreams() {

			if string(s.Protocol()) == protocol {
				return true
			}
		}
	}
	return false
}

//func for checking if host has any connection open
func HasConnection() bool {
	for _, c := range Host.Network().Conns() {
		if c != nil {
			return true
		}
	}
	return false
}

//func to start a stream with a peer only if there is no stream open and return the stream in any cases
func streamStart(ctx context.Context, peerid peer.ID, ProtocolID string) network.Stream {

	stream := GetStreamsFromPeerProto(peerid, ProtocolID)

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
