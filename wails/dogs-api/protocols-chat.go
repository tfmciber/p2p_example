package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (c *P2Papp) SendTextHandler(text string, rendezvous string) bool {

	////get time and date dd/mm/yyyy hh:mm
	t := time.Now()
	date := t.Format("02/01/2006 15:04")
	c.fmtPrintln(date)
	var err error
	ok := true
	c.writeDataRendFunc(c.textproto, rendezvous, func(stream network.Stream) {

		_, err = stream.Write([]byte(rendezvous + "$" + text + "$" + date))
		if err != nil {
			ok = false
			c.disconnectHost(stream, err, string(stream.Protocol()))
		}

	})
	return ok

}

func (c *P2Papp) receiveTexthandler(stream network.Stream) {

	buff := make([]byte, 2000)
	for {

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

		fmt.Printf("[%s] %s = %s \n", rendezvous, stream.Conn().RemotePeer(), text)
		runtime.EventsEmit(c.ctx, "receiveMessage", rendezvous, text, stream.Conn().RemotePeer().String(), date)
	}

}
