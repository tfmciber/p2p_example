package main

import (
	"context"
	"fmt"
	"log"
	filepath "path/filepath"
	"sync"
	"time"

	"strings"

	"github.com/gen2brain/malgo"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/libp2p/go-libp2p/p2p/muxer/mplex"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	quic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	tcp "github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/multiformats/go-multiaddr"
)

// Host is the libp2p host

type P2Papp struct {
	Host             host.Host
	mu               sync.Mutex
	ctx              context.Context
	cancelRendezvous context.CancelFunc
	priv             crypto.PrivKey
	kdht             *dht.IpfsDHT
	data             map[string]struct {
		peers []peer.ID
		timer uint
	}

	direcmessages []peer.ID
	refresh       uint
	preferquic    bool
	rendezvousS   chan string
	textproto     protocol.ID
	audioproto    protocol.ID
	benchproto    protocol.ID
	cmdproto      protocol.ID
	fileproto     protocol.ID
	//chanel of struct of 2 strings
	useradded chan bool
	chatadded chan string
}

type PathFilename struct {
	Path     string `json:"path"`
	Filename string `json:"filename"`
}

type User struct {
	Ip     string `json:"ip"`
	Status bool   `json:"status"`
}
type Users struct {
	Chat  string `json:"chat"`
	Peers []User `json:"user"`
}

func (c *P2Papp) ListUsers() []Users {

	var users []Users

	for chat, peers := range c.data {

		var aux []User

		for _, peer := range peers.peers {

			if peer != "" {

				status := c.Host.Network().Connectedness(peer) == network.Connected
				c.fmtPrintln("status chat [", chat, " ]", status, c.Host.Network().Connectedness(peer), peer.String())
				aux = append(aux, User{Ip: peer.String(), Status: status})
			}

		}

		users = append(users, Users{Chat: chat, Peers: aux})

	}

	return users

}

func (c *P2Papp) SelectFiles() []PathFilename {

	var pathFilenames []PathFilename

	file, err := runtime.OpenMultipleFilesDialog(c.ctx, runtime.OpenDialogOptions{})
	if err != nil {
		return nil
	}
	//get file sizes
	i := 0
	for _, f := range file {
		filename := filepath.Base(f)
		pathFilenames = append(pathFilenames, PathFilename{Path: f, Filename: filename})

		i += 1
	}

	return pathFilenames
}
func (c *P2Papp) startup(ctx context.Context) {
	c.ctx = ctx
	c.DataChanged()
}
func (c *P2Papp) Add(key string, value peer.ID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.fmtPrintln("Adding", key, value)
	//if key not in data, add it
	if _, ok := c.data[key]; !ok {
		c.data[key] = struct {
			peers []peer.ID
			timer uint
		}{peers: []peer.ID{value}, timer: c.refresh}
		go func() {

			c.chatadded <- key

		}()

	} else {
		c.data[key] = struct {
			peers []peer.ID
			timer uint
		}{peers: append(c.data[key].peers, value), timer: c.refresh}

	}

	if _, ok := c.data[""]; !ok {
		c.data[""] = struct {
			peers []peer.ID
			timer uint
		}{peers: []peer.ID{value}, timer: c.refresh}
	} else {
		//append peer only if not already in list of peer.ID in c.data[""]
		if !contains(c.data[""].peers, value) {
			c.data[""] = struct {
				peers []peer.ID
				timer uint
			}{peers: append(c.data[""].peers, value), timer: c.refresh}
		}

	}

	if value != "" {
		go func() {
			//aux := []string{key, value.String()}
			c.useradded <- true
		}()

	}
}

func (c *P2Papp) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.fmtPrintln("clear")
	c.data = make(map[string]struct {
		peers []peer.ID
		timer uint
	})

	c.priv = nil

	if c.Host != nil {
		c.Host.Close()
	}
	if c.kdht != nil {
		c.kdht.Close()
	}

}
func (c *P2Papp) Get(key string) ([]peer.ID, bool) {

	c.mu.Lock()

	defer c.mu.Unlock()
	c.fmtPrintln("Get", key, c.data[key].peers)
	id := c.GetPeerIDfromstring(key)
	if id != "" {

		if contains(c.data[""].peers, c.GetPeerIDfromstring(key)) {
			return []peer.ID{c.GetPeerIDfromstring(key)}, true
		} else {
			return nil, false
		}
	} else {
		return c.data[key].peers, false
	}
}
func (c *P2Papp) GetRend() []string {

	c.mu.Lock()

	defer c.mu.Unlock()
	//return all keys of c.data

	keys := make([]string, 0, len(c.data))
	for k := range c.data {
		keys = append(keys, k)
	}
	c.fmtPrintln("keys=", keys)
	return keys

}
func (c *P2Papp) GetTimer(key string) uint {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.data[key].timer
}
func (c *P2Papp) SetTimer(key string, value uint) {
	c.mu.Lock()
	defer c.mu.Unlock()
	aux := c.data[key]
	c.data[key] = struct {
		peers []peer.ID
		timer uint
	}{peers: aux.peers, timer: value}

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
func (c *P2Papp) ListChats() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	var chats []string
	for k := range c.data {
		if k != "" {
			chats = append(chats, k)
		}
	}
	return chats
}

//func that alerts if c.data has changed
func (c *P2Papp) DataChanged() {

	go func() {
		for {
			select {
			case <-c.chatadded:
				c.fmtPrintln("chatadded")
				aux := c.ListChats()

				runtime.EventsEmit(c.ctx, "updateChats", aux)
			case <-c.useradded:
				c.fmtPrintln("user added")
				aux := c.ListUsers()

				runtime.EventsEmit(c.ctx, "updateUsers", aux)
			case <-time.After(30 * time.Second):
				aux := c.ListUsers()
				runtime.EventsEmit(c.ctx, "updateUsers", aux)

			case <-c.ctx.Done():

				return
			}
		}
	}()
}

func (c *P2Papp) GetData() map[string]struct {
	peers []peer.ID
	timer uint
} {
	return c.data
}

// function to create a host with a private key and a resource manager to limit the number of connections and streams per peer and per protocol
func (c *P2Papp) NewHost() string {

	/*
		limiter := rcmgr.InfiniteLimits

		rcm, err := rcmgr.NewResourceManager(rcmgr.NewFixedLimiter(limiter))
		if err != nil {
			log.Fatal(err)
		}
	*/
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

	c.fmtPrintln("[*] Creating Host")
	c.Host, err = libp2p.New(
		// Use the keypair we generated
		libp2p.Identity(c.priv),
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/udp/0/quic", "/ip4/0.0.0.0/tcp/0"),

		// support TLS connections
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		// support noise connections
		libp2p.Security(noise.ID, noise.New),
		// support any other default transports (TCP,quic)
		DefaultTransports,

		libp2p.ResourceManager(rcm),
		libp2p.UserAgent("P2P_Example"),
		libp2p.NATPortMap(),
		libp2p.EnableRelay(),
		libp2p.EnableHolePunching(),
		//libp2p.EnableNATService(),

		libp2p.ChainOptions(
			libp2p.Muxer("/yamux/1.0.0", yamux.DefaultTransport),
			libp2p.Muxer("/mplex/6.7.0", mplex.DefaultTransport),
		),
	)
	if err != nil {
		c.fmtPrintln("Error creating host: ", err)
		panic(err)

	}

	c.Host.SetStreamHandler(c.textproto, c.receiveTexthandler)
	c.Host.SetStreamHandler(c.audioproto, c.receiveAudioHandler)
	c.Host.SetStreamHandler(c.fileproto, c.receiveFilehandler)
	c.Host.SetStreamHandler(c.benchproto, c.receiveBenchhandler)
	c.Host.SetStreamHandler(c.cmdproto, c.receiveCommandhandler)

	c.fmtPrintln("\t[*] Host listening on: ", c.Host.Addrs())
	c.fmtPrintln("\t[*] Starting Relay system")

	_, err1 := relay.New(c.Host)
	if err1 != nil {
		log.Printf("Failed to instantiate the relay: %v", err)

	}
	c.fmtPrintln("\t[*] Host Done")
	return c.Host.ID().String()

}

func (c *P2Papp) connectToPeer(ctx context.Context, peeraddr peer.AddrInfo, rendezvous string, preferQUIC bool, start bool) bool {

	if c.Host.Network().Connectedness(peeraddr.ID) == network.Connected {
		c.fmtPrintln("\t\t\t Already connected to ", peeraddr.ID, peeraddr.Addrs)
		return true
	}

	c.fmtPrintln("\t\t[*] Connecting to: ", peeraddr.ID)

	err := c.Host.Connect(c.ctx, peeraddr)

	if err == nil {

		c.fmtPrintln("\t\t\t Successfully connected to ", peeraddr.ID, peeraddr.Addrs)
		c.Add(rendezvous, peeraddr.ID)

		c.setTransport(peeraddr.ID, preferQUIC)
		stream, err1 := c.Host.NewStream(ctx, peeraddr.ID, c.cmdproto)
		if err1 != nil {
			c.fmtPrintln("\t\t\t\t Failed to create stream to ", peeraddr.ID, err1)
			return false
		}
		c.fmtPrintln("\t\t\t\t Successfully created stream to ", peeraddr.ID)
		ms := "rendezvous$" + rendezvous
		stream.Write([]byte(ms))
		if start {
			c.startStreams(rendezvous, peeraddr.ID)
		}
		return true
	}
	c.fmtPrintln("\t\t\t Failed to connect to ", peeraddr.ID, err)
	return false

}

func (c *P2Papp) connectToPeers(peeraddrs []peer.AddrInfo, rendezvous string, preferQUIC bool, start bool, ctx context.Context, ctx2 context.Context) []peer.ID {

	var failed []peer.ID
	var wg sync.WaitGroup
	var mu sync.Mutex
	var end = make(chan bool)
	go func() {
		for _, peeraddr := range peeraddrs {

			wg.Add(1)
			go func(peeraddr peer.AddrInfo) {
				defer wg.Done()
				if !c.connectToPeer(c.ctx, peeraddr, rendezvous, preferQUIC, start) {
					mu.Lock()
					failed = append(failed, peeraddr.ID)
					mu.Unlock()
				}

			}(peeraddr)

		}
		wg.Wait()
		end <- true
	}()
	select {
	case <-ctx.Done():
		c.fmtPrintln("[*] global ctx done")
		return nil
	case <-ctx2.Done():
		c.fmtPrintln("[*] ctx2 done")
		return nil
	case <-end:
		return failed

	}

}
func (c *P2Papp) setTransport(peerid peer.ID, preferQUIC bool) bool {

	// get peer addrinfo from id
	peeraddr := c.Host.Peerstore().PeerInfo(peerid)
	var addrs []multiaddr.Multiaddr

	conns := c.Host.Network().ConnsToPeer(peeraddr.ID)
	for _, conn := range conns {
		c.fmtPrintln("Currently Using : ", conn.ConnState().Transport)

		// if we want to use quic and the connection is quic, return
		if conn.ConnState().Transport == "quic" && preferQUIC {
			c.fmtPrintln("[*] Using QUIC protocol")
			return true
		}
		// if we want to use tcp and the connection is tcp, return
		if conn.ConnState().Transport == "tcp" && !preferQUIC {
			c.fmtPrintln("[*] Using TCP protocol")

			return true
		}
		//conn is tcp and we want quic
		if conn.ConnState().Transport == "tcp" {
			c.fmtPrintln("[*] Changing to QUIC protocol")
			addrs = selectAddrs(peeraddr.Addrs, true, false)

		}
		//conn is quic and we want tcp
		if conn.ConnState().Transport == "quic" {

			c.fmtPrintln("[*] Changing to TCP protocol")
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

		c.fmtPrintln("Error connecting to ", addrs, err)
		return false
	}

	connectionss := c.Host.Network().ConnsToPeer(peeraddr.ID)

	if len(connectionss) == 0 {
		c.fmtPrintln("Error connecting to ", peeraddr.ID, connectionss)
		return false
	}
	singleconn := connectionss[0]

	if preferQUIC && singleconn.ConnState().Transport == "quic" {
		c.fmtPrintln("[*] Succesfully changed to QUIC")
		return true

	} else if !preferQUIC && singleconn.ConnState().Transport == "tcp" {
		c.fmtPrintln("[*] Succesfully changed to TCP")
		return true

	} else {
		c.fmtPrintln("Error changing transport")
		return false
	}

}
func (c *P2Papp) disconnectPeer(peerid string) {

	c.fmtPrintln("[*] Disconnecting peer ", peerid)

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

func (c *P2Papp) Reconnect(rendezvous string) {

	peerids, _ := c.Get(rendezvous)
	var peeraddrs []peer.AddrInfo
	for _, peerid := range peerids {
		peeraddrs = append(peeraddrs, c.Host.Peerstore().PeerInfo(peerid))
	}
	c.connectToPeers(peeraddrs, rendezvous, true, true, c.ctx, context.Background())

}
func (c *P2Papp) GetPeerIDfromstring(peerid string) peer.ID {

	for _, v := range c.Host.Network().Conns() {

		if v.RemotePeer().String() == peerid {

			return v.RemotePeer()
		}

	}
	return peer.ID("")

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
			c.SendTextHandler(param2, rendezvous)
		case cmd == "file":
			go c.SendFile(rendezvous, param2)
		case cmd == "call":
			initAudio(ctxmalgo)
			c.sendAudioHandler(rendezvous)
		case cmd == "stopcall":
			quitAudio(ctxmalgo)
		case cmd == "audio": //iniciar audio
			c.recordAudio(ctxmalgo, rendezvous, quitchan)
		case cmd == "stopaudio": //para audio y enviar
			quitchan <- true

		case cmd == "stats":
			c.hostStats()
		case cmd == "conns":
			c.listCons()
		case cmd == "streams":
			c.listStreams()
		case cmd == "disconn":
			c.disconnectPeer(rendezvous)
		case cmd == "id":
			c.fmtPrintln("ID: ", c.Host.ID())
		case cmd == "benchmark":
			nMess := 1000
			nBytes := 1024
			times := 1
			peerid := c.GetPeerIDfromstring(rendezvous)
			if peerid != "" {
				c.benchTCPQUIC(peerid, nBytes, nMess, times)
			}

		case cmd == "help":
			c.fmtPrintln("Commands:  \n mdns$rendezvous : Discover peers using Multicast DNS \n dht$rendezvous : Discover peers using DHT \n remove$rendezvous : Remove rendezvous from DHT \n clear : Disconnect all peers \n text$rendezvous$text : Send text to peers \n file$rendezvous$filepath : Send file to peers \n call$rendezvous : Call peers \n stopcall : Stop call \n audio$rendezvous : Record audio and send to peer \n stopaudio : Stop recording audio \n users : List all users \n conns : List all connections \n streams : List all streams \n disconn$peerid : Disconnect peer \n benchmark$times$nMessages$nBytes : Benchmark TCP/QUIC \n help : Show this help")
		case cmd == "exit":
			c.fmtPrintln("Exiting...")

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
		c.fmtPrintln("System:",
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
		c.fmtPrintln("Transient:",
			"\n\t memory:", stat.Memory,
			"\n\t numFD:", stat.NumFD,
			"\n\t connsIn:", stat.NumConnsInbound,
			"\n\t connsOut:", stat.NumConnsOutbound,
			"\n\t streamIn:", stat.NumStreamsInbound,
			"\n\t streamOut:", stat.NumStreamsOutbound)
		return nil
	})

}
