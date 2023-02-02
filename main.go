package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"context"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

func main() {

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		fmt.Println("\r- Exiting Program")
		Host.Close()
		os.Exit(0)
	}()

	fmt.Println("Exampldsdsae P2P code ")

	config, err := ParseFlags()

	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var rcm network.ResourceManager
	Host, rcm = NewHost(ctx, "/chat/1.1.0")
	if err != nil {
		panic(err)
	}
	fmt.Println("Host created. We are:", Host.ID())

	var FoundPeersDHT <-chan peer.AddrInfo
	var FoundPeersMDNS <-chan peer.AddrInfo

	if config.mdns {

		FoundPeersMDNS = FindPeersMDNS(config.RendezvousString)
		go ConnecToPeers(ctx, FoundPeersMDNS, "/chat/1.1.0")
	}

	if config.dht {

		FoundPeersDHT = discoverPeers(ctx, Host, config.RendezvousString)
		go ConnecToPeers(ctx, FoundPeersDHT, "/chat/1.1.0")

	}

	//FoundPeers := merge(FoundPeersDHT, FoundPeersMDNS)
	SendTextHandler()

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
