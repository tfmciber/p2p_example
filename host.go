package main

import (
	"context"
	"fmt"
	"time"

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
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
)

var SendStreams = make(map[string]network.Stream)
var Host host.Host

//function executes terminal typed in comands
func execCommnad(ctx context.Context) {

	for {

		cmd := <-cmdChan
		//crear/llamar a las funciones para iniciar texto/audio/listar usuarios conectados/desactivar mic/sileciar/salir
		switch {
		case strings.Contains(cmd, "mdns"):
			fmt.Println("connect-mdns")
			cadena := "etesksdla321323121"
			FoundPeersMDNS = FindPeersMDNS(cadena)
			go ConnecToPeers(ctx, FoundPeersMDNS)
			fmt.Println("connect-mdns")
			SendTextHandler()
		case strings.Contains(cmd, "dht"):
			cadena := "etesksdla3213123121"
			FoundPeersDHT = discoverPeers(ctx, Host, cadena)
			go ConnecToPeers(ctx, FoundPeersDHT)

			SendTextHandler()

		case strings.Contains(cmd, "audio"):
			fmt.Print("audio")
			startInit(ctx1)
			SendAudioHandler()
		case strings.Contains(cmd, "users"):
			fmt.Print("audio")
		case strings.Contains(cmd, "mute"):
			fmt.Print("mute")
		case strings.Contains(cmd, "quit"):
			fmt.Print("mute")
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
		libp2p.Transport(tcp.NewTCPTransport),
	)

	//var idht *dht.IpfsDHT

	if err != nil {
		panic(err)
	}

	h, err := libp2p.New(
		// Use the keypair we generated
		libp2p.Identity(priv),

		// Multiple listen addresses
		libp2p.ListenAddrStrings(

			// regular tcp connections
			fmt.Sprintf("/ip4/0.0.0.0/udp/0/quic"), // a UDP endpoint for the QUIC transport If errors regarding buffer run sudo sysctl -w net.core.rmem_max=2500000
			fmt.Sprintf("/ip4/0.0.0.0/tcp/0"),
		),
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
	for _, ProtocolID := range protocols {

		h.SetStreamHandler(protocol.ID(ProtocolID), handleStream)
	}

	return h, rcm
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
		streamStart(ctx, peer.ID)

	}

}
func streamStart(ctx context.Context, peerid peer.ID) {
	for _, ProtocolID := range protocols {

		stream, err := Host.NewStream(ctx, peerid, protocol.ID(ProtocolID))
		fmt.Print("new stream")

		if err != nil {
			fmt.Println("Stream failed:", err)

			continue
		} else {
			SendStreams[stream.ID()] = stream
		}
	}
}

func debug(rcm network.ResourceManager) {

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
	}
}
