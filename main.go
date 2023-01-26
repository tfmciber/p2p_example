package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"context"

	"github.com/libp2p/go-libp2p/core/peer"
	// "time"
	// "github.com/libp2p/go-libp2p/p2p/net/connmgr"
	// "github.com/libp2p/go-libp2p"
	// "github.com/libp2p/go-libp2p/core/crypto"
	// "github.com/libp2p/go-libp2p/core/host"
	// "github.com/libp2p/go-libp2p/core/routing"
	// "github.com/libp2p/go-libp2p/p2p/security/noise"
	// libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
)

func main() {

	fmt.Println("Example P2P code ")

	config, err := ParseFlags()

	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	Host := NewHost(ctx, config.Lport, "/chat/1.1.0")

	fmt.Println("Host created. We are:", Host.ID())
	fmt.Println(Host.Addrs())

	var FoundPeersDHT <-chan peer.AddrInfo
	var FoundPeersMDNS <-chan peer.AddrInfo

	if config.mdns {
		fmt.Println("Finding Peers using Multicast DNS")
		FoundPeersMDNS = FindPeersMDNS(Host, config.RendezvousString)

	}

	if config.dht {
		kademliaDHT := SetandJoinDHT(ctx, Host, config.BootstrapPeers)
		FoundPeersDHT = FindPeersDHT(ctx, kademliaDHT, config.RendezvousString)
		fmt.Println(FoundPeersDHT)
	}

	FoundPeers := merge(FoundPeersDHT, FoundPeersMDNS)

	ConnecToPeers(ctx, Host, FoundPeers, "/chat/1.1.0")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT)

	select {
	case <-stop:
		Host.Close()
		os.Exit(0)

	}
}
