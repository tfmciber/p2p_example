package main

import (
	"fmt"
	"io"
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

		messageSizeStr := fillString(fmt.Sprintf("%d", messageSize), 32)
		protocolstr := fillString(fmt.Sprintf("%d", protocol), 32)

		stream, err := c.Host.NewStream(c.ctx, peerid, c.benchproto)
		if err != nil {
			fmt.Println("stream is nil")
			return
		}

		stream.Write([]byte(protocolstr))
		stream.Write([]byte(messageSizeStr))
		sent := 0

		sendBuffer := make([]byte, messageSize)
		//recvBuffer := make([]byte, 32)
		start := time.Now()
		for time.Now().Sub(start) < 1*time.Second {

			n, err := stream.Write(sendBuffer)
			if err != nil {
				stream, err = c.Host.NewStream(c.ctx, peerid, c.benchproto)
				if err != nil {
					fmt.Println("stream is nil")
					return
				}
			}

			sent += n

		}
		elapsed := time.Since(start)
		appendToCSV("./bench.csv", []string{stream.Conn().ConnState().Transport, fmt.Sprintf("%d", sent), fmt.Sprintf("%f", elapsed.Seconds()), fmt.Sprintf("%d", messageSize)})
		stream.Close()
		stream.Reset()

	}
}

func (c *P2Papp) receiveBenchhandler(stream network.Stream) {

	messageSize := make([]byte, 32)
	stream.Read(messageSize)

	messageSizenum, _ := strconv.Atoi(strings.Trim(string(messageSize), ":"))

	protocol := make([]byte, 32)
	stream.Read(protocol)
	protocolnum, _ := strconv.Atoi(strings.Trim(string(protocol), ":"))

	if messageSizenum == 0 || protocolnum == 0 {
		fmt.Println("Invalid message size or number of messages", messageSizenum, protocolnum)
		stream.Close()
		stream.Reset()
		return
	}

	if protocolnum == 1 {
		//ensure that the stream protocol is TCP
		if stream.Conn().ConnState().Transport != "tcp" {
			fmt.Println("Invalid protocol")
			stream.Close()
			stream.Reset()
			return

		}
	} else if protocolnum == 2 {
		//ensure that the stream protocol is QUIC
		if stream.Conn().ConnState().Transport != "quic" {
			fmt.Println("Invalid protocol")
			stream.Close()
			stream.Reset()
			return

		}
	}

	receiveBuffer := make([]byte, messageSizenum)
	total := 0
	var err error
	var n int
	for err == nil {
		n, err = io.Reader.Read(stream, receiveBuffer)
		total += n
	}

	stream.Close()
	stream.Reset()
	return
}

func (c *P2Papp) benchTCPQUIC(peerid peer.ID, nBytes int, nMess int, times int) {

	fmt.Println("[*] Starting Benchmark with", nMess, "messages of", nBytes, "bytes")
	createFile("./bench.csv")
	appendToCSV("./bench.csv", []string{"Protocol", "Bytes", "Time", "Size"})
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
