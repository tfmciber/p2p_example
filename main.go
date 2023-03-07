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
	"github.com/libp2p/go-libp2p/core/network"
)

func main() {
	quic := true
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
	var rcm network.ResourceManager
	Host, rcm = newHost(hostctx, priv, *debug)
	kademliaDHT := initDHT(hostctx, Host)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println("Host created. We are:", Host.ID())

	// Go routines

	go dhtRoutine(ctx, rendezvousS, kademliaDHT, quic)
	go readStdin()

	go execCommnad(ctx, mctx, priv)
	go func() {
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
	}()

	select {}
}
