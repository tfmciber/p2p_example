package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (c *P2Papp) SendTextHandler(text string, rendezvous string) bool {
	c.fmtPrintln("SendTextHandler " + text + " " + rendezvous)
	////get time and date dd/mm/yyyy hh:mm
	t := time.Now()
	date := t.Format("02/01/2006 15:04")

	ok := true
	message := (rendezvous + "$" + text + "$" + date)

	x, y := c.Get(rendezvous)
	if x == nil {
		return false
	}
	if y == true {
		//we are sending a direct message
		c.AddDm(x[0])
	}

	c.fmtPrintln(fmt.Sprintf("[*] %s [%s] %s  \n", date, rendezvous, text))

	c.writeDataRendFunc(c.textproto, rendezvous, func(stream network.Stream) {

		n, err := stream.Write([]byte(message))
		c.fmtPrintln(fmt.Sprintf("Enviamos [*] %s [%s] %s = %s,%d \n", date, rendezvous, c.Host.ID(), text, n))
		if err != nil {
			ok = false
			c.disconnectHost(stream, err, string(stream.Protocol()))
		}

	})
	return ok

}

func (c *P2Papp) receiveTexthandler(stream network.Stream) {

	for {
		buff := make([]byte, 2000)
		_, err := stream.Read(buff)

		if err != nil {

			c.disconnectHost(stream, err, string(stream.Protocol()))
			return

		}

		data := strings.SplitN(string(buff[:]), "$", 3)
		var rendezvous string
		var text string
		var date string
		if len(data) > 1 {
			rendezvous = data[0]
			text = data[1]

		}
		if len(data) > 2 {
			date = data[2]
		} else {
			t := time.Now()
			date = t.Format("02/01/2006 15:04")
		}
		if rendezvous == c.Host.ID().String() {
			// if we receive our ID as rendezvous, it means we are receiving a direct message
			c.AddDm(stream.Conn().RemotePeer())
			rendezvous = stream.Conn().RemotePeer().String()
		}
		c.fmtPrintln(fmt.Sprintf("[*] %s [%s] %s = %s \n", date, rendezvous, stream.Conn().RemotePeer(), text))
		runtime.EventsEmit(c.ctx, "receiveMessage", rendezvous, text, stream.Conn().RemotePeer().String(), date)

	}

}

func (c *P2Papp) AddDm(peerid peer.ID) {

	if !contains(c.direcmessages, peerid) {
		c.direcmessages = append(c.direcmessages, peerid)
		//convert []peer.ID to []string
		var peerids []string
		for _, v := range c.direcmessages {
			peerids = append(peerids, v.String())
		}
		runtime.EventsEmit(c.ctx, "directMessage", peerids)

	}

}
