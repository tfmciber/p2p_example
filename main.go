package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gen2brain/malgo"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
)

func main() {

	debug := flag.Bool("debug", false, "debug mode, Prints host stats")
	refreshTime := flag.Uint("refresh", 5, "Minutes to refresh the DHT")
	quic := flag.Bool("quic", false, "Use QUIC transport")
	filename := flag.String("config", "./config.json", "Config file")
	flag.Parse()
	fmt.Println("[*] Starting Application [*]")
	fmt.Println("\t[*] Debug mode:", *debug, "Refresh time:", *refreshTime, "QUIC:", *quic, "Config file:", *filename)

	var priv crypto.PrivKey

	priv = initPriv(*filename)

	hostctx = context.Background()

	interrupts()

	mctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {

	})
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer func() {
		_ = mctx.Uninit()
		mctx.Free()
	}()
	var rcm network.ResourceManager
	Host, rcm = newHost(hostctx, priv)
	kademliaDHT := initDHT(hostctx, Host)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println("Host created. We are:", Host.ID())

	// Go routines

	go dhtRoutine(ctx, rendezvousS, kademliaDHT, *quic, *refreshTime)
	go readStdin()

	go execCommnad(ctx, mctx)
	if *debug {
		go hostStats(rcm)
	}

	select {}
}
