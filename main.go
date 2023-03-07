package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gen2brain/malgo"
	"github.com/libp2p/go-libp2p/core/crypto"
)

var rendezvousS chan string

func main() {

	debug := flag.Bool("debug", false, "debug mode, generate new identity and no listen address")
	flag.Parse()
	fmt.Println("[*] Starting Application [*]")
	filename := "./config.json"
	var priv crypto.PrivKey
	if *debug {
		priv, _, _ = crypto.GenerateKeyPair(
			crypto.Ed25519, // Select your key type. Ed25519 are nice short
			-1,             // Select key length when possible (i.e. RSA).
		)

	} else {
		priv = initPriv(filename)
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
	Host, _ = newHost(hostctx, priv, *debug)
	kademliaDHT := initDHT(hostctx, Host)

	fmt.Println("Host created. We are:", Host.ID())

	// Go routines

	// connecting to peers every 5 min or when chan rendezvousS is writted

	go func() {
		for {
			select {
			case <-time.After(5 * time.Minute):
				connectToPeers(hostctx, Host, kademliaDHT)
			case <-rendezvousS:
				connectToPeers(hostctx, Host, kademliaDHT)
			}
		}
	}()
	go readStdin()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go execCommnad(ctx, mctx, priv, kademliaDHT)

	select {}
}
