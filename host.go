package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/routing"

	//"github.com/libp2p/go-libp2p/core/routing"

	"github.com/libp2p/go-libp2p/p2p/muxer/mplex"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
)

var textChan = make(chan string)
var streams = make(map[string]network.Stream)

func DisconnectHost(stream network.Stream) {

	delete(streams, stream.ID())
	fmt.Printf("Hay %d Clientes conectados \n", len(streams))
	for k, v := range streams {
		fmt.Println(k, "value is", v)
	}

}

func handleStream(stream network.Stream) {

	// Create a buffer stream for non blocking read and write.

	fmt.Print(stream.Protocol())

	go ReadStdin()
	go readData(stream)
	go writeToStreams()

}
func writeToStreams() {

	for {
		data := <-textChan
		for _, stream := range streams {
			w := bufio.NewWriter(stream)
			_, err := w.WriteString(fmt.Sprintf("%s\n", data))
			if err != nil {
				fmt.Println("Error writing to buffer")
				panic(err)
			}

			err = w.Flush()
			if err != nil {
				fmt.Println("Error flushing buffer")
				panic(err)
			}

		}

	}
}
func ReadStdin() {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		Data, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin")
			panic(err)
		}

		textChan <- Data

	}
}

func readData(stream network.Stream) {
	for {
		r := bufio.NewReader(stream)
		str, err := r.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from buffer removing stream from list")
			DisconnectHost(stream)
		}

		if str == "" {
			return
		}
		if str != "\n" {
			// Green console colour: 	\x1b[32m
			// Reset console colour: 	\x1b[0m
			fmt.Printf("\x1b[32m%s\x1b[0m> ", str)
		}

	}
}

func writeData(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin")
			panic(err)
		}

		_, err = rw.WriteString(fmt.Sprintf("%s\n", sendData))
		if err != nil {
			fmt.Println("Error writing to buffer")
			panic(err)
		}
		err = rw.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer")
			panic(err)
		}
	}
}
func NewHost(ctx context.Context, Lport int, ProtocolID string) host.Host {

	priv, _, err := crypto.GenerateKeyPair(
		crypto.Ed25519, // Select your key type. Ed25519 are nice short
		-1,             // Select key length when possible (i.e. RSA).
	)
	if err != nil {
		panic(err)
	}

	var idht *dht.IpfsDHT

	connmgr, err := connmgr.NewConnManager(
		100, // Lowwater
		400, // HighWater,
		connmgr.WithGracePeriod(time.Minute),
	)

	if err != nil {
		panic(err)
	}

	h, err := libp2p.New(
		// Use the keypair we generated
		libp2p.Identity(priv),

		// Multiple listen addresses
		libp2p.ListenAddrStrings(

			//fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", Lport),      // regular tcp connections
			fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic", Lport), // a UDP endpoint for the QUIC transport If errors regarding buffer run sudo sysctl -w net.core.rmem_max=2500000

		),
		// support TLS connections
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		// support noise connections
		libp2p.Security(noise.ID, noise.New),
		// support any other default transports (TCP)
		libp2p.DefaultTransports,

		// Let's prevent our peer from having too many
		// connections by attaching a connection manager.
		libp2p.ConnectionManager(connmgr),
		// Attempt to open ports using uPNP for NATed hosts.
		libp2p.NATPortMap(),
		// Let this host use the DHT to find other hosts
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			idht, err = dht.New(ctx, h)
			return idht, err
		}),
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

func ConnecToPeers(ctx context.Context, host host.Host, peerChan <-chan peer.AddrInfo, ProtocolID string) {

	for peer := range peerChan {
		fmt.Print("HELLLO")
		if peer.ID == host.ID() {
			continue
		}
		if err := host.Connect(ctx, peer); err != nil {
			fmt.Println("Connection failed:", err)
			continue
		}
		//fmt.Println("Connecting to:", peer)

		stream, err := host.NewStream(ctx, peer.ID, protocol.ID(ProtocolID))
		if err != nil {
			fmt.Println("Connection failed:", err)

			continue
		} else {
			streams[stream.ID()] = stream
			handleStream(stream)

		}

	}

}
