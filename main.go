package main

import (
	"context"
	"fmt"
)

var protocols = []string{"/audio/1.1.0", "/chat/1.1.0"}

func main() {

	filename := "./config.json"
	priv := initPriv(filename)

	hostctx = context.Background()

	Interrupts()
	initCTX()

	Host, _ = NewHost(hostctx, priv)

	fmt.Println("Host created. We are:", Host.ID())

	// Go routines

	go ReadStdin()

	// Start State machine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go execCommnad(ctx)

	select {}
}
