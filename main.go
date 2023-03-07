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

var rendezvousS = make(chan string, 1)

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
	Host, _ = newHost(hostctx, priv, *debug)
	kademliaDHT := initDHT(hostctx, Host)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println("Host created. We are:", Host.ID())

	// Go routines

	//go func for when a channel, "aux" is written create a new nuction that runs every 5 minutes appends the value written to the channel to a list and then runs the function
	// for all the values in the list that wherent run in the last 5 minutes

	go func() {
		var allRedenzvous = map[string]int{}
		for {
			select {
			case <-time.After(1 * time.Minute):

				for rendezvous, time := range allRedenzvous {
					if time == 0 {
						rendezvousS <- rendezvous
					} else {
						allRedenzvous[rendezvous]--
						fmt.Println("Rendezvous", rendezvous, "restarting in ", allRedenzvous[rendezvous])
					}
				}

			case aux := <-rendezvousS:

				FoundPeersDHT := discoverPeers(ctx, kademliaDHT, Host, aux)
				failed := connecToPeers(ctx, FoundPeersDHT, aux, quic, true)
				connectRelay(aux)
				connectthrougRelays(failed, aux)
				allRedenzvous[aux] = 5
			}
		}
	}()

	go readStdin()

	go execCommnad(ctx, mctx, priv)

	select {}
}
