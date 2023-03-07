package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"strings"

	"github.com/gen2brain/malgo"
	"github.com/libp2p/go-libp2p"

	"github.com/libp2p/go-libp2p/config"
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
var cmdChan = make(chan string)
var hostctx context.Context

// Ren variable to store the rendezvous string and the peers associated to it
var Ren = make(map[string][]peer.ID) //map of peers associated to a rendezvous string

// function to create a host with a private key and a resource manager to limit the number of connections and streams per peer and per protocol
func newHost(ctx context.Context, priv crypto.PrivKey, nolisteners bool) (host.Host, network.ResourceManager) {

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
	var addr config.Option
	if nolisteners {
		addr = libp2p.NoListenAddrs
	} else {
		addr = libp2p.ListenAddrStrings("/ip4/0.0.0.0/udp/0/quic", "/ip4/0.0.0.0/tcp/0")

	}
	if err != nil {
		panic(err)
	}
	fmt.Println("[*] Creating Host")
	h, err := libp2p.New(
		// Use the keypair we generated
		libp2p.Identity(priv),

		addr,

		// support TLS connections
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		// support noise connections
		libp2p.Security(noise.ID, noise.New),

		// support any other default transports (TCP,quic)
		DefaultTransports,
		libp2p.ResourceManager(rcm),
		libp2p.UserAgent("P2P_Example"),
		//libp2p.NATPortMap(),
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
	var protocols = []string{"/audio/1.1.0", "/chat/1.1.0", "/file/1.1.0", "/bench/1.1.0"}
	for _, ProtocolID := range protocols {

		h.SetStreamHandler(protocol.ID(ProtocolID), handleStream)
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

//func to notify if  an new peer has connected using the rendezvous string

// func to connect to peers found in rendezvous
func connecToPeers(ctx context.Context, peerChan <-chan peer.AddrInfo, rendezvous string, preferQUIC bool, start bool) []peer.AddrInfo {
	fmt.Println("[*] Connecting to peers")
	var peersFound []peer.AddrInfo
	var failed []peer.AddrInfo
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
			stream, err1 := Host.NewStream(ctx, peeraddr.ID, "/chat/1.1.0")
			if err1 != nil {
				fmt.Println("Error creating stream: ", err1)
				continue
			}
			if start {
				startStreams(rendezvous, peeraddr, stream)
			}

			n, err := stream.Write([]byte("/cmd/" + rendezvous + "/"))
			fmt.Println("Wrote ", n, " bytes to stream, err: ", err)

		} else {
			fmt.Println("\t\t\tError connecting to ", peeraddr.ID)
			failed = append(failed, peeraddr)
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
	Host.Peerstore().AddAddrs(peeraddr.ID, addrs, time.Hour*1)
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

			FoundPeersMDNS := findPeersMDNS(rendezvous)
			go connecToPeersMDNS(ctx, FoundPeersMDNS, rendezvous, quic, false)

		case cmd == "dht":

			rendezvousS <- rendezvous
		case cmd == "remove":
			deleteRendezvous <- rendezvous

		case cmd == "clear":
			disconnectAll()

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
			fmt.Printf("Command %s not valid \n", cmd)
			fmt.Print("Valid commands are: \n")
			fmt.Print("mdns$rendezvous \n")
			fmt.Print("dht$rendezvous \n")
			fmt.Print("text$rendezvous:text \n")
			fmt.Print("file$rendezvous$absolutepath \n")
			fmt.Print("call$rendezvous \n")
			fmt.Print("stopcall \n")
			fmt.Print("audio$rendezvous \n")
			fmt.Print("stopaudio \n")
			fmt.Print("allusers \n")
			fmt.Print("conns \n")
			fmt.Print("streams \n")
			fmt.Print("bench$rendezvous \n")
			fmt.Print("benchmark$rendezvous$times$number of messages$number of bytes \n")

		}
	}
}
