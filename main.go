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
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

func main() {

	debug := flag.Bool("debug", false, "debug mode, Prints host stats")
	seed := flag.Int("seed", 0, "Seed for the random number generator for debug mode")
	refreshTime := flag.Uint("refresh", 15, "Minutes to refresh the DHT")
	quic := flag.Bool("quic", false, "Use QUIC transport")
	filename := flag.String("config", "./config.json", "Config file")
	mtype := flag.Bool("req", false, "debug mode, Prints host stats")
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

	var textproto = protocol.ID("/text/1.0.0")
	var audioproto = protocol.ID("/audio/1.0.0")
	var benchproto = protocol.ID("/bench/1.0.0")
	var cmdproto = protocol.ID("/cmd/1.0.0")
	var fileproto = protocol.ID("/file/1.0.0")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	P := P2Papp{data: make(map[string]struct {
		peers []peer.ID
		timer uint
	}), ctx: ctx, rendezvousS: make(chan string, 1), priv: priv, textproto: textproto, audioproto: audioproto, benchproto: benchproto, cmdproto: cmdproto, fileproto: fileproto}

	P.newHost()
	P.initDHT()
	P.interrupts()

	fmt.Println("Host created. We are:", P.Host.ID())

	// Go routines
	var cmdChan = make(chan string)

	go P.dhtRoutine(*quic, *refreshTime, *mtype)

	go readStdin(cmdChan)

	go P.execCommnad(mctx, *quic, cmdChan)

	cmdChan <- "dht$llkhkjhkñ"

	select {}
}
