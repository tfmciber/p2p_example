package main

import (
	"context"
	"fmt"
	"os"
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
	peer   peer.AddrInfo
	online bool
}

var Ren = make(map[string][]peer.ID) //map of peers associated to a rendezvous string
var Peers = make(map[peer.ID]Peer)   //map of peers
var found chan bool

//function executes terminal typed in comands

func execCommnad(ctx context.Context, ctxmalgo *malgo.AllocatedContext, priv crypto.PrivKey) {
	var quitchan chan bool

	for {

		cmds := strings.SplitN(<-cmdChan, "$", 5)
		var cmd string
		var rendezvous string
		var param2, param3, param4, param5 string

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
		if len(cmds) > 5 {
			param5 = cmds[5]
		}
		quic := true
		//crear/llamar a las funciones para iniciar texto/audio/listar usuarios conectados/desactivar mic/sileciar/salir
		switch {

		case cmd == "mdns":

			FoundPeersMDNS = FindPeersMDNS(rendezvous)
			go ConnecToPeersMDNS(ctx, FoundPeersMDNS, rendezvous, quic, false)

		case cmd == "dht":

			FoundPeersDHT = discoverPeers(ctx, Host, rendezvous)
			ConnecToPeers(ctx, FoundPeersDHT, rendezvous, quic, false)

		case cmd == "clear":
			DisconnectAll()

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

			listUSers()
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
			times := 100
			mdns := "mdns"

			if param2 != "" {
				mdns = param2
			}

			if param3 != "" {
				times, _ = strconv.Atoi(param3)

			}

			if param4 != "" {
				nMess, _ = strconv.Atoi(param4)
			}
			if param5 != "" {
				nBytes, _ = strconv.Atoi(param5)

			}
			benchTCPQUIC(ctx, mdns, rendezvous, times, nBytes, nMess)

		default:
			fmt.Printf("Comnad %s not valid \n", cmd)
			fmt.Print("Valid commands are: \n")
			fmt.Print("mdns$rendezvous \n")
			fmt.Print("dht$rendezvous \n")
			fmt.Print("text$rendezvous:text \n")
			fmt.Print("file$rendezvous$absolutepath \n")
			fmt.Print("call$rendezvous \n")
			fmt.Print("stopcall \n")
			fmt.Print("audio$rendezvous \n")
			fmt.Print("stopaudio \n")
			fmt.Print("users$rendezvous \n")
			fmt.Print("allusers \n")
			fmt.Print("conns \n")
			fmt.Print("streams \n")
			fmt.Print("bench$rendezvous \n")
			fmt.Print("benchmark$rendezvous$mdns|dht$times$number of messages$number of bytes \n")

		}
	}
}

// function to create a host with a private key and a resource manager to limit the number of connections and streams per peer and per protocol
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
		libp2p.EnableRelay(),
		libp2p.EnableHolePunching(),

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

// func to see if var Peers is in Host conn
func CheckCoon() {
	found := false
	for _, v := range Ren {
		for _, p := range v {

			for _, c := range Host.Network().Conns() {
				if c.RemotePeer() == p {
					if !Peers[c.RemotePeer()].online {
						fmt.Println("User ", c.RemotePeer(), " connected")
						Peers[c.RemotePeer()] = Peer{peer: Host.Network().Peerstore().PeerInfo(c.RemotePeer()), online: true}
						fmt.Println("found", c.RemotePeer())
					}
					found = true
					break
				}
			}

			if !found {
				if Peers[p].online {
					fmt.Println("User ", p, " disconnected")
					Peers[p] = Peer{peer: Host.Network().Peerstore().PeerInfo(p), online: false}
				}
			}
		}
	}
}

// func to get peer.Addrinfo from peer.ID
func GetPeerInfo(id peer.ID) peer.AddrInfo {
	return Host.Network().Peerstore().PeerInfo(id)
}

// func t o notify on disconection

func Notifyonconnect() {
	Host.Network().Notify(&network.NotifyBundle{
		ConnectedF: func(net network.Network, conn network.Conn) {

			Peers[conn.RemotePeer()] = Peer{peer: net.Peerstore().PeerInfo(conn.RemotePeer()), online: true}

		},
	})
}

// funct to list all curennt users in rendezvous
func listUSers() {

	fmt.Println("Users connected:")
	for _, v := range Peers {
		fmt.Println(v)
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

// func to connect to peers found in rendezvous
func ConnecToPeers(ctx context.Context, peerChan <-chan peer.AddrInfo, rendezvous string, preferQUIC bool, start bool) {

	var peersFound []peer.AddrInfo
	for peer := range peerChan {

		if peer.ID == Host.ID() {
			continue
		}
		//check if peer has addresses
		if len(peer.Addrs) > 0 {
			peersFound = append(peersFound, peer)
			fmt.Println("found:", peer.ID)
		}

	}
	fmt.Print("Bootstrap finished\n")

	for _, peeraddr := range peersFound {
		err := Host.Connect(ctx, peeraddr)
		if err == nil {

			fmt.Println("Successfully connected to ", peeraddr.ID, peeraddr.Addrs)
			if !Contains(Ren[rendezvous], peeraddr.ID) {
				Ren[rendezvous] = append(Ren[rendezvous], peeraddr.ID)
			}
			Peers[peeraddr.ID] = Peer{peer: peeraddr, online: true}
			setTransport(ctx, peeraddr.ID, preferQUIC)
			if start {
				startStreams(rendezvous, peeraddr)
			}

		} else {
			fmt.Println("Error connecting to ", peeraddr.Addrs, err)
		}

	}
	fmt.Println("\t [*] Finished peer discovery, ", len(peersFound), " peers found, ", len(Ren[rendezvous]), " peers connected")

}

func setTransport(ctx context.Context, peerid peer.ID, preferQUIC bool) {

	// get peer addrinfo from id
	peeraddr := Host.Peerstore().PeerInfo(peerid)
	var addrs []multiaddr.Multiaddr

	conn := Host.Network().ConnsToPeer(peeraddr.ID)
	for _, c := range conn {
		fmt.Println("Currently Using : ", c.ConnState().Transport)

		// if we want to use quic and the connection is quic, return
		if c.ConnState().Transport == "quic" && preferQUIC {
			fmt.Println("[*] Using QUIC protocol")
			return
		}
		// if we want to use tcp and the connection is tcp, return
		if c.ConnState().Transport == "tcp" && !preferQUIC {
			fmt.Println("[*] Using TCP protocol")

			return
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

		fmt.Print("Not Supported")
		os.Exit(0)

	}

	Host.Peerstore().ClearAddrs(peeraddr.ID)
	Host.Peerstore().AddAddrs(peeraddr.ID, addrs, time.Hour*1)
	CloseConns(peeraddr.ID)

	var NewPeer peer.AddrInfo
	NewPeer.ID = peeraddr.ID
	NewPeer.Addrs = addrs

	err := Host.Connect(ctx, NewPeer)

	if err != nil {

		fmt.Println("Error connecting to ", peeraddr.Addrs, err)
	} else {

		conn := Host.Network().ConnsToPeer(peeraddr.ID)[0]
		if preferQUIC && conn.ConnState().Transport == "quic" {
			fmt.Println("Using QUIC")

		} else if !preferQUIC && conn.ConnState().Transport == "tcp" {
			fmt.Println("Using TCP")

		} else {
			fmt.Println("Error changing transport")
		}

	}

}
func SetPeersTRansport(ctx context.Context, preferQUIC bool) {

	for _, v := range Peers {
		setTransport(ctx, v.peer.ID, preferQUIC)

	}

}

func startStreams(rendezvous string, peeraddr peer.AddrInfo) {

	stream := streamStart(hostctx, peeraddr.ID, "/chat/1.1.0")
	stream.Write([]byte("/cmd/" + rendezvous + "/"))
	go ReceiveTexthandler(stream)
	stream2 := streamStart(hostctx, peeraddr.ID, "/audio/1.1.0")
	go ReceiveAudioHandler(stream2)

}
func CloseConns(ID peer.ID) {
	for _, v := range Host.Network().ConnsToPeer(ID) {
		v.Close()
	}
}

// func to get all streams with a peer of a given protcol
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
func ConnecToPeersMDNS(ctx context.Context, peerChan <-chan peer.AddrInfo, rendezvous string, preferQUIC bool, preferTCP bool) {
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
				if !Contains(Ren[rendezvous], peer.ID) {

					Ren[rendezvous] = append(Ren[rendezvous], peer.ID)
				}

				//start stream of text and audio

			}

		}
	}
}

// func to disconnect from all peers and close connections
func DisconnectAll() {
	for _, v := range Host.Network().Conns() {

		for _, s := range v.GetStreams() {
			s.Close()
			s.Reset()

		}
		v.Close()

	}
	//clear Ren
	Ren = make(map[string][]peer.ID)
	//clear Peers
	Peers = make(map[peer.ID]Peer)

}

// func for getting all streams from host of a given protocol
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

// func for checking if host has any stream open of a protocol
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

// func for checking if host has any connection open
func HasConnection() bool {
	for _, c := range Host.Network().Conns() {
		if c != nil {
			return true
		}
	}
	return false
}

// func to start a stream with a peer only if there is no stream open and return the stream in any cases
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
