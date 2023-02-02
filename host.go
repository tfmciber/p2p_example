package main

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"

	//"github.com/libp2p/go-libp2p/core/routing"

	"github.com/libp2p/go-libp2p/p2p/muxer/mplex"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	quic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
)

var SendStreams = make(map[string]network.Stream)
var Host host.Host

func NewHost(ctx context.Context, ProtocolID string) host.Host {
	var DefaultTransports = libp2p.ChainOptions(

		libp2p.Transport(quic.NewTransport),
		libp2p.Transport(tcp.NewTCPTransport),
	)
	priv, _, err := crypto.GenerateKeyPair(
		crypto.Ed25519, // Select your key type. Ed25519 are nice short
		-1,             // Select key length when possible (i.e. RSA).
	)
	if err != nil {
		panic(err)
	}

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

	h.SetStreamHandler(protocol.ID(ProtocolID), handleStream)

	return h
}

func ConnecToPeers(ctx context.Context, peerChan <-chan peer.AddrInfo, ProtocolID string) {

	for peer := range peerChan {

		if peer.ID == Host.ID() {
			continue
		}

		if err := Host.Connect(ctx, peer); err != nil {
			fmt.Println("Connection failed:", err)
			continue
		}
		fmt.Println("Connected to:", peer.ID.Pretty())

		stream, err := Host.NewStream(ctx, peer.ID, protocol.ID(ProtocolID))

		if err != nil {
			fmt.Println("Stream failed:", err)

			continue
		} else {
			SendStreams[stream.ID()] = stream
		}

	}

}
