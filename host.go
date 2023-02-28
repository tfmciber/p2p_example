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
	peer   peer.AddrInfo
	online bool
}

var Ren = make(map[string][]peer.ID) //map of peers associated to a rendezvous string
var Peers = make(map[peer.ID]Peer)   //map of peers
var found chan bool

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
		stream, err1 := Host.NewStream(ctx, peeraddr.ID, "/chat/1.1.0")
		if err1 != nil {
			fmt.Println("Error connecting to ", peeraddr.Addrs, err1)
		}

		if err1 == nil {
			stream.Write([]byte("/cmd/" + rendezvous + "/"))
			fmt.Println("Successfully connected to ", peeraddr.ID, peeraddr.Addrs)
			if !Contains(Ren[rendezvous], peeraddr.ID) {
				Ren[rendezvous] = append(Ren[rendezvous], peeraddr.ID)
			}
			Peers[peeraddr.ID] = Peer{peer: peeraddr, online: true}
			setTransport(ctx, peeraddr.ID, preferQUIC)
			if start {
				startStreams(rendezvous, peeraddr, stream)
			}

		} else {
			fmt.Println("Error connecting to ", peeraddr.Addrs, err)
		}

	}
	fmt.Println("\t [*] Finished peer discovery, ", len(peersFound), " peers found, ", len(Ren[rendezvous]), " peers connected")

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
	Host.Peerstore().AddAddrs(peeraddr.ID, addrs, time.Hour*1)
	CloseConns(peeraddr.ID)

	var NewPeer peer.AddrInfo
	NewPeer.ID = peeraddr.ID
	NewPeer.Addrs = addrs

	err := Host.Connect(ctx, NewPeer)

	if err != nil {

		fmt.Println("Error connecting to ", addrs, err)
		return false
	} else {

		conn := Host.Network().ConnsToPeer(peeraddr.ID)[0]
		if preferQUIC && conn.ConnState().Transport == "quic" {
			fmt.Println("[*] Succesfully changed to QUIC")
			return true

		} else if !preferQUIC && conn.ConnState().Transport == "tcp" {
			fmt.Println("[*] Succesfully changed to TCP")
			return true

		} else {
			fmt.Println("Error changing transport")
			return false
		}

	}

}
func execCommnad(ctx context.Context, ctxmalgo *malgo.AllocatedContext, priv crypto.PrivKey) {
	var quitchan chan bool

	for {

		cmds := strings.SplitN(<-cmdChan, "$", 5)
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

			FoundPeersMDNS = FindPeersMDNS(rendezvous)
			go ConnecToPeersMDNS(ctx, FoundPeersMDNS, rendezvous, quic, false)

		case cmd == "dht":

			FoundPeersDHT = discoverPeers(ctx, Host, rendezvous)
			ConnecToPeers(ctx, FoundPeersDHT, rendezvous, quic, true)

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

			if param2 != "" {
				times, _ = strconv.Atoi(param2)

			}
			if param3 != "" {
				nMess, _ = strconv.Atoi(param3)
			}
			if param4 != "" {
				nBytes, _ = strconv.Atoi(param4)

			}
			if times < 1 {
				times = 1
			}
			if nMess < 1 {
				nMess = 1
			}
			if nBytes < 1 {
				nBytes = 1
			}
			benchTCPQUIC(ctx, rendezvous, times, nBytes, nMess)

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
			fmt.Print("benchmark$rendezvous$times$number of messages$number of bytes \n")

		}
	}
}
