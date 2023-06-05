package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"strings"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"

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
	filemu           sync.Mutex
	queueFiles       map[string][]string
	queueFilesMutex  sync.Mutex
	ctx              context.Context
	cancelRendezvous map[string]context.CancelFunc
	priv             crypto.PrivKey
	kdht             *dht.IpfsDHT

	data          map[string]HostData
	direcmessages map[string]DmData

	refresh     uint
	preferquic  bool
	rendezvousS chan string
	textproto   protocol.ID
	audioproto  protocol.ID
	benchproto  protocol.ID
	cmdproto    protocol.ID
	fileproto   protocol.ID
	useradded   chan bool
	updateDHT   chan bool
	reloadChat  chan string
	chatadded   chan string
	key         []byte
	messages    map[string][]Message
}
type HostData struct {
	Peers  []peer.ID `json:"Peers"`
	Timer  uint      `json:"Timer"`
	Status bool      `json:"Status"`
}
type DmData struct {
	Status bool `json:"Status"`
}
type PathFilename struct {
	Path     string `json:"path"`
	Filename string `json:"filename"`
	Progress int    `json:"progress"`
}

type User struct {
	Ip     string `json:"ip"`
	Status bool   `json:"status"`
}
type Users struct {
	Chat  string `json:"chat"`
	Peers []User `json:"user"`
}
type Statistics struct {
	SysMemory               int64 `json:"sysMemory"`
	SysNumFD                int   `json:"sysNumFD"`
	SysNumConnsInbound      int   `json:"sysNumConnsInbound"`
	SysNumConnsOutbound     int   `json:"sysNumConnsOutbound"`
	SysNumStreamsInbound    int   `json:"sysNumStreamsInbound"`
	SysNumStreamsOutbound   int   `json:"sysNumStreamsOutbound"`
	TransMemory             int64 `json:"transMemory"`
	TransNumFD              int   `json:"transNumFD"`
	TransNumConnsInbound    int   `json:"transNumConnsInbound"`
	TransNumConnsOutbound   int   `json:"transNumConnsOutbound"`
	TransNumStreamsInbound  int   `json:"transNumStreamsInbound"`
	TransNumStreamsOutbound int   `json:"transNumStreamsOutbound"`
}

func (c *P2Papp) ListUsers() []Users {

	var users []Users

	for chat, peers := range c.data {

		var aux []User

		for _, peer := range peers.Peers {

			if peer != "" {

				status := c.Host.Network().Connectedness(peer) == network.Connected
				c.fmtPrintln("status chat [", chat, " ]", status, c.Host.Network().Connectedness(peer), peer.String())
				aux = append(aux, User{Ip: peer.String(), Status: status})
			}

		}
		if len(aux) == 0 {
			aux = []User{}
		}
		users = append(users, Users{Chat: chat, Peers: aux})

	}

	return users

}
func (c *P2Papp) FakeUsers() []Users {

	var users []Users

	var aux []User

	for i := 0; i < 20; i++ {

		aux = append(aux, User{Ip: fmt.Sprintf("user-%d", i), Status: true})

	}

	users = append(users, Users{Chat: "", Peers: aux})

	return users

}

func (c *P2Papp) startup(ctx context.Context) {
	c.ctx = ctx

	c.DataChanged()

}

func (c *P2Papp) close(ctx context.Context) bool {

	if c.Host == nil {
		return false
	}

	c.Close()
	return false

}
func (c *P2Papp) Close() {
	fmt.Println("saving")
	c.saveData("direcmessages", c.direcmessages)
	c.saveData("data", c.data)
	c.saveData("message", c.messages)
	c.ctx.Done()
	c.Host.Close()
	c.kdht.Close()

}
func (c *P2Papp) saveall() {
	go func() {
		for {

			select {
			case <-c.ctx.Done():
				return
			case <-time.After(time.Second * 600):

				if c.Host != nil {

					c.saveData("direcmessages", c.direcmessages)
					c.saveData("data", c.data)
					c.saveData("message", c.messages)
				}

			}

		}
	}()
}

func (c *P2Papp) SetPeers(key string, values []peer.ID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if key != "" {
		c.data[key] = HostData{Peers: values, Timer: c.refresh, Status: true}
	}
}
func (c *P2Papp) Add(key string, value peer.ID) {
	c.mu.Lock()

	c.fmtPrintln("Adding", key, value)
	//if key not in data, add it
	if _, ok := c.data[key]; !ok {
		if key != "" {
			c.data[key] = HostData{Peers: []peer.ID{value}, Timer: c.refresh, Status: true}
		} else {
			c.data[key] = HostData{Peers: []peer.ID{}, Timer: c.refresh, Status: true}
		}

	} else {
		if !contains(c.data[key].Peers, value) {
			c.data[key] = HostData{Peers: append(c.data[key].Peers, value), Timer: c.refresh, Status: true}
		} else {
			c.data[key] = HostData{Status: true, Peers: c.data[key].Peers, Timer: c.data[key].Timer}
		}
	}

	if _, ok := c.data[""]; !ok {
		c.data[""] = HostData{Peers: []peer.ID{value}, Timer: c.refresh, Status: true}
	} else {
		if !contains(c.data[""].Peers, value) {
			c.data[""] = HostData{Peers: append(c.data[""].Peers, value), Timer: c.refresh, Status: true}
		}

	}
	c.mu.Unlock()
	c.EmitEvent("updateChats", c.GetData())
	if value != "" {
		c.EmitEvent("updateUsers", c.ListUsers())
	}

}

func (c *P2Papp) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]HostData)

	c.priv = nil

	if c.Host != nil {
		c.Host.Close()
	}
	if c.kdht != nil {
		c.kdht.Close()
	}

}
func (c *P2Papp) Get(key string, getonlineonly bool) ([]peer.ID, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	id, err := peer.Decode(key)
	if err == nil {

		if contains(c.data[""].Peers, id) {
			if getonlineonly {
				if c.Host.Network().Connectedness(id) == network.Connected {
					return []peer.ID{id}, true
				}

			}
			return []peer.ID{id}, true

		} else {

			return nil, false
		}
	} else {

		if !c.checkRend(key) {
			return nil, false
		}
		var auxpeers []peer.ID
		for _, peer := range c.data[key].Peers {
			if peer != "" {
				if getonlineonly {
					if c.Host.Network().Connectedness(peer) == network.Connected {
						auxpeers = append(auxpeers, peer)
					}
				} else {
					auxpeers = append(auxpeers, peer)
				}

			}
		}
		return auxpeers, false
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

	return keys

}

func (c *P2Papp) checkRend(rend string) bool {
	if _, ok := c.data[rend]; ok {
		return true
	}
	return false
}
func (c *P2Papp) checkUser(user string) bool {

	for _, us := range c.data[""].Peers {
		if us.String() == user {
			return true
		}
	}
	return false
}
func (c *P2Papp) GetTimer(key string) uint {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.data[key].Timer
}
func (c *P2Papp) SetTimer(key string, value uint) {
	c.mu.Lock()
	defer c.mu.Unlock()
	aux := c.data[key]
	c.data[key] = HostData{Peers: aux.Peers, Timer: value, Status: aux.Status}

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

func (c *P2Papp) GetData() map[string]HostData {
	return c.data
}

// function to create a host with a private key and a resource manager to limit the number of connections and streams per peer and per protocol
func (c *P2Papp) NewHost() string {

	limiter := rcmgr.InfiniteLimits

	rcm, err := rcmgr.NewResourceManager(rcmgr.NewFixedLimiter(limiter))
	if err != nil {
		log.Fatal(err)
	}
	/*
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
	*/
	var DefaultTransports = libp2p.ChainOptions(

		libp2p.Transport(tcp.NewTCPTransport),

		libp2p.Transport(quic.NewTransport),
	)

	c.fmtPrintln("[*] Creating Host", c.priv)
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
		libp2p.EnableNATService(),

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
	c.Host.SetStreamHandler(c.fileproto, c.receiveFilehandler)
	c.Host.SetStreamHandler(c.benchproto, c.receiveBenchhandler)
	c.Host.SetStreamHandler(c.cmdproto, c.receiveCommandhandler)

	c.fmtPrintln("\t[*] Host listening on: ", c.Host.Addrs())
	c.fmtPrintln("\t[*] Host ID : ", c.Host.ID().String())
	c.fmtPrintln("\t[*] Starting Relay system")

	_, err1 := relay.New(c.Host)
	if err1 != nil {
		log.Printf("Failed to instantiate the relay: %v", err)

	}

	c.fmtPrintln("\t[*] Host Done")
	c.Host.Network().Notify(&network.NotifyBundle{
		DisconnectedF: func(net network.Network, conn network.Conn) {
			if contains(c.data[""].Peers, conn.RemotePeer()) {
				//say why we disconnected

				peerinfo := c.Host.Peerstore().PeerInfo(conn.RemotePeer())
				//try to reconnect
				err = c.Host.Connect(c.ctx, peerinfo)
				if err != nil {
					c.fmtPrintln("Error reconnecting to peer:", err)

				}

				c.useradded <- true
			}

		},
		ConnectedF: func(net network.Network, conn network.Conn) {
			if contains(c.data[""].Peers, conn.RemotePeer()) {
				c.fmtPrintln("[*] Connected to ", conn.RemotePeer())
				c.useradded <- true
			}
		},
	})

	//start dht
	c.InitDHT()
	c.DhtRoutine(c.preferquic)
	c.saveall()

	return c.Host.ID().String()

}

func (c *P2Papp) connectToPeer(ctx context.Context, peeraddr peer.AddrInfo, rendezvous string, preferQUIC bool, start bool) bool {

	if c.Host.Network().Connectedness(peeraddr.ID) == network.Connected {
		c.fmtPrintln("\t\t\t Already connected to ", peeraddr.ID, peeraddr.Addrs)
		c.Add(rendezvous, peeraddr.ID)
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

	c.fmtPrintln("\t[*] Connecting to peers")
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
		c.fmtPrintln("[*] end connectToPeers")
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

			}
			v.Close()
		}
	}

}

func (c *P2Papp) Reconnect(rendezvous string) {

	peerids, _ := c.Get(rendezvous, false)
	var peeraddrs []peer.AddrInfo
	for _, peerid := range peerids {
		peeraddrs = append(peeraddrs, c.Host.Peerstore().PeerInfo(peerid))
	}
	c.connectToPeers(peeraddrs, rendezvous, true, true, c.ctx, context.Background())

}

func (c *P2Papp) execCommnad(quic bool, cmdChan chan string) {

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
		case cmd == "stats":
			c.HostStats()
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
			peerid, err := peer.Decode(rendezvous)
			if err == nil {
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

func (c *P2Papp) HostStats() {
	go func() {

		for {
			var stats Statistics
			var oldstats Statistics
			select {
			case <-c.ctx.Done():
				return
			case <-time.After(60 * time.Second):
				rcm := c.Host.Network().ResourceManager()

				rcm.ViewSystem(func(scope network.ResourceScope) error {
					stat := scope.Stat()

					stats.SysMemory = stat.Memory
					stats.SysNumFD = stat.NumFD
					stats.SysNumConnsInbound = stat.NumConnsInbound
					stats.SysNumConnsOutbound = stat.NumConnsOutbound
					stats.SysNumStreamsInbound = stat.NumStreamsInbound
					stats.SysNumStreamsOutbound = stat.NumStreamsOutbound

					return nil
				})
				rcm.ViewTransient(func(scope network.ResourceScope) error {
					stat := scope.Stat()

					stats.TransMemory = stat.Memory
					stats.TransNumFD = stat.NumFD
					stats.TransNumConnsInbound = stat.NumConnsInbound
					stats.TransNumConnsOutbound = stat.NumConnsOutbound
					stats.TransNumStreamsInbound = stat.NumStreamsInbound
					stats.TransNumStreamsOutbound = stat.NumStreamsOutbound

					return nil
				})

				//check if there is a change in the stats
				if !compare(stats, oldstats) {
					c.EmitEvent("Statistics", stats)
				}

				oldstats = stats

			}
		}
	}()
}
