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

		sendBuffer = bytes.Repeat([]byte("a"), messageSize)
		recvBuffer := make([]byte, 32)
		start := time.Now()
		for j := 0; j < numMessages; j++ {

			_, err := stream.Write(sendBuffer)
			if err != nil {
				fmt.Println("Write failed: restarting ", err)
				stream.Reset()
				stream = c.streamStart(peerid, c.benchproto)

			}

		}
		_, err := stream.Read(recvBuffer)
		if err != nil {
			fmt.Println("Read failed: restarting ", err)
		}

		numread, _ := strconv.Atoi(strings.Trim(string(recvBuffer), ":"))

		elapsed := time.Since(start)

		if numread != 0 {
			appendToCSV("./bench.csv", []string{stream.Conn().ConnState().Transport, fmt.Sprintf("%d", numread), fmt.Sprintf("%d ", elapsed.Microseconds()), fmt.Sprintf("%d ", messageSize)})
		} else {
			fmt.Println("Invalid message size or number of messages", recvBuffer)
		}
		stream.Close()
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

	receiveBuffer := make([]byte, messageSizenum)
	total := 0
	for i := 0; i < numMessagesnum; i++ {

		n, err := stream.Read(receiveBuffer)
		if err != nil {
			fmt.Println("Error reading from stream", err)

		}

		total += n
	}
	//convert total to byte array

	fmt.Println("Total bytes received", total)
	totalstr := fillString(fmt.Sprintf("%d", total), 32)
	_, err := stream.Write([]byte(totalstr))
	if err != nil {
		fmt.Println("Error writing to stream", err)
	}

	stream.Reset()
	stream.Close()

}

func (c *P2Papp) benchTCPQUIC(peerid peer.ID, nBytes int, nMess int, times int) {

	fmt.Println("[*] Starting Benchmark with", nMess, "messages of", nBytes, "bytes")

	//Get all strings from the Host data

	fmt.Println("[*] Starting Benchmark with", nMess, "messages of", nBytes, "bytes")
	fmt.Println("\t[*] Starting TCP Benchmark")
	c.benchProto(peerid, nMess, false, 64, 1024, 64, times)
	fmt.Println("\t[*] Starting UDP Benchmark")
	c.benchProto(peerid, nMess, true, 64, 1024, 64, times)

}

func (c *P2Papp) benchProto(peerid peer.ID, nMess int, udp_tcp bool, start int, end int, step int, times int) {

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

	all := (start + end) / 2 * int(end/step) * times
	bar := progressbar.Default(100)
	for j := start; j < end+1; j += step {
		for i := 0; i < times; i++ {
			c.sendBench(nMess, j, prot, peerid)
			time.Sleep(100 * time.Millisecond)
		}
		total += j * times
		progress := int((float64(total)) / (float64(all)) * 100)
		if progress%1 == 0 && progress != last {
			bar.Add(progress - last)
		}
		last = progress

	}

	bar.Add(100 - last)
	fmt.Println("[*] Benchmark finished")

}
