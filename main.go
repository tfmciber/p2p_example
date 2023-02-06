package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gen2brain/malgo"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

var cmdChan = make(chan string)

func main() {

	protocols := []string{"/audio/1.1.0", "/chat/1.1.0"}
	filename := "./config.json"

	Interrupts()
	priv := initPriv(filename)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var rcm network.ResourceManager
	Host, rcm = NewHost(ctx, protocols, priv)

	fmt.Println("Host created. We are:", Host.ID())
	ctx1, err1 := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {

	})
	if err1 != nil {
		fmt.Println(err1)
		os.Exit(1)
	}
	defer func() {
		_ = ctx1.Uninit()
		ctx1.Free()
	}()

	config, err := ParseFlags()

	if err != nil {
		panic(err)
	}
	// Go routines
	go execCommnad()
	var FoundPeersDHT <-chan peer.AddrInfo
	var FoundPeersMDNS <-chan peer.AddrInfo

	if config.mdns {

		FoundPeersMDNS = FindPeersMDNS(config.RendezvousString)
		go ConnecToPeers(ctx, FoundPeersMDNS, protocols)
	}

	if config.dht {

		FoundPeersDHT = discoverPeers(ctx, Host, config.RendezvousString)
		go ConnecToPeers(ctx, FoundPeersDHT, protocols)

	}

	//FoundPeers := merge(FoundPeersDHT, FoundPeersMDNS)
	SendTextHandler()
	//SendAudioHandler()

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
