package main

import (
	"context"
	"fmt"

	"github.com/gen2brain/malgo"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"

	//"github.com/libp2p/go-libp2p/core/routing"

	"strings"

	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/libp2p/go-libp2p/p2p/muxer/mplex"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	quic "github.com/libp2p/go-libp2p/p2p/transport/quic"
)

var Host host.Host

//function executes terminal typed in comands
func execCommnad(ctx context.Context, ctxmalgo *malgo.AllocatedContext) {

	for {

		cmd := <-cmdChan
		//crear/llamar a las funciones para iniciar texto/audio/listar usuarios conectados/desactivar mic/sileciar/salir
		switch {
		case strings.Contains(cmd, "mdns:"):
			cadena := strings.TrimPrefix(cmd, "mdns:")
			FoundPeersMDNS = FindPeersMDNS(cadena)
			go ConnecToPeers(ctx, FoundPeersMDNS)
		case strings.Contains(cmd, "dht:"):
			cadena := strings.TrimPrefix(cmd, "dht:")
			FoundPeersDHT = discoverPeers(ctx, Host, cadena)
			go ConnecToPeers(ctx, FoundPeersDHT)
		case strings.Contains(cmd, "text:"):
			cadena := strings.TrimPrefix(cmd, "text:")
			textChan <- []byte(cadena)
		case strings.Contains(cmd, "file:"):
			filename := strings.TrimPrefix(cmd, "file:")
			sendFile(filename)
		case strings.Contains(cmd, "audio"):
			initAudio(ctxmalgo)
		case strings.Contains(cmd, "users"):
			listUSers()
		case strings.Contains(cmd, "conns"):
			listCons()
		case strings.Contains(cmd, "streams"):
			listStreams()
		default:
			fmt.Printf("Comnad %s not valid \n", cmd)
		}
	}
}

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

		libp2p.Transport(quic.NewTransport),
	//	libp2p.Transport(tcp.NewTCPTransport),
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

		//libp2p.ListenAddrStrings("/ip4/0.0.0.0/udp/0/quic"),
		// support TLS connections
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		// support noise connections
		libp2p.Security(noise.ID, noise.New),

		// support any other default transports (TCP,quic)
		DefaultTransports,
		libp2p.ResourceManager(rcm),

		// Let's prevent our peer from having too many
		// connections by attaching a connection manager.
		//libp2p.ConnectionManager(connmgr),
		// Attempt to open ports using uPNP for NATed hosts.
		libp2p.NATPortMap(),
		// Let this host use the DHT to find other hosts
		// libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
		// 	idht, err = dht.New(ctx, h)
		// 	return idht, err
		// }),
		// If you want to help other peers to figure out if they are behind
		// NATs, you can launch the server-side of AutoNAT too (AutoRelay
		// already runs the client)
		//
		// This service is highly rate-limited and should not cause any
		// performance issues.
		libp2p.ChainOptions(
			libp2p.Muxer("/yamux/1.0.0", yamux.DefaultTransport),
			libp2p.Muxer("/mplex/6.7.0", mplex.DefaultTransport),
		),
		libp2p.EnableNATService(),
	)
	if err != nil {
		panic(err)
	}
	var protocols = []string{"/audio/1.1.0", "/chat/1.1.0", "/file/1.1.0"}
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

// func to list all streams open
func listStreams() {
	fmt.Println("Streams open:")
	for _, v := range Host.Network().Conns() {
		//list all streams for the connection
		for _, s := range v.GetStreams() {
			fmt.Println(s)
		}

	}
}

// func t o notify on disconection

func Notifyondisconnect() {
	Host.Network().Notify(&network.NotifyBundle{
		DisconnectedF: func(net network.Network, conn network.Conn) {
			fmt.Println("Disconnected from:", conn.RemotePeer())
		},
	})
}

// funct to list all curennt users
func listUSers() {
	fmt.Println("Users connected:")
	for _, v := range Host.Network().Peers() {
		fmt.Println(v)
	}
}

func ConnecToPeers(ctx context.Context, peerChan <-chan peer.AddrInfo) {

	for peer := range peerChan {

		if peer.ID == Host.ID() {
			continue
		}

		if err := Host.Connect(ctx, peer); err != nil {
			fmt.Println("Connection failed:", err)
			continue
		}

		fmt.Println("Connected to:", peer.ID.Pretty())

	}

}

//func to check if stream is bidirectional not using dirboth

func isBidirectional(stream network.Stream) bool {
	return stream.Stat().Direction == network.DirOutbound

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
func HasStream() bool {
	for _, c := range Host.Network().Conns() {
		for _, s := range c.GetStreams() {
			if s != nil {
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
	if HasStreamOfProtocol(ProtocolID) == false {

		stream, err := Host.NewStream(ctx, peerid, protocol.ID(ProtocolID))

		if err != nil {
			fmt.Println("Stream failed:", err)

		}
		return stream
	}
	return GetStreamsFromProtocol(ProtocolID)[0]
}
