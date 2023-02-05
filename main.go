package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"context"

	"github.com/gen2brain/malgo"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

func main() {

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

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
	CaptureDevice := initCaptureDevice(ctx1)
	PlayDevice := initPlaybackDevice(ctx1)
	startDevice(CaptureDevice)

	startDevice(PlayDevice)

	go func() {
		<-quit
		fmt.Println("\r- Exiting Program")
		Host.Close()
		os.Exit(0)
	}()

	config, err := ParseFlags()

	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	protocols := []string{"/audio/1.1.0", "/chat/1.1.0"}

	var rcm network.ResourceManager
	Host, rcm = NewHost(ctx, protocols)
	if err != nil {
		panic(err)
	}
	fmt.Println("Host created. We are:", Host.ID())

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
	SendAudioHandler()

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
