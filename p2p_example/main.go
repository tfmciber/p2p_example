package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gen2brain/malgo"
)

func main() {

	filename := "./config.json"
	priv := initPriv(filename)

	hostctx = context.Background()

	Interrupts()

	mctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {

	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer func() {
		_ = mctx.Uninit()
		mctx.Free()
	}()
	Host, _ = NewHost(hostctx, priv)

	fmt.Println("Host created. We are:", Host.ID())

	// Go routines

	go ReadStdin()

	// call CheckCoon every 5 seconds
	go func() {
		for {
			CheckCoon()
			time.Sleep(5 * time.Second)
		}
	}()

	//go Notifyondisconnect()

	// Start State machine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go execCommnad(ctx, mctx, priv)

	select {}
}
