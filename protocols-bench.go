package main

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/schollz/progressbar/v3"
)

func sendBench(numMessages int, messageSize int, rendezvous string) {

	protocol := "/bench/1.1.0"
	numMessagesStr := fillString(fmt.Sprintf("%d", numMessages), 32)
	messageSizeStr := fillString(fmt.Sprintf("%d", messageSize), 32)
	writeDataRend([]byte(numMessagesStr), protocol, rendezvous, false)
	writeDataRend([]byte(messageSizeStr), protocol, rendezvous, false)
	sendBuffer := make([]byte, messageSize)
	sendBuffer = bytes.Repeat([]byte("a"), messageSize)
	for i := 0; i < numMessages; i++ {

		writeDataRend(sendBuffer, protocol, rendezvous, false)
	}
	closeStreams(protocol)
}

func receiveBenchhandler(stream network.Stream) {

	numMessages := make([]byte, 32)
	stream.Read(numMessages)

	numMessagesnum, _ := strconv.Atoi(strings.Trim(string(numMessages), ":"))

	messageSize := make([]byte, 32)
	stream.Read(messageSize)

	messageSizenum, _ := strconv.Atoi(strings.Trim(string(messageSize), ":"))

	if numMessagesnum == 0 || messageSizenum == 0 {
		fmt.Println("Invalid message size or number of messages")
		stream.Reset()
		stream.Close()
		return
	}

	start := time.Now()
	receiveBuffer := make([]byte, messageSizenum)
	for i := 0; i < numMessagesnum; i++ {
		stream.Read(receiveBuffer)
	}
	elapsed := time.Since(start)
	fmt.Print("\r \t [*] Benchmarked ", stream.Conn().ConnState().Transport, " Protocol (", numMessagesnum, messageSizenum, ") in ", elapsed)

	appendToCSV("./bench.csv", []string{stream.Conn().ConnState().Transport, fmt.Sprintf("%d", numMessagesnum), fmt.Sprintf("%d", messageSizenum), fmt.Sprintf("%d ", elapsed.Microseconds())})

	stream.Reset()
	stream.Close()

}

func benchTCPQUIC(ctx context.Context, rendezvous string, times, nBytes int, nMess int) {

	fmt.Println("[*] Starting Benchmark with", nMess, "messages of", nBytes, "bytes", times, "times")
	fmt.Println("\t[*] Starting QUIC Benchmark")

	if !hasPeers(rendezvous) {
		fmt.Println("No online peers found")
		return
	}

	fmt.Println(onlinePeers(rendezvous))

	fmt.Println("[*] Starting Benchmark with", nMess, "messages of", nBytes, "bytes", times, "times")
	fmt.Println("\t[*] Starting QUIC Benchmark")
	if !setPeersTRansport(ctx, true) {
		fmt.Println("Error Changing Peers Transport")
		return
	}
	total := 0
	last := 0
	//sucesion aritemtica

	all := (64 + 1024) / 2 * int(nBytes/64) * times
	bar := progressbar.Default(100)
	for j := 64; j < nBytes+1; j += 64 {
		for i := 0; i < times; i++ {
			sendBench(nMess, j, rendezvous)
			total += j
			time.Sleep(10 * time.Millisecond)
		}
		progress := int((float64(total)) / (float64(all)) * 100)
		if progress%1 == 0 && progress != last {
			bar.Add(progress - last)
		}
		last = progress

	}

	fmt.Println("\t[*] Starting TCP Benchmark")

	if !setPeersTRansport(ctx, false) {
		fmt.Println("Error Changing Peers Transport")
		return
	}
	total = 0

	last = 0
	bar = progressbar.Default(100)
	for j := 64; j < nBytes+1; j += 64 {
		for i := 0; i < times; i++ {
			sendBench(nMess, j, rendezvous)
			total += j
			time.Sleep(10 * time.Millisecond)
		}

		progress := int((float64(total)) / (float64(all)) * 100)
		if progress%1 == 0 && progress != last {
			bar.Add(progress - last)
		}
		last = progress

	}

	fmt.Println("\t[*] Benchmark finished")

}
