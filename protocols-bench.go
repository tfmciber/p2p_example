package main

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/schollz/progressbar/v3"
)

// procol int 1 = TCP, 2 = QUIC
func (c *P2Papp) sendBench(numMessages int, messageSize int, protocol int, peerid peer.ID) {

	if c.Host.Network().Connectedness(peerid) == network.Connected {

		numMessagesStr := fillString(fmt.Sprintf("%d", numMessages), 32)
		messageSizeStr := fillString(fmt.Sprintf("%d", messageSize), 32)
		protocolstr := fillString(fmt.Sprintf("%d", protocol), 32)

		stream := c.streamStart(peerid, c.benchproto)

		if stream == nil {
			fmt.Println("stream is nil")
			return
		}
		stream.Write([]byte(numMessagesStr))
		stream.Write([]byte(messageSizeStr))
		stream.Write([]byte(protocolstr))

		sendBuffer := make([]byte, messageSize)
		recvBuffer := make([]byte, messageSize)
		sendBuffer = bytes.Repeat([]byte("a"), messageSize)

		for j := 0; j < numMessages; j++ {

			start := time.Now()
			stream.Write(sendBuffer)
			n, _ := stream.Read(recvBuffer)
			elapsed := time.Since(start)
			appendToCSV("./bench.csv", []string{stream.Conn().ConnState().Transport, fmt.Sprintf("%d", n), fmt.Sprintf("%d ", elapsed.Microseconds())})

		}

		c.closeStreams(string(c.benchproto))
	}
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
		stream.Write(receiveBuffer)
		end := time.Since(start)
		fmt.Println(" \t [*] Message ", i, " of ", numMessagesnum, " took ", end.Microseconds(), " microseconds")
	}

	fmt.Println(" \t [*] Benchmarked ", stream.Conn().ConnState().Transport, " Protocol (", stream.Conn().ConnState().Transport, ")", messageSizenum)

	stream.Reset()
	stream.Close()

}

func (c *P2Papp) benchTCPQUIC(peerid peer.ID, nBytes int, nMess int) {

	fmt.Println("[*] Starting Benchmark with", nMess, "messages of", nBytes, "bytes")

	//Get all strings from the Host data

	fmt.Println("[*] Starting Benchmark with", nMess, "messages of", nBytes, "bytes")
	fmt.Println("\t[*] Starting TCP Benchmark")
	c.benchProto(peerid, nMess, false, 1024, 65536, 1024)
	fmt.Println("\t[*] Starting UDP Benchmark")
	c.benchProto(peerid, nMess, true, 1024, 65536, 1024)

}

func (c *P2Papp) benchProto(peerid peer.ID, nMess int, udp_tcp bool, start int, end int, step int) {

	if !c.setTransport(peerid, udp_tcp) {
		fmt.Println("Error Changing Peer Transport")
		return
	}
	total := 0
	last := 0
	//sucesion aritemtica
	prot := 1
	if udp_tcp {
		prot = 2
	}

	all := (start + end) / 2 * int(end/step)
	bar := progressbar.Default(100)
	for j := start; j < end+1; j += step {

		c.sendBench(nMess, j, prot, peerid)
		total += j
		progress := int((float64(total)) / (float64(all)) * 100)
		if progress%1 == 0 && progress != last {
			bar.Add(progress - last)
		}
		last = progress

	}

	bar.Add(100 - last)
	fmt.Println("[*] Benchmark finished")

}
