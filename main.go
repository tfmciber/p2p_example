package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	mrand "math/rand"
	"os"

	"github.com/gen2brain/malgo"
	"github.com/libp2p/go-libp2p/core/crypto"
)

func main() {

	debug := flag.Bool("debug", false, "debug mode, Prints host stats")
	seed := flag.Int("seed", 0, "Seed for the random number generator for debug mode")
	refreshTime := flag.Uint("refresh", 50, "Minutes to refresh the DHT")
	quic := flag.Bool("quic", false, "Use QUIC transport")
	filename := flag.String("config", "./config.json", "Config file")
	flag.Parse()
	fmt.Println("[*] Starting Application [*]")

	fmt.Println("\t[*] Debug mode:", *debug, "Refresh time:", *refreshTime, "QUIC:", *quic, "Config file:", *filename)

	var priv crypto.PrivKey
	if *debug {

		seed := mrand.New(mrand.NewSource(int64(*seed)))
		priv, _, _ = crypto.GenerateECDSAKeyPair(seed)
	} else {
		priv = initPriv(*filename)
	}

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

	Host, _ = newHost(hostctx, priv)
	kademliaDHT := initDHT(hostctx, Host)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println("Host created. We are:", Host.ID())

	// Go routines

	go dhtRoutine(ctx, rendezvousS, kademliaDHT, *quic, *refreshTime)
	go readStdin()

	go execCommnad(ctx, mctx, *quic)

	select {}
}
