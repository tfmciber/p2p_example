package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"strings"

	"github.com/gen2brain/malgo"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"

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

type P2Papp struct {
	Host host.Host
	mu   sync.Mutex
	ctx  context.Context
	priv crypto.PrivKey
	kdht *dht.IpfsDHT
	data map[string]struct {
		peers []peer.ID
		timer uint
	}

	rendezvousS chan string
	textproto   protocol.ID
	audioproto  protocol.ID
	benchproto  protocol.ID
	cmdproto    protocol.ID
	fileproto   protocol.ID
}

func (c *P2Papp) Add(key string, value peer.ID) {
	c.mu.Lock()
	aux := c.data[key].peers
	for _, a := range aux {
		if a == value {
			return
		}
	}
	aux = append(aux, value)
	c.data[key] = struct {
		peers []peer.ID
		timer uint
	}{peers: aux, timer: c.data[key].timer}

	c.mu.Unlock()
}
func (c *P2Papp) Clear() {
	c.mu.Lock()
	c.data = make(map[string]struct {
		peers []peer.ID
		timer uint
	})
	c.mu.Unlock()
}
func (c *P2Papp) Get(key string) []peer.ID {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.data[key].peers
}
func (c *P2Papp) GetTimer(key string) uint {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.data[key].timer
}
func (c *P2Papp) SetTimer(key string, value uint) {
	c.mu.Lock()
	aux := c.data[key]
	c.data[key] = struct {
		peers []peer.ID
		timer uint
	}{peers: aux.peers, timer: value}

	c.mu.Unlock()

}
func (c *P2Papp) GetKeys() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	keys := make([]string, len(c.data))
	i := 0
	for k := range c.data {
		keys[i] = k
		i++
	}
	return keys
}

// function to create a host with a private key and a resource manager to limit the number of connections and streams per peer and per protocol
func (c *P2Papp) newHost() {

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
	c.Host, err = libp2p.New(
		// Use the keypair we generated
		libp2p.Identity(c.priv),
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/udp/5000/quic", "/ip4/0.0.0.0/tcp/5000"),

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

	c.Host.SetStreamHandler(c.textproto, c.receiveTexthandler)
	c.Host.SetStreamHandler(c.audioproto, c.receiveAudioHandler)
	c.Host.SetStreamHandler(c.fileproto, c.receiveFilehandler)
	c.Host.SetStreamHandler(c.benchproto, c.receiveBenchhandler)
	c.Host.SetStreamHandler(c.cmdproto, c.receiveCommandhandler)

	fmt.Println("\t[*] Host listening on: ", c.Host.Addrs())
	fmt.Println("\t[*] Starting Relay system")

	_, err1 := relay.New(c.Host)
	if err1 != nil {
		log.Printf("Failed to instantiate the relay: %v", err)

	}

	fmt.Println("[*] Host Created")

}

// func to connect to peers found in rendezvous
func (c *P2Papp) receivePeersDHT(peerChan <-chan peer.AddrInfo, rendezvous string) []peer.AddrInfo {
	fmt.Println("[*] Connecting to peers")
	var peersFound []peer.AddrInfo
	fmt.Println("\t[*] Receiving peers")
	for peerr := range peerChan {

		if peerr.ID == c.Host.ID() {
			continue
		}
		peersFound = append(peersFound, peerr)
		fmt.Println("\t\t[*] Peer Found:", peer.Encode(peerr.ID))

	}

	fmt.Println("[*] Finished peer discovery, ", len(peersFound), " peers found, ", len(c.Get(rendezvous)), " peers connected")
	return peersFound

}
func (c *P2Papp) connectToPeer(ctx context.Context, peeraddr peer.AddrInfo, rendezvous string, preferQUIC bool, start bool) bool {

	if c.Host.Network().Connectedness(peeraddr.ID) == network.Connected {
		return true
	}

	fmt.Println("\t\t[*] Connecting to: ", peeraddr.ID)

	err := c.Host.Connect(c.ctx, peeraddr)

	if err == nil {

		fmt.Println("\t\t\t Successfully connected to ", peeraddr.ID, peeraddr.Addrs)
		c.Add(rendezvous, peeraddr.ID)

		c.setTransport(peeraddr.ID, preferQUIC)
		stream, err1 := c.Host.NewStream(ctx, peeraddr.ID, c.cmdproto)
		if err1 != nil {
			fmt.Println("\t\t\t\t Failed to create stream to ", peeraddr.ID, err1)
			return false
		}
		fmt.Println("\t\t\t\t Successfully created stream to ", peeraddr.ID)
		ms := "rendezvous$" + rendezvous
		stream.Write([]byte(ms))
		if start {
			c.startStreams(rendezvous, peeraddr.ID)
		}
		return true
	}
	fmt.Println("\t\t\t Failed to connect to ", peeraddr.ID, err)
	return false

}

func (c *P2Papp) connectToPeers(peeraddrs []peer.AddrInfo, rendezvous string, preferQUIC bool, start bool) []peer.ID {

	var failed []peer.ID
	for _, peeraddr := range peeraddrs {

		if !c.connectToPeer(c.ctx, peeraddr, rendezvous, preferQUIC, start) {
			failed = append(failed, peeraddr.ID)
		}
	}
	return failed

}
func (c *P2Papp) setTransport(peerid peer.ID, preferQUIC bool) bool {

	// get peer addrinfo from id
	peeraddr := c.Host.Peerstore().PeerInfo(peerid)
	var addrs []multiaddr.Multiaddr

	conns := c.Host.Network().ConnsToPeer(peeraddr.ID)
	for _, conn := range conns {
		fmt.Println("Currently Using : ", conn.ConnState().Transport)

		// if we want to use quic and the connection is quic, return
		if conn.ConnState().Transport == "quic" && preferQUIC {
			fmt.Println("[*] Using QUIC protocol")
			return true
		}
		// if we want to use tcp and the connection is tcp, return
		if conn.ConnState().Transport == "tcp" && !preferQUIC {
			fmt.Println("[*] Using TCP protocol")

			return true
		}
		//conn is tcp and we want quic
		if conn.ConnState().Transport == "tcp" {
			fmt.Println("[*] Changing to QUIC protocol")
			addrs = selectAddrs(peeraddr.Addrs, true, false)

		}
		//conn is quic and we want tcp
		if conn.ConnState().Transport == "quic" {

			fmt.Println("[*] Changing to TCP protocol")
			addrs = selectAddrs(peeraddr.Addrs, false, true)

		}

	}
	if len(addrs) == 0 {

		fmt.Print("Not Supported", peeraddr.Addrs, preferQUIC, "current transport", conns)
		return false

	}

	c.Host.Peerstore().ClearAddrs(peeraddr.ID)
	c.closeConns(peeraddr.ID)

	var NewPeer peer.AddrInfo
	NewPeer.ID = peeraddr.ID
	NewPeer.Addrs = addrs

	err := c.Host.Connect(c.ctx, NewPeer)

	if err != nil {

		fmt.Println("Error connecting to ", addrs, err)
		return false
	}

	connectionss := c.Host.Network().ConnsToPeer(peeraddr.ID)

	if len(connectionss) == 0 {
		fmt.Println("Error connecting to ", peeraddr.ID, connectionss)
		return false
	}
	singleconn := connectionss[0]

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
func (c *P2Papp) disconnectPeer(peerid string) {

	fmt.Println("[*] Disconnecting peer ", peerid)

	for _, v := range c.Host.Network().Conns() {

		if v.RemotePeer().String() == peerid {

			for _, stream := range v.GetStreams() {

				stream.Close()
				stream.Reset()
			}
			v.Close()
		}
	}

}

func (c *P2Papp) execCommnad(ctxmalgo *malgo.AllocatedContext, quic bool, cmdChan chan string) {
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

			FoundPeersMDNS := c.findPeersMDNS(rendezvous)
			go c.connecToPeersMDNS(FoundPeersMDNS, rendezvous, quic, false)

		case cmd == "dht":
			c.rendezvousS <- rendezvous

		case cmd == "clear":
			c.clear()
		case cmd == "text":
			c.sendTextHandler(param2, rendezvous)
		case cmd == "file":
			go c.sendFile(rendezvous, param2)
		case cmd == "call":
			initAudio(ctxmalgo)
			c.sendAudioHandler(rendezvous)
		case cmd == "stopcall":
			quitAudio(ctxmalgo)
		case cmd == "audio": //iniciar audio
			c.recordAudio(ctxmalgo, rendezvous, quitchan)
		case cmd == "stopaudio": //para audio y enviar
			quitchan <- true
		case cmd == "users":
			c.listallUSers()
		case cmd == "stats":
			c.hostStats()
		case cmd == "conns":
			c.listCons()
		case cmd == "streams":
			c.listStreams()
		case cmd == "disconn":
			c.disconnectPeer(rendezvous)
		case cmd == "id":
			fmt.Println("ID: ", c.Host.ID())
		case cmd == "benchmark":
			nMess := 2048
			nBytes := 1024
			times := 10
			c.benchTCPQUIC(rendezvous, times, nBytes, nMess)
		case cmd == "help":
			fmt.Println("Commands:  \n mdns$rendezvous : Discover peers using Multicast DNS \n dht$rendezvous : Discover peers using DHT \n remove$rendezvous : Remove rendezvous from DHT \n clear : Disconnect all peers \n text$rendezvous$text : Send text to peers \n file$rendezvous$filepath : Send file to peers \n call$rendezvous : Call peers \n stopcall : Stop call \n audio$rendezvous : Record audio and send to peer \n stopaudio : Stop recording audio \n users : List all users \n conns : List all connections \n streams : List all streams \n disconn$peerid : Disconnect peer \n benchmark$times$nMessages$nBytes : Benchmark TCP/QUIC \n help : Show this help")
		default:
			fmt.Printf("Command %s not valid, see help for valid commands \n", cmd)

		}
	}
}

func (c *P2Papp) hostStats() {

	//get rcm from host

	rcm := c.Host.Network().ResourceManager()
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
