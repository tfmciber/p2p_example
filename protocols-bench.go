package main

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
)

func sendBench(numMessages int, messageSize int, rendezvous string) {

	//send only if there is a peer connected

	protocol := "/bench/1.1.0"

	//fmt.Println("Starting connection Benchmark ", numMessages, " messages of size ", messageSize, " bytes each, please ensure that only one peer is connected to the host ")

	numMessagesStr := fillString(fmt.Sprintf("%d", numMessages), 32)

	messageSizeStr := fillString(fmt.Sprintf("%d", messageSize), 32)
	WriteDataRend([]byte(numMessagesStr), protocol, rendezvous, false)
	WriteDataRend([]byte(messageSizeStr), protocol, rendezvous, false)
	//start := time.Now()
	sendBuffer := make([]byte, messageSize)
	sendBuffer = bytes.Repeat([]byte("a"), messageSize)
	for i := 0; i < numMessages; i++ {

		WriteDataRend(sendBuffer, protocol, rendezvous, false)
	}

	//elapsed := time.Since(start)

	//fmt.Println("Bench completed in ", elapsed, "at an average speed of ", float64(numMessages*messageSize)/float64(elapsed.Milliseconds()), " Mbytes/sec")
	closeStreams(protocol)
}

func ReceiveBenchhandler(stream network.Stream) {
	fmt.Println("Receiving Benchmark data")
	numMessages := make([]byte, 32)
	stream.Read(numMessages)

	numMessagesnum, _ := strconv.Atoi(strings.Trim(string(numMessages), ":"))

	messageSize := make([]byte, 32)
	stream.Read(messageSize)

	messageSizenum, _ := strconv.Atoi(strings.Trim(string(messageSize), ":"))

	start := time.Now()
	receiveBuffer := make([]byte, messageSizenum)
	for i := 0; i < numMessagesnum; i++ {
		stream.Read(receiveBuffer)
	}
	elapsed := time.Since(start)
	fmt.Println("Benchmarked ", stream.Conn().ConnState().Transport, " Protocol, receiving ", numMessagesnum, messageSizenum, " bytes in ", elapsed, "at an average speed of ", float64(numMessagesnum*messageSizenum)/float64(elapsed.Milliseconds()), " Mbytes/sec")

	appendToCSV("./bench.csv", []string{stream.Conn().ConnState().Transport, fmt.Sprintf("%d", numMessagesnum), fmt.Sprintf("%d", messageSizenum), fmt.Sprintf("%d ", elapsed.Microseconds())})

	stream.Reset()
	stream.Close()

}

func benchTCPQUIC(ctx context.Context, mdns string, rendezvous string, nBytes int, nMess int) {
	fmt.Println("QUIC benchmark with", mdns)

	DisconnectAll()
	if mdns == "mdns" {
		FoundPeersMDNS = FindPeersMDNS(rendezvous)
		go ConnecToPeersMDNS(ctx, FoundPeersMDNS, rendezvous, true)

	} else {
		FoundPeersDHT = discoverPeers(ctx, Host, rendezvous)
		ConnecToPeers(ctx, FoundPeersDHT, rendezvous, true)

	}
	<-time.After(4 * time.Second)
	if len(Ren[rendezvous]) == 0 {
		fmt.Println("No peers found")
		return

	} else {
		fmt.Println("Starting benchmark with")
		for _, p := range Ren[rendezvous] {
			fmt.Println(Peers[p].peer)
		}
	}

	for j := 64; j < nBytes+1; j += 64 {
		for i := 0; i < 10; i++ {
			sendBench(nMess, j, rendezvous)
			time.Sleep(10 * time.Millisecond)
		}
	}

	fmt.Println("TCP benchmark")
	DisconnectAll()

	if mdns == "mdns" {
		FoundPeersMDNS = FindPeersMDNS(rendezvous)
		go ConnecToPeersMDNS(ctx, FoundPeersMDNS, rendezvous, false)

	} else {
		FoundPeersDHT = discoverPeers(ctx, Host, rendezvous)
		ConnecToPeers(ctx, FoundPeersDHT, rendezvous, false)
	}

	<-time.After(4 * time.Second)

	if len(Ren[rendezvous]) == 0 {
		fmt.Println("No peers found")
		return

	} else {
		fmt.Println("Starting benchmark with")
		for _, p := range Ren[rendezvous] {
			fmt.Println(Peers[p].peer)
		}
	}

	fmt.Println("Benchmarking with ", nMess, " messages of ", nBytes, " bytes")
	for j := 64; j < nBytes+1; j += 64 {
		for i := 0; i < 10; i++ {

			sendBench(nMess, j, rendezvous)
			time.Sleep(10 * time.Millisecond)
		}
	}
	fmt.Println("Benchmark finished")
}
