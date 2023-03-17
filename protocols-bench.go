package main

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/schollz/progressbar/v3"
)

// procol int 1 = TCP, 2 = QUIC
func (c *P2Papp) sendBench(numMessages int, messageSize int, protocol int, rendezvous string) {

	numMessagesStr := fillString(fmt.Sprintf("%d", numMessages), 32)
	messageSizeStr := fillString(fmt.Sprintf("%d", messageSize), 32)
	protocolstr := fillString(fmt.Sprintf("%d", protocol), 32)
	c.writeDataRend([]byte(numMessagesStr), c.benchproto, rendezvous, false)
	c.writeDataRend([]byte(messageSizeStr), c.benchproto, rendezvous, false)
	c.writeDataRend([]byte(protocolstr), c.benchproto, rendezvous, false)
	sendBuffer := make([]byte, messageSize)
	sendBuffer = bytes.Repeat([]byte("a"), messageSize)

	for j := 0; j < numMessages; j++ {

		c.writeDataRend(sendBuffer, c.benchproto, rendezvous, false)
	}

	c.closeStreams(string(c.benchproto))
}

func (c *P2Papp) receiveBenchhandler(stream network.Stream) {

	numMessages := make([]byte, 32)
	stream.Read(numMessages)

	numMessagesnum, _ := strconv.Atoi(strings.Trim(string(numMessages), ":"))

	messageSize := make([]byte, 32)
	stream.Read(messageSize)

	messageSizenum, _ := strconv.Atoi(strings.Trim(string(messageSize), ":"))

	protocol := make([]byte, 32)
	stream.Read(protocol)
	protocolnum, _ := strconv.Atoi(strings.Trim(string(protocol), ":"))

	if numMessagesnum == 0 || messageSizenum == 0 || protocolnum == 0 {
		fmt.Println("Invalid message size or number of messages", numMessagesnum, messageSizenum, protocolnum)
		//stream.Reset()
		stream.Close()
		return
	}

	if protocolnum == 1 {
		//ensure that the stream protocol is TCP
		if stream.Conn().ConnState().Transport != "tcp" {
			fmt.Println("Invalid protocol")
			stream.Reset()

		}
	} else if protocolnum == 2 {
		//ensure that the stream protocol is QUIC
		if stream.Conn().ConnState().Transport != "quic" {
			fmt.Println("Invalid protocol")
			stream.Reset()

		}
	}

	fmt.Println("[*] Received Benchmark request with", numMessagesnum, "messages of", messageSizenum, "bytes")
	receiveBuffer := make([]byte, messageSizenum)

	for i := 0; i < numMessagesnum; i++ {
		start := time.Now()
		stream.Read(receiveBuffer)
		elapsed := time.Since(start)
		appendToCSV("./bench.csv", []string{stream.Conn().ConnState().Transport, fmt.Sprintf("%d", numMessagesnum), fmt.Sprintf("%d", messageSizenum), fmt.Sprintf("%d ", elapsed.Microseconds())})

	}

	fmt.Println(" \t [*] Benchmarked ", stream.Conn().ConnState().Transport, " Protocol (", stream.Conn().ConnState().Transport, ")", messageSizenum)

	stream.Reset()
	stream.Close()

}

func (c *P2Papp) benchTCPQUIC(rendezvous string, nBytes int, nMess int) {

	fmt.Println("[*] Starting Benchmark with", nMess, "messages of", nBytes, "bytes")

	//Get all strings from the Host data

	oldTimers := make(map[string]uint)
	ren := c.GetKeys()

	for _, r := range ren {
		oldTimers[r] = c.GetTimer(r)
		c.SetTimer(r, 9999999999999)
	}
	fmt.Println("[*] Starting Benchmark with", nMess, "messages of", nBytes, "bytes")
	fmt.Println("\t[*] Starting TCP Benchmark")
	c.benchProto(rendezvous, nBytes, nMess, false, 64, 1024, 192)
	fmt.Println("\t[*] Starting UDP Benchmark")
	c.benchProto(rendezvous, nBytes, nMess, true, 64, 1024, 192)

	for _, r := range ren {
		c.SetTimer(r, oldTimers[r])
	}

}

func (c *P2Papp) benchProto(rendezvous string, nBytes int, nMess int, udp_tcp bool, start int, end int, step int) {

	peers := c.onlinePeers(rendezvous)
	if len(peers) == 0 {
		fmt.Println("No online peers found, attempting to reconnect")
		c.Reconnect(rendezvous)
		peers := c.onlinePeers(rendezvous)
		if len(peers) == 0 {
			fmt.Println("No online peers found, aborting")
			return
		}
	}

	if !c.setPeersTRansport(c.ctx, rendezvous, udp_tcp) {
		fmt.Println("Error Changing Peers Transport")
		return
	}
	total := 0
	last := 0
	//sucesion aritemtica
	prot := 1
	if udp_tcp {
		prot = 2
	}

	all := (start + end) / 2 * int(nBytes/step)
	bar := progressbar.Default(100)
	for j := start; j < nBytes+1; j += step {

		c.sendBench(nMess, j, prot, rendezvous)
		total += j
		progress := int((float64(total)) / (float64(all)) * 100)
		if progress%1 == 0 && progress != last {
			bar.Add(progress - last)
		}
		last = progress

	}
	fmt.Println("[*] Benchmark finished")

}
