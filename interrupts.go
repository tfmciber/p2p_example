package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func Interrupts() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-quit
		fmt.Println("\r- Exiting Program")
		Host.Close()
		os.Exit(0)
	}()
}
