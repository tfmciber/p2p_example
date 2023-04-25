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

	c.writeDataRendFunc(c.textproto, rendezvous, func(stream network.Stream) {

		n, err := stream.Write([]byte(message))
		c.fmtPrintln(fmt.Sprintf("Sent [*] %s [%s] %s = %s,%d \n", date, rendezvous, c.Host.ID(), text, n))
		if err != nil {
			ok = false
			c.disconnectHost(stream, err, string(stream.Protocol()))
		}

	})
	return ok

}
func (c *P2Papp) LeaveChat(rendezvous string) {
	c.fmtPrintln("[*] LeaveChat " + rendezvous)
	//check if rendezvus exists
	if rendezvous == "" {
		c.fmtPrintln("rendezvous is empty")
		c.chatadded <- rendezvous
		c.useradded <- true
		return
	}
	peerid := c.GetPeerIDfromstring(rendezvous)
	isrend := c.checkRend(rendezvous)

	if !isrend && peerid == "" {
		c.fmtPrintln("rendezvous does not exist or is not a peerid")
		c.chatadded <- rendezvous
		c.useradded <- true
		return
	}

	c.writeDataRendFunc(c.cmdproto, rendezvous, func(stream network.Stream) {
		tries := 0
	leave:
		n, err := stream.Write([]byte("leave$" + rendezvous))
		if err != nil {
			tries++
			peerid := stream.Conn().RemotePeer()
			c.fmtPrintln("leave sent to "+rendezvous+" "+c.Host.ID().String()+" "+fmt.Sprintf("%d", n)+" "+peerid.String(), err)
			stream.Close()

			stream = c.streamStart(peerid, c.cmdproto)
			if tries < 3 {

				goto leave
			}
		}
		fmt.Println(n, err)
		c.fmtPrintln("leave sent to " + rendezvous + " " + c.Host.ID().String() + " " + fmt.Sprintf("%d", n))

	})

	//remove rendezvous from list

	if isrend {
		c.fmtPrintln("rendezvous deleted "+rendezvous, "c.data:", c.data)
		c.leaveChat(rendezvous)

	}

	if peerid != "" {
		c.fmtPrintln("DM deleted " + peerid)
		c.leaveChat(peerid.String())
	}
	c.chatadded <- rendezvous
	c.useradded <- true

}
func (c *P2Papp) deleteDm(peerid peer.ID) {
	if contains(c.direcmessages, peerid) {
		for i, p := range c.direcmessages {
			if p == peerid {
				c.direcmessages = append(c.direcmessages[:i], c.direcmessages[i+1:]...)
			}
		}
		var peerids []string
		for _, v := range c.direcmessages {
			peerids = append(peerids, v.String())
		}
		if len(peerids) == 0 {
			peerids = []string{}
		}
		c.trashchats[peerid.String()] = true
		runtime.EventsEmit(c.ctx, "directMessage", peerids)
	}
	c.newThrash(peerid.String(), true)
}
func (c *P2Papp) leaveChat(rendezvous string) {
	c.fmtPrintln("LeaveChat " + rendezvous)
	delete(c.data, rendezvous)

	c.newThrash(rendezvous, true)

}
func (c *P2Papp) DeleteChat(rendezvous string) {
	c.fmtPrintln("DeleteChat " + rendezvous)

	c.newThrash(rendezvous, false)

}

func (c *P2Papp) newThrash(key string, add bool) {
	var aux []string
	//add key to trashchats

	c.trashchats[key] = add
	// convert map to slice for true values
	for k, g := range c.trashchats {
		if g == true {
			aux = append(aux, k)
		}
	}
	if len(aux) == 0 {
		aux = []string{}
	}
	c.EmitEvent("newThrash", aux)
}
func (c *P2Papp) receiveTexthandler(stream network.Stream) {

	for {
		buff := make([]byte, 2000)
		_, err := stream.Read(buff)

		//if err is not EOF
		if err != nil {
			if err.Error() != "EOF" {

				c.disconnectHost(stream, err, string(stream.Protocol()))
				return
			}

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
		c.fmtPrintln(fmt.Sprintf("received message [*] %s [%s] %s = %s \n", date, rendezvous, stream.Conn().RemotePeer(), text))
		c.EmitEvent("receiveMessage", rendezvous, text, stream.Conn().RemotePeer().String(), date)
		stream.Close()
		return

	}

}
